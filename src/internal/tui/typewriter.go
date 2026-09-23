package tui

// full = tout le texte, pos = lettres deja montrees
type Typewriter struct {
	full []rune
	pos  int
}

// NewTypewriter cree un typewriter pour le texte s
// Au debut rien n'est visible, pos vaut 0
func NewTypewriter(s string) *Typewriter {
	return &Typewriter{full: []rune(s)}
}

// Tick montre n lettres de plus et rend le texte visible
// n negatif = 0. depasse jamais la fin du texte
func Tick(t *Typewriter, n int) string {
	if t == nil {
		return ""
	}
	if n < 0 {
		n = 0
	}
	t.pos += n
	if t.pos > len(t.full) {
		t.pos = len(t.full)
	}
	return string(t.full[:t.pos])
}

// Skip : affiche all
func Skip(t *Typewriter) string {
	if t == nil {
		return ""
	}
	t.pos = len(t.full)
	return string(t.full)
}

// IsDone ->  repond a : tout est write ?.
func IsDone(t *Typewriter) bool {
	if t == nil {
		return true
	}
	return t.pos >= len(t.full)
}

// VisibleText rend le texte visible sans avancer
// Pratique pour redessiner la meme frame sans reveler de lettre
func VisibleText(t *Typewriter) string {
	if t == nil || t.pos <= 0 {
		return ""
	}
	if t.pos > len(t.full) {
		t.pos = len(t.full)
	}
	return string(t.full[:t.pos])
}

// VisibleLen compte les lettres visibles
// Sert a brancher les couleurs : WriteColoredRunes(..., VisibleLen(tw), ...)
func VisibleLen(t *Typewriter) int {
	return len([]rune(VisibleText(t)))
}

// ResetTypewriter recommence a zero / Sans argument ca rejoue juste le meme
func ResetTypewriter(t *Typewriter, s ...string) {
	if t == nil {
		return
	}
	if len(s) > 0 {
		t.full = []rune(s[0])
	}
	t.pos = 0
}
