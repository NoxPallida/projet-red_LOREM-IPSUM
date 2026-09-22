package menu

import (
	"strconv"

	"runa/internal/character"
	"runa/internal/enemies"
	"runa/internal/guild"
	"runa/internal/spawner"
	"runa/internal/tui"
	"runa/internal/world"
)

// tileScale : chaque case monde vaut 2x2 cellules -> zoom avant,
// deux fois moins de cases visibles (surtout en vertical).
const tileScale = 2

// worldMapLayer renvoie une FuncLayer qui dessine la fenêtre de tiles
// centrée sur le joueur, directement sur le canvas plein écran, avec
// les monstres vivants superposés sur la carte.
func worldMapLayer(w *world.World, p *world.Player, sp *spawner.Spawner) tui.FuncLayer {
	return tui.FuncLayer{
		Visible: true,
		Draw: func(c *tui.Canvas) {
			halfW, halfH := c.W/(2*tileScale), c.H/(2*tileScale)
			for sy := 0; sy < c.H; sy++ {
				for sx := 0; sx < c.W; sx++ {
					wx := p.X + (sx/tileScale - halfW)
					wy := p.Y + (sy/tileScale - halfH)

					if enemy, found := sp.EnemyAt(wx, wy); found {
						drawEnemyCell(c, sx, sy, enemy)
						continue
					}

					tile, loaded := w.TileAt(wx, wy)
					if !loaded {
						c.Set(sx, sy, ' ')
						continue
					}
					c.SetStyled(sx, sy, tile.Symbol, tile.FG, "")
				}
			}
			// Joueur : pave 2x2 blanc brillant au centre.
			for dy := 0; dy < tileScale; dy++ {
				for dx := 0; dx < tileScale; dx++ {
					c.SetStyled(halfW*tileScale+dx, halfH*tileScale+dy, '█', tui.FGBrightWhite, "")
				}
			}
		},
	}
}

// drawEnemyCell peint UN pixel écran (sx, sy) appartenant au pavé 2x2
// d'un monstre : un bloc plein coloré selon l'espèce, sauf le coin
// haut-gauche du pavé qui porte une lettre d'identification.
// sx%tileScale==0 && sy%tileScale==0 repère ce coin car les blocs sont
// alignés sur la grille écran depuis sx=0 (division entière sx/tileScale).
func drawEnemyCell(c *tui.Canvas, sx, sy int, e *enemies.EnemyInstance) {
	fg := enemyColor(e)
	if sx%tileScale == 0 && sy%tileScale == 0 {
		c.SetStyled(sx, sy, enemyGlyph(e), fg, "")
		return
	}
	c.SetStyled(sx, sy, '█', fg, "")
}

// enemyGlyph : une seule lettre par monstre, basée sur son ID de
// template plutôt que sur son nom (stable même si le nom affiché change).
func enemyGlyph(e *enemies.EnemyInstance) rune {
	switch e.Template.ID {
	case "rat":
		return 'r'
	case "wolf":
		return 'w'
	case "boar":
		return 'b'
	case "troll":
		return 'T'
	case "goblin":
		return 'g'
	default:
		return 'm'
	}
}

// enemyColor distingue les espèces par couleur, en plus de la lettre :
// plus le monstre est costaud, plus la couleur tranche.
func enemyColor(e *enemies.EnemyInstance) string {
	switch e.Template.ID {
	case "rat":
		return tui.FGLightYellow
	case "wolf":
		return tui.FGLightRed
	case "boar":
		return tui.FGRed
	case "troll":
		return tui.FGMagenta
	case "goblin":
		return tui.FGGreen
	default:
		return tui.FGLightRed
	}
}

// displayInfo construit le HUD en Layer classique,
// centré en haut, composité PAR-DESSUS la carte (dedans, pas à côté).
func displayInfo(c *tui.Canvas, ch *character.Character, gs *guild.GuildStatus) tui.Layer {
	x := (c.W - 32) / 2
	if x < 0 {
		x = 0
	}
	l := tui.NewLayer(x, 1, 32, 6)
	tui.DrawBoxWithTitle(l.C, 0, 0, 32, 6, ch.Name)
	l.C.Write(2, 2, "HP: ")
	hpText := strconv.Itoa(int(ch.Hp)) + "/" + strconv.Itoa(int(ch.HpMax))
	l.C.WriteStyled(6, 2, hpText, tui.FGLightRed, "")
	// Niveau en jaune sur la ligne 3
	lvlText := strconv.Itoa((int(ch.Level())))
	lvlStr := "Lvl: " + lvlText
	l.C.WriteStyled(2, 3, lvlStr, tui.FGYellow, "")
	// XP en cyan juste à côté du niveau, sur la même ligne
	xpNeeded := ch.XpNeededForNext()
	xpCurrent := strconv.Itoa(int(ch.XP))
	xpInfo := "XP: " + xpCurrent
	if xpNeeded == 0 {
		// Niveau 1 : affichage simplifié
		xpInfo += "/0"
	} else {
		xpTotal := strconv.Itoa(xpNeeded)
		xpInfo += "/" + xpTotal
	}
	// Position X pour l'XP : juste après "Lvl: ZZ" + un espace
	xpCol := 2 + len(lvlStr) + 1
	l.C.WriteStyled(xpCol, 3, xpInfo, tui.FGCyan, "")
	// Guild rank on line 4
	l.C.Write(2, 4, "Rank: ")
	l.C.WriteStyled(8, 4, gs.Rank.String(), tui.FGGreen, "")
	return l
}

// encounterLayer displays a message when the player bumps into a static monster.
func encounterLayer(c *tui.Canvas, e *enemies.EnemyInstance) tui.Layer {
	w, h := 40, 4
	x := (c.W - w) / 2
	if x < 0 {
		x = 0
	}
	l := tui.NewLayer(x, c.H-h-1, w, h) // at bottom of screen, outside HUD
	tui.DrawBoxWithTitle(l.C, 0, 0, w, h, "Encounter")
	text := e.Template.Name + " (Lvl " + strconv.Itoa(int(e.Level)) + ") blocks the path"
	l.C.Write(2, 2, text)
	return l
}
