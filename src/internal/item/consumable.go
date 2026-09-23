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
	HealPotion         = NewConsumable("Heal Potion", KindPotion, 0, 3, Common, 64)
	SmallHealPotion    = NewConsumable("Small Heal Potion", KindPotion, 0, 1, Common, 64)
	LargeHealPotion    = NewConsumable("Large Heal Potion", KindPotion, 0, 5, Common, 64)
	TitanicHealPotion  = NewConsumable("Titanic Heal Potion", KindPotion, 0, 8, Common, 64)
	PoisonPotion       = NewConsumable("Poison Potion", KindPotion, 0, 6, Common, 64)
	FireballBook       = NewConsumable("Fireball Book", KindSpellBook, 0, 25, Rare, 1)
	HealBook           = NewConsumable("Heal Book", KindSpellBook, 0, 20, Uncommon, 1)
	ShadowBoltBook     = NewConsumable("Shadow Bolt Book", KindSpellBook, 0, 30, Uncommon, 1)
	PoisonDartBook     = NewConsumable("Poison Dart Book", KindSpellBook, 0, 18, Common, 1)
	ChainLightningBook = NewConsumable("Chain Lightning Book", KindSpellBook, 0, 55, Rare, 1)
	IceBarrierBook     = NewConsumable("Ice Barrier Book", KindSpellBook, 0, 25, Uncommon, 1)
	DivineSmiteBook    = NewConsumable("Divine Smite Book", KindSpellBook, 0, 65, Epic, 1)
	LifeDrainBook      = NewConsumable("Life Drain Book", KindSpellBook, 0, 28, Uncommon, 1)
	EarthquakeBook     = NewConsumable("Earthquake Book", KindSpellBook, 0, 80, Epic, 1)
)
