package enemies

import (
	"testing"
)

func TestEnemyTemplatesAndDrops(t *testing.T) {
	enemyIDs := []string{"rat", "wolf", "boar", "troll", "goblin"}

	for _, id := range enemyIDs {
		tmpl, ok := GetTemplate(id)
		if !ok {
			t.Fatalf("enemy template %q not found", id)
		}
		if len(tmpl.Drops) == 0 {
			t.Fatalf("enemy template %q has no item drops configured", id)
		}

		inst, ok := NewEnemyInstance(id, 1)
		if !ok {
			t.Fatalf("failed to create instance for %q", id)
		}
		_ = inst.RollDrops()
	}
}
