package menu

import (
	"os"

	"runa/internal/character"
	"runa/internal/tui"
	"runa/internal/world"
)

// StartGame construit le monde à partir de la carte de la ville,
// place le joueur à son point de spawn, puis lance la boucle de jeu.
// C'est la seule fonction que main.go a besoin d'appeler.
func StartGame(ch *character.Character) error {
	loader, pm := world.LoadTownMap()
	w := world.NewWorld(loader)
	p := &world.Player{X: pm.StartX, Y: pm.StartY}
	w.EnsureLoaded(p.X, p.Y, 2) // charge les chunks autour du spawn avant le 1er rendu

	return RunGame(ch, w, p)
}

func RunGame(ch *character.Character, w *world.World, p *world.Player) error {
	return tui.RunLoop(os.Stdin, os.Stdout, 0, 0, func(c *tui.Canvas, ev *tui.Event) bool {
		if ev != nil {
			if dir, ok := directionFromEvent(*ev); ok {
				if world.MovePlayer(w, p, dir) {
					w.EnsureLoaded(p.X, p.Y, 2) // rayon de 2 chunks autour du joueur
				}
			}
			if ev.K == tui.KeyEsc {
				return true
			}
		}

		layers := []tui.FuncLayer{worldMapLayer(w, p)}
		c.DrawFuncLayers(layers)
		c.DrawLayer(hudLayer(ch))

		return false
	})
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
