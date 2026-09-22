package spell

import (
	"testing"
)

func TestAllClassesAndSubclasses(t *testing.T) {
	if len(AllClasses) != 3 {
		t.Fatalf("expected 3 races, got %d", len(AllClasses))
	}

	expectedRaces := map[Class]string{
		Human: "Humain",
		Elf:   "Elfe",
		Dwarf: "Nain",
	}

	for _, c := range AllClasses {
		expectedName, ok := expectedRaces[c]
		if !ok {
			t.Fatalf("unexpected race %v", c)
		}
		if c.String() != expectedName {
			t.Fatalf("race %v String() = %q, want %q", c, c.String(), expectedName)
		}

		subs := SubclassesForClass(c)
		if len(subs) < 2 {
			t.Fatalf("expected at least 2 subclasses for race %v, got %d", c, len(subs))
		}
		for _, s := range subs {
			if s.String() == "" || s.String() == "Aventurier" {
				t.Fatalf("subclass %v has invalid string %q", s, s.String())
			}
		}
	}
}
