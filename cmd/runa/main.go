package main

import (
	"fmt"
	"os"
	"runa/internal/menu"
	"runa/internal/tui"
)

// introPath : cinematique jouee derriere la boite PLAY au lancement.
// Absent = pas d'intro, le jeu demarre direct.
const introPath = "cinematic/videoplayback.cine"

func main() {
	// Mode raw : chaque touche (ESPACE, fleches...) arrive aussitot.
	// Sans ca (mode cooked), ESPACE n'est delivre qu'apres ENTREE.
	// L'affichage reste bon car Flush/FlushStyled emettent "\r\n"
	// (OPOST etant coupe en raw, "\n" seul ferait un escalier).
	if tui.IsTerminal(os.Stdin) {
		st, err := tui.MakeRaw(os.Stdin)
		if err == nil {
			defer func() {
				_ = tui.Restore(os.Stdin, st)
			}()
		}
	}
	// Intro si le fichier existe, sinon on passe direct au jeu.
	if _, err := os.Stat(introPath); err == nil {
		if err := menu.ShowIntro(os.Stdout, os.Stdin, introPath); err != nil {
			fmt.Fprintln(os.Stderr, "red:", err)
			os.Exit(1)
		}
	}
	if err := menu.Run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "red:", err)
		os.Exit(1)
	}
}
