package tui

import (
	"fmt"
	"os"
	"strconv"

	"golang.org/x/term"
)

func Size() (width, height int, err error) {
	return SizeFrom(os.Stdout)
}

// SizeFrom renvoie taille du terminal
// Ordre : term.GetSize, sinon variables environnement COLUMNS/LINES
func SizeFrom(f *os.File) (width, height int, err error) {
	if f == nil {
		return 0, 0, fmt.Errorf("tui: fichier nil, taille illisible")
	}
	if w, h, err := term.GetSize(int(f.Fd())); err == nil && w > 0 && h > 0 {
		return w, h, nil
	}
	if w, h, ok := sizeFromEnv(); ok {
		return w, h, nil
	}
	return 0, 0, fmt.Errorf("tui: taille terminal illisible (pas un terminal ? renseigne COLUMNS/LINES)")
}

// sizeFromEnv lit COLUMNS/LINES
func sizeFromEnv() (w, h int, ok bool) {
	cw, errW := strconv.Atoi(os.Getenv("COLUMNS")) // Atoi convert string into int
	ch, errH := strconv.Atoi(os.Getenv("LINES"))
	if errW != nil || errH != nil || cw <= 0 || ch <= 0 {
		return 0, 0, false
	}
	return cw, ch, true
}

// IsTerminal dit si f est un vrai terminal ou juste un pipe / fichier.
// sert a choisir le mode d'affichage :
// vrai terminal -> alt-screen + lecture clavier,
// pas de terminal (CI, pipe, test) ( permet de pas return d'erreur lors des test)
func IsTerminal(f *os.File) bool {
	if f == nil {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

// MakeRaw passe le terminal en mode brut : chaque touche arrive
// En mode normal ("cooked"), c'est le terminal qui attend la ligne.
// Rend l'ancien etat : l'appelant doit faire defer Restore juste apres,
// sinon le terminal reste casse quand le programme quitte.
func MakeRaw(f *os.File) (*term.State, error) {
	return term.MakeRaw(int(f.Fd()))
}

// Restore remet le terminal comme avant MakeRaw.
// s peut etre nil : dans ce cas ca ne fait rien, sans erreur.
func Restore(f *os.File, s *term.State) error {
	if s == nil {
		return nil
	}
	return term.Restore(int(f.Fd()), s)
}
