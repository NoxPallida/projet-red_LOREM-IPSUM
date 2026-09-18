package character

import (
	"math/rand"
)

func RandomInt(min, max int) int {
	return rand.Intn(max-min+1) + min // pour faciliter le use du random : min = chiffre min, max = chiffre max
}

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

func (s Speed) Speed() int {
	switch s {
	case Human:
		return RandomInt(5, 10)
	case Elf:
		return RandomInt(7, 12)
	case Dwarf:
		return RandomInt(1, 5)
	default:
		return 0
	}
}

func (s Strength) Strength() int {
	switch s {
	case Human:
		return RandomInt(1, 10)
	case Elf:
		return RandomInt(1, 5)
	case Dwarf:
		return RandomInt(5, 15)
	default:
		return 0
	}
}

func (m Mana) Mana() int {
	switch m {
	case Human:
		return RandomInt(1, 50)
	case Elf:
		return RandomInt(25, 75)
	case Dwarf:
		return RandomInt(10, 30)
	default:
		return 0
	}
}

type Character struct {
	Name              string
	Class             Class
	Level             int
	Hp                int
	HpMax             int
	Mana              int
	Speed             int
	Strength          int
	FreePotionClaimed bool
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
	HPMax := 100
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
	}
}
