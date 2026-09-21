package tui

import "strings"

// cell c'est le contenu d'UNE case : une lettre + sa couleur
// fg et bg sont des codes ANSI ("") veut dire couleur du terminal)
// Exemple : fg="\x1b[31m" (rouge), bg="" (fond par defaut)
type cell struct {
	r  rune
	fg string
	bg string
}

// Canvas c'est la memoire video : une grille de cases qu'on redessine
// a chaque frame puis qu'on envoie d'un coup a l'ecran
// L'ordre de dessin compte : le dernier qui ecrit gagne, case par case
// C'est comme ca qu'une boite de dialogue cache la carte en dessous
type Canvas struct {
	W, H  int
	cells []cell
}

// NewCanvas cree un canvas vide de largeur w et hauteur h
func NewCanvas(w, h int) *Canvas {
	c := &Canvas{W: w, H: h, cells: make([]cell, w*h)}
	c.Clear()
	return c
}

// Clear remplit tout le canvas avec des espaces sans couleur
func (c *Canvas) Clear() {
	for i := range c.cells {
		c.cells[i] = cell{r: ' '}
	}
}

// Set ecrit une seule lettre sans couleur a la position (x, y)
// La couleur de la case est effacee : redessiner du texte normal
// par-dessus du texte colore remet la couleur par defaut
// Si c'est en dehors de l'ecran, c'est ignore sans planter :
// un calque peut depasser du canvas sans faire crasher l'affichage
func (c *Canvas) Set(x, y int, r rune) {
	c.SetStyled(x, y, r, "", "")
}

// SetStyled ecrit une lettre AVEC sa couleur
// fg = couleur du texte, bg = couleur du fond, "" = defaut
func (c *Canvas) SetStyled(x, y int, r rune, fg, bg string) {
	if c == nil {
		return
	}
	if x < 0 || x >= c.W || y < 0 || y >= c.H {
		return
	}
	c.cells[y*c.W+x] = cell{r: r, fg: fg, bg: bg}
}

// Write ecrit du texte sans couleur a la position (x, y)
// une case par lettre
// On compte en lettres et pas en octets, donc les accents ne decalent pas la mise en page
func (c *Canvas) Write(x, y int, s string) {
	i := 0
	for _, r := range s {
		c.Set(x+i, y, r)
		i++
	}
}

// WriteStyled ecrit du texte colore
// Pour du texte multicolore (balises /RED/), voir WriteMarkup
func (c *Canvas) WriteStyled(x, y int, s, fg, bg string) {
	i := 0
	for _, r := range s {
		c.SetStyled(x+i, y, r, fg, bg)
		i++
	}
}

// String aplati le canvas en texte brut, ligne par ligne
// Les couleurs sont ignorees : pratique pour les tests (comparer du texte sans codes ANSI parasites).
func (c *Canvas) String() string {
	var b strings.Builder
	for y := 0; y < c.H; y++ {
		for x := 0; x < c.W; x++ {
			b.WriteRune(c.cells[y*c.W+x].r)
		}
		if y < c.H-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// RenderStyled aplati le canvas en texte AVEC les couleurs ANSI
// Un code n'est emis que quand la couleur change, et chaque ligne
// finit par un reset
// A envoyer avec FlushStyled (pas Flush, qui attend du texte brut)
func (c *Canvas) RenderStyled() string {
	const reset = "\x1b[0m"
	var b strings.Builder
	for y := 0; y < c.H; y++ {
		curFG, curBG := "", ""
		for x := 0; x < c.W; x++ {
			cl := c.cells[y*c.W+x]
			if cl.fg != curFG || cl.bg != curBG {
				b.WriteString(reset)
				b.WriteString(cl.fg)
				b.WriteString(cl.bg)
				curFG, curBG = cl.fg, cl.bg
			}
			b.WriteRune(cl.r)
		}
		if curFG != "" || curBG != "" {
			b.WriteString(reset)
		}
		if y < c.H-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}
