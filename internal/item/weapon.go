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
	RustySword      = NewSword("Rusty Sword", 5, 0, 2, 0, Common, 1)
	WarAxe          = NewAxe("War Axe", 25, 15, 30, Rare, 1)
	ApprenticeStaff = NewStaff("Apprentice Staff", 10, 12, 24, Uncommon, 1)
)
