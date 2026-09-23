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
// Variante : DialogueWith(c, out, in, "QUETE", fondVideo, texts...)
// pour un titre et un fond perso, DialogueKeys(...) sur un canal existant.

import (
	"io"
	"time"
)

// dialogueSpeed c'est le temps entre chaque lettre.
const dialogueSpeed = 25 * time.Millisecond

// Dialogue affiche un ou plusieurs textes a la suite (titre DIALOGUE,
// sans fond). Voir le commentaire du fichier pour le contrat complet.
func Dialogue(c *Canvas, out io.Writer, in io.Reader, texts ...string) bool {
	return DialogueWith(c, out, in, "DIALOGUE", nil, texts...)
}

// DialogueWith : comme Dialogue, mais titre parametrable et fond
// anime derriere (nil = aucun fond, juste la boite).
func DialogueWith(c *Canvas, out io.Writer, in io.Reader, title string, behind Draw, texts ...string) bool {
	if c == nil || out == nil || len(texts) == 0 {
		return false
	}
	// In nil = on affiche juste le premier texte sans attendre.
	if in == nil {
		showOne(c, out, title, behind, texts[0])
		return false
	}

	// Pompe a touches partagee (voir prompt.go).
	keys, stop := PumpKeys(in)
	defer stop()
	return DialogueKeys(c, out, keys, title, behind, texts...)
}

// DialogueKeys : comme DialogueWith mais sur un canal de touches
// existant. A utiliser quand l'appelant pompe deja les touches
// (meme raison qu'AskKeys) : creer une 2e pompe volerait des touches.
func DialogueKeys(c *Canvas, out io.Writer, keys <-chan Event, title string, behind Draw, texts ...string) bool {
	if c == nil || out == nil || keys == nil || len(texts) == 0 {
		return false
	}
	draw := func(box TextBox, colors []string, n int) error {
		if behind != nil {
			c.Clear()
			behind(c)
		}
		DrawTextBox(c, box, colors, n, "[ESPACE] suite  [q] quitter")
		return FlushStyled(out, c)
	}

	for idx, s := range texts {
		plain, colors := ParseMarkup(s)
		box := LayoutBottomBox(c, title, plain)
		tw := NewTypewriter(plain)

		// Etape 1 : ecriture lettre par lettre, interruptible.
		for !IsDone(tw) {
			Tick(tw, 1)
			if err := draw(box, colors, VisibleLen(tw)); err != nil {
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
					_ = draw(box, colors, len([]rune(plain)))
				}
				// Les autres touches restent dans le canal : la
				// suivante servira a l'etape d'apres, dans l'ordre.
				// (Pas de drain : chaque appui = une etape.)
			case <-time.After(dialogueSpeed):
			}
		}

		// Etape 2 : texte complet, attend ESPACE/ENTREE pour suite.
		_ = draw(box, colors, len([]rune(plain)))

		// Dernier texte : on sort apres un appui (comme quitter).
		// Textes du milieu : on enchaine.
		// On redessine en attendant la touche pour que le fond
		// (ex : video) continue de jouer au lieu de geler.
		redraw := func() {
			_ = draw(box, colors, len([]rune(plain)))
		}
		if idx == len(texts)-1 {
			if waitNext(keys, redraw) {
				return false
			}
			return true
		}
		if waitNext(keys, redraw) {
			continue
		}
		return true
	}
	return false
}

// showOne dessine un texte d'un coup sans boucle (in == nil).
func showOne(c *Canvas, out io.Writer, title string, behind Draw, s string) {
	plain, colors := ParseMarkup(s)
	box := LayoutBottomBox(c, title, plain)
	if behind != nil {
		c.Clear()
		behind(c)
	}
	DrawTextBox(c, box, colors, len([]rune(plain)), "[ESPACE] suite  [q] quitter")
	_ = FlushStyled(out, c)
}

// waitNext attend ESPACE/ENTREE (suite) ou q/ECHAP (quitter) ou EOF.
// Entre deux touches, tick() redessine (toutes les 100ms) pour que
// le fond continue (ex : video) au lieu de geler sur la derniere frame.
// Rend true = suite, false = quitter.
func waitNext(keys <-chan Event, tick func()) bool {
	for {
		select {
		case ev, ok := <-keys:
			if !ok {
				return false
			}
			if IsQuit(ev) {
				return false
			}
			if IsConfirm(ev) {
				return true
			}
		case <-time.After(100 * time.Millisecond):
			if tick != nil {
				tick()
			}
		}
	}
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
