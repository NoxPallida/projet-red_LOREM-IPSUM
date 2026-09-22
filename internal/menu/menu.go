package menu

import (
	"io"
	"os"

	"runa/internal/character"
	"runa/internal/tui"
)

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

// Setup joue l'intro-histoire (video en boucle + pseudo) et rend
// le pseudo tape (X). Sans fichier video : rend "", nil.
func Setup(in io.Reader, out io.Writer, SetupPath string) (string, error) {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	if _, err := os.Stat(SetupPath); err != nil {
		return "", nil
	}
	return ShowStory(in, out, SetupPath)
}

func Run(in io.Reader, out io.Writer, name string) error {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}

	if err := tui.EnterAltScreen(out); err != nil {
		return err
	}
	defer func() {
		_ = tui.ExitAltScreen(out)
	}()

	// Suite normale : on entre dans la map monde (map.txt).
	return StartGame(character.InitCharacter(name, character.Human))
}
