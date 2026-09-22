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

	// Pompe a touches partagee (voir prompt.go).
	keys, stop := PumpKeys(in)
	defer stop()

	for idx, s := range texts {
		plain, colors := ParseMarkup(s)
		box := LayoutBottomBox(c, "DIALOGUE", plain)
		tw := NewTypewriter(plain)

		// Etape 1 : ecriture lettre par lettre, interruptible.
		for !IsDone(tw) {
			Tick(tw, 1)
			DrawTextBox(c, box, colors, VisibleLen(tw), "[ESPACE] suite  [q] quitter")
			if err := FlushStyled(out, c); err != nil {
				return true
			}
			// Attend dialogueSpeed mais guette ESPACE / q / ECHAP.
			select {
			case ev, ok := <-keys:
				if !ok {
					return true
				}
				if IsQuit(ev) {
					return true
				}
				if IsConfirm(ev) {
					Skip(tw)
					DrawTextBox(c, box, colors, len([]rune(plain)), "[ESPACE] suite  [q] quitter")
					_ = FlushStyled(out, c)
				}
				// Les autres touches restent dans le canal : la
				// suivante servira a l'etape d'apres, dans l'ordre.
				// (Pas de drain : chaque appui = une etape.)
			case <-time.After(dialogueSpeed):
			}
		}

		// Etape 2 : texte complet, attend ESPACE/ENTREE pour suite.
		DrawTextBox(c, box, colors, len([]rune(plain)), "[ESPACE] suite  [q] quitter")
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

// showOne dessine un texte d'un coup sans boucle (in == nil).
func showOne(c *Canvas, out io.Writer, s string, _ io.Reader) {
	plain, colors := ParseMarkup(s)
	box := LayoutBottomBox(c, "DIALOGUE", plain)
	DrawTextBox(c, box, colors, len([]rune(plain)), "[ESPACE] suite  [q] quitter")
	_ = FlushStyled(out, c)
}

// waitNext attend ESPACE/ENTREE (suite) ou q/ECHAP (quitter) ou EOF.
// Rend true = suite, false = quitter.
func waitNext(keys <-chan Event) bool {
	for ev := range keys {
		if IsQuit(ev) {
			return false
		}
		if IsConfirm(ev) {
			return true
		}
	}
	return false
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
