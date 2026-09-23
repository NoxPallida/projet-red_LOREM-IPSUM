package inventory

import (
	"errors"
	"runa/src/internal/item"
)

const (
	BaseCapacityI = 10
	BaseCapacityC = 100
	UpgradeBonus  = 10
	MaxUpgrades   = 3
)

type Slot struct {
	Item     item.Item
	Quantity uint8
}

type Inventory struct {
	Slots    []Slot
	Capacity uint8
	Upgrades uint8
}

// Chest est le coffre de stockage d'une maison : exactement la meme
// logique que Inventory (slots, piles, capacite), sans la dupliquer.
// Comme Class = spell.Class, c'est un alias, pas un nouveau type :
// toutes les methodes d'Inventory (AddItem, RemoveItem...) marchent
// direct dessus. Seule difference d'usage : Upgrades n'est jamais
// utilise pour un coffre (capacite fixe).
type Chest = Inventory

func NewChest() Chest {
	return Chest{
		Slots:    make([]Slot, 0, BaseCapacityC),
		Capacity: BaseCapacityC,
		Upgrades: 0,
	}
}

func NewInventory() Inventory {
	return Inventory{
		Slots:    make([]Slot, 0, BaseCapacityI),
		Capacity: BaseCapacityI,
		Upgrades: 0,
	}
}

func (inv *Inventory) UpgradeCapacity() error {
	if inv.Upgrades >= MaxUpgrades {
		return errors.New("Max inventory upgrades reached (3/3) !")
	}
	inv.Capacity += UpgradeBonus
	inv.Upgrades += 1
	return nil
}

func (inv *Inventory) AddItem(itm item.Item, quantity int) error {
	if quantity <= 0 {
		return errors.New("Invalid quantity")
	}

	maxStack := uint8(itm.MaxStack())
	remaining := uint8(quantity)

	// 1. Remplir les slots existants contenant déjà cet objet
	for i := range inv.Slots {
		if inv.Slots[i].Item.Name() == itm.Name() && inv.Slots[i].Quantity < maxStack {
			spaceInSlot := maxStack - inv.Slots[i].Quantity
			if remaining <= spaceInSlot {
				inv.Slots[i].Quantity += remaining
				return nil
			}
			inv.Slots[i].Quantity = maxStack
			remaining -= spaceInSlot
		}
	}

	// 2. Créer de nouveaux slots si nécessaire
	for remaining > 0 {
		if uint8(len(inv.Slots)) >= inv.Capacity {
			return errors.New("Inventory full ! Impossible to add the item !")
		}

		toAdd := remaining
		if toAdd > maxStack {
			toAdd = maxStack
		}

		inv.Slots = append(inv.Slots, Slot{
			Item:     itm,
			Quantity: toAdd,
		})

		remaining -= toAdd
	}

	return nil
}

func (inv *Inventory) RemoveItem(itemName string, quantity uint8) error {
	if !inv.HasItem(itemName, quantity) {
		return errors.New("Invalid quantity")
	}

	remaining := quantity
	for i := len(inv.Slots) - 1; i >= 0; i-- {
		if inv.Slots[i].Item.Name() == itemName {
			if inv.Slots[i].Quantity > remaining {
				inv.Slots[i].Quantity -= remaining
				return nil
			}

			remaining -= inv.Slots[i].Quantity
			inv.Slots = append(inv.Slots[:i], inv.Slots[i+1:]...)

			if remaining == 0 {
				return nil
			}
		}
	}

	return nil
}

func (inv *Inventory) HasItem(itemName string, quantity uint8) bool {
	var count uint8 = 0
	for _, slot := range inv.Slots {
		if slot.Item.Name() == itemName {
			count += slot.Quantity
		}
	}
	return count >= quantity
}

func (inv *Inventory) IsFull() bool {
	return uint8(len(inv.Slots)) >= inv.Capacity
}
