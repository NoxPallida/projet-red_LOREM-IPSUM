package main

import (
	"fmt"
	"os"
	"runa/internal/menu"
)

// introPath : cinematique jouee derriere la boite PLAY au lancement.
// Absent = pas d'intro, le jeu demarre direct.
const introPath = "cinematic/videoplayback.cine"

func main() {
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
