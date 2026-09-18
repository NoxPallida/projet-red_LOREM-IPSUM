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
	AdventurerHat   = NewEquipment("Adventurer's Hat", SlotHead, 10, 5, 10, Common, 1)
	AdventurerTunic = NewEquipment("Adventurer's Tunic", SlotChest, 25, 5, 10, Common, 1)
	AdventurerPants = NewEquipment("Adventurer's Pants", SlotLeg, 20, 5, 10, Common, 1)
	AdventurerBoots = NewEquipment("Adventurer's Boots", SlotFeet, 15, 5, 10, Common, 1)
)
