package menu

import (
	"io"
	"time"

	"runa/internal/character"
	"runa/internal/tui"
	"runa/internal/world"
)

// deadFont : lettres 5x7 en blocs pleins pour l'ecran de mort.
var deadFont = map[rune][]string{
	'D': {
		"████ ",
		"█   █",
		"█   █",
		"█   █",
		"█   █",
		"█   █",
		"████ ",
	},
	'E': {
		"█████",
		"█    ",
		"█    ",
		"████ ",
		"█    ",
		"█    ",
		"█████",
	},
	'A': {
		"  █  ",
		" █ █ ",
		"█   █",
		"█████",
		"█   █",
		"█   █",
		"█   █",
	},
}

// showDeathScreen : fond noir, "DEAD" geant en rouge, puis on attend
// ESPACE ou ENTREE. EOF (tests) : on rend la main sans attendre.
func showDeathScreen(in io.Reader, out io.Writer, c *tui.Canvas) {
	tui.FillStyled(c, 0, 0, c.W, c.H, ' ', "", tui.BGBlack)
	word := "DEAD"
	bw := len(word)*6 - 1
	ox := (c.W - bw) / 2
	oy := (c.H - 7) / 2
	for i, r := range []rune(word) {
		for dy, row := range deadFont[r] {
			for dx, cr := range row {
				if cr == ' ' {
					continue
				}
				c.SetStyled(ox+i*6+dx, oy+dy, cr, tui.FGRed, "")
			}
		}
	}
	hint := "ESPACE pour renaitre"
	hx := (c.W - len([]rune(hint))) / 2
	if hx < 0 {
		hx = 0
	}
	c.Write(hx, oy+9, hint)
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
