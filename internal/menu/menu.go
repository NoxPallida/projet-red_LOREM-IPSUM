package menu

import (
	"fmt"
	"io"
	"os"
	"time"

	"runa/internal/tui"
)

const dialogText = "Placeholder : bla bla bleuh bul ezyzkfkf ..."

func Run(in io.Reader, out io.Writer) error {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}

	// Taille dynamique du terminal sinon err ( skill issue )
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

	// Boite en bas
	boxW := w - 4
	if boxW < 10 {
		boxW = w
	}
	if boxW < 10 {
		boxW = 10
	}
	boxH := 7
	x := 2
	y := h - boxH - 1
	if y < 0 {
		y = 0
	}

	// dessiner la boite + ecrire texte
	tw := tui.NewTypewriter(dialogText)
	for !tui.IsDone(tw) {
		tui.Tick(tw, 1)
		drawDialog(c, x, y, boxW, boxH, tui.VisibleText(tw))
		if err := tui.Flush(out, c.String()); err != nil {
			return err
		}
		time.Sleep(25 * time.Millisecond)
	}

	// attendre ENTREE pour close
	drawDialog(c, x, y, boxW, boxH, tui.VisibleText(tw))
	if err := tui.Flush(out, c.String()); err != nil {
		return err
	}
	for {
		ev, err := tui.ReadKey(in)
		if err != nil {
			return tui.Flush(out, c.String())
		}
		if ev.K == tui.KeyQuit || ev.K == tui.KeyEsc {
			return tui.Flush(out, c.String())
		}
		if ev.K == tui.KeyEnter {
			break
		}
	}

	// clear box
	c.Clear()
	if err := tui.Flush(out, c.String()); err != nil {
		return err
	}
	return nil
}

// drawDialog redessine tout sur le meme canvas ( sa save des perf de pas clear a chaque fois )
func drawDialog(c *tui.Canvas, x, y, boxW, boxH int, text string) {
	c.Clear()
	tui.DrawBoxWithTitle(c, x, y, boxW, boxH, "DIALOGUE")
	lines := cutLines(text, boxW-4)
	maxLines := boxH - 3
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	for i, line := range lines {
		c.Write(x+2, y+1+i, line)
	}
	c.Write(x+2, y+boxH-2, "[ENTREE] fermer  [q] quitter")
}

// cutLines coupe le texte en morceaux de width lettres.
// Simple : on coupe aux lettres, pas aux mots.
func cutLines(s string, width int) []string {
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
