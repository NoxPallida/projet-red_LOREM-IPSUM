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
		return "Unknown"
	}
}

// AllClasses liste toutes les races définies dans le jeu.
var AllClasses = []Class{Human, Elf, Dwarf}

// SubclassesForClass renvoie les spécialisations définies pour la race donnée.
func SubclassesForClass(c Class) []Subclass {
	switch c {
	case Human:
		return []Subclass{SubclassHumanCQC, SubclassHumanMage}
	case Elf:
		return []Subclass{SubclassElfArcher, SubclassElfSpiritMage}
	case Dwarf:
		return []Subclass{SubclassDwarfWarrior, SubclassDwarfBerserker}
	default:
		return nil
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
	SubclassDwarfWarrior
	SubclassDwarfBerserker
)

func (s Subclass) String() string {
	switch s {
	case SubclassHumanCQC:
		return "CQC Warrior"
	case SubclassHumanMage:
		return "Mage"
	case SubclassElfArcher:
		return "Archer"
	case SubclassElfSpiritMage:
		return "Spirit Mage"
	case SubclassDwarfWarrior:
		return "Dwarf Warrior"
	case SubclassDwarfBerserker:
		return "Berserker"
	default:
		return "Adventurer"
	}
}

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

// BaseAttackFor renvoie l'attaque de base (niveau 1) propre à la classe et sous-classe.
func BaseAttackFor(class Class, subclass Subclass) Spell {
	for _, s := range registry {
		if s.Method == ObtainByLevel && s.MinLevel <= 1 {
			if s.Class == class && s.Subclass == subclass {
				return s
			}
		}
	}
	for _, s := range registry {
		if s.Method == ObtainByLevel && s.MinLevel <= 1 {
			if (s.Class == class || s.Class == ClassAny) && (s.Subclass == subclass || s.Subclass == SubclassAny) {
				return s
			}
		}
	}
	return Punch
}

// --- Sorts et attaques définis ---
// id name dmg mana method lvl race classe
var (
	// Level 1: Class base attacks (0 mana cost)
	Punch       = NewSpell("punch", "Punch", 2, 0, ObtainByLevel, 2, Human, SubclassHumanCQC)
	ArcaneSpark = NewSpell("arcane_spark", "Arcane Spark", 2, 0, ObtainByLevel, 1, Human, SubclassHumanMage)
	QuickShot   = NewSpell("quick_shot", "Quick Shot", 2, 0, ObtainByLevel, 1, Elf, SubclassElfArcher)
	SpiritOrb   = NewSpell("spirit_orb", "Spirit Orb", 2, 0, ObtainByLevel, 1, Elf, SubclassElfSpiritMage)
	HeavyStrike = NewSpell("heavy_strike", "Heavy Strike", 2, 0, ObtainByLevel, 1, Dwarf, SubclassDwarfWarrior)
	FuriousBlow = NewSpell("furious_blow", "Furious Strike", 2, 0, ObtainByLevel, 1, Dwarf, SubclassDwarfBerserker)

	// Level 10: Advanced skills (mana cost)
	Slash       = NewSpell("slash", "Simple Slash", 15, 5, ObtainByLevel, 10, Human, SubclassHumanCQC)
	ArcaneBolt  = NewSpell("arcane_bolt", "Arcane Bolt", 14, 8, ObtainByLevel, 10, Human, SubclassHumanMage)
	PreciseShot = NewSpell("precise_shot", "Precise Shot", 14, 6, ObtainByLevel, 10, Elf, SubclassElfArcher)
	NatureHeal  = NewSpell("nature_heal", "Nature Heal", 0, 12, ObtainByLevel, 10, Elf, SubclassElfSpiritMage)
	ShieldBash  = NewSpell("shield_bash", "Shield Bash", 15, 5, ObtainByLevel, 10, Dwarf, SubclassDwarfWarrior)
	RageStrike  = NewSpell("rage_strike", "Rage Strike", 18, 8, ObtainByLevel, 10, Dwarf, SubclassDwarfBerserker)

	// Level 9-50 : Human Mage, la voie la plus fournie en sorts (progression arcanique complète)
	FrostShard     = NewSpell("frost_shard", "Frost Shard", 24, 12, ObtainByLevel, 9, Human, SubclassHumanMage)
	LightningSurge = NewSpell("lightning_surge", "Lightning Surge", 34, 15, ObtainByLevel, 15, Human, SubclassHumanMage)
	ArcaneNova     = NewSpell("arcane_nova", "Arcane Nova", 48, 20, ObtainByLevel, 24, Human, SubclassHumanMage)
	MeteorShard    = NewSpell("meteor_shard", "Meteor Shard", 65, 26, ObtainByLevel, 35, Human, SubclassHumanMage)
	ArchmagesWrath = NewSpell("archmages_wrath", "Archmage's Wrath", 85, 32, ObtainByLevel, 50, Human, SubclassHumanMage)

	// Level 9-50 : Elf Spirit Mage, magie de la nature, aussi fournie que le mage humain
	ThornLash       = NewSpell("thorn_lash", "Thorn Lash", 22, 11, ObtainByLevel, 9, Elf, SubclassElfSpiritMage)
	WildfireBloom   = NewSpell("wildfire_bloom", "Wildfire Bloom", 32, 14, ObtainByLevel, 15, Elf, SubclassElfSpiritMage)
	AncestralSurge  = NewSpell("ancestral_surge", "Ancestral Surge", 46, 19, ObtainByLevel, 24, Elf, SubclassElfSpiritMage)
	TempestCall     = NewSpell("tempest_call", "Tempest Call", 62, 24, ObtainByLevel, 35, Elf, SubclassElfSpiritMage)
	WorldTreesWrath = NewSpell("world_trees_wrath", "World Tree's Wrath", 80, 30, ObtainByLevel, 50, Elf, SubclassElfSpiritMage)

	// Level 20 : un seul sort avancé de plus pour les voies peu magiques
	CrossCut         = NewSpell("cross_cut", "Cross Cut", 26, 9, ObtainByLevel, 20, Human, SubclassHumanCQC)
	PiercingArrow    = NewSpell("piercing_arrow", "Piercing Arrow", 25, 9, ObtainByLevel, 20, Elf, SubclassElfArcher)
	EarthbreakerSlam = NewSpell("earthbreaker_slam", "Earthbreaker Slam", 27, 9, ObtainByLevel, 20, Dwarf, SubclassDwarfWarrior)
	BloodfuryRampage = NewSpell("bloodfury_rampage", "Bloodfury Rampage", 30, 10, ObtainByLevel, 20, Dwarf, SubclassDwarfBerserker)

	// Spells obtained by book
	Fireball       = NewSpell("fireball", "Fireball", 20, 10, ObtainByBook, 0, ClassAny, SubclassAny)
	Heal           = NewSpell("heal", "Heal", 0, 15, ObtainByBook, 0, ClassAny, SubclassAny)
	ShadowBolt     = NewSpell("shadow_bolt", "Shadow Bolt", 25, 12, ObtainByBook, 0, ClassAny, SubclassAny)
	PoisonDart     = NewSpell("poison_dart", "Poison Dart", 18, 8, ObtainByBook, 0, ClassAny, SubclassAny)
	ChainLightning = NewSpell("chain_lightning", "Chain Lightning", 35, 18, ObtainByBook, 0, ClassAny, SubclassAny)
	DivineSmite    = NewSpell("divine_smite", "Divine Smite", 40, 20, ObtainByBook, 0, ClassAny, SubclassAny)
	LifeDrain      = NewSpell("life_drain", "Life Drain", 22, 11, ObtainByBook, 0, ClassAny, SubclassAny)
	Earthquake     = NewSpell("earthquake", "Earthquake", 50, 25, ObtainByBook, 0, ClassAny, SubclassAny)
)
