package character

import (
	"math"
	"math/rand"
	"runa/internal/spell"
)

type Speed int
type Mana int
type Strength int

// Class est un alias sur spell.Class : une seule définition existe
// (dans spell, pour éviter un import cycle), mais reste utilisable
// ici sous le nom character.Class / character.Human / etc.
type Class = spell.Class

const (
	Human = spell.Human
	Elf   = spell.Elf
	Dwarf = spell.Dwarf
)

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
	Level             uint8
	XP                uint16
	Hp                uint16
	HpMax             uint16
	Mana              uint16
	Speed             uint8
	Strength          uint8
	FreePotionClaimed bool
	Money             uint16
}

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

	return &Character{
		Name:              FormattedName,
		Class:             class,
		Level:             1,
		XP:                0,
		Hp:                HPMax,
		HpMax:             HPMax,
		Mana:              Mana(class).Mana(),
		Speed:             Speed(class).Speed(),
		Strength:          Strength(class).Strength(),
		FreePotionClaimed: false,
		Money:             100,
	}
}

func TotalXpForLevel(level uint8) int {
	if level <= 1 {
		return 0
	}
	return int(10 * (math.Pow(float64(level-1), 1.8)))
}

func (c *Character) XpNeededForNext() int {
	return TotalXpForLevel(c.Level+1) - TotalXpForLevel(c.Level)
}
func (c *Character) AddXP(amount uint16) bool {
	c.XP += amount
	levelUp := false

	for {
		xpNeeded := c.XpNeededForNext()
		if int(c.XP) >= xpNeeded && c.Level < 100 {
			c.XP -= uint16(xpNeeded)
			c.Level += 1
			levelUp = true
		} else {
			break
		}
	}
	return levelUp
}
