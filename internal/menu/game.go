package menu

import (
	"os"
	"strconv"
	"strings"
	"time"

	"runa/internal/character"
	"runa/internal/enemies"
	"runa/internal/guild"
	"runa/internal/spawner"
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
// place le joueur à son point de spawn, peuple la carte de monstres
// selon le rang de départ (F), puis lance la boucle de jeu.
// C'est la seule fonction que main.go a besoin d'appeler.
func StartGame(ch *character.Character) error {
	loader, pm := world.LoadWorldMap()
	w := world.NewWorld(loader)
	p := &world.Player{X: pm.StartX, Y: pm.StartY}
	w.EnsureLoaded(p.X, p.Y, loadRadius()) // charge autour du spawn avant le 1er rendu

	gs := guild.NewGuildStatus()
	sp := spawner.NewSpawner(pm.Tiles, gs.Rank, time.Now().UnixNano())

	return RunGame(ch, w, p, sp, gs, pm.StartX, pm.StartY)
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
func RunGame(ch *character.Character, w *world.World, p *world.Player, sp *spawner.Spawner, gs *guild.GuildStatus, spawnX, spawnY int) error {
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

	// pending : le monstre statique que le joueur vient de percuter,
	// en attente du système de combat (pas encore implémenté). Tant
	// qu'il est non-nil, le déplacement dans sa direction reste bloqué.
	// Remis à nil dès qu'aucun axe du pas courant ne bute plus dessus.
	var pending *enemies.EnemyInstance

	// inIn non-nil = le joueur est DANS un batiment : la carte monde
	// est remplacee par la piece, le reste (HUD, rencontres) pareil.
	var inIn *InteriorSession

	// render redessine TOUT l'écran : fond effacé, carte (ou piece),
	// HUD, et la rencontre en cours par-dessus si elle existe.
	// Déclaré ici (avant tryAxis) car tryAxis l'appelle après un combat.
	render := func() {
		c.Clear()
		if inIn != nil {
			drawInterior(c, inIn)
		} else {
			layers := []tui.FuncLayer{worldMapLayer(w, p, sp)}
			c.DrawFuncLayers(layers)
		}
		c.DrawLayer(displayInfo(c, ch, gs))
		if pending != nil {
			c.DrawLayer(encounterLayer(c, pending))
		}
		_ = tui.FlushStyled(out, c)
	}

	// tryAxis tente un déplacement sur UN axe (dir) : si un monstre
	// vivant occupe la case visée, le joueur se tourne vers lui sans
	// avancer et la rencontre reste en attente (TODO combat). Sinon,
	// le déplacement suit son cours normal via world.MovePlayer.
	tryAxis := func(dir world.Direction) bool {
		nx, ny := nextPos(p, dir)
		if enemy, found := sp.EnemyAt(nx, ny); found {
			p.Dir = dir
			cb := RunCombat(in, out, c, ch, enemy, kitty)

			if cb.PlayerWon {
				sp.Kill(nx, ny, gs.Rank)
				ch.GainExp(enemy.XPDrop)
				guild.RegisterKill(gs, enemy.Template.ID) // fait avancer les quêtes actives ; le rendu se fait à la guilde
				for _, it := range cb.DroppedItems {
					_ = ch.Inventory.AddItem(it, 1)
				}
				// Butin : on le dit au joueur au lieu de l'ajouter en
				// silence (le log du combat est efface par render()).
				if len(cb.DroppedItems) > 0 {
					if kitty {
						tui.PopKitty(out)
					}
					names := make([]string, 0, len(cb.DroppedItems))
					for _, it := range cb.DroppedItems {
						names = append(names, it.Name())
					}
					tui.Dialogue(c, out, in,
						"Victoire ! +"+strconv.Itoa(int(enemy.XPDrop))+" XP.",
						"Butin : "+strings.Join(names, ", ")+".")
					if kitty {
						tui.PushKitty(out)
					}
				}
			}
			// defaite : ecran de mort puis respawn au spawn avec 50 % des PV.
			if ch.Hp == 0 {
				showDeathScreen(in, out, c)
				respawn(p, spawnX, spawnY, ch)
				pending = nil
				w.EnsureLoaded(p.X, p.Y, radius)
			}

			render() // redessine la carte par-dessus l'écran de combat
			return false
		}
		return world.MovePlayer(w, p, dir)
	}

	// step avance de (dx, dy) : x puis y, chacun teste sa case.
	// Dans un batiment : meme principe sur la grille de la piece ;
	// arriver sur la porte interieure fait sortir (retour dehors).
	// Dehors : arriver sur une porte 'E' fait entrer dans le batiment.
	step := func(dx, dy int) {
		pending = nil // reset : seul un axe bloqué par un monstre le remet à jour ci-dessous
		if inIn != nil {
			if dx < 0 {
				inIn.move(world.West)
			} else if dx > 0 {
				inIn.move(world.East)
			}
			if dy < 0 {
				inIn.move(world.North)
			} else if dy > 0 {
				inIn.move(world.South)
			}
			if inIn.onDoor() {
				p.X, p.Y = inIn.EntryX, inIn.EntryY
				inIn = nil
				w.EnsureLoaded(p.X, p.Y, radius)
			}
			return
		}
		moved := false
		if dx < 0 {
			moved = tryAxis(world.West) || moved
		} else if dx > 0 {
			moved = tryAxis(world.East) || moved
		}
		if dy < 0 {
			moved = tryAxis(world.North) || moved
		} else if dy > 0 {
			moved = tryAxis(world.South) || moved
		}
		if moved {
			w.EnsureLoaded(p.X, p.Y, radius)
			if sess := enterInterior(w, p); sess != nil {
				inIn = sess
			}
		}
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
			// Touche 'e' ou 'E' : ouvrir l'inventaire
			if ev.K == tui.KeyRune && (ev.R == 'e' || ev.R == 'E') && !rel {
				if kitty {
					tui.PopKitty(out)
				}
				RunInventoryMenu(in, out, c, ch, render)
				if kitty {
					tui.PushKitty(out)
				}
				render()
				pressed = true
				continue
			}
			key, dx, dy, ok := normDir(ev)
			if !ok {
				// Espace ou Entrée : coffre si adjacent, sinon PNJ si adjacent.
				if (ev.K == tui.KeyEnter || (ev.K == tui.KeyRune && ev.R == ' ')) && !rel {
					if inIn != nil && inIn.isNearChest() {
						if kitty {
							tui.PopKitty(out)
						}
						RunChestMenu(in, out, c, ch, &ch.Chest, render)
						if kitty {
							tui.PushKitty(out)
						}
						render()
						pressed = true
					} else if inIn != nil && inIn.isNearNPC() {
						if kitty {
							tui.PopKitty(out)
						}
						// Le PNJ lache sa vanne d'abord. S'il est parti
						// (q/ECHAP), on n'ouvre pas son menu derriere.
						openMenu := true
						if line, ok := npcBanter(inIn.In.Kind); ok {
							openMenu = !tui.Dialogue(c, out, in, inIn.In.NPC.Name+" : "+line)
						}
						if openMenu {
							switch inIn.In.Kind {
							case world.ZoneShop:
								RunShopMenu(in, out, c, ch, render)
							case world.ZoneForge:
								RunForgeMenu(in, out, c, ch, render)
							case world.ZoneGuild:
								RunGuildMenu(in, out, c, ch, gs, render)
							}
						}
						if kitty {
							tui.PushKitty(out)
						}
						render()
						pressed = true
					}
				}
				continue
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

// nextPos calcule la case visée par un déplacement d'UN axe, SANS bouger
// le joueur ni consulter World — juste de l'arithmétique, pour pouvoir
// vérifier la présence d'un monstre avant de déléguer à MovePlayer.
func nextPos(p *world.Player, dir world.Direction) (int, int) {
	x, y := p.X, p.Y
	switch dir {
	case world.North:
		y--
	case world.South:
		y++
	case world.East:
		x++
	case world.West:
		x--
	}
	return x, y
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
		case 'w', 'W':
			return -1, 1, true
		case 'c', 'C':
			return 1, 1, true
		}
	}
	return 0, 0, false
}
