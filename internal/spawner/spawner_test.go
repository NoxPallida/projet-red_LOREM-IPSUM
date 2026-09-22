package spawner

import (
	"testing"

	"runa/internal/guild"
	"runa/internal/world"
)

func TestTrainingGroundGoblin(t *testing.T) {
	// Création d'une grille 200x200 simulant map.txt
	tiles := make([][]world.Tile, 200)
	for y := range tiles {
		tiles[y] = make([]world.Tile, 200)
		for x := range tiles[y] {
			if x >= 43 && x <= 51 && y >= 173 && y <= 181 {
				tiles[y][x] = world.Tile{Kind: world.TileSand, Walkable: true}
			} else {
				tiles[y][x] = world.Tile{Kind: world.TileGrass, Walkable: true}
			}
		}
	}

	sp := NewSpawner(tiles, guild.RankF, 42)

	// Vérifie la présence du gobelin au terrain d'entraînement
	mob, found := sp.EnemyAt(47, 177)
	if !found {
		t.Fatalf("Gobelin non trouvé à (47, 177)")
	}
	if mob.Template.ID != "goblin" {
		t.Errorf("Mob attendu 'goblin', reçu %q", mob.Template.ID)
	}

	// Tuer le gobelin doit le faire réapparaître au terrain d'entraînement
	sp.Kill(47, 177, guild.RankF)
	respawned, foundAfter := sp.EnemyAt(47, 177)
	if !foundAfter {
		t.Fatalf("Gobelin devrait réapparaître à (47, 177)")
	}
	if respawned.Template.ID != "goblin" {
		t.Errorf("Respawn attendu 'goblin', reçu %q", respawned.Template.ID)
	}
}
