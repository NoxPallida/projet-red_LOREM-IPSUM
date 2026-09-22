package cine

// Player : joue un Movie charge avec Load, en plein terminal.
// ESPACE / ECHAP / ENTREE passe la cinematique, EOF l'arrete aussi.

import (
	"fmt"
	"io"
	"strings"
	"time"

	"runa/internal/tui"
)

// Play affiche chaque frame a 1/FPS seconde.
// Le film garde sa taille (W x H), colle en haut a gauche du canvas.
// Sans overlay : appelle PlayWith avec nil.
func Play(out io.Writer, in io.Reader, m Movie) error {
	return PlayWith(out, in, m, nil)
}

// PlayWith joue comme Play, mais appelle overlay(c) apres chaque frame.
// L'overlay dessine PAR-DESSUS la video (boite, titre...) : il est
// redessine a chaque frame donc il reste fixe pendant que ca defile.
// overlay nil = Play normal.
func PlayWith(out io.Writer, in io.Reader, m Movie, overlay func(c *tui.Canvas)) error {
	if len(m.Frames) == 0 {
		return fmt.Errorf("cine: film vide")
	}
	if m.FPS <= 0 {
		return fmt.Errorf("cine: fps invalide %d", m.FPS)
	}
	w, h, err := tui.Size()
	if err != nil || w <= 0 || h <= 0 {
		w, h = m.W, m.H
	}
	c := tui.NewCanvas(w, h)
	if err := tui.EnterAltScreen(out); err != nil {
		return err
	}
	defer func() {
		_ = tui.ExitAltScreen(out)
	}()

	// Touche de skip en fond : ReadKey bloque, donc goroutine.
	// On ferme stop a la premiere touche ou a EOF.
	stop := make(chan struct{})
	go func() {
		for {
			ev, err := tui.ReadKey(in)
			if err != nil {
				close(stop)
				return
			}
			if ev.K == tui.KeyEsc || ev.K == tui.KeyEnter || (ev.K == tui.KeyRune && ev.R == ' ') {
				close(stop)
				return
			}
		}
	}()

	frameDur := time.Second / time.Duration(m.FPS)
	for i := range m.Frames {
		c.Clear()
		DrawFrame(c, m, i)
		if overlay != nil {
			overlay(c)
		}
		if err := tui.FlushStyled(out, c); err != nil {
			return err
		}
		select {
		case <-stop:
			return nil
		case <-time.After(frameDur):
		}
	}
	return nil
}

// DrawFrame pose la frame i du film sur le canvas (le fond seul),
// centree horizontalement et verticalement par rapport a la taille du canvas.
// L'appelant Clear avant, et dessine son overlay apres.
func DrawFrame(c *tui.Canvas, m Movie, i int) {
	if c == nil || i < 0 || i >= len(m.Frames) || i >= len(m.ColorFrames) {
		return
	}
	offsetX := (c.W - m.W) / 2
	offsetY := (c.H - m.H) / 2

	clines := strings.Split(m.ColorFrames[i], "\n")
	for y, line := range strings.Split(m.Frames[i], "\n") {
		targetY := y + offsetY
		x := 0
		for _, r := range line {
			targetX := x + offsetX
			fg := ""
			if y < len(clines) {
				cr := []rune(clines[y])
				if x < len(cr) {
					fg = tui.FGByIndex(hexVal(cr[x]))
				}
			}
			c.SetStyled(targetX, targetY, r, fg, "")
			x++
		}
	}
}
