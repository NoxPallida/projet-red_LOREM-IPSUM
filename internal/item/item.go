package item

type ItemType int

const (
	TypeWeapon ItemType = iota
	TypeEquipment
	TypeConsumable
	TypeLoot
)

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
	PriceSell() int
	PriceBuy() int
}

// Base regroupe les champs communs à tous les items.
type Base struct {
	ItemName      string
	ItemPriceSell int
}

func (b Base) Name() string   { return b.ItemName }
func (b Base) PriceSell() int { return b.ItemPriceSell }
