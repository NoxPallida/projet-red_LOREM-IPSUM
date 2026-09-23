package character

import (
	"math"
	"math/rand"
	"runa/internal/inventory"
	"runa/internal/item"
	"runa/internal/spell"
)

type Speed int
type Mana int
type Strength int

// Class est un alias sur spell.Class : une seule définition existe
// (dans spell, pour éviter un import cycle), mais reste utilisable
// ici sous le nom character.Class / character.Human / etc.
type Class = spell.Class
type Subclass = spell.Subclass

const (
	Human = spell.Human
	Elf   = spell.Elf
	Dwarf = spell.Dwarf
)

var AllClasses = spell.AllClasses
var SubclassesForClass = spell.SubclassesForClass

func (s Speed) Speed() uint8 {
	switch Class(s) {
	case Human:
		return uint8(rand.Intn(10-4) + 5)
	case Elf:
		return uint8(rand.Intn(12-6) + 7)
	case Dwarf:
		return uint8(rand.Intn(5-0) + 1)
	default:
		return 0
	}
}

func (s Strength) Strength() uint8 {
	switch Class(s) {
	case Human:
		return uint8(rand.Intn(10-0) + 1)
	case Elf:
		return uint8(rand.Intn(5-0) + 1)
	case Dwarf:
		return uint8(rand.Intn(15-4) + 5)
	default:
		return 0
	}
}

func (m Mana) Mana() uint16 {
	switch Class(m) {
	case Human:
		return uint16(rand.Intn(50-0) + 1)
	case Elf:
		return uint16(rand.Intn(75-24) + 25)
	case Dwarf:
		return uint16(rand.Intn(30-9) + 10)
	default:
		return 0
	}
}

type Character struct {
	Name              string
	Class             Class
	Subclass          Subclass
	level             uint8 // privé : accès via Level(), pour satisfaire guild.Member/etc.
	XP                uint16
	Hp                uint16
	HpMax             uint16
	Mana              uint16
	ManaMax           uint16
	Speed             uint8
	Strength          uint8
	FreePotionClaimed bool
	money             uint16 // privé : accès via Money()/SpendMoney()/EarnMoney()
	Inventory         inventory.Inventory
	Chest             inventory.Chest
	EquippedWeapon    *item.Weapon
	EquippedArmor     [4]*item.Equipment // indexé par item.SlotHead..SlotFeet, nil = vide
	KnownSpells       map[string]bool
}

// Level expose le niveau du personnage. Nécessaire pour satisfaire
// guild.Member (et plus tard toute interface qui a besoin du niveau).
func (c *Character) Level() uint8 { return c.level }

// Money expose l'argent du personnage. Nécessaire pour shop.Customer
// et forge.Crafter (qui attendent Money() uint16, pas un champ).
func (c *Character) Money() uint16 { return c.money }

// SpendMoney retire de l'argent si le personnage en a assez ;
// renvoie false sans rien modifier sinon.
func (c *Character) SpendMoney(amount uint16) bool {
	if c.money < amount {
		return false
	}
	c.money -= amount
	return true
}

// EarnMoney ajoute de l'argent, avec une protection contre le
// dépassement de capacité d'un uint16 (65535 max).
func (c *Character) EarnMoney(amount uint16) {
	total := uint32(c.money) + uint32(amount)
	if total > math.MaxUint16 {
		total = math.MaxUint16
	}
	c.money = uint16(total)
}

// HasClaimedFreePotion / ClaimFreePotion : requis par shop.Customer.
func (c *Character) HasClaimedFreePotion() bool { return c.FreePotionClaimed }
func (c *Character) ClaimFreePotion()           { c.FreePotionClaimed = true }

func FormatName(name string) string {
	if len(name) == 0 {
		return "Traveler"
	}
	if len(name) > 15 {
		return "Traveler"
	}
	full_letters := ""
	for _, letter := range name {
		if 'a' <= letter && letter <= 'z' || 'A' <= letter && letter <= 'Z' {
			full_letters += string(letter)
		}
	}
	fin_name := ""
	for i, lettre := range full_letters {
		if i == 0 && 'a' <= lettre && lettre <= 'z' {
			fin_name += string(lettre - 32)
		} else if i >= 1 && 'A' <= lettre && lettre <= 'Z' {
			fin_name += string(lettre + 32)
		} else {
			fin_name += string(lettre)
		}
	}
	if len(fin_name) == 0 {
		return "Traveler"
	}
	return fin_name
}

func InitCharacter(name string, class Class) *Character {
	FormattedName := FormatName(name)
	var HPMax uint16
	switch class {
	case Human:
		HPMax = 100
	case Elf:
		HPMax = 80
	case Dwarf:
		HPMax = 120
	}

	manaRoll := Mana(class).Mana()

	ch := &Character{
		Name:              FormattedName,
		Class:             class,
		level:             1,
		XP:                0,
		Hp:                HPMax,
		HpMax:             HPMax,
		Mana:              manaRoll,
		ManaMax:           manaRoll, // le perso commence toujours à mana plein
		KnownSpells:       make(map[string]bool),
		Speed:             Speed(class).Speed(),
		Strength:          Strength(class).Strength(),
		FreePotionClaimed: false,
		money:             100,
		Inventory:         inventory.NewInventory(),
		Chest:             inventory.NewChest(),
	}
	ch.LearnAvailableSpells()
	return ch
}

func TotalXpForLevel(level uint8) int {
	if level <= 1 {
		return 0
	}
	return int(10 * (math.Pow(float64(level-1), 1.8)))
}

func (c *Character) XpNeededForNext() int {
	return TotalXpForLevel(c.level+1) - TotalXpForLevel(c.level)
}

func (c *Character) AddXP(amount uint16) bool {
	c.XP += amount
	levelUp := false

	for {
		xpNeeded := c.XpNeededForNext()
		if int(c.XP) >= xpNeeded && c.level < 100 {
			c.XP -= uint16(xpNeeded)
			c.level += 1
			levelUp = true

			// --- Bonus de niveau ---
			c.Strength += 1 // +1 Atk (Force) par niveau
			c.ManaMax += 3  // +3 Mana Max par niveau
			c.Mana += 3
			c.LearnAvailableSpells()
		} else {
			break
		}
	}
	return levelUp
}

// GainExp est l'équivalent d'AddXP mais en uint32, pour satisfaire
// guild.Member (les récompenses de quêtes hauts rangs pourraient
// dépasser la portée d'un uint16 un jour). Découpe l'ajout par
// tranches de MaxUint16 pour ne jamais dépasser AddXP.
func (c *Character) GainExp(amount uint16) {
	c.AddXP(amount)
}

// TotalAttack combine la force de base, le bonus de niveau (+1 par niveau au-dessus du niv 1)
// et les dégâts de l'arme équipée.
func (c *Character) TotalAttack() uint16 {
	total := uint16(c.Strength) + uint16(c.level-1)
	if c.EquippedWeapon != nil {
		total += uint16(c.EquippedWeapon.Damage)
	}
	return total
}

// GetManaMax retourne le mana max de base + le bonus du niveau (+3 par niveau au-dessus du niv 1)
// + le bonus de mana fourni par l'arme équipée.
func (c *Character) GetManaMax() uint16 {
	total := c.ManaMax + uint16(c.level-1)*3
	if c.EquippedWeapon != nil {
		total += uint16(c.EquippedWeapon.Mana)
	}
	return total
}

// EquipWeapon équipe l'arme (retirée de l'inventaire par l'appelant).
// L'ancienne arme repart dans l'inventaire au lieu d'être perdue.
func (c *Character) EquipWeapon(w item.Weapon) {
	c.ensureInventory()
	if old := c.EquippedWeapon; old != nil {
		_ = c.Inventory.AddItem(*old, 1)
	}
	cp := w
	c.EquippedWeapon = &cp
}

// UnequipWeapon déséquipe l'arme vers l'inventaire.
// Rend false si vide ou inventaire plein.
func (c *Character) UnequipWeapon() bool {
	if c.EquippedWeapon == nil {
		return false
	}
	c.ensureInventory()
	if err := c.Inventory.AddItem(*c.EquippedWeapon, 1); err != nil {
		return false
	}
	c.EquippedWeapon = nil
	return true
}

// EquipArmor équipe une pièce d'armure (slot déterminé par la pièce).
// L'ancienne pièce repart à l'inventaire, le bonus HP s'applique
// au max ET au courant (sinon l'équiper à pleine vie ne servirait à rien).
func (c *Character) EquipArmor(e item.Equipment) {
	c.ensureInventory()
	slot := e.Slot
	if slot < item.SlotHead || slot > item.SlotFeet {
		return
	}
	_ = c.Inventory.RemoveItem(e.Name(), 1)
	if old := c.EquippedArmor[slot]; old != nil {
		_ = c.Inventory.AddItem(*old, 1)
	}
	cp := e
	c.EquippedArmor[slot] = &cp
	c.HpMax += uint16(e.HPBonus)
	c.Hp += uint16(e.HPBonus)
}

// UnequipArmor déséquipe un slot vers l'inventaire.
// Rend false si vide ou inventaire plein. Le bonus HP est retiré
// (courant clampé au nouveau max).
func (c *Character) UnequipArmor(slot item.EquipmentSlot) bool {
	if slot < item.SlotHead || slot > item.SlotFeet {
		return false
	}
	old := c.EquippedArmor[slot]
	if old == nil {
		return false
	}
	c.ensureInventory()
	if err := c.Inventory.AddItem(*old, 1); err != nil {
		return false
	}
	if bonus := uint16(old.HPBonus); c.HpMax > bonus {
		c.HpMax -= bonus
	} else {
		c.HpMax = 0
	}
	if c.Hp > c.HpMax {
		c.Hp = c.HpMax
	}
	c.EquippedArmor[slot] = nil
	return true
}

// LearnSpell satisfait item.Learner (cf. item-effect.go, UseSpellBook) :
// renvoie false si le sort est déjà connu, pour que le livre ne soit
// jamais "gâché" sur un sort déjà appris.
func (c *Character) LearnSpell(spellID string) bool {
	if c.KnownSpells == nil {
		c.KnownSpells = make(map[string]bool)
	}
	if c.KnownSpells[spellID] {
		return false
	}
	c.KnownSpells[spellID] = true
	return true
}

// BaseAttack renvoie le sort/l'attaque de base propre à la classe et sous-classe du personnage.
func (c *Character) BaseAttack() spell.Spell {
	return spell.BaseAttackFor(c.Class, c.Subclass)
}

// LearnAvailableSpells apprend automatiquement les sorts débloqués selon la classe, sous-classe et niveau.
func (c *Character) LearnAvailableSpells() {
	if c.KnownSpells == nil {
		c.KnownSpells = make(map[string]bool)
	}
	for _, s := range spell.AvailableSpells(c.Class, c.Subclass, c.level) {
		c.KnownSpells[s.ID] = true
	}
}

// SetSubclass configure la sous-classe du personnage et met à jour ses sorts.
func (c *Character) SetSubclass(sub Subclass) {
	c.Subclass = sub
	newKnown := make(map[string]bool)
	for id, known := range c.KnownSpells {
		if s, ok := spell.GetSpell(id); ok && s.Method == spell.ObtainByBook && known {
			newKnown[id] = true
		}
	}
	c.KnownSpells = newKnown
	c.LearnAvailableSpells()
}
