package enemies

import (
	"math"
	"math/rand"

	"runa/src/internal/item"
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
// ID name hp atk speed xp
var (
	Rat = NewEnemyTemplate("rat", "Rat", 3, 1, 8, 2,
		0.11, 0.10, 0.15,
		Attack{Name: "Bite", Damage: 1},
	).WithDrops(
		DropRate{Item: item.RatHide, Chance: 0.60, Quantity: 1},
		DropRate{Item: item.SmallHealPotion, Chance: 0.05, Quantity: 1},
	)

	Slime = NewEnemyTemplate("slime", "Slime", 20, 0, 2, 1,
		0.75, 0, 0.05,
		Attack{Name: "Wobble", Damage: 0},
	).WithDrops(
		DropRate{Item: item.SlimeMucus, Chance: 1, Quantity: 2},
	)

	Goblin = NewEnemyTemplate("goblin", "Goblin", 10, 2, 4, 5,
		0.13, 0.10, 0.16,
		Attack{Name: "Club Strike", Damage: 3},
		Attack{Name: "Rock Throw", Damage: 2},
	).WithDrops(
		DropRate{Item: item.SmallHealPotion, Chance: 0.10, Quantity: 1},
		DropRate{Item: item.GoblinSkin, Chance: 0.30, Quantity: 1},
		DropRate{Item: item.FireballBook, Chance: 0.05, Quantity: 1},
	)

	Hobgoblin = NewEnemyTemplate("hobgoblin", "Hobgoblin", 17, 4, 6, 10,
		0.16, 0.11, 0.18,
		Attack{Name: "Club Strike", Damage: 5},
		Attack{Name: "Rock Throw", Damage: 4},
	).WithDrops(
		DropRate{Item: item.SmallHealPotion, Chance: 0.15, Quantity: 1},
		DropRate{Item: item.HealPotion, Chance: 0.05, Quantity: 1},
		DropRate{Item: item.GoblinSkin, Chance: 0.40, Quantity: 2},
		DropRate{Item: item.ShadowBoltBook, Chance: 0.05, Quantity: 1},
	)

	Wolf = NewEnemyTemplate("wolf", "Wolf", 25, 5, 10, 10,
		0.18, 0.12, 0.22,
		Attack{Name: "Bite", Damage: 5},
		Attack{Name: "Claw", Damage: 4},
	).WithDrops(
		DropRate{Item: item.WolfFur, Chance: 0.75, Quantity: 1},
		DropRate{Item: item.WolfClaw, Chance: 0.15, Quantity: 3},
		DropRate{Item: item.SmallHealPotion, Chance: 0.05, Quantity: 1},
		DropRate{Item: item.HealBook, Chance: 0.10, Quantity: 1},
	)

	Boar = NewEnemyTemplate("boar", "Boar", 40, 7, 6, 15,
		0.20, 0.15, 0.25,
		Attack{Name: "Charge", Damage: 8},
	).WithDrops(
		DropRate{Item: item.BoarLeather, Chance: 0.80, Quantity: 1},
		DropRate{Item: item.BoarTusk, Chance: 0.10, Quantity: 1},
		DropRate{Item: item.SmallHealPotion, Chance: 0.15, Quantity: 1},
	)

	Troll = NewEnemyTemplate("troll", "Troll", 120, 15, 4, 50,
		0.25, 0.20, 0.30,
		Attack{Name: "Club Smash", Damage: 15},
		Attack{Name: "Crush", Damage: 20},
	).WithDrops(
		DropRate{Item: item.TrollHide, Chance: 1.00, Quantity: 1},
		DropRate{Item: item.HealPotion, Chance: 0.20, Quantity: 1},
		DropRate{Item: item.EarthquakeBook, Chance: 0.05, Quantity: 1}, // Thématique physique/terre
	)

	Kobold = NewEnemyTemplate("kobold", "Kobold", 14, 3, 7, 6,
		0.13, 0.11, 0.17,
		Attack{Name: "Dagger Stab", Damage: 3},
		Attack{Name: "Sneak Attack", Damage: 4},
	).WithDrops(
		DropRate{Item: item.KoboldFang, Chance: 0.55, Quantity: 1},
		DropRate{Item: item.SmallHealPotion, Chance: 0.08, Quantity: 1},
		DropRate{Item: item.PoisonDartBook, Chance: 0.08, Quantity: 1}, // Petit mob sournois
	)

	SkeletonWarrior = NewEnemyTemplate("skeleton_warrior", "Skeleton Warrior", 45, 8, 5, 18,
		0.17, 0.13, 0.19,
		Attack{Name: "Bone Slash", Damage: 8},
		Attack{Name: "Rattle Strike", Damage: 6},
	).WithDrops(
		DropRate{Item: item.BoneShard, Chance: 0.70, Quantity: 2},
		DropRate{Item: item.HealPotion, Chance: 0.10, Quantity: 1},
		DropRate{Item: item.LifeDrainBook, Chance: 0.06, Quantity: 1}, // Thématique Mort-vivant
	)

	Harpy = NewEnemyTemplate("harpy", "Harpy", 35, 6, 14, 16,
		0.16, 0.12, 0.20,
		Attack{Name: "Talon Dive", Damage: 6},
		Attack{Name: "Shriek", Damage: 4},
	).WithDrops(
		DropRate{Item: item.HarpyFeather, Chance: 0.65, Quantity: 2},
		DropRate{Item: item.SmallHealPotion, Chance: 0.10, Quantity: 1},
		DropRate{Item: item.IceBarrierBook, Chance: 0.07, Quantity: 1}, // Sort utilitaire/vent/glace
	)

	Orc = NewEnemyTemplate("orc", "Orc", 55, 9, 5, 20,
		0.19, 0.14, 0.23,
		Attack{Name: "Axe Cleave", Damage: 9},
		Attack{Name: "Shoulder Bash", Damage: 6},
	).WithDrops(
		DropRate{Item: item.OrcHide, Chance: 0.70, Quantity: 1},
		DropRate{Item: item.OrcTusk, Chance: 0.25, Quantity: 1},
		DropRate{Item: item.HealPotion, Chance: 0.08, Quantity: 1},
	)

	Ogre = NewEnemyTemplate("ogre", "Ogre", 160, 18, 3, 70,
		0.24, 0.18, 0.28,
		Attack{Name: "Club Smash", Damage: 18},
		Attack{Name: "Ground Slam", Damage: 22},
	).WithDrops(
		DropRate{Item: item.OgreClub, Chance: 0.40, Quantity: 1},
		DropRate{Item: item.HealPotion, Chance: 0.20, Quantity: 1},
		DropRate{Item: item.EarthquakeBook, Chance: 0.08, Quantity: 1}, // Sort lourd de choc
	)

	Minotaur = NewEnemyTemplate("minotaur", "Minotaur", 200, 22, 7, 90,
		0.25, 0.19, 0.29,
		Attack{Name: "Horn Charge", Damage: 22},
		Attack{Name: "Axe Swing", Damage: 18},
	).WithDrops(
		DropRate{Item: item.MinotaurHorn, Chance: 0.35, Quantity: 1},
		DropRate{Item: item.LargeHealPotion, Chance: 0.15, Quantity: 1},
		DropRate{Item: item.DivineSmiteBook, Chance: 0.05, Quantity: 1}, // Gros dégâts de zone / sacrer
	)

	Wyrm = NewEnemyTemplate("wyrm", "Wyrm", 260, 28, 6, 130,
		0.27, 0.21, 0.31,
		Attack{Name: "Venom Bite", Damage: 24},
		Attack{Name: "Tail Sweep", Damage: 28},
	).WithDrops(
		DropRate{Item: item.WyrmScale, Chance: 0.40, Quantity: 1},
		DropRate{Item: item.LargeHealPotion, Chance: 0.20, Quantity: 1},
		DropRate{Item: item.ChainLightningBook, Chance: 0.08, Quantity: 1}, // Créature magique avancée
	)

	Wyvern = NewEnemyTemplate("wyvern", "Wyvern", 400, 35, 9, 250,
		0.30, 0.24, 0.35,
		Attack{Name: "Ruby Gaze", Damage: 30},
		Attack{Name: "Serpent Coil", Damage: 35},
		Attack{Name: "Wing Slash", Damage: 25},
	).WithDrops(
		DropRate{Item: item.WyvernScale, Chance: 0.60, Quantity: 1},
		DropRate{Item: item.WyvernFang, Chance: 0.25, Quantity: 1},
		DropRate{Item: item.TitanicHealPotion, Chance: 0.30, Quantity: 1},
		DropRate{Item: item.ChainLightningBook, Chance: 0.12, Quantity: 1}, // Boss / Élite
		DropRate{Item: item.DivineSmiteBook, Chance: 0.08, Quantity: 1},
	)
)
