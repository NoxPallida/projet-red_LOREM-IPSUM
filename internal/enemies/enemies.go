package enemies

import (
	"math"
	"math/rand"

	"runa/internal/item"
)

type Attack struct {
	Name     string
	Damage   uint16
	ManaCost uint16 // 0 pour une attaque physique sans coût
}

// DropRate décrit un objet qui peut tomber à la mort du monstre.
type DropRate struct {
	Item     item.Item
	Chance   float64 // probabilité entre 0.0 et 1.0 (ex: 0.70 = 70%)
	Quantity uint8
}

// EnemyTemplate décrit un TYPE de monstre : ses stats de base au niveau 1,
// et son taux de croissance par niveau. Deux monstres du même type mais
// de niveaux différents partagent le même template.
//
// Les *Growth sont exprimés en fraction (0.20 = +20% par niveau au-dessus
// du niveau 1), pour que chaque monstre puisse scaler différemment
// (un troll monte plus vite en HP qu'un rat, par exemple).
type EnemyTemplate struct {
	ID         string // identifiant stable, aligné sur guild.Quest.MonsterID
	Name       string
	BaseHP     uint16
	BaseAtk    uint16
	BaseSpeed  uint8 // ne scale pas avec le niveau
	BaseXPDrop uint16
	HPGrowth   float64
	AtkGrowth  float64
	XPGrowth   float64
	Attacks    []Attack
	Drops      []DropRate
}

// registry centralise tous les templates de monstres, indexés par ID.
var registry = make(map[string]EnemyTemplate)

// WithDrops associe une liste de drops possibles au template.
func (t EnemyTemplate) WithDrops(drops ...DropRate) EnemyTemplate {
	t.Drops = drops
	registry[t.ID] = t
	return t
}

// NewEnemyTemplate centralise la création ET l'enregistrement, pour ne
// jamais avoir un monstre défini mais introuvable par ID.
func NewEnemyTemplate(id, name string, baseHP, baseAtk uint16, baseSpeed uint8, baseXPDrop uint16, hpGrowth, atkGrowth, xpGrowth float64, attacks ...Attack) EnemyTemplate {
	t := EnemyTemplate{
		ID:         id,
		Name:       name,
		BaseHP:     baseHP,
		BaseAtk:    baseAtk,
		BaseSpeed:  baseSpeed,
		BaseXPDrop: baseXPDrop,
		HPGrowth:   hpGrowth,
		AtkGrowth:  atkGrowth,
		XPGrowth:   xpGrowth,
		Attacks:    attacks,
	}
	registry[id] = t
	return t
}

// GetTemplate recherche un template par son ID.
func GetTemplate(id string) (EnemyTemplate, bool) {
	t, ok := registry[id]
	return t, ok
}

// scale applique le taux de croissance à une valeur de base pour un niveau
// donné. Le niveau 1 renvoie toujours exactement la base, sans arrondi
// parasite. Le résultat est plafonné à math.MaxUint16 pour ne jamais
// déborder silencieusement sur un monstre de très haut niveau.
func scale(base uint16, growth float64, level uint8) uint16 {
	if level <= 1 {
		return base
	}
	factor := 1 + growth*float64(level-1)
	val := float64(base) * factor
	if val > math.MaxUint16 {
		val = math.MaxUint16
	}
	return uint16(val)
}

// HPAtLevel, AtkAtLevel, XPDropAtLevel : les 3 stats qui montent avec le
// niveau. La vitesse (BaseSpeed) n'a volontairement pas d'équivalent :
// elle reste fixe, propre à l'espèce plutôt qu'au niveau individuel.
func (t EnemyTemplate) HPAtLevel(level uint8) uint16 {
	return scale(t.BaseHP, t.HPGrowth, level)
}

func (t EnemyTemplate) AtkAtLevel(level uint8) uint16 {
	return scale(t.BaseAtk, t.AtkGrowth, level)
}

func (t EnemyTemplate) XPDropAtLevel(level uint8) uint16 {
	return scale(t.BaseXPDrop, t.XPGrowth, level)
}

// EnemyInstance est un monstre concret rencontré en combat : un template
// + un niveau + des HP courants qui diminuent au fil des coups reçus.
// Les stats (MaxHP, Atk, Speed, XPDrop) sont calculées UNE FOIS à la
// création, pas recalculées à chaque combat : un monstre ne "level up"
// pas en plein combat.
type EnemyInstance struct {
	Template EnemyTemplate
	Level    uint8
	MaxHP    uint16
	HP       uint16
	Atk      uint16
	Speed    uint8
	XPDrop   uint16
}

// NewEnemyInstance fait apparaître un monstre d'un type et d'un niveau
// donnés, en appliquant le scaling du template. false si templateID
// est inconnu.
func NewEnemyInstance(templateID string, level uint8) (*EnemyInstance, bool) {
	t, ok := GetTemplate(templateID)
	if !ok {
		return nil, false
	}
	maxHP := t.HPAtLevel(level)
	return &EnemyInstance{
		Template: t,
		Level:    level,
		MaxHP:    maxHP,
		HP:       maxHP,
		Atk:      t.AtkAtLevel(level),
		Speed:    t.BaseSpeed,
		XPDrop:   t.XPDropAtLevel(level),
	}, true
}

// IsAlive dit si le monstre peut encore combattre.
func (e *EnemyInstance) IsAlive() bool {
	return e.HP > 0
}

// TakeDamage inflige des dégâts, sans jamais descendre sous 0
// (uint16 wraparound sinon, ce qui ferait "revivre" le monstre à HP max).
func (e *EnemyInstance) TakeDamage(amount uint16) {
	if amount >= e.HP {
		e.HP = 0
		return
	}
	e.HP -= amount
}

// RandomAttack choisit une attaque au hasard parmi celles du template.
// Si aucune attaque n'est définie, renvoie une attaque générique basée
// sur Atk, pour qu'un monstre sans liste d'attaques reste jouable.
func (e *EnemyInstance) RandomAttack() Attack {
	if len(e.Template.Attacks) == 0 {
		return Attack{Name: "Strike", Damage: e.Atk}
	}
	return e.Template.Attacks[rand.Intn(len(e.Template.Attacks))]
}

// RollDrops tire au sort les objets laissés par le monstre à sa mort.
func (e *EnemyInstance) RollDrops() []item.Item {
	var dropped []item.Item
	for _, d := range e.Template.Drops {
		if rand.Float64() <= d.Chance {
			qty := d.Quantity
			if qty == 0 {
				qty = 1
			}
			for q := uint8(0); q < qty; q++ {
				dropped = append(dropped, d.Item)
			}
		}
	}
	return dropped
}

// --- Monstres définis ---
// Les ID ("rat", "wolf", "boar", "troll", "goblin") sont alignés sur
// guild.Quest.MonsterID, pour que RegisterKill(g, enemy.Template.ID)
// fonctionne directement une fois le combat implémenté.

var (
	Rat = NewEnemyTemplate("rat", "Rat", 10, 2, 8, 5,
		0.15, 0.10, 0.20,
		Attack{Name: "Bite", Damage: 2},
	).WithDrops(
		DropRate{Item: item.RavenFeather, Chance: 0.60, Quantity: 1},
		DropRate{Item: item.SmallHealPotion, Chance: 0.25, Quantity: 1},
	)

	Wolf = NewEnemyTemplate("wolf", "Wolf", 25, 5, 10, 10,
		0.18, 0.12, 0.22,
		Attack{Name: "Bite", Damage: 5},
		Attack{Name: "Claw", Damage: 4},
	).WithDrops(
		DropRate{Item: item.WolfFur, Chance: 0.75, Quantity: 1},
		DropRate{Item: item.SmallHealPotion, Chance: 0.20, Quantity: 1},
	)

	Boar = NewEnemyTemplate("boar", "Boar", 40, 7, 6, 15,
		0.20, 0.15, 0.25,
		Attack{Name: "Charge", Damage: 8},
	).WithDrops(
		DropRate{Item: item.BoarLeather, Chance: 0.80, Quantity: 1},
		DropRate{Item: item.HealPotion, Chance: 0.30, Quantity: 1},
	)

	Troll = NewEnemyTemplate("troll", "Troll", 120, 15, 4, 50,
		0.25, 0.20, 0.30,
		Attack{Name: "Club Smash", Damage: 15},
		Attack{Name: "Crush", Damage: 20},
	).WithDrops(
		DropRate{Item: item.TrollHide, Chance: 1.00, Quantity: 1},
		DropRate{Item: item.LargeHealPotion, Chance: 0.50, Quantity: 1},
		DropRate{Item: item.FireballBook, Chance: 0.20, Quantity: 1},
	)

	Goblin = NewEnemyTemplate("goblin", "Goblin", 15, 3, 7, 8,
		0.15, 0.10, 0.20,
		Attack{Name: "Club Strike", Damage: 3},
		Attack{Name: "Rock Throw", Damage: 2},
	).WithDrops(
		DropRate{Item: item.SmallHealPotion, Chance: 0.40, Quantity: 1},
		DropRate{Item: item.HealPotion, Chance: 0.20, Quantity: 1},
		DropRate{Item: item.WolfFur, Chance: 0.30, Quantity: 1},
	)
)
