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
			if dir, ok := directionFromKey(ev.K); ok {
				if w.TryMove(p, dir) {
					w.EnsureLoaded(p.X, p.Y, 2) // rayon de 2 chunks autour du joueur
				}
			}
			if ev.K == tui.KeyQuit {
				return true
			}
		}

		layers := []tui.FuncLayer{worldMapLayer(w, p)}
		c.DrawFuncLayers(layers)
		c.DrawLayer(hudLayer(ch))

		return false
	})
}

// directionFromKey traduit une touche fléchée en Direction.
// Les 4 directions façon vieux Pokémon : haut/bas/gauche/droite uniquement.
func directionFromKey(k tui.Key) (world.Direction, bool) {
	switch k {
	case tui.KeyUp:
		return world.North, true
	case tui.KeyDown:
		return world.South, true
	case tui.KeyLeft:
		return world.West, true
	case tui.KeyRight:
		return world.East, true
	default:
		return 0, false
	}
}
