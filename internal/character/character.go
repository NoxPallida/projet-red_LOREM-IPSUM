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
	EquippedWeapon    *item.Weapon
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
		return "Michel Ier"
	}
	if len(name) > 15 {
		return "Michel IIe"
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
		return "Michel Ier"
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

	return &Character{
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
	}
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

// TotalAttack combine la force du personnage et les dégâts de son arme
// équipée (0 si aucune arme). C'est cette valeur que combat.go utilise
// pour l'attaque de base.
func (c *Character) TotalAttack() uint16 {
	total := uint16(c.Strength)
	if c.EquippedWeapon != nil {
		total += uint16(c.EquippedWeapon.Damage)
	}
	return total
}

// EquipWeapon change l'arme équipée du personnage.
func (c *Character) EquipWeapon(w item.Weapon) {
	c.EquippedWeapon = &w
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
