package menu

import (
	"os"
	"time"

	"runa/internal/character"
	"runa/internal/tui"
	"runa/internal/world"
)

// moveTick : pas + rendu toutes les 100ms.
// Touches maintenues (répétition terminal) = mouvement continu,
// plusieurs touches dans la même fenêtre = diagonale.
const moveTick = 100 * time.Millisecond

// StartGame construit le monde à partir de la grande carte (map.txt),
// place le joueur à son point de spawn, puis lance la boucle de jeu.
// C'est la seule fonction que main.go a besoin d'appeler.
func StartGame(ch *character.Character) error {
	loader, pm := world.LoadWorldMap()
	w := world.NewWorld(loader)
	p := &world.Player{X: pm.StartX, Y: pm.StartY}
	w.EnsureLoaded(p.X, p.Y, 2) // charge les chunks autour du spawn avant le 1er rendu

	return RunGame(ch, w, p)
}

func RunGame(ch *character.Character, w *world.World, p *world.Player) error {
	in := os.Stdin
	out := os.Stdout

	sw, sh, err := tui.Size()
	if err != nil {
		return err
	}
	c := tui.NewCanvas(sw, sh)

	if err := tui.EnterAltScreen(out); err != nil {
		return err
	}
	defer func() {
		_ = tui.ExitAltScreen(out)
	}()

	// Pompe à touches (ReadKey bloque) -> canal, comme Dialogue.
	keys := make(chan tui.Event, 64)
	done := make(chan struct{})
	defer close(done)
	go func() {
		defer close(keys)
		for {
			ev, err := tui.ReadKey(in)
			if err != nil {
				select {
				case keys <- tui.Event{K: tui.KeyEsc}:
				case <-done:
				}
				return
			}
			select {
			case keys <- ev:
			case <-done:
				return
			}
		}
	}()

	render := func() {
		c.Clear()
		layers := []tui.FuncLayer{worldMapLayer(w, p)}
		c.DrawFuncLayers(layers)
		c.DrawLayer(hudLayer(c, ch))
		_ = tui.FlushStyled(out, c)
	}
	render()

	ticker := time.NewTicker(moveTick)
	defer ticker.Stop()
	for range ticker.C {
		quit := false
		dx, dy := 0, 0
	drain:
		for {
			select {
			case ev, ok := <-keys:
				if !ok {
					return nil
				}
				if ev.K == tui.KeyEsc {
					quit = true
					break drain
				}
				if ddx, ddy, ok := dirDelta(ev); ok {
					dx += ddx
					dy += ddy
				}
			default:
				break drain
			}
		}
		if quit {
			return nil
		}
		// Diagonale : on applique x puis y, chacun teste sa case.
		// Touches opposées (q+d) s'annulent.
		moved := false
		if dx < 0 {
			moved = world.MovePlayer(w, p, world.West) || moved
		} else if dx > 0 {
			moved = world.MovePlayer(w, p, world.East) || moved
		}
		if dy < 0 {
			moved = world.MovePlayer(w, p, world.North) || moved
		} else if dy > 0 {
			moved = world.MovePlayer(w, p, world.South) || moved
		}
		if moved {
			w.EnsureLoaded(p.X, p.Y, 2) // rayon de 2 chunks autour du joueur
		}
		render()
	}
	return nil
}

// dirDelta traduit une touche en déplacement (dx, dy).
// Flèches + ZQSD, comme directionFromEvent.
func dirDelta(ev tui.Event) (int, int, bool) {
	dir, ok := directionFromEvent(ev)
	if !ok {
		return 0, 0, false
	}
	switch dir {
	case world.North:
		return 0, -1, true
	case world.South:
		return 0, 1, true
	case world.East:
		return 1, 0, true
	case world.West:
		return -1, 0, true
	}
	return 0, 0, false
}

// directionFromEvent traduit une touche en Direction.
// Fleches + ZQSD (majuscules acceptees), facon vieux Pokemon :
// haut/bas/gauche/droite uniquement. Q = ouest, PAS quitter.
func directionFromEvent(ev tui.Event) (world.Direction, bool) {
	switch ev.K {
	case tui.KeyUp:
		return world.North, true
	case tui.KeyDown:
		return world.South, true
	case tui.KeyLeft:
		return world.West, true
	case tui.KeyRight:
		return world.East, true
	}
	if ev.K == tui.KeyRune {
		switch ev.R {
		case 'z', 'Z':
			return world.North, true
		case 's', 'S':
			return world.South, true
		case 'q', 'Q':
			return world.West, true
		case 'd', 'D':
			return world.East, true
		}
	}
	return 0, false
}
