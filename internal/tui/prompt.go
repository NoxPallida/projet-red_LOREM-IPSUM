package tui

// prompt.go : les briques partagees des interactions clavier.
// Pompe a touches, boite de texte commune, et fonction Ask.
// dialogue.go et menu/intro.go les utilisent au lieu de
// dupliquer chacun leur pompe, leur layout et leur saisie.

import (
	"io"
	"sync"
	"time"
	"unicode"
)

// Draw c'est juste "une fonction qui dessine sur le canvas".
// Ca sert de fond (ex : une frame video) derriere une boite :
// on la passe aux fonctions qui gerent la boucle, au lieu de
// recopier la boucle Clear -> fond -> boite -> flush partout.
type Draw func(c *Canvas)

// PumpKeys lit les touches en fond et les pousse dans un canal,
// car ReadKey bloque. Rend le canal + stop() a defer : stop tue
// la goroutine. EOF = on envoie KeyEsc (quitter proprement)
// puis on ferme le canal.
func PumpKeys(in io.Reader) (<-chan Event, func()) {
	keys := make(chan Event, 16)
	done := make(chan struct{})
	var once sync.Once
	stop := func() { once.Do(func() { close(done) }) }
	go func() {
		defer close(keys)
		for {
			ev, err := ReadKey(in)
			if err != nil {
				select {
				case keys <- Event{K: KeyEsc}:
				case <-done:
				}
				return
			}
			if ev.K == KeyNone {
				continue
			}
			select {
			case keys <- ev:
			case <-done:
				return
			}
		}
	}()
	return keys, stop
}

// IsQuit dit si la touche doit quitter : q ou ECHAP.
func IsQuit(ev Event) bool {
	if ev.K == KeyEsc {
		return true
	}
	return ev.K == KeyRune && (ev.R == 'q' || ev.R == 'Q')
}

// IsConfirm dit si la touche doit valider/passer : ESPACE ou ENTREE.
func IsConfirm(ev Event) bool {
	if ev.K == KeyEnter {
		return true
	}
	return ev.K == KeyRune && ev.R == ' '
}

// TextBox c'est une boite de texte prete a dessiner : position,
// taille, lignes decoupees, titre. Se construit avec LayoutBottomBox.
type TextBox struct {
	X, Y, W, H int
	Inner      int
	Lines      []string
	Title      string
}

// LayoutBottomBox calcule une boite de dialogue en bas de l'ecran
// pour le texte donne (titre affiche en haut du cadre).
func LayoutBottomBox(c *Canvas, title, plain string) TextBox {
	inner := c.W - 8
	if inner < 10 {
		inner = 10
	}
	lines := WrapText(plain, inner)
	bw := c.W - 4
	if bw < 12 {
		bw = c.W
	}
	if bw < 12 {
		bw = 12
	}
	bh := len(lines) + 3
	if bh < 5 {
		bh = 5
	}
	if bh > c.H {
		bh = c.H
	}
	maxLines := bh - 3
	if maxLines < 1 {
		maxLines = 1
	}
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	x := 2
	if x+bw > c.W {
		x = 0
	}
	y := c.H - bh - 1
	if y < 0 {
		y = 0
	}
	return TextBox{X: x, Y: y, W: bw, H: bh, Inner: inner, Lines: lines, Title: title}
}

// DrawTextBox redessine la boite + les n premieres lettres en couleur.
// hint = ligne d'aide en bas du cadre ("" = aucune).
// La boite est creuse comme les autres : seul le cadre est dessine,
// mais on efface d'abord son rectangle (sinon le fond baverait).
func DrawTextBox(c *Canvas, box TextBox, colors []string, n int, hint string) {
	FillRect(c, box.X, box.Y, box.W, box.H, ' ')
	DrawBoxWithTitle(c, box.X, box.Y, box.W, box.H, box.Title)
	drawn := 0
	for li, line := range box.Lines {
		if drawn >= n {
			break
		}
		base := li * box.Inner
		pos := 0
		for _, r := range line {
			if drawn >= n {
				break
			}
			fg := ""
			if base+pos < len(colors) {
				fg = colors[base+pos]
			}
			c.SetStyled(box.X+2+pos, box.Y+1+li, r, fg, "")
			pos++
			drawn++
		}
	}
	if hint != "" {
		c.Write(box.X+2, box.Y+box.H-2, hint)
	}
}

// Ask demande un texte au joueur dans une inputbox centree.
// title = titre du cadre, prompt = question affichee, maxLen = limite.
// behind dessine le fond derriere (nil = ecran vide).
// Rend (texte, true) si ENTREE avec texte non vide,
// ("", false) si ECHAP, EOF ou erreur d'affichage.
// Lettres/chiffres/-/_ acceptes, BACKSPACE efface, curseur qui cligne.
func Ask(c *Canvas, out io.Writer, in io.Reader, title, prompt string, maxLen int, behind Draw) (string, bool) {
	if c == nil || out == nil || in == nil {
		return "", false
	}
	keys, stop := PumpKeys(in)
	defer stop()
	return AskKeys(c, out, keys, title, prompt, maxLen, behind)
}

// AskKeys c'est Ask mais sur un canal de touches existant.
// A utiliser quand l'appelant pompe deja les touches (ex : l'intro
// enchaine histoire + pseudo) : creer une 2e pompe sur le meme
// clavier volerait des touches a la 1re, car les deux liraient en
// meme temps. Ici pas de course : un seul lecteur.
func AskKeys(c *Canvas, out io.Writer, keys <-chan Event, title, prompt string, maxLen int, behind Draw) (string, bool) {
	if c == nil || out == nil || keys == nil {
		return "", false
	}
	if maxLen < 1 {
		maxLen = 1
	}
	var buf []rune
	frame := 0
	for {
		c.Clear()
		if behind != nil {
			behind(c)
		}
		drawAskBox(c, title, prompt, string(buf), frame)
		if err := FlushStyled(out, c); err != nil {
			return "", false
		}
		select {
		case ev, ok := <-keys:
			if !ok {
				return "", false
			}
			switch {
			case ev.K == KeyEnter:
				if len(buf) > 0 {
					return string(buf), true
				}
			case ev.K == KeyBackspace:
				if len(buf) > 0 {
					buf = buf[:len(buf)-1]
				}
			case ev.K == KeyEsc:
				if len(buf) > 0 {
					buf = nil
				}
			case ev.K == KeyRune:
				if len(buf) < maxLen && (unicode.IsLetter(ev.R) || unicode.IsDigit(ev.R) || ev.R == '-' || ev.R == '_') {
					buf = append(buf, ev.R)
				}
			}
		case <-time.After(50 * time.Millisecond):
			frame++
		}
	}
}

// drawAskBox dessine l'inputbox centree : titre, question,
// "> texte + curseur qui cligne" (frame = compteur d'images).
func drawAskBox(c *Canvas, title, prompt, text string, frame int) {
	bw := 44
	if bw > c.W {
		bw = c.W
	}
	bh := 7
	if bh > c.H {
		bh = c.H
	}
	x, y := CenteredBox(c.W, c.H, bw, bh)
	FillRect(c, x, y, bw, bh, ' ')
	DrawBoxWithTitle(c, x, y, bw, bh, title)
	c.Write(x+3, y+2, prompt)
	cur := " "
	if frame%8 < 4 {
		cur = "█"
	}
	c.Write(x+3, y+4, "> "+text+cur)
}
