package tui

import (
	"io"
	"strings"
)

const (
	altScreenOn  = "\x1b[?1049h" // ne touche pas scrollback
	altScreenOff = "\x1b[?1049l"
	cursorHide   = "\x1b[?25l"
	cursorShow   = "\x1b[?25h"
	homeCursor   = "\x1b[H" // start screen
)

// EnterAltScreen switches to the alternate buffer and hides the cursor
func EnterAltScreen(w io.Writer) error {
	_, err := io.WriteString(w, altScreenOn+cursorHide)
	return err
}

// ExitAltScreen restores the normal buffer,  Deferred in menu.Run so the terminal
func ExitAltScreen(w io.Writer) error {
	_, err := io.WriteString(w, altScreenOff+cursorShow)
	return err
}

// nlToCRLF convertit les fins de ligne pour le mode raw.
// En raw OPOST est coupe : "\n" seul ne revient plus en debut
// de ligne (effet escalier). "\r\n" marche dans les deux modes
// (en cooked le terminal le replie pareil), donc on convertit
// toujours : un seul endroit a maintenir pour tout l'affichage.
func nlToCRLF(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\n", "\r\n")
}

// Flush writes a full frame : move the cursor home and overwrite the previous frame
func Flush(w io.Writer, frame string) error {
	_, err := io.WriteString(w, homeCursor+nlToCRLF(frame))
	return err
}

// FlushStyled envoie un canvas AVEC ses couleurs
// Meme chose que Flush mais avec RenderStyled : a utiliser des qu'on
// a pose du style (WriteStyled, WriteMarkup, boite coloree...)
// Sans style sur le canvas, le resultat est identique a Flush
func FlushStyled(w io.Writer, c *Canvas) error {
	if c == nil {
		return nil
	}
	_, err := io.WriteString(w, homeCursor+nlToCRLF(c.RenderStyled()))
	return err
}
