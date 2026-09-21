package menu

import (
	"strconv"

	"runa/internal/character"
	"runa/internal/tui"
	"runa/internal/world"
)

// worldMapLayer renvoie une FuncLayer qui dessine la fenêtre de tiles
// centrée sur le joueur, directement sur le canvas plein écran.
func worldMapLayer(w *world.World, p *world.Player) tui.FuncLayer {
	return tui.FuncLayer{
		Visible: true,
		Draw: func(c *tui.Canvas) {
			halfW, halfH := c.W/2, c.H/2
			for sy := 0; sy < c.H; sy++ {
				for sx := 0; sx < c.W; sx++ {
					wx := p.X + (sx - halfW)
					wy := p.Y + (sy - halfH)
					tile, loaded := w.TileAt(wx, wy)
					if !loaded {
						c.Set(sx, sy, ' ')
						continue
					}
					c.SetStyled(sx, sy, tile.Symbol, tile.FG, "")
				}
			}
			c.SetStyled(halfW, halfH, '@', tui.FGBrightWhite, "")
		},
	}
}

// hudLayer construit le HUD en Layer classique,
// positionné en haut à gauche, composité PAR-DESSUS la carte.
func hudLayer(ch *character.Character) tui.Layer {
	l := tui.NewLayer(0, 0, 30, 4)
	tui.DrawBoxWithTitle(l.C, 0, 0, 30, 4, ch.Name)
	l.C.Write(2, 2, "HP: ")
	hpText := strconv.Itoa(int(ch.Hp)) + "/" + strconv.Itoa(int(ch.HpMax))
	l.C.WriteStyled(6, 2, hpText, tui.FGLightRed, "")
	return l
}
