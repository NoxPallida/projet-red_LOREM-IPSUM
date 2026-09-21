package menu

// Demo post-intro : fabrique une fausse erreur avec NewError et
// l'affiche X secondes en haut a droite (bord rouge, fond rouge
// clair, texte blanc), pour verifier que ca marche. Temporaire.

import (
	"fmt"
	"io"
	"path"
	"strconv"
	"time"

	"runa/internal/tui"
)

// testErrorSecs = duree d'affichage de la demo.
const testErrorSecs = 3

// showTestError dessine la msgbox d'erreur et attend X secondes.
func showTestError(c *tui.Canvas, out io.Writer) {
	if c == nil || out == nil {
		return
	}
	// Fausse erreur : le fichier est capture tout seul par NewError.
	testErr := tui.NewError("demo post-intro", fmt.Errorf("asset manquant"))
	lines := []string{
		"fichier: " + path.Base(testErr.File) + ":" + strconv.Itoa(testErr.Line),
		"op: " + testErr.Operation,
		"cause: " + testErr.Err.Error(),
	}

	// Boite en haut a droite, taille selon le texte.
	bw := 4
	for _, l := range lines {
		if len([]rune(l))+6 > bw {
			bw = len([]rune(l)) + 6
		}
	}
	bh := len(lines) + 4
	x := c.W - bw - 1
	if x < 0 {
		x = 0
	}
	y := 1

	// Fond rouge clair + bord degrade rouge + texte blanc.
	tui.FillStyled(c, x, y, bw, bh, ' ', tui.FGBrightWhite, tui.BGLightRed)
	tui.DrawBoxGradient(c, x, y, bw, bh, "ERREUR TEST", tui.Gradient("FIRE"), tui.BGLightRed)
	for i, l := range lines {
		c.WriteStyled(x+3, y+2+i, l, tui.FGBrightWhite, tui.BGLightRed)
	}
	_ = tui.FlushStyled(out, c)
	time.Sleep(testErrorSecs * time.Second)
}
