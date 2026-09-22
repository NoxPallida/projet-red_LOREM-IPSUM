package character

import (
	"runa/internal/inventory"
	"runa/internal/item"
)

// Ce fichier adapte l'API réelle d'inventory.Inventory (quantités en int,
// erreurs explicites) à la forme attendue par shop.Customer / forge.Crafter
// (bool/pas de retour), sans modifier inventory.go qui reste autonome
// et testable indépendamment de Character.

// CanFitItem vérifie SANS ajouter : soit un slot existant a de la place
// pour cet item, soit un nouveau slot est disponible. Nécessaire pour
func (c *Character) ensureInventory() {
	if c.Inventory.Capacity == 0 {
		c.Inventory = inventory.NewInventory()
	}
}

// CanFitItem vérifie SANS ajouter : soit un slot existant a de la place
// pour cet item, soit un nouveau slot est disponible. Nécessaire pour
// respecter l'ordre strict de shop.Buy/forge.Craft (vérifier avant d'agir).
func (c *Character) CanFitItem(i item.Item) bool {
	c.ensureInventory()
	maxStack := i.MaxStack()
	for _, slot := range c.Inventory.Slots {
		if slot.Item.Name() == i.Name() && slot.Quantity < maxStack {
			return true
		}
	}
	return uint8(len(c.Inventory.Slots)) < c.Inventory.Capacity
}

// AddItem ajoute une unité de l'item. L'erreur est ignorée volontairement :
// shop.Buy et forge.Craft appellent TOUJOURS CanFitItem juste avant,
// donc un échec ici ne devrait jamais arriver en pratique.
func (c *Character) AddItem(i item.Item) {
	c.ensureInventory()
	_ = c.Inventory.AddItem(i, 1)
}

// RemoveItem retire une unité de l'item, par nom. Utilisé par shop.Sell.
func (c *Character) RemoveItem(i item.Item) bool {
	return c.Inventory.RemoveItem(i.Name(), 1) == nil
}

// CountItem renvoie combien d'exemplaires de cet item le personnage possède.
// Nécessaire pour forge.Craft (recettes multi-matériaux).
func (c *Character) CountItem(i item.Item) uint8 {
	var count uint8
	for _, slot := range c.Inventory.Slots {
		if slot.Item.Name() == i.Name() {
			count += slot.Quantity
		}
	}
	return count
}

// RemoveItemQty retire une quantité précise. Utilisé par forge.Craft
// pour consommer les matériaux d'une recette.
func (c *Character) RemoveItemQty(i item.Item, qty uint8) bool {
	return c.Inventory.RemoveItem(i.Name(), qty) == nil
}

// CanUpgradeInventory / UpgradeInventory : requis par shop.Customer.
func (c *Character) CanUpgradeInventory() bool {
	return c.Inventory.Upgrades < inventory.MaxUpgrades
}

func (c *Character) UpgradeInventory() {
	_ = c.Inventory.UpgradeCapacity()
}
