package menu

import (
	"io"
	"time"

	"runa/internal/character"
	"runa/internal/tui"
	"runa/internal/world"
)

// deadBanner : banniere DEAD en blocs pleins et ombres pour l'ecran de mort.
var deadBanner = []string{
	" ██████████   ██████████   █████████   ██████████",
	"▒▒███▒▒▒▒███ ▒▒███▒▒▒▒▒█  ███▒▒▒▒▒███ ▒▒███▒▒▒▒███",
	" ▒███   ▒▒███ ▒███  █ ▒  ▒███    ▒███  ▒███   ▒▒███",
	" ▒███    ▒███ ▒██████    ▒███████████  ▒███    ▒███",
	" ▒███    ▒███ ▒███▒▒█    ▒███▒▒▒▒▒███  ▒███    ▒███",
	" ▒███    ███  ▒███ ▒   █ ▒███    ▒███  ▒███    ███",
	" ██████████   ██████████ █████   █████ ██████████",
	" ▒▒▒▒▒▒▒▒▒▒   ▒▒▒▒▒▒▒▒▒▒ ▒▒▒▒▒   ▒▒▒▒▒ ▒▒▒▒▒▒▒▒▒▒",
}

// showDeathScreen : fond noir, "DEAD" geant en rouge, puis on attend
// ESPACE ou ENTREE. EOF (tests) : on rend la main sans attendre.
func showDeathScreen(in io.Reader, out io.Writer, c *tui.Canvas) {
	tui.FillStyled(c, 0, 0, c.W, c.H, ' ', "", tui.BGBlack)
	bannerW := 51
	ox := (c.W - bannerW) / 2
	if ox < 0 {
		ox = 0
	}
	oy := (c.H - len(deadBanner) - 3) / 2
	if oy < 0 {
		oy = 0
	}
	for dy, row := range deadBanner {
		for dx, cr := range []rune(row) {
			if cr == ' ' {
				continue
			}
			fg := tui.FGLightRed
			if cr == '▒' {
				fg = tui.FGRed
			}
			c.SetStyled(ox+dx, oy+dy, cr, fg, "")
		}
	}
	hint := "Press SPACE to respawn"
	hx := (c.W - len([]rune(hint))) / 2
	if hx < 0 {
		hx = 0
	}
	c.WriteStyled(hx, oy+len(deadBanner)+2, hint, tui.FGBrightWhite, "")
	_ = tui.FlushStyled(out, c)
	for {
		ev, rel, have, err := tui.PollKey(in)
		if err != nil {
			return
		}
		if !have {
			time.Sleep(keySleep)
			continue
		}
		if rel {
			continue
		}
		if ev.K == tui.KeyEnter || (ev.K == tui.KeyRune && ev.R == ' ') {
			return
		}
	}
}

// respawn replace le joueur au spawn avec 50 % de ses PV max.
func respawn(p *world.Player, spawnX, spawnY int, ch *character.Character) {
	p.X, p.Y = spawnX, spawnY
	ch.Hp = ch.HpMax / 2
	if ch.Hp == 0 {
		ch.Hp = 1
	}
}
