package character

import (
	"testing"

	"runa/internal/inventory"
	"runa/internal/item"
	"runa/internal/shop"
)

func TestInitCharacterInventoryCapacity(t *testing.T) {
	ch := InitCharacter("Hero", Human)
	if ch.Inventory.Capacity != inventory.BaseCapacity {
		t.Fatalf("expected initial inventory capacity %d, got %d", inventory.BaseCapacity, ch.Inventory.Capacity)
	}
	if len(ch.Inventory.Slots) != 0 {
		t.Fatalf("expected empty inventory slots, got %d", len(ch.Inventory.Slots))
	}
	if !ch.CanFitItem(item.HealPotion) {
		t.Fatalf("expected CanFitItem to return true for new character")
	}

	res, name := shop.Buy(ch, item.HealPotion)
	if res != shop.Success {
		t.Fatalf("expected shop.Buy to succeed, got %v", res)
	}
	if name != item.HealPotion.Name() {
		t.Fatalf("expected item %q, got %q", item.HealPotion.Name(), name)
	}
	if len(ch.Inventory.Slots) != 1 {
		t.Fatalf("expected 1 item in inventory, got %d", len(ch.Inventory.Slots))
	}
}
