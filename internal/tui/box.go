package tui

// DrawBox dessine un cadre sur le canvas c.
// x, y = position du coin haut-gauche. w, h = largeur et hauteur.
// Les boites sont creuses : seul le cadre est dessine, l'interieur
// n'est jamais touche.
func DrawBox(c *Canvas, x, y, w, h int) {
	if c == nil {
		return
	}
	if w < 2 {
		return
	}
	if h < 2 {
		return
	}
	// Coins.
	c.Set(x, y, '┌')
	c.Set(x+w-1, y, '┐')
	c.Set(x, y+h-1, '└')
	c.Set(x+w-1, y+h-1, '┘')
	// Lignes haut et bas.
	for i := 1; i < w-1; i++ {
		c.Set(x+i, y, '─')
		c.Set(x+i, y+h-1, '─')
	}
	// Lignes gauche et droite.
	for j := 1; j < h-1; j++ {
		c.Set(x, y+j, '│')
		c.Set(x+w-1, y+j, '│')
	}
}

// dessine un cadre et ecrit le titre
func DrawBoxWithTitle(c *Canvas, x, y, w, h int, title string) {
	DrawBox(c, x, y, w, h)
	if c == nil {
		return
	}
	if title == "" {
		return
	}
	if w < 7 {
		return
	}

	// On coupe le titre si trop long
	letters := []rune(title)
	max := w - 6
	if len(letters) > max {
		letters = letters[:max]
	}
	c.Write(x+3, y, string(letters))
}

// FillRect remplit un rectangle avec le caractere r
func FillRect(c *Canvas, x, y, w, h int, r rune) {
	if c == nil {
		return
	}
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			c.Set(x+i, y+j, r)
		}
	}
}

// CenteredBox donne la position pour centrer un cadre
func CenteredBox(screenW, screenH, w, h int) (x, y int) {
	x = (screenW - w) / 2
	if x < 0 {
		x = 0
	}
	y = (screenH - h) / 2
	if y < 0 {
		y = 0
	}
	return x, y
}
