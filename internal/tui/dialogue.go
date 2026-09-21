package tui

// Dialogue : boite de dialogue prete a l'emploi.
// Affiche un ou plusieurs textes en bas de l'ecran, lettre par
// lettre (balises /COULEUR/ comprises).
//
// Controles : ESPACE pendant l'ecriture = skip (tout afficher
// d'un coup). ESPACE ou ENTREE quand c'est deja tout affiche =
// passer au texte suivant. S'il n'y a plus de texte, on revient
// au jeu. ECHAP / q = quitter le dialogue sans passer a la suite.
// Rend true si quitte (ECHAP/q), false si suite normale.
//
// La boite est creuse comme toutes les boites tui : seul le cadre
// est dessine, jamais l'interieur. Ici on efface juste le rectangle
// de la boite avant (sinon le texte serait illisible sur le fond).
// Exemple : Dialogue(c, out, in, "/RED/gg/RED/ test", "suite")

import (
	"io"
	"time"
)

// dialogueSpeed c'est le temps entre chaque lettre.
const dialogueSpeed = 25 * time.Millisecond

// Dialogue affiche un ou plusieurs textes a la suite.
// Voir le commentaire du fichier pour le contrat complet.
func Dialogue(c *Canvas, out io.Writer, in io.Reader, texts ...string) bool {
	if c == nil || out == nil || len(texts) == 0 {
		return false
	}
	// In nil = on affiche juste le premier texte sans attendre.
	if in == nil {
		showOne(c, out, texts[0], nil)
		return false
	}

	// On lit les touches en fond : ReadKey bloque, donc goroutine
	// qui pousse dans un canal. EOF = on envoie ECHAP (quitter
	// proprement, comme avant) puis on ferme. done permet a la
	// goroutine de mourir quand Dialogue rend la main.
	keys := make(chan Event, 16)
	done := make(chan struct{})
	defer close(done)
	go func() {
		defer close(keys)
		for {
			ev, err := ReadKey(in)
			if err != nil {
				select {
				case keys <- Event{K: KeyEsc}:
				case <-done:
				}
				return
			}
			select {
			case keys <- ev:
			case <-done:
				return
			}
		}
	}()

	for idx, s := range texts {
		plain, colors := ParseMarkup(s)
		box := layoutDialogue(c, plain)
		tw := NewTypewriter(plain)

		// Etape 1 : ecriture lettre par lettre, interruptible.
		for !IsDone(tw) {
			Tick(tw, 1)
			drawDialogue(c, box, colors, VisibleLen(tw))
			if err := FlushStyled(out, c); err != nil {
				return true
			}
			// Attend dialogueSpeed mais guette ESPACE / q / ECHAP.
			select {
			case ev, ok := <-keys:
				if !ok {
					return true
				}
				if isQuit(ev) {
					return true
				}
				if isSkip(ev) {
					Skip(tw)
					drawDialogue(c, box, colors, len([]rune(plain)))
					_ = FlushStyled(out, c)
				}
				// Les autres touches restent dans le canal : la
				// suivante servira a l'etape d'apres, dans l'ordre.
				// (Pas de drain : chaque appui = une etape.)
			case <-time.After(dialogueSpeed):
			}
		}

		// Etape 2 : texte complet, attend ESPACE/ENTREE pour suite.
		drawDialogue(c, box, colors, len([]rune(plain)))
		_ = FlushStyled(out, c)

		// Dernier texte : on sort apres un appui (comme quitter).
		// Textes du milieu : on enchaine.
		if idx == len(texts)-1 {
			if waitNext(keys) {
				return false
			}
			return true
		}
		if waitNext(keys) {
			continue
		}
		return true
	}
	return false
}

// layoutDialogue calcule ou et comment dessiner la boite.
type dialogueBox struct {
	x, y, bw, bh, inner int
	lines               []string
}

func layoutDialogue(c *Canvas, plain string) dialogueBox {
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
	return dialogueBox{x: x, y: y, bw: bw, bh: bh, inner: inner, lines: lines}
}

// showOne dessine un texte d'un coup sans boucle (in == nil).
func showOne(c *Canvas, out io.Writer, s string, _ io.Reader) {
	plain, colors := ParseMarkup(s)
	box := layoutDialogue(c, plain)
	drawDialogue(c, box, colors, len([]rune(plain)))
	_ = FlushStyled(out, c)
}

// isSkip dit si la touche doit sauter l'anim : ESPACE ou ENTREE.
func isSkip(ev Event) bool {
	if ev.K == KeyEnter {
		return true
	}
	if ev.K == KeyRune && ev.R == ' ' {
		return true
	}
	return false
}

// isQuit dit si la touche doit quitter : q ou ECHAP.
func isQuit(ev Event) bool {
	if ev.K == KeyEsc {
		return true
	}
	if ev.K == KeyRune && (ev.R == 'q' || ev.R == 'Q') {
		return true
	}
	return false
}

// waitNext attend ESPACE/ENTREE (suite) ou q/ECHAP (quitter) ou EOF.
// Rend true = suite, false = quitter.
func waitNext(keys <-chan Event) bool {
	for ev := range keys {
		if isQuit(ev) {
			return false
		}
		if isSkip(ev) {
			return true
		}
	}
	return false
}

// drawDialogue redessine la boite + les n premieres lettres.
func drawDialogue(c *Canvas, box dialogueBox, colors []string, n int) {
	FillRect(c, box.x, box.y, box.bw, box.bh, ' ')
	DrawBoxWithTitle(c, box.x, box.y, box.bw, box.bh, "DIALOGUE")
	drawn := 0
	for li, line := range box.lines {
		if drawn >= n {
			break
		}
		base := li * box.inner
		pos := 0
		for _, r := range line {
			if drawn >= n {
				break
			}
			fg := ""
			if base+pos < len(colors) {
				fg = colors[base+pos]
			}
			c.SetStyled(box.x+2+pos, box.y+1+li, r, fg, "")
			pos++
			drawn++
		}
	}
	c.Write(box.x+2, box.y+box.bh-2, "[ESPACE] suite  [q] quitter")
}

// WrapText coupe s en morceaux de width lettres.
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
