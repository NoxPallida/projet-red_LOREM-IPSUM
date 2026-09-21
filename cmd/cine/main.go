package main

import (
	"fmt"
	"image"
	_ "image/png"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"runa/internal/cine"
)

const ramp = " .:-=+*#%░▒▓█"

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Convertir : go run ./cmd/cine video.mp4 [-o film.cine] [-w 80] [-h 22] [-fps 10] [-maxf 300] [-sat 2.0]")
		fmt.Println("Jouer     : go run ./cmd/cine play film.cine")
		os.Exit(2)
	}

	if os.Args[1] == "play" {
		if len(os.Args) < 3 {
			fmt.Println("Jouer : go run ./cmd/cine play film.cine")
			os.Exit(2)
		}
		m, err := cine.Load(os.Args[2])
		if err != nil {
			fmt.Println("erreur :", err)
			os.Exit(1)
		}
		if err := cine.Play(os.Stdout, os.Stdin, m); err != nil {
			fmt.Println("erreur :", err)
			os.Exit(1)
		}
		return
	}

	in := os.Args[1]
	out := ""
	w, h, fps, maxf := 80, 22, 10, 300
	sat := 2.0
	args := os.Args[2:]

	for i := 0; i < len(args); i++ {
		get := func() string {
			if i+1 >= len(args) {
				return ""
			}
			i++
			return args[i]
		}

		switch args[i] {
		case "-o":
			out = get()
		case "-w":
			w, _ = strconv.Atoi(get())
		case "-h":
			h, _ = strconv.Atoi(get())
		case "-fps":
			fps, _ = strconv.Atoi(get())
		case "-maxf":
			maxf, _ = strconv.Atoi(get())
		case "-sat":
			sat, _ = strconv.ParseFloat(get(), 64)
		default:
			fmt.Println("option inconnue :", args[i])
			os.Exit(2)
		}
	}

	if w <= 0 || h <= 0 || fps <= 0 || maxf <= 0 || sat <= 0 {
		fmt.Println("w, h, fps, maxf et sat doivent etre > 0")
		os.Exit(2)
	}

	if out == "" {
		out = strings.TrimSuffix(in, filepath.Ext(in)) + ".cine"
	}

	if err := convert(in, out, w, h, fps, maxf, sat); err != nil {
		fmt.Println("erreur :", err)
		os.Exit(1)
	}
}

// convert fait tout : ffmpeg -> png -> ascii + couleurs -> .cine
func convert(in, out string, w, h, fps, maxf int, sat float64) error {
	if _, err := os.Stat(in); err != nil {
		return fmt.Errorf("video illisible : %w", err)
	}
	tmp, err := os.MkdirTemp("", "cine")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	// ffmpeg : une image par frame, taille exacte, fps voulu.
	// exec.Command prend le programme PUIS chaque argument separe
	// (pas une phrase entiere : ca ne passe pas par un shell).
	filter := "scale=" + strconv.Itoa(w) + ":" + strconv.Itoa(h) + ",fps=" + strconv.Itoa(fps)
	cmd := exec.Command("ffmpeg", "-y", "-v", "error", "-i", in, "-vf", filter, filepath.Join(tmp, "f_%04d.png"))
	if msg, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg : %w (%s)", err, strings.TrimSpace(string(msg)))
	}

	paths, err := filepath.Glob(filepath.Join(tmp, "f_*.png"))
	if err != nil || len(paths) == 0 {
		return fmt.Errorf("aucune frame sortie de ffmpeg")
	}
	sort.Strings(paths)
	if len(paths) > maxf {
		paths = paths[:maxf]
	}

	m := cine.Movie{W: w, H: h, FPS: fps}
	for _, p := range paths {
		lines, clines, err := imageToASCII(p, w, h, sat)
		if err != nil {
			return err
		}
		m.Frames = append(m.Frames, strings.Join(lines, "\n"))
		m.ColorFrames = append(m.ColorFrames, strings.Join(clines, "\n"))
	}
	if err := cine.Save(out, m); err != nil {
		return err
	}
	abs, _ := filepath.Abs(out)
	fmt.Printf("OK : %s (%d frames %dx%d a %d fps)\n", abs, len(m.Frames), w, h, fps)
	return nil
}

// rgbTable donne le RVB de chaque index 0-15, dans l'ordre de tui.FGPalette.
// Le pixel prend la couleur la plus proche (distance au carre).
var rgbTable = [16][3]int{
	{0, 0, 0},       // BLACK
	{170, 0, 0},     // RED
	{0, 170, 0},     // GREEN
	{170, 85, 0},    // YELLOW
	{0, 0, 170},     // BLUE
	{170, 0, 170},   // MAGENTA
	{0, 170, 170},   // CYAN
	{170, 170, 170}, // WHITE
	{85, 85, 85},    // GRAY
	{255, 85, 85},   // LIGHTRED
	{85, 255, 85},   // LIGHTGREEN
	{255, 255, 85},  // LIGHTYELLOW
	{85, 85, 255},   // LIGHTBLUE
	{255, 85, 255},  // LIGHTMAGENTA
	{85, 255, 255},  // LIGHTCYAN
	{255, 255, 255}, // BRIGHTWHITE
}

// nearest rend l'index 0-15 de la couleur la plus proche de (r, g, b).
func nearest(r, g, b int) int {
	best, bestD := 0, -1
	for i, c := range rgbTable {
		dr, dg, db := r-c[0], g-c[1], b-c[2]
		d := dr*dr + dg*dg + db*db
		if bestD < 0 || d < bestD {
			best, bestD = i, d
		}
	}
	return best
}

// hexDigit rend le chiffre hexa 0-9A-F de l'index 0-15.
func hexDigit(i int) rune {
	if i < 10 {
		return rune('0' + i)
	}
	return rune('A' + i - 10)
}

// boost sature le pixel : eloigne chaque canal de la moyenne.
// sat=1 = original, 2 = deux fois plus vif. La moyenne ne change pas,
// donc la lettre (niveau de gris) reste la meme, seule la couleur vive.
// Sans ca, une video terne donne que du gris + jaune.
func boost(r, g, b int, sat float64) (int, int, int) {
	m := float64(r+g+b) / 3
	clamp := func(c int) int {
		v := int(m + sat*(float64(c)-m))
		if v < 0 {
			return 0
		}
		if v > 255 {
			return 255
		}
		return v
	}
	return clamp(r), clamp(g), clamp(b)
}

// imageToASCII lit un png et rend H lignes de W lettres + H lignes
// de W chiffres hexa (la couleur de chaque case, saturee x sat).
func imageToASCII(path string, w, h int, sat float64) (lines, clines []string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, nil, fmt.Errorf("decode %s : %w", filepath.Base(path), err)
	}
	bounds := img.Bounds()
	iw, ih := bounds.Dx(), bounds.Dy()
	levels := []rune(ramp)
	lines = make([]string, h)
	clines = make([]string, h)
	for y := 0; y < h; y++ {
		row := make([]rune, w)
		crow := make([]rune, w)
		for x := 0; x < w; x++ {
			// Echantillonne le pixel correspondant dans l'image d'origine.
			r, g, b, _ := img.At(bounds.Min.X+x*iw/w, bounds.Min.Y+y*ih/h).RGBA()
			r8, g8, b8 := int(r>>8), int(g>>8), int(b>>8)
			gray := uint8((r8 + g8 + b8) / 3) // 0 (noir) .. 255 (blanc)
			row[x] = levels[int(gray)*(len(levels)-1)/255]
			br, bg, bb := boost(r8, g8, b8, sat)
			crow[x] = hexDigit(nearest(br, bg, bb))
		}
		lines[y] = string(row)
		clines[y] = string(crow)
	}
	return lines, clines, nil
}
