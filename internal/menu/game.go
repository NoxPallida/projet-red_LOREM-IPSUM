package menu

import (
	"os"
	"time"

	"runa/internal/character"
	"runa/internal/tui"
	"runa/internal/world"
)

// Répétition pilotée par le jeu (mode kitty seulement) :
// pas d'attente OS, les touches tenues avancent toutes seules.
const (
	keySleep    = 5 * time.Millisecond   // pause entre deux sondages
	repeatEvery = 80 * time.Millisecond  // ~12 pas/s touche tenue
	holdExpire  = 500 * time.Millisecond // sécurité : touche silencieuse = relâchée
)

// StartGame construit le monde à partir de la grande carte (map.txt),
// place le joueur à son point de spawn, puis lance la boucle de jeu.
// C'est la seule fonction que main.go a besoin d'appeler.
func StartGame(ch *character.Character) error {
	loader, pm := world.LoadWorldMap()
	w := world.NewWorld(loader)
	p := &world.Player{X: pm.StartX, Y: pm.StartY}
	w.EnsureLoaded(p.X, p.Y, loadRadius()) // charge autour du spawn avant le 1er rendu

	return RunGame(ch, w, p)
}

// loadRadius couvre tout l'écran visible + marge, même en grand terminal.
// Taille lue une fois ici (StartGame) et une fois dans RunGame.
func loadRadius() int {
	sw, sh, err := tui.Size()
	if err != nil {
		return 6 // repli : couvre ~256 colonnes
	}
	return world.LoadRadius(sw/(2*tileScale), sh/(2*tileScale))
}

// RunGame : boucle simple, sans goroutine ni canal ni defer.
// On sonde les touches sans bloquer, 1 pas par appui, + répétition
// pilotée par le jeu quand le terminal dit qui est TENU (kitty).
// a/e/w/c = diagonales (AZERTY : a en haut à gauche de zq, e en haut
// à droite, w en bas à gauche, c en bas à droite).
// Sans kitty (Windows conhost, vieux terminaux, pipes) : 1 touche =
// 1 pas, exactement comme avant. Quitter : ECHAP (ou entrée fermée).
// Le nettoyage (kitty + écran) est fait à la main avant chaque sortie.
func RunGame(ch *character.Character, w *world.World, p *world.Player) error {
	in := os.Stdin
	out := os.Stdout

	sw, sh, err := tui.Size()
	if err != nil {
		return err
	}
	radius := world.LoadRadius(sw/(2*tileScale), sh/(2*tileScale))
	c := tui.NewCanvas(sw, sh)

	if err := tui.EnterAltScreen(out); err != nil {
		return err
	}

	// Kitty ? appui+relâchement suivis ; sinon classique.
	kitty := false
	if tui.IsTerminal(in) && tui.IsTerminal(out) {
		kitty = tui.QueryKitty(in, out)
		if kitty {
			tui.PushKitty(out)
		}
	}
	leave := func() {
		if kitty {
			tui.PopKitty(out)
		}
		_ = tui.ExitAltScreen(out)
	}

	// step avance de (dx, dy) : x puis y, chacun teste sa case.
	step := func(dx, dy int) {
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
			w.EnsureLoaded(p.X, p.Y, radius)
		}
	}
	render := func() {
		c.Clear()
		layers := []tui.FuncLayer{worldMapLayer(w, p)}
		c.DrawFuncLayers(layers)
		c.DrawLayer(hudLayer(c, ch))
		_ = tui.FlushStyled(out, c)
	}

	held := map[tui.Event]time.Time{} // touches tenues (kitty) -> dernier signe de vie
	lastStep := time.Now()
	render()
	for {
		// 1. Tout ce qui est arrivé, sans bloquer.
		pressed := false
		for {
			ev, rel, have, err := tui.PollKey(in)
			if err != nil {
				leave()
				return nil // entrée fermée : on quitte proprement
			}
			if !have {
				break
			}
			if ev.K == tui.KeyNone {
				continue // bruit avalé
			}
			if ev.K == tui.KeyEsc && !rel {
				leave()
				return nil
			}
			key, dx, dy, ok := normDir(ev)
			if !ok {
				continue // espace, entrée... : rien à faire en jeu
			}
			if !kitty {
				// Classique : 1 touche = 1 pas, de suite.
				step(dx, dy)
				lastStep = time.Now()
				pressed = true
				continue
			}
			if rel {
				delete(held, key)
			} else {
				if _, known := held[key]; !known {
					step(dx, dy) // appui : pas immédiat
					lastStep = time.Now()
				}
				held[key] = time.Now()
			}
			pressed = true
		}
		// 2. Touches oubliées (relâchement manqué) : poubelle.
		now := time.Now()
		for k, seen := range held {
			if now.Sub(seen) > holdExpire {
				delete(held, k)
			}
		}
		// 3. Répétition tenue : toutes les 80ms, pas d'attente OS.
		if kitty && len(held) > 0 && now.Sub(lastStep) >= repeatEvery {
			dx, dy := heldVector(held)
			step(dx, dy)
			lastStep = now
			pressed = true
		}
		if pressed {
			render()
		}
		time.Sleep(keySleep)
	}
}

// normDir stabilise la touche (majuscules -> minuscules pour le suivi)
// et rend son déplacement. a/e/w/c = diagonales AZERTY.
func normDir(ev tui.Event) (key tui.Event, dx, dy int, ok bool) {
	key = ev
	if ev.K == tui.KeyRune && ev.R >= 'A' && ev.R <= 'Z' {
		key.R += 'a' - 'A'
	}
	dx, dy, ok = dirDelta8(key)
	return key, dx, dy, ok
}

// heldVector combine les touches tenues : z+q tenus = diagonale,
// q+d tenus s'annulent. Testé en unitaire.
func heldVector(held map[tui.Event]time.Time) (int, int) {
	dx, dy := 0, 0
	for k := range held {
		ddx, ddy, ok := dirDelta8(k)
		if ok {
			dx += ddx
			dy += ddy
		}
	}
	return dx, dy
}

// dirDelta8 traduit une touche en déplacement (dx, dy).
// ZQSD + flèches + diagonales AZERTY : a = haut-gauche,
// e = haut-droite, w = bas-gauche, c = bas-droite.
func dirDelta8(ev tui.Event) (int, int, bool) {
	if ev.K == tui.KeyUp {
		return 0, -1, true
	}
	if ev.K == tui.KeyDown {
		return 0, 1, true
	}
	if ev.K == tui.KeyLeft {
		return -1, 0, true
	}
	if ev.K == tui.KeyRight {
		return 1, 0, true
	}
	if ev.K == tui.KeyRune {
		switch ev.R {
		case 'z', 'Z':
			return 0, -1, true
		case 's', 'S':
			return 0, 1, true
		case 'q', 'Q':
			return -1, 0, true
		case 'd', 'D':
			return 1, 0, true
		case 'a', 'A':
			return -1, -1, true
		case 'e', 'E':
			return 1, -1, true
		case 'w', 'W':
			return -1, 1, true
		case 'c', 'C':
			return 1, 1, true
		}
	}
	return 0, 0, false
}
