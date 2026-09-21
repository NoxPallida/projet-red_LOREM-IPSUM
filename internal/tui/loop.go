package tui

import (
	"fmt"
	"io"
	"os"
)

// Frame renders one frame onto c. ev is nil for the first frame,
// then holds the key pressed since the previous frame.
// Return quit=true to leave RunLoop.
type Frame func(c *Canvas, ev *Event) (quit bool)

// RunLoop c'est lui qui va "drive" le engine en loopant pour render
// w,h <= 0  va juste auto-detect via Size (terminal reel, sinon COLUMNS/LINES)
func RunLoop(in io.Reader, out io.Writer, w, h int, frame Frame) error {
	if out == nil {
		out = os.Stdout
	}
	if frame == nil {
		return nil
	}
	if w <= 0 || h <= 0 {
		var err error
		w, h, err = resolveSize(w, h)
		if err != nil {
			return err
		}
	}
	c := NewCanvas(w, h)

	if err := EnterAltScreen(out); err != nil {
		return err
	}
	defer func() {
		_ = ExitAltScreen(out)
	}()

	//	c.Clear()
	if quit := frame(c, nil); quit {
		return FlushStyled(out, c)
	}
	if err := FlushStyled(out, c); err != nil {
		return err
	}

	for {
		ev, err := ReadKey(in)
		if err != nil {
			// EOF / closed stdin: render last state and exit
			return FlushStyled(out, c)
		}
		if ev.K == KeyEsc {
			return FlushStyled(out, c)
		}
		// Pas de quit sur 'q' : en jeu Q = aller a l'ouest (ZQSD).
		// Seul ECHAP quitte ; le callback frame decide du reste.
		e := ev
		c.Clear()
		if quit := frame(c, &e); quit {
			return FlushStyled(out, c)
		}
		if err := FlushStyled(out, c); err != nil {
			return err
		}
	}
}

func resolveSize(w, h int) (int, int, error) {
	if w > 0 && h > 0 {
		return w, h, nil
	}
	sw, sh, err := Size()
	if err != nil {
		return 0, 0, NewError("tui: impossible de lire la taille (%w)", err)
	}
	if w <= 0 {
		w = sw
	}
	if h <= 0 {
		h = sh
	}
	if w <= 0 || h <= 0 {
		return 0, 0, NewError(fmt.Sprintf("tui: taille invalide w=%d h=%d", w, h), nil)
	}
	return w, h, nil
}
