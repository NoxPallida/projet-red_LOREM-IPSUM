package spell

// Class représente la race du personnage. Défini ici (et non dans character)
// pour éviter un import cycle : character importe spell, jamais l'inverse.
type Class int

const (
	Human Class = iota
	Elf
	Dwarf
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

// Subclass représente la spécialisation du personnage, dépendante de sa race.
type Subclass int

const (
	SubclassAny Subclass = iota - 1 // pas de restriction de sous-classe
	SubclassHumanCQC
	SubclassHumanMage
	SubclassElfArcher
	SubclassElfSpiritMage
	// SubclassDwarf... : à définir avec le groupe, cf. README
)

// ClassAny : pas de restriction de race (ex: Coup de poing).
const ClassAny Class = -1

// ObtainMethod distingue comment un sort peut être obtenu.
type ObtainMethod int

const (
	ObtainByLevel ObtainMethod = iota
	ObtainByBook
)

// Spell décrit un sort/compétence, indépendamment du personnage qui le possède.
type Spell struct {
	ID       string
	Name     string
	Damage   uint8
	ManaCost uint8

	Method   ObtainMethod
	MinLevel uint8 // pertinent uniquement si Method == ObtainByLevel

	Class    Class    // ClassAny si accessible à toutes les races
	Subclass Subclass // SubclassAny si accessible à toutes les sous-classes
}

var registry = make(map[string]Spell)

// NewSpell centralise la création ET l'enregistrement, pour ne jamais
// avoir un sort défini mais introuvable par ID.
func NewSpell(id, name string, damage, manaCost uint8, method ObtainMethod, minLevel uint8, class Class, subclass Subclass) Spell {
	s := Spell{
		ID:       id,
		Name:     name,
		Damage:   damage,
		ManaCost: manaCost,
		Method:   method,
		MinLevel: minLevel,
		Class:    class,
		Subclass: subclass,
	}
	registry[id] = s
	return s
}

func GetSpell(id string) (Spell, bool) {
	s, ok := registry[id]
	return s, ok
}

// AvailableSpells renvoie les sorts qu'un personnage peut apprendre par
// montée de niveau, selon sa race, sa sous-classe et son niveau actuel.
// Ne renvoie jamais les sorts ObtainByBook (ceux-là passent uniquement
// par item.UseSpellBook).
func AvailableSpells(class Class, subclass Subclass, level uint8) []Spell {
	var out []Spell
	for _, s := range registry {
		if s.Method != ObtainByLevel {
			continue
		}
		if level < s.MinLevel {
			continue
		}
		if s.Class != ClassAny && s.Class != class {
			continue
		}
		if s.Subclass != SubclassAny && s.Subclass != subclass {
			continue
		}
		out = append(out, s)
	}
	return out
}

// --- Sorts définis ---

var (
	Punch    = NewSpell("punch", "Coup de poing", 5, 0, ObtainByLevel, 1, ClassAny, SubclassAny)
	Fireball = NewSpell("fireball", "Boule de feu", 20, 15, ObtainByBook, 0, ClassAny, SubclassAny)

	Slash      = NewSpell("slash", "Taillade", 15, 5, ObtainByLevel, 3, Human, SubclassHumanCQC)
	ArcaneBolt = NewSpell("arcane_bolt", "Projectile arcanique", 12, 10, ObtainByLevel, 3, Human, SubclassHumanMage)

	PreciseShot = NewSpell("precise_shot", "Tir précis", 14, 6, ObtainByLevel, 3, Elf, SubclassElfArcher)
	NatureHeal  = NewSpell("nature_heal", "Soin de la nature", 0, 12, ObtainByLevel, 3, Elf, SubclassElfSpiritMage)
)
