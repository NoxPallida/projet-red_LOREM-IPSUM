package menu

import (
	"testing"
	"time"

	"runa/internal/tui"
)

func TestHeldVector(t *testing.T) {
	now := time.Now()
	held := func(evs ...tui.Event) map[tui.Event]time.Time {
		m := map[tui.Event]time.Time{}
		for _, ev := range evs {
			m[ev] = now
		}
		return m
	}
	cases := []struct {
		name   string
		held   map[tui.Event]time.Time
		dx, dy int
	}{
		{"vide", held(), 0, 0},
		{"z seul", held(tui.RuneEvent('z')), 0, -1},
		{"z+q tenus", held(tui.RuneEvent('z'), tui.RuneEvent('q')), -1, -1},
		{"q+d s'annulent", held(tui.RuneEvent('q'), tui.RuneEvent('d')), 0, 0},
		{"a = diagonale", held(tui.RuneEvent('a')), -1, -1},
		{"fleche haut", held(tui.Event{K: tui.KeyUp}), 0, -1},
	}
	for _, tc := range cases {
		if dx, dy := heldVector(tc.held); dx != tc.dx || dy != tc.dy {
			t.Errorf("%s = %d,%d, want %d,%d", tc.name, dx, dy, tc.dx, tc.dy)
		}
	}
}

func TestNormDir(t *testing.T) {
	key, dx, dy, ok := normDir(tui.RuneEvent('Z'))
	if !ok || dx != 0 || dy != -1 || key.R != 'z' {
		t.Errorf("normDir(Z) = %+v,%d,%d,%v", key, dx, dy, ok)
	}
	if _, _, _, ok := normDir(tui.RuneEvent(' ')); ok {
		t.Error("normDir(espace) devrait être rejeté")
	}
}
