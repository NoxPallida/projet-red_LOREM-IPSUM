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

// Setup joue l'intro-histoire (video en boucle + pseudo), le dialogue
// et la machine a sous (race et mana), puis rend le personnage cree.
// Sans fichier video : rend un personnage par defaut.
func Setup(in io.Reader, out io.Writer, SetupPath string) (*character.Character, error) {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	if _, err := os.Stat(SetupPath); err != nil {
		return character.InitCharacter("Voyageur", character.Human), nil
	}
	return ShowStory(in, out, SetupPath)
}

func Run(in io.Reader, out io.Writer, ch *character.Character) error {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	if ch == nil {
		ch = character.InitCharacter("Voyageur", character.Human)
	}

	if err := tui.EnterAltScreen(out); err != nil {
		return err
	}
	defer func() {
		_ = tui.ExitAltScreen(out)
	}()

	// Suite normale : on entre dans la map monde (map.txt).
	return StartGame(ch)
}
