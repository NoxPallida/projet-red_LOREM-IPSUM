package character

import (
	"math/rand"
)

type Class int
type Speed int
type Mana int
type Strength int

const (
	Human = 0
	Elf   = 1
	Dwarf = 2
)

func (c Class) String() string {
	switch c {
	case Human:
		return "Human"
	case Elf:
		return "Elf"
	case Dwarf:
		return "Dwarf"
	default:
		return "Unknown species"
	}
}

func (s Speed) Speed() uint8 {
	switch s {
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
	switch s {
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
	switch m {
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
			full_letters += string(letter) // mise en forme qu'avec les lettres dans le cas où d'autres caractères sont la
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
		Hp:                HPMax,
		HpMax:             HPMax,
		Mana:              Mana(class).Mana(),
		Speed:             Speed(class).Speed(),
		Strength:          Strength(class).Strength(),
		FreePotionClaimed: false,
		Money:             100,
	}
}
