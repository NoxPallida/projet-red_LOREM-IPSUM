package item

type EquipmentSlot int

const (
	SlotHead EquipmentSlot = iota
	SlotChest
	SlotLeg
	SlotFeet
)

type Equipment struct {
	Base
	Slot         EquipmentSlot
	HPBonus      uint8
	ItemPriceBuy uint8
}

func (e Equipment) Type() ItemType {
	return TypeEquipment
}

func (e Equipment) PriceBuy() uint8 { return e.ItemPriceBuy }

func NewEquipment(name string, slot EquipmentSlot, hpBonus, pricesell, pricebuy uint8, rarity Rarity, maxStack uint8) Equipment {
	return Equipment{
		Base: Base{
			ItemName:      name,
			ItemPriceSell: pricesell,
			ItemRarity:    rarity,
			ItemMaxStack:  maxStack,
		},
		Slot:         slot,
		HPBonus:      hpBonus,
		ItemPriceBuy: pricebuy,
	}
}

var (
	StarterHat        = NewEquipment("Starter Hat", SlotHead, 2, 0, 1, Common, 1)
	StarterTunic      = NewEquipment("Starter Tunic", SlotChest, 5, 0, 1, Common, 1)
	StarterPants      = NewEquipment("Starter Pants", SlotLeg, 4, 0, 1, Common, 1)
	StarterBoots      = NewEquipment("Starter Boots", SlotFeet, 2, 0, 1, Common, 1)
	LeatherCap        = NewEquipment("Leather Cap", SlotHead, 4, 4, 8, Uncommon, 1)
	LeatherTunic      = NewEquipment("Leather Tunic", SlotChest, 10, 8, 16, Uncommon, 1)
	LeatherPants      = NewEquipment("Leather Pants", SlotLeg, 8, 7, 14, Uncommon, 1)
	LeatherBoots      = NewEquipment("Leather Boots", SlotFeet, 4, 4, 8, Uncommon, 1)
	OrcHelmet         = NewEquipment("Orc Helmet", SlotHead, 8, 10, 20, Rare, 1)
	OrcChestplate     = NewEquipment("Orc Chestplate", SlotChest, 18, 20, 40, Rare, 1)
	OrcGreaves        = NewEquipment("Orc Greaves", SlotLeg, 14, 16, 32, Rare, 1)
	OrcBoots          = NewEquipment("Orc Boots", SlotFeet, 8, 10, 20, Rare, 1)
	HarpyCowl         = NewEquipment("Harpy Cowl", SlotHead, 6, 13, 25, Rare, 1)
	HarpyRobe         = NewEquipment("Harpy Robe", SlotChest, 14, 25, 50, Rare, 1)
	HarpyBreeches     = NewEquipment("Harpy Breeches", SlotLeg, 10, 19, 38, Rare, 1)
	HarpySandals      = NewEquipment("Harpy Sandals", SlotFeet, 6, 13, 25, Rare, 1)
	MinotaurHelm      = NewEquipment("Minotaur Helm", SlotHead, 16, 30, 60, Epic, 1)
	MinotaurCuirass   = NewEquipment("Minotaur Cuirass", SlotChest, 32, 60, 120, Epic, 1)
	MinotaurLeggings  = NewEquipment("Minotaur Leggings", SlotLeg, 24, 45, 90, Epic, 1)
	MinotaurStompers  = NewEquipment("Minotaur Stompers", SlotFeet, 16, 30, 60, Epic, 1)
	WyrmScaleCrown    = NewEquipment("Wyrm Scale Crown", SlotHead, 14, 35, 70, Epic, 1)
	WyrmScaleHauberk  = NewEquipment("Wyrm Scale Hauberk", SlotChest, 28, 70, 140, Epic, 1)
	WyrmScaleGreaves  = NewEquipment("Wyrm Scale Greaves", SlotLeg, 22, 50, 100, Epic, 1)
	WyrmScaleBoots    = NewEquipment("Wyrm Scale Boots", SlotFeet, 14, 35, 70, Epic, 1)
	WyvernHelm        = NewEquipment("Wyvern Helm", SlotHead, 25, 130, 255, Legendary, 1)
	WyvernBreastplate = NewEquipment("Wyvern Breastplate", SlotChest, 50, 130, 255, Legendary, 1)
	WyvernGreaves     = NewEquipment("Wyvern Greaves", SlotLeg, 40, 130, 255, Legendary, 1)
	WyvernSabatons    = NewEquipment("Wyvern Sabatons", SlotFeet, 25, 130, 255, Legendary, 1)
)
