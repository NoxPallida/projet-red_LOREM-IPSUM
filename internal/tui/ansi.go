package tui

import "io"

const (
	altScreenOn  = "\x1b[?1049h" // ne touche pas scrollback
	altScreenOff = "\x1b[?1049l"
	cursorHide   = "\x1b[?25l"
	cursorShow   = "\x1b[?25h"
	homeCursor   = "\x1b[H" // start screen
)

// EnterAltScreen switches to the alternate buffer and hides the cursor.
func EnterAltScreen(w io.Writer) error {
	_, err := io.WriteString(w, altScreenOn+cursorHide)
	return err
}

// ExitAltScreen restores the normal buffer,  Deferred in menu.Run so the terminal
func ExitAltScreen(w io.Writer) error {
	_, err := io.WriteString(w, altScreenOff+cursorShow)
	return err
}

// Flush writes a full frame : move the cursor home and overwrite the previous frame
func Flush(w io.Writer, frame string) error {
	_, err := io.WriteString(w, homeCursor+frame)
	return err
}
