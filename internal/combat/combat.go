package combat

import (
	"fmt"

	"runa/internal/character"
	"runa/internal/enemies"
	"runa/internal/item"
	"runa/internal/spell"
)

// AttackOption décrit UNE entrée du menu "Attaque" : soit l'attaque de
// base à mains nues/à l'arme (SpellID == ""), soit un sort connu.
type AttackOption struct {
	Name     string
	Damage   uint16
	ManaCost uint16
	SpellID  string // "" pour l'attaque de base
}

// AttackOptions liste tout ce que le joueur peut choisir sous "Attaque" :
// l'attaque de base (Strength + arme équipée) toujours en premier,
// puis chaque sort appris. C'est cette liste que menu/combat.go affichera.
func AttackOptions(c *character.Character) []AttackOption {
	opts := []AttackOption{
		{Name: "Attaque de base", Damage: c.TotalAttack(), ManaCost: 0},
	}
	for id, known := range c.KnownSpells {
		if !known {
			continue
		}
		if s, ok := spell.GetSpell(id); ok {
			opts = append(opts, AttackOption{
				Name:     s.Name,
				Damage:   uint16(s.Damage),
				ManaCost: uint16(s.ManaCost),
				SpellID:  s.ID,
			})
		}
	}
	return opts
}

// AvailableConsumables liste les consommables actuellement dans
// l'inventaire du joueur, pour peupler le menu "Objet".
func AvailableConsumables(c *character.Character) []item.Item {
	var out []item.Item
	for _, slot := range c.Inventory.Slots {
		if slot.Item.Type() == item.TypeConsumable {
			out = append(out, slot.Item)
		}
	}
	return out
}

// ActiveEffect suit UN effet de statut en cours sur l'ennemi (poison,
// bleed...) : combien de dégâts par tick, combien de tours il reste
// à courir. Vit uniquement le temps du combat, jamais persisté ailleurs.
type ActiveEffect struct {
	Type      item.EffectType
	Amount    uint16
	Remaining uint8
}

// Combat représente UN affrontement en cours, tour par tour, entre le
// joueur et un monstre statique rencontré sur la carte.
type Combat struct {
	Player       *character.Character
	Enemy        *enemies.EnemyInstance
	EnemyEffects []ActiveEffect // poison/bleed actifs sur l'ennemi, tick à chaque tour
	Log          []string       // messages affichables tels quels par menu/combat.go
	Over         bool
	PlayerWon    bool
	PlayerFled   bool
}

// NewCombat démarre un combat. Le joueur commence TOUJOURS à mana
// pleine, quel que soit son mana avant le combat (règle explicite).
func NewCombat(p *character.Character, e *enemies.EnemyInstance) *Combat {
	p.Mana = p.ManaMax
	return &Combat{Player: p, Enemy: e}
}

type ActionResult int

const (
	ActionOK ActionResult = iota
	ErrNotEnoughMana
	ErrUnknownAttack
	ErrItemNotOwned
	ErrCombatOver
)

// PlayerAttack exécute une attaque du joueur, identifiée par son SpellID
// ("" = attaque de base). Dégâts directs à l'ennemi, puis effets attachés
// à l'arme équipée si c'est l'attaque de base (ex: le Bleed d'une hache),
// puis résolution du reste du tour (statuts, riposte, régén de mana).
func (cb *Combat) PlayerAttack(spellID string) ActionResult {
	if cb.Over {
		return ErrCombatOver
	}

	opt, ok := findOption(cb.Player, spellID)
	if !ok {
		return ErrUnknownAttack
	}
	if cb.Player.Mana < opt.ManaCost {
		return ErrNotEnoughMana
	}
	cb.Player.Mana -= opt.ManaCost

	cb.Enemy.TakeDamage(opt.Damage)
	cb.log("Vous utilisez %s : %d dégâts.", opt.Name, opt.Damage)

	// L'attaque de base porte les effets de l'arme équipée (ex: bleed
	// d'une hache). Un sort n'a pas d'arme à consulter : ses éventuels
	// effets viendraient de spell.Spell lui-même, pas encore modélisés.
	if spellID == "" && cb.Player.EquippedWeapon != nil {
		cb.applyItemEffects(*cb.Player.EquippedWeapon)
	}

	if cb.checkEnemyDefeated() {
		return ActionOK
	}
	cb.endPlayerTurn()
	return ActionOK
}

// UseItem consomme un consommable de l'inventaire du joueur, applique
// ses effets (cf. applyItemEffects), puis enchaîne comme une attaque :
// utiliser un objet prend aussi le tour du joueur.
func (cb *Combat) UseItem(consumable item.Item) ActionResult {
	if cb.Over {
		return ErrCombatOver
	}
	if !cb.Player.RemoveItem(consumable) {
		return ErrItemNotOwned
	}

	cb.applyItemEffects(consumable)
	cb.log("Vous utilisez %s.", consumable.Name())

	cb.endPlayerTurn()
	return ActionOK
}

// applyItemEffects répartit les effets attachés à un item (via
// item.EffectsOf) : le soin s'applique immédiatement au joueur, tout
// le reste (Poison, Bleed, futurs effets offensifs) devient un statut
// actif sur l'ennemi qui tique chaque tour. Si un effet du même Type
// est déjà actif sur l'ennemi, il est RAFRAÎCHI (Amount/Remaining
// remplacés) plutôt qu'empilé — sinon frapper en boucle avec une arme
// à effet permettrait de cumuler des dizaines de stacks de bleed et de
// one-shot n'importe quel monstre. Partagée par UseItem (consommables)
// et PlayerAttack (effets de l'arme équipée), pour que les deux sources
// d'effets soient traitées identiquement.
func (cb *Combat) applyItemEffects(i item.Item) {
	for _, eff := range item.EffectsOf(i) {
		if eff.Type == item.EffectHeal {
			cb.Player.Hp += uint16(eff.Amount)
			if cb.Player.Hp > cb.Player.HpMax {
				cb.Player.Hp = cb.Player.HpMax
			}
			continue
		}

		if idx, found := cb.findEnemyEffect(eff.Type); found {
			cb.EnemyEffects[idx].Amount = uint16(eff.Amount)
			cb.EnemyEffects[idx].Remaining = eff.Duration
			cb.log("%s : %s est rafraîchi.", cb.Enemy.Template.Name, eff.Type.String())
			continue
		}

		cb.EnemyEffects = append(cb.EnemyEffects, ActiveEffect{
			Type:      eff.Type,
			Amount:    uint16(eff.Amount),
			Remaining: eff.Duration,
		})
		cb.log("%s est affligé de %s.", cb.Enemy.Template.Name, eff.Type.String())
	}
}

// findEnemyEffect cherche un effet actif d'un Type donné parmi les
// statuts en cours sur l'ennemi, pour permettre le rafraîchissement
// plutôt que l'empilement dans applyItemEffects.
func (cb *Combat) findEnemyEffect(t item.EffectType) (int, bool) {
	for i, ae := range cb.EnemyEffects {
		if ae.Type == t {
			return i, true
		}
	}
	return 0, false
}

// Flee tente une fuite. Pas de stat d'évasion côté monstre pour
// l'instant : la fuite réussit toujours, mais consomme quand même
// le tour (cohérent avec "choisir Fuir" comme une action à part entière).
func (cb *Combat) Flee() ActionResult {
	if cb.Over {
		return ErrCombatOver
	}
	cb.Over = true
	cb.PlayerFled = true
	cb.log("Vous prenez la fuite.")
	return ActionOK
}

// endPlayerTurn résout tout ce qui suit l'action du joueur, dans l'ordre :
// 1. les statuts actifs sur l'ennemi tiquent (peut le tuer sans riposte),
// 2. si l'ennemi survit, il riposte avec une attaque aléatoire,
// 3. si le joueur survit, il régénère 5% de son mana max.
// Chaque étape s'arrête si le combat vient de se terminer.
func (cb *Combat) endPlayerTurn() {
	cb.tickEnemyEffects()
	if cb.checkEnemyDefeated() {
		return
	}
	cb.enemyTurn()
	if cb.Over {
		return
	}
	cb.regenMana()
}

// tickEnemyEffects applique un tour de dégâts pour chaque statut actif
// sur l'ennemi, puis décrémente sa durée restante. Un effet à 0 tour
// restant est retiré. Le poison/bleed peut donc achever l'ennemi ici,
// avant même que le joueur n'attaque à nouveau.
func (cb *Combat) tickEnemyEffects() {
	if len(cb.EnemyEffects) == 0 {
		return
	}
	var remaining []ActiveEffect
	for _, ae := range cb.EnemyEffects {
		cb.Enemy.TakeDamage(ae.Amount)
		ae.Remaining--
		cb.log("%s subit %d dégâts de %s (%d tour(s) restant(s)).", cb.Enemy.Template.Name, ae.Amount, ae.Type.String(), ae.Remaining)
		if ae.Remaining > 0 {
			remaining = append(remaining, ae)
		}
	}
	cb.EnemyEffects = remaining
}

// checkEnemyDefeated marque le combat comme gagné si l'ennemi est mort,
// et log le message correspondant. Renvoie true si c'est le cas, pour
// que l'appelant puisse arrêter la résolution du tour immédiatement.
func (cb *Combat) checkEnemyDefeated() bool {
	if !cb.Enemy.IsAlive() {
		cb.Over = true
		cb.PlayerWon = true
		cb.log("%s est vaincu !", cb.Enemy.Template.Name)
		return true
	}
	return false
}

// enemyTurn fait jouer l'ennemi : une attaque choisie au hasard dans
// son Set (déjà géré par enemies.RandomAttack), appliquée au joueur.
func (cb *Combat) enemyTurn() {
	atk := cb.Enemy.RandomAttack()
	if atk.Damage >= cb.Player.Hp {
		cb.Player.Hp = 0
	} else {
		cb.Player.Hp -= atk.Damage
	}
	cb.log("%s utilise %s : %d dégâts.", cb.Enemy.Template.Name, atk.Name, atk.Damage)

	if cb.Player.Hp == 0 {
		cb.Over = true
		cb.PlayerWon = false
		cb.log("Vous êtes vaincu...")
	}
}

// regenMana rend 5% du mana MAX du joueur, arrondi vers le bas, avec un
// minimum de 1 point si le joueur a du mana max (sinon 5% d'un petit
// total arrondirait toujours à 0 et la régén serait invisible).
func (cb *Combat) regenMana() {
	if cb.Player.ManaMax == 0 {
		return
	}
	regen := uint16(float64(cb.Player.ManaMax) * 0.05)
	if regen == 0 {
		regen = 1
	}
	cb.Player.Mana += regen
	if cb.Player.Mana > cb.Player.ManaMax {
		cb.Player.Mana = cb.Player.ManaMax
	}
}

func findOption(c *character.Character, spellID string) (AttackOption, bool) {
	for _, opt := range AttackOptions(c) {
		if opt.SpellID == spellID {
			return opt, true
		}
	}
	return AttackOption{}, false
}

func (cb *Combat) log(format string, args ...any) {
	cb.Log = append(cb.Log, fmt.Sprintf(format, args...))
}
