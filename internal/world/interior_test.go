package world

import (
	"testing"
)

func TestInteriorDimensionsAndNPC(t *testing.T) {
	in := BuildInterior(ZoneShop)
	if in.W != 27 || in.H != 15 {
		t.Errorf("Dimensions intérieur = %dx%d, want 27x15", in.W, in.H)
	}
	if in.NPC.Name != "Marchand" {
		t.Errorf("NPC Shop attendu 'Marchand', reçu %q", in.NPC.Name)
	}

	inGuild := BuildInterior(ZoneGuild)
	if inGuild.NPC.Name != "Maire de guilde" {
		t.Errorf("NPC Guild attendu 'Maire de guilde', reçu %q", inGuild.NPC.Name)
	}

	inForge := BuildInterior(ZoneForge)
	if inForge.NPC.Name != "Forgeron" {
		t.Errorf("NPC Forge attendu 'Forgeron', reçu %q", inForge.NPC.Name)
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
