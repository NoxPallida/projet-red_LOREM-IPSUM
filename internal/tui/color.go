package tui

// Chaque nom existe en version texte (FG...) et fond (BG...)
// "" = (defaut)
const (
	FGBlack        = "\x1b[30m"
	FGRed          = "\x1b[31m"
	FGGreen        = "\x1b[32m"
	FGYellow       = "\x1b[33m"
	FGBlue         = "\x1b[34m"
	FGMagenta      = "\x1b[35m"
	FGCyan         = "\x1b[36m"
	FGWhite        = "\x1b[37m"
	FGGray         = "\x1b[90m"
	FGLightRed     = "\x1b[91m"
	FGLightGreen   = "\x1b[92m"
	FGLightYellow  = "\x1b[93m"
	FGLightBlue    = "\x1b[94m"
	FGLightMagenta = "\x1b[95m"
	FGLightCyan    = "\x1b[96m"
	FGBrightWhite  = "\x1b[97m"

	BGBlack        = "\x1b[40m"
	BGRed          = "\x1b[41m"
	BGGreen        = "\x1b[42m"
	BGYellow       = "\x1b[43m"
	BGBlue         = "\x1b[44m"
	BGMagenta      = "\x1b[45m"
	BGCyan         = "\x1b[46m"
	BGWhite        = "\x1b[47m"
	BGGray         = "\x1b[100m"
	BGLightRed     = "\x1b[101m"
	BGLightGreen   = "\x1b[102m"
	BGLightYellow  = "\x1b[103m"
	BGLightBlue    = "\x1b[104m"
	BGLightMagenta = "\x1b[105m"
	BGLightCyan    = "\x1b[106m"
	BGBrightWhite  = "\x1b[107m"
)

// fgByName retrouve le code texte d'un nom de balise
// false si le nom n'existe pas (la balise sera ecrite telle quelle)
func fgByName(name string) (string, bool) {
	switch name {
	case "BLACK":
		return FGBlack, true
	case "RED":
		return FGRed, true
	case "GREEN":
		return FGGreen, true
	case "YELLOW":
		return FGYellow, true
	case "BLUE":
		return FGBlue, true
	case "MAGENTA":
		return FGMagenta, true
	case "CYAN":
		return FGCyan, true
	case "WHITE":
		return FGWhite, true
	case "GRAY":
		return FGGray, true
	case "LIGHTRED":
		return FGLightRed, true
	case "LIGHTGREEN":
		return FGLightGreen, true
	case "LIGHTYELLOW":
		return FGLightYellow, true
	case "LIGHTBLUE":
		return FGLightBlue, true
	case "LIGHTMAGENTA":
		return FGLightMagenta, true
	case "LIGHTCYAN":
		return FGLightCyan, true
	case "BRIGHTWHITE":
		return FGBrightWhite, true
	}
	return "", false
}

// Degrades prets a l'emploi : des listes de couleurs dans l'ordre
// On les passe a WriteGradient ou DrawBoxGradient qui les repetent
// en boucle sur les lettres / le cadre
var (
	GradientFire   = []string{FGRed, FGLightRed, FGYellow, FGLightYellow}
	GradientOcean  = []string{FGBlue, FGCyan, FGLightCyan, FGBrightWhite}
	GradientForest = []string{FGGreen, FGLightGreen, FGLightYellow}
	GradientSunset = []string{FGRed, FGLightMagenta, FGLightYellow}
	GradientShadow = []string{FGGray, FGWhite, FGBrightWhite}
)

func Gradient(name string) []string {
	switch name {
	case "FIRE":
		return GradientFire
	case "OCEAN":
		return GradientOcean
	case "FOREST":
		return GradientForest
	case "SUNSET":
		return GradientSunset
	case "SHADOW":
		return GradientShadow
	}
	return nil
}

// ParseMarkup decoupe un texte a balises en texte brut + couleur par lettre
// Balise = /NOM/ avec NOM dans la palette (ex : /RED/)
// La balise bascule la couleur : la 1re l'allume, la 2e identique l'eteint
// /RESET/ (ou //) eteint toujours. Balise inconnue = ecrite telle quelle
// Exemple : "test /RED/yo/RESET/ gg" -> "test yo gg" avec "yo" en rouge
//
// Le tableau colors est aligne avec les lettres de plain :
// colors[i] est la couleur de la ieme lettre
func ParseMarkup(s string) (plain string, colors []string) {
	runes := []rune(s)
	var out []rune
	var cols []string
	cur := ""
	i := 0
	for i < len(runes) {
		if runes[i] == '/' {
			// Reset court : // eteint la couleur
			if i+1 < len(runes) && runes[i+1] == '/' {
				cur = ""
				i += 2
				continue
			}
			// Cherche /NOM/ : '/' + lettres majuscules + '/'
			j := i + 1
			for j < len(runes) && runes[j] >= 'A' && runes[j] <= 'Z' {
				j++
			}
			if j > i+1 && j < len(runes) && runes[j] == '/' {
				name := string(runes[i+1 : j])
				if name == "RESET" {
					cur = ""
					i = j + 1
					continue
				}
				if code, ok := fgByName(name); ok {
					if cur == code {
						cur = "" // 2e fois : on eteint (bascule)
					} else {
						cur = code
					}
					i = j + 1
					continue
				}
			}
		}
		out = append(out, runes[i])
		cols = append(cols, cur)
		i++
	}
	return string(out), cols
}

// SplitTypewriter prepare un texte a balises pour le Typewriter
// plain va dans NewTypewriter, colors suit chaque lettre
// Exemple d'usage, frame par frame :
//
//	plain, colors := SplitTypewriter("test /RED/yo/RESET/ gg")
//	tw := NewTypewriter(plain)
//	// ... Tick(tw, 1) puis :
//	WriteColoredRunes(c, x, y, plain, colors, VisibleLen(tw), "")
func SplitTypewriter(s string) (plain string, colors []string) {
	return ParseMarkup(s)
}

// WriteMarkup ecrit un texte a balises /COULEUR/ sur le canvas
// bg = couleur de fond pour tout le texte ("" = defaut)
func WriteMarkup(c *Canvas, x, y int, s, bg string) {
	plain, colors := ParseMarkup(s)
	WriteColoredRunes(c, x, y, plain, colors, len([]rune(plain)), bg)
}

// WriteColoredRunes ecrit les n premieres lettres de plain avec
// leur couleur (tableau aligne venu de ParseMarkup)
// C'est ce qui branche le Typewriter aux couleurs : n = lettres visibles
func WriteColoredRunes(c *Canvas, x, y int, plain string, colors []string, n int, bg string) {
	if c == nil {
		return
	}
	i := 0
	for _, r := range plain {
		if i >= n {
			break
		}
		fg := ""
		if i < len(colors) {
			fg = colors[i]
		}
		c.SetStyled(x+i, y, r, fg, bg)
		i++
	}
}

// WriteGradient ecrit du texte en degrade : les couleurs de fgs
// tournent en boucle sur les lettres (fgs[0], fgs[1], ..., fgs[0]...)
// fgs vide = texte normal sans couleur
func WriteGradient(c *Canvas, x, y int, s, bg string, fgs []string) {
	if len(fgs) == 0 {
		c.Write(x, y, s)
		return
	}
	i := 0
	for _, r := range s {
		c.SetStyled(x+i, y, r, fgs[i%len(fgs)], bg)
		i++
	}
}

// FillStyled remplit un rectangle avec une lettre ET une couleur
// Sert aux fonds de boite coloree (ex : ' ' blanc sur fond rouge clair)
func FillStyled(c *Canvas, x, y, w, h int, r rune, fg, bg string) {
	if c == nil || w <= 0 || h <= 0 {
		return
	}
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			c.SetStyled(x+i, y+j, r, fg, bg)
		}
	}
}

// DrawBoxGradient dessine un cadre dont le bord tourne en degrade
// fgs vide = cadre normal (appel a DrawBox). L'interieur est rempli
// avec bg pour teinter la boite. title en blanc eclatant
func DrawBoxGradient(c *Canvas, x, y, w, h int, title string, fgs []string, bg string) {
	if len(fgs) == 0 {
		DrawBox(c, x, y, w, h)
		return
	}
	if c == nil || w < 2 || h < 2 {
		return
	}
	// Fond teinte d'abord, bord par-dessus
	FillStyled(c, x, y, w, h, ' ', "", bg)
	// Cadre dans l'ordre horaire en partant du coin haut-gauche,
	// pour que le degrade fasse le tour sans cassure
	k := 0
	put := func(px, py int, r rune) {
		c.SetStyled(px, py, r, fgs[k%len(fgs)], bg)
		k++
	}
	put(x, y, '┌')
	for i := 1; i < w-1; i++ {
		put(x+i, y, '─')
	}
	put(x+w-1, y, '┐')
	for j := 1; j < h-1; j++ {
		put(x+w-1, y+j, '│')
	}
	put(x+w-1, y+h-1, '┘')
	for i := w - 2; i >= 1; i-- {
		put(x+i, y+h-1, '─')
	}
	put(x, y+h-1, '└')
	for j := h - 2; j >= 1; j-- {
		put(x, y+j, '│')
	}
	if title != "" && w >= 7 {
		letters := []rune(" " + title + " ")
		max := w - 4
		if len(letters) > max {
			letters = letters[:max]
		}
		c.WriteStyled(x+2, y, string(letters), FGBrightWhite, bg)
	}
}
