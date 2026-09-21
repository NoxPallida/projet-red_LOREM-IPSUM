package menu

import (
	"fmt"
	"io"
	"os"

	"runa/internal/tui"
)

const dialogText = "Placeholder : bla bla bleuh bul ezyzkfkf ..."

// Run ouvre l'ecran du menu avec une boite de dialogue en bas.
//
// in c'est le clavier : tout ce que le joueur tape arrive par la.
// En jeu on passe os.Stdin, en test un strings.Reader avec les
// touches qu'on veut simuler ("q", "\n"...).
//
// out c'est l'ecran : tout ce qu'on affiche est ecrit dedans.
// En jeu on passe os.Stdout, en test un bytes.Buffer qu'on relit
// pour verifier ce qui a ete dessine.
//
// nil = defaut : in devient os.Stdin, out devient os.Stdout.
func Run(in io.Reader, out io.Writer) error {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}

	// Taille dynamique du terminal sinon err (skill issue).
	w, h, err := tui.Size()
	if err != nil {
		return fmt.Errorf("menu: %w", err)
	}

	c := tui.NewCanvas(w, h)

	if err := tui.EnterAltScreen(out); err != nil {
		return err
	}
	defer func() {
		_ = tui.ExitAltScreen(out)
	}()

	if tui.Dialogue(c, out, in, dialogText) {
		return nil
	}

	// clear box
	c.Clear()
	if err := tui.Flush(out, c.String()); err != nil {
		return err
	}
	return nil
}
