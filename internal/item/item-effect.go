package item

// EffectType liste les effets disponibles. Facile à étendre plus tard
// (burn, stun, buff...).
type EffectType int

const (
	EffectPoison EffectType = iota
	EffectBleed
	EffectHeal
)

func (e EffectType) String() string {
	switch e {
	case EffectPoison:
		return "Poison"
	case EffectBleed:
		return "Bleed"
	case EffectHeal:
		return "Heal"
	default:
		return "None"
	}
}

// Effect décrit un effet de statut générique : combien de dégâts/soin
type Effect struct {
	Type     EffectType
	Amount   uint8
	Duration uint8 // nombre de tours
}

// effectRegistry associe le nom d'un item à ses effets.
// Volontairement découplé de Weapon/Equipment/Loot/Consumable :
// n'importe quel type qui implémente Item peut recevoir des effets
// sans avoir à modifier sa struct.
var effectRegistry = make(map[string][]Effect)

// AttachEffect lie un ou plusieurs effets à n'importe quel item.
func AttachEffect(i Item, effects ...Effect) {
	effectRegistry[i.Name()] = append(effectRegistry[i.Name()], effects...)
}

// EffectsOf renvoie les effets attachés à un item (vide si aucun).
func EffectsOf(i Item) []Effect {
	return effectRegistry[i.Name()]
}

// HasEffect indique si un item porte un type d'effet donné.
func HasEffect(i Item, t EffectType) bool {
	for _, e := range EffectsOf(i) {
		if e.Type == t {
			return true
		}
	}
	return false
}

func init() {
	AttachEffect(PoisonPotion, Effect{Type: EffectPoison, Amount: 10, Duration: 3})
	AttachEffect(HealPotion, Effect{Type: EffectHeal, Amount: 50, Duration: 1})
	AttachEffect(WarAxe, Effect{Type: EffectBleed, Amount: 5, Duration: 4})
}
