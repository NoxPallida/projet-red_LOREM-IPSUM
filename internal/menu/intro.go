package menu

// Intro : cinematique de lancement avec une boite PLAY par-dessus.
// La video defile derriere, la boite reste fixe au milieu.

import (
	"io"
	"os"

	"runa/internal/cine"
	"runa/internal/tui"
)

// ShowIntro joue le film path avec une boite PLAY au milieu.
// q / ECHAP / ENTREE passe l'intro.Erreur si le fichier est illisible.
func ShowIntro(out io.Writer, in io.Reader, path string) error {
	if out == nil {
		out = os.Stdout
	}
	if in == nil {
		in = os.Stdin
	}
	m, err := cine.Load(path)
	if err != nil {
		return err
	}
	return cine.PlayWith(out, in, m, drawPlayBox)
}

// drawPlayBox dessine la boite PLAY centree, par-dessus la frame.
// Que PLAY, rien d'autre : pas de titre, pas d'indice.
// Boite opaque : l'interieur est rempli en noir, on ne voit plus
// la video a travers. Seul le texte PLAY reste, blanc sur noir.
func drawPlayBox(c *tui.Canvas) {
	if c == nil {
		return
	}
	bw, bh := 26, 5
	if bw > c.W {
		bw = c.W
	}
	if bh > c.H {
		bh = c.H
	}
	x, y := tui.CenteredBox(c.W, c.H, bw, bh)
	// Fond noir opaque d'abord : cache la video derriere la boite.
	tui.FillStyled(c, x, y, bw, bh, ' ', "", tui.BGBlack)
	tui.DrawBox(c, x, y, bw, bh)
	label := []rune("PLAY")
	lx := x + (bw-len(label))/2
	if lx < x+1 {
		lx = x + 1
	}
	c.WriteStyled(lx, y+bh/2, "PLAY", tui.FGBrightWhite, tui.BGBlack)
}
