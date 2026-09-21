package tui

// Dialogue : boite de dialogue prete a l'emploi.
// Affiche s en bas de l'ecran, lettre par lettre (balises /COULEUR/
// comprises), puis attend ENTREE. Rend true si le joueur veut
// quitter (q / ECHAP), false sinon.
//
// La boite est creuse comme toutes les boites tui : seul le cadre
// est dessine, jamais l'interieur. Ici on efface juste le rectangle
// de la boite avant (sinon le texte serait illisible sur le fond).
// Exemple : Dialogue(c, out, in, "/RED/gg/RED/ test")

import (
	"io"
	"time"
)

// dialogueSpeed c'est le temps entre chaque lettre.
const dialogueSpeed = 25 * time.Millisecond

// Dialogue affiche s dans une boite en bas du canvas.
// Voir le commentaire du fichier pour le contrat complet.
func Dialogue(c *Canvas, out io.Writer, in io.Reader, s string) bool {
	if c == nil || out == nil {
		return true
	}
	plain, colors := ParseMarkup(s)
	inner := c.W - 8
	if inner < 10 {
		inner = 10
	}
	lines := WrapText(plain, inner)
	bw := c.W - 4
	if bw < 12 {
		bw = c.W
	}
	if bw < 12 {
		bw = 12
	}
	bh := len(lines) + 3
	if bh < 5 {
		bh = 5
	}
	if bh > c.H {
		bh = c.H
	}
	maxLines := bh - 3
	if maxLines < 1 {
		maxLines = 1
	}
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	x := 2
	if x+bw > c.W {
		x = 0
	}
	y := c.H - bh - 1
	if y < 0 {
		y = 0
	}

	// Etape 1 : le texte s'ecrit tout seul.
	tw := NewTypewriter(plain)
	for !IsDone(tw) {
		Tick(tw, 1)
		drawDialogue(c, x, y, bw, bh, inner, lines, colors, VisibleLen(tw))
		if err := FlushStyled(out, c); err != nil {
			return true
		}
		time.Sleep(dialogueSpeed)
	}

	// Etape 2 : texte complet, on attend ENTREE.
	drawDialogue(c, x, y, bw, bh, inner, lines, colors, len([]rune(plain)))
	if err := FlushStyled(out, c); err != nil {
		return true
	}
	if in == nil {
		return false
	}
	for {
		ev, err := ReadKey(in)
		if err != nil {
			return true
		}
		if ev.K == KeyQuit || ev.K == KeyEsc {
			return true
		}
		if ev.K == KeyEnter {
			return false
		}
	}
}

// drawDialogue redessine la boite + les n premieres lettres.
// n = lettres revelees par le typewriter.
func drawDialogue(c *Canvas, x, y, bw, bh, inner int, lines []string, colors []string, n int) {
	FillRect(c, x, y, bw, bh, ' ')
	DrawBoxWithTitle(c, x, y, bw, bh, "DIALOGUE")
	drawn := 0
	for li, line := range lines {
		if drawn >= n {
			break
		}
		base := li * inner
		pos := 0
		for _, r := range line {
			if drawn >= n {
				break
			}
			fg := ""
			if base+pos < len(colors) {
				fg = colors[base+pos]
			}
			c.SetStyled(x+2+pos, y+1+li, r, fg, "")
			pos++
			drawn++
		}
	}
	c.Write(x+2, y+bh-2, "[ENTREE] suite  [q] quitter")
}

// WrapText coupe s en morceaux de width lettres.
// Simple : on coupe aux lettres, pas aux mots.
func WrapText(s string, width int) []string {
	if width <= 0 {
		return []string{s}
	}
	letters := []rune(s)
	out := []string{}
	for i := 0; i < len(letters); i += width {
		end := i + width
		if end > len(letters) {
			end = len(letters)
		}
		out = append(out, string(letters[i:end]))
	}
	if len(out) == 0 {
		return []string{""}
	}
	return out
}
