package world

import (
	"testing"
)

func TestInteriorDimensionsAndNPC(t *testing.T) {
	in := BuildInterior(ZoneShop)
	if in.W != 27 || in.H != 15 {
		t.Errorf("Dimensions intérieur = %dx%d, want 27x15", in.W, in.H)
	}
	if in.NPC.Name != "Merchant" {
		t.Errorf("NPC Shop expected 'Merchant', got %q", in.NPC.Name)
	}

	inGuild := BuildInterior(ZoneGuild)
	if inGuild.NPC.Name != "Guild Master" {
		t.Errorf("NPC Guild expected 'Guild Master', got %q", inGuild.NPC.Name)
	}

	inForge := BuildInterior(ZoneForge)
	if inForge.NPC.Name != "Blacksmith" {
		t.Errorf("NPC Forge expected 'Blacksmith', got %q", inForge.NPC.Name)
	}

	// Vérifier que les zones du village sont distinctes (1 marchand, 1 maire, 1 forgeron)
	// et que le terrain d'entraînement (42..52) n'a plus de zone
	if z, ok := ZoneAt(45, 175); ok && z.Kind == ZoneShop {
		t.Errorf("Le terrain d'entraînement ne devrait plus être un Shop")
	}

	zShop, okShop := ZoneAt(60, 175)
	if !okShop || zShop.Kind != ZoneShop {
		t.Errorf("Bâtiment haut-gauche devrait être ZoneShop")
	}

	zGuild, okGuild := ZoneAt(75, 175)
	if !okGuild || zGuild.Kind != ZoneGuild {
		t.Errorf("Bâtiment haut-droite devrait être ZoneGuild")
	}

	zForge, okForge := ZoneAt(60, 180)
	if !okForge || zForge.Kind != ZoneForge {
		t.Errorf("Bâtiment bas-gauche devrait être ZoneForge")
	}
}
