package cine

// Format .cine couleur (texte, lisible) :
//
//	CINE2              <- magic, verifie au chargement
//	80 22 10 150       <- largeur hauteur fps nombre_de_frames
//	----FRAME----      <- separateur
//	<22 lignes de 80 lettres>
//	----COLOR----      <- couleurs de la frame
//	<22 lignes de 80 chiffres hexa 0-9A-F = index 0-15
//	dans tui.FGPalette>

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Magic ouvre chaque fichier .cine.
const Magic = "CINE2"

// FrameSep separe les frames.
const FrameSep = "----FRAME----"

// ColorSep separe les lettres des couleurs dans une frame.
const ColorSep = "----COLOR----"

// Movie c'est une cinematique ascii couleur prete a jouer.
type Movie struct {
	W, H, FPS int
	Frames    []string // une frame = H lignes collees avec "\n"
	// Une entree par frame comme Frames : H lignes de W chiffres
	// hexa collees avec "\n".
	ColorFrames []string
}

// Save ecrit un film en .cine.
// Chaque frame doit faire H lignes de W lettres + W chiffres hexa.
func Save(path string, m Movie) error {
	if m.W <= 0 || m.H <= 0 || m.FPS <= 0 {
		return fmt.Errorf("cine: dimensions invalides w=%d h=%d fps=%d", m.W, m.H, m.FPS)
	}
	if len(m.ColorFrames) != len(m.Frames) {
		return fmt.Errorf("cine: il manque les couleurs (%d frames, %d blocs couleur)", len(m.Frames), len(m.ColorFrames))
	}
	var b strings.Builder
	b.WriteString(Magic + "\n")
	b.WriteString(strconv.Itoa(m.W) + " " + strconv.Itoa(m.H) + " " + strconv.Itoa(m.FPS) + " " + strconv.Itoa(len(m.Frames)) + "\n")
	for i, f := range m.Frames {
		lines := strings.Split(f, "\n")
		if len(lines) != m.H {
			return fmt.Errorf("cine: frame %d a %d lignes au lieu de %d", i, len(lines), m.H)
		}
		for _, l := range lines {
			if len([]rune(l)) != m.W {
				return fmt.Errorf("cine: frame %d fait %d lettres au lieu de %d", i, len([]rune(l)), m.W)
			}
		}
		clines := strings.Split(m.ColorFrames[i], "\n")
		if len(clines) != m.H {
			return fmt.Errorf("cine: couleurs frame %d : %d lignes au lieu de %d", i, len(clines), m.H)
		}
		for _, l := range clines {
			if !isHexLine(l, m.W) {
				return fmt.Errorf("cine: couleurs frame %d invalides", i)
			}
		}
		b.WriteString(FrameSep + "\n")
		b.WriteString(f + "\n")
		b.WriteString(ColorSep + "\n")
		b.WriteString(m.ColorFrames[i] + "\n")
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// isHexLine dit si s fait W chiffres hexa 0-9A-F.
func isHexLine(s string, w int) bool {
	runes := []rune(s)
	if len(runes) != w {
		return false
	}
	for _, r := range runes {
		if !(r >= '0' && r <= '9' || r >= 'A' && r <= 'F') {
			return false
		}
	}
	return true
}

// hexVal rend l'index 0-15 d'un chiffre hexa, 0 si invalide.
func hexVal(r rune) int {
	switch {
	case r >= '0' && r <= '9':
		return int(r - '0')
	case r >= 'A' && r <= 'F':
		return int(r-'A') + 10
	}
	return 0
}

// Load lit un .cine et verifie magic + dimensions + frames + couleurs.
func Load(path string) (Movie, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Movie{}, err
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) < 2 || lines[0] != Magic {
		return Movie{}, fmt.Errorf("cine: mauvais magic (pas un .cine couleur ?)")
	}
	nums := strings.Fields(lines[1])
	if len(nums) != 4 {
		return Movie{}, fmt.Errorf("cine: header invalide")
	}
	vals := make([]int, 4)
	for i, s := range nums {
		v, err := strconv.Atoi(s)
		if err != nil || v <= 0 {
			return Movie{}, fmt.Errorf("cine: header invalide")
		}
		vals[i] = v
	}
	m := Movie{W: vals[0], H: vals[1], FPS: vals[2]}
	want := vals[3]
	pos := 2
	for len(m.Frames) < want {
		if pos >= len(lines) || lines[pos] != FrameSep {
			return Movie{}, fmt.Errorf("cine: frame %d manquante", len(m.Frames))
		}
		pos++
		if pos+m.H > len(lines) {
			return Movie{}, fmt.Errorf("cine: frame %d tronquee", len(m.Frames))
		}
		for _, l := range lines[pos : pos+m.H] {
			if len([]rune(l)) != m.W {
				return Movie{}, fmt.Errorf("cine: frame %d fait %d lettres au lieu de %d", len(m.Frames), len([]rune(l)), m.W)
			}
		}
		m.Frames = append(m.Frames, strings.Join(lines[pos:pos+m.H], "\n"))
		pos += m.H
		if pos >= len(lines) || lines[pos] != ColorSep {
			return Movie{}, fmt.Errorf("cine: couleurs frame %d manquantes", len(m.Frames)-1)
		}
		pos++
		if pos+m.H > len(lines) {
			return Movie{}, fmt.Errorf("cine: couleurs frame %d tronquees", len(m.Frames)-1)
		}
		for _, l := range lines[pos : pos+m.H] {
			if !isHexLine(l, m.W) {
				return Movie{}, fmt.Errorf("cine: couleurs frame %d invalides", len(m.Frames)-1)
			}
		}
		m.ColorFrames = append(m.ColorFrames, strings.Join(lines[pos:pos+m.H], "\n"))
		pos += m.H
	}
	return m, nil
}
