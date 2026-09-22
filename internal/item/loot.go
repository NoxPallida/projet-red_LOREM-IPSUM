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

// name pricesell rarity maxstack
var (
	WolfFur      = NewLoot("Wolf Fur", Common, 2, 64)
	WolfClaw     = NewLoot("Wolf Claw", Uncommon, 2, 50)
	TrollHide    = NewLoot("Troll Hide", Rare, 7, 20)
	BoarLeather  = NewLoot("Boar Leather", Uncommon, 3, 50)
	BoarTusk     = NewLoot("Boar Tusk", Rare, 10, 20)
	RatHide      = NewLoot("Rat Hide", Common, 1, 64)
	GoblinSkin   = NewLoot("Goblin Skin", Common, 2, 64)
	SlimeMucus   = NewLoot("Slime Mucus", Common, 3, 64)
	KoboldFang   = NewLoot("Kobold Fang", Common, 2, 64)
	OrcTusk      = NewLoot("Orc Tusk", Uncommon, 5, 50)
	OrcHide      = NewLoot("Orc Hide", Uncommon, 4, 50)
	BoneShard    = NewLoot("Bone Shard", Rare, 2, 20)
	HarpyFeather = NewLoot("Harpy Feather", Rare, 4, 20)
	OgreClub     = NewLoot("Ogre Club Fragment", Epic, 12, 10)
	MinotaurHorn = NewLoot("Minotaur Horn", Epic, 18, 10)
	WyrmScale    = NewLoot("Wyrm Scale", Epic, 30, 10)
	WyvernScale  = NewLoot("Wyvern Scale", Legendary, 100, 5)
	WyvernFang   = NewLoot("Wyvern Fang", Legendary, 120, 5)
)
