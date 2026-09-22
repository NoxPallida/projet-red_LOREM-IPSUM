package main

import (
	"fmt"
	"os"

	"golang.org/x/term"

	"runa/internal/audio"
	"runa/internal/menu"
	"runa/internal/tui"
)

// introPath : cinematique jouee derriere la boite PLAY au lancement.
// Absent = pas d'intro, le jeu demarre direct.
const introPath = "assets/cinematic/videoplayback.cine"

func main() {
	// Mode raw : chaque touche (ESPACE, fleches...) arrive aussitot.
	// Sans ca (mode cooked), ESPACE n'est delivre qu'apres ENTREE.
	// L'affichage reste bon car Flush/FlushStyled emettent "\r\n"
	// (OPOST etant coupe en raw, "\n" seul ferait un escalier).
	// Pas de defer : on restaure le terminal à la main avant de sortir.
	var rawState *term.State
	if tui.IsTerminal(os.Stdin) {
		if st, err := tui.MakeRaw(os.Stdin); err == nil {
			rawState = st
		}
	}
	code := run()
	if rawState != nil {
		_ = tui.Restore(os.Stdin, rawState)
	}
	os.Exit(code)
}

// run fait le travail et rend un code de sortie (0 = ok).
func run() int {
	// Musique des l'intro : decompressee en memoire, jouee en fond.
	// Sans carte son ca previent et le jeu continue sans musique.
	stopMusic := audio.StartOST()
	defer stopMusic()

	// Intro si le fichier existe, sinon on passe direct au jeu.
	if _, err := os.Stat(introPath); err == nil {
		if err := menu.ShowIntro(os.Stdout, os.Stdin, introPath); err != nil {
			fmt.Fprintln(os.Stderr, "red:", err)
			return 1
		}
	}
	if err := menu.Run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "red:", err)
		return 1
	}
	return 0
}
