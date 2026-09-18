package item

type Loot struct {
	Base
}

func (l Loot) Type() ItemType { return TypeLoot }

func NewLoot(name string, rarity Rarity, pricesell, maxStack uint8) Loot {
	return Loot{
		Base: Base{
			ItemName:      name,
			ItemPriceSell: pricesell,
			ItemRarity:    rarity,
			ItemMaxStack:  maxStack,
		},
	}
}

var (
	WolfFur      = NewLoot("Wolf Fur", Common, 2, 64)
	TrollHide    = NewLoot("Troll Hide", Rare, 7, 20)
	BoarLeather  = NewLoot("Boar Leather", Uncommon, 3, 50)
	RavenFeather = NewLoot("Raven Feather", Common, 1, 64)
)
