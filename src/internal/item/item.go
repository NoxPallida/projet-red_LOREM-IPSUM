package item

type ItemType int
type Rarity int

const (
	TypeWeapon ItemType = iota
	TypeEquipment
	TypeConsumable
	TypeLoot
)

const (
	Common Rarity = iota
	Uncommon
	Rare
	Epic
	Legendary
)

func (r Rarity) String() string {
	switch r {
	case Common:
		return "Common"
	case Uncommon:
		return "Uncommon"
	case Rare:
		return "Rare"
	case Epic:
		return "Epic"
	case Legendary:
		return "Legendary"
	default:
		return "???"
	}
}

func (t ItemType) String() string {
	switch t {
	case TypeWeapon:
		return "Weapon"
	case TypeEquipment:
		return "Equipment"
	case TypeConsumable:
		return "Consumable"
	case TypeLoot:
		return "Loot"
	default:
		return "Inconnu"
	}
}

type Item interface {
	Name() string
	Type() ItemType
	PriceSell() uint8
	Rarity() Rarity
	MaxStack() uint8
}

// Base regroupe les champs communs à tous les items.
type Base struct {
	ItemName      string
	ItemPriceSell uint8
	ItemRarity    Rarity
	ItemMaxStack  uint8
}

func (b Base) Name() string     { return b.ItemName }
func (b Base) PriceSell() uint8 { return b.ItemPriceSell }
func (b Base) Rarity() Rarity   { return b.ItemRarity }
func (b Base) MaxStack() uint8  { return b.ItemMaxStack }
