package item

type WeaponType int

const (
	SwordType WeaponType = iota
	AxeType
	StaffType
)

type Weapon struct {
	Base
	Kind         WeaponType
	Damage       uint8
	Mana         uint8
	ItemPriceBuy uint8
}

func (w Weapon) Type() ItemType { return TypeWeapon }

func (w Weapon) PriceBuy() uint8 { return w.ItemPriceBuy }

func newWeapon(name string, kind WeaponType, damage, mana, pricesell, pricebuy uint8, rarity Rarity, maxStack uint8) Weapon {
	return Weapon{
		Base: Base{
			ItemName:      name,
			ItemPriceSell: pricesell,
			ItemRarity:    rarity,
			ItemMaxStack:  maxStack,
		},
		Kind:         kind,
		Damage:       damage,
		Mana:         mana,
		ItemPriceBuy: pricebuy,
	}
}

// NewAxe : uniquement des dégâts, pas de mana.
func NewAxe(name string, damage, pricesell, pricebuy uint8, rarity Rarity, maxStack uint8) Weapon {
	return newWeapon(name, AxeType, damage, 0, pricesell, pricebuy, rarity, maxStack)
}

// NewStaff : uniquement du mana, pas de dégâts.
func NewStaff(name string, mana, pricesell, pricebuy uint8, rarity Rarity, maxStack uint8) Weapon {
	return newWeapon(name, StaffType, 0, mana, pricesell, pricebuy, rarity, maxStack)
}

// NewSword : peut avoir les deux stats.
func NewSword(name string, damage, mana, pricesell, pricebuy uint8, rarity Rarity, maxStack uint8) Weapon {
	return newWeapon(name, SwordType, damage, mana, pricesell, pricebuy, rarity, maxStack)
}

var (
	// AXES  - Uniquement dégâts, 0 mana
	StarterAxe          = NewAxe("Starter Axe", 5, 1, 2, Common, 1)
	GoblinCleaver       = NewAxe("Goblin Cleaver", 12, 4, 8, Common, 1)
	BoarSplitter        = NewAxe("Boar Splitter", 20, 8, 16, Uncommon, 1)
	WarAxe              = NewAxe("War Axe", 25, 15, 30, Rare, 1)
	OrcGreataxe         = NewAxe("Orc Greataxe", 35, 25, 50, Rare, 1)
	MinotaurExecutioner = NewAxe("Minotaur Executioner", 55, 60, 120, Epic, 1)
	WyvernDecapitator   = NewAxe("Wyvern Decapitator", 85, 130, 255, Legendary, 1)

	// STAFFS - Uniquement mana, 0 dégâts
	StarterStaff        = NewStaff("Starter Staff", 5, 1, 2, Common, 1)
	SlimeScepter        = NewStaff("Slime Scepter", 10, 4, 8, Common, 1)
	ApprenticeStaff     = NewStaff("Apprentice Staff", 15, 12, 24, Uncommon, 1)
	HarpyWand           = NewStaff("Harpy Wand", 25, 20, 40, Uncommon, 1)
	BoneStaff           = NewStaff("Bone Staff", 40, 35, 70, Rare, 1)
	WyrmRod             = NewStaff("Wyrm Rod", 60, 90, 180, Epic, 1)
	ArchmageWyvernStaff = NewStaff("Archmage Wyvern Staff", 90, 130, 255, Legendary, 1)

	// SWORDS - Mana ET dégâts
	StarterSword    = NewSword("Starter Sword", 5, 0, 1, 2, Common, 1)
	KoboldDagger    = NewSword("Kobold Dagger", 10, 0, 3, 6, Common, 1)
	WolfToothBlade  = NewSword("Wolf Tooth Blade", 18, 0, 10, 20, Uncommon, 1)
	Spellblade      = NewSword("Runic Spellblade", 28, 15, 40, 80, Rare, 1)
	TrollSlayer     = NewSword("Troll Slayer Greatsword", 38, 0, 50, 100, Rare, 1)
	WyrmSaber       = NewSword("Wyrm Scale Saber", 52, 10, 100, 200, Epic, 1)
	WyvernFangBlade = NewSword("Wyvern Fang Blade", 75, 15, 130, 255, Legendary, 1)

	// BOWS - Déclarés avec NewSword just bc
	StarterBow       = NewSword("Starter Bow", 5, 0, 1, 2, Common, 1)
	RatRunnerBow     = NewSword("Rat Skin Shortbow", 9, 2, 3, 6, Common, 1)
	BoarRecurveBow   = NewSword("Boar Sinew Bow", 16, 0, 9, 18, Uncommon, 1)
	HarpyFeatherBow  = NewSword("Harpy Feather Bow", 24, 5, 20, 40, Uncommon, 1)
	BoneLongbow      = NewSword("Bone Splinter Longbow", 36, 0, 45, 90, Rare, 1)
	MinotaurGreatbow = NewSword("Minotaur Horn Greatbow", 54, 0, 110, 220, Epic, 1)
	WyvernStalkerBow = NewSword("Wyvern Stalker Bow", 72, 20, 130, 255, Legendary, 1)
)
