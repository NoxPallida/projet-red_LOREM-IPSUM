package tui

import (
	"golang.org/x/term"
	"os"
)

func Size() (width, height int, err error) {
	return term.GetSize(int(os.Stdout.Fd()))
}
