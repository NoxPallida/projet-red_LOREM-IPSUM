package item

type ConsumableKind int

const (
	KindPotion ConsumableKind = iota
	KindFood
	KindSpellBook
)

type Consumable struct {
	Base
	Kind         ConsumableKind
	ItemPriceBuy uint8
}

func (c Consumable) Type() ItemType { return TypeConsumable }

func (c Consumable) PriceBuy() uint8 { return c.ItemPriceBuy }

func NewConsumable(name string, kind ConsumableKind, pricesell, pricebuy uint8, rarity Rarity, maxStack uint8) Consumable {
	return Consumable{
		Base: Base{
			ItemName:      name,
			ItemPriceSell: pricesell,
			ItemRarity:    rarity,
			ItemMaxStack:  maxStack,
		},
		Kind:         kind,
		ItemPriceBuy: pricebuy,
	}
}

var (
	HealPotion   = NewConsumable("Heal Potion", KindPotion, 0, 3, Common, 64)
	PoisonPotion = NewConsumable("Poison Potion", KindPotion, 0, 6, Common, 64)
	FireballBook = NewConsumable("Fireball Book", KindSpellBook, 0, 25, Rare, 1)
)
