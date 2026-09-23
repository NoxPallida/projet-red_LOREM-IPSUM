package assets

import (
	"embed"
	"strings"
)

//go:embed ascii/*.txt
var asciiFS embed.FS

// Art rend les lignes de assets/ascii/<name>.txt (sans l'extension).
// Retourne nil si le fichier manque : l'appelant n'affiche rien.
// Les lignes vides de fin sont retirees, les '\r' nettoyes.
func Art(name string) []string {
	b, err := asciiFS.ReadFile("ascii/" + name + ".txt")
	if err != nil {
		return nil
	}
	lines := strings.Split(string(b), "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], "\r")
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}
