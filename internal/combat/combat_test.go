package combat

import (
	"testing"

	"runa/internal/character"
	"runa/internal/enemies"
	"runa/internal/spell"
)

func TestClassBaseAttackOptions(t *testing.T) {
	// Test Mage base attack
	mage := character.InitCharacter("Gandalf", spell.Human)
	mage.SetSubclass(spell.SubclassHumanMage)

	opts := AttackOptions(mage)
	if len(opts) == 0 {
		t.Fatalf("expected at least 1 attack option")
	}
	if opts[0].Name != "Arcane Spark" {
		t.Fatalf("expected base attack to be 'Arcane Spark', got %q", opts[0].Name)
	}
	if opts[0].ManaCost != 0 {
		t.Fatalf("expected base attack mana cost to be 0, got %d", opts[0].ManaCost)
	}

	// Test Archer base attack
	archer := character.InitCharacter("Legolas", spell.Elf)
	archer.SetSubclass(spell.SubclassElfArcher)

	optsArcher := AttackOptions(archer)
	if optsArcher[0].Name != "Quick Shot" {
		t.Fatalf("expected base attack to be 'Quick Shot', got %q", optsArcher[0].Name)
	}

	// Test Berserker base attack
	dwarf := character.InitCharacter("Gimli", spell.Dwarf)
	dwarf.SetSubclass(spell.SubclassDwarfBerserker)

	optsDwarf := AttackOptions(dwarf)
	if optsDwarf[0].Name != "Furious Strike" {
		t.Fatalf("expected base attack to be 'Furious Strike', got %q", optsDwarf[0].Name)
	}
}

func TestLevelUpUnlocksSpells(t *testing.T) {
	mage := character.InitCharacter("Gandalf", spell.Human)
	mage.SetSubclass(spell.SubclassHumanMage)

	// Initially level 1: only base attack
	opts1 := AttackOptions(mage)
	if len(opts1) != 1 {
		t.Fatalf("expected 1 attack at level 1, got %d", len(opts1))
	}

	// Add enough XP to level up to 3
	for mage.Level() < 3 {
		mage.AddXP(50)
	}

	opts3 := AttackOptions(mage)
	if len(opts3) < 2 {
		t.Fatalf("expected at least 2 attacks at level 3, got %d", len(opts3))
	}

	foundArcaneBolt := false
	for _, opt := range opts3 {
		if opt.Name == "Arcane Bolt" {
			foundArcaneBolt = true
		}
	}
	if !foundArcaneBolt {
		t.Fatalf("expected Arcane Bolt to be unlocked at level 3")
	}
}

func TestCombatDefeatAndDrops(t *testing.T) {
	hero := character.InitCharacter("Hero", spell.Human)
	hero.Subclass = spell.SubclassHumanCQC
	hero.LearnAvailableSpells()

	rat, ok := enemies.NewEnemyInstance("rat", 1)
	if !ok {
		t.Fatalf("failed to create rat instance")
	}

	cb := NewCombat(hero, rat)
	// Kill rat directly
	rat.TakeDamage(100)
	cb.checkEnemyDefeated()

	if !cb.Over || !cb.PlayerWon {
		t.Fatalf("expected combat to be won")
	}
}
