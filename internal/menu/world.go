package menu

import (
	"strconv"

	"runa/internal/character"
	"runa/internal/tui"
	"runa/internal/world"
)

// tileScale : chaque case monde vaut 2x2 cellules -> zoom avant,
// deux fois moins de cases visibles (surtout en vertical).
const tileScale = 2

// worldMapLayer renvoie une FuncLayer qui dessine la fenêtre de tiles
// centrée sur le joueur, directement sur le canvas plein écran.
func worldMapLayer(w *world.World, p *world.Player) tui.FuncLayer {
	return tui.FuncLayer{
		Visible: true,
		Draw: func(c *tui.Canvas) {
			halfW, halfH := c.W/(2*tileScale), c.H/(2*tileScale)
			for sy := 0; sy < c.H; sy++ {
				for sx := 0; sx < c.W; sx++ {
					wx := p.X + (sx/tileScale - halfW)
					wy := p.Y + (sy/tileScale - halfH)
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

// hudLayer construit le HUD en Layer classique,
// centré en haut, composité PAR-DESSUS la carte (dedans, pas à côté).
func hudLayer(c *tui.Canvas, ch *character.Character) tui.Layer {
	x := (c.W - 30) / 2
	if x < 0 {
		x = 0
	}
	l := tui.NewLayer(x, 1, 30, 4)
	tui.DrawBoxWithTitle(l.C, 0, 0, 30, 4, ch.Name)
	l.C.Write(2, 2, "HP: ")
	hpText := strconv.Itoa(int(ch.Hp)) + "/" + strconv.Itoa(int(ch.HpMax))
	l.C.WriteStyled(6, 2, hpText, tui.FGLightRed, "")
	return l
}
