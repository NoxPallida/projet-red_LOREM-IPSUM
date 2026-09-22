package menu

// Intro : cinematique de lancement avec une boite PLAY par-dessus.
// La video defile derriere, la boite reste fixe au milieu.

import (
	"fmt"
	"io"
	"os"
	"time"

	"runa/internal/cine"
	"runa/internal/tui"
)

// ShowIntro joue le film path avec une boite PLAY au milieu.
// q / ECHAP / ENTREE passe l'intro.Erreur si le fichier est illisible.
func ShowIntro(out io.Writer, in io.Reader, path string) error {
	if out == nil {
		out = os.Stdout
	}
	if in == nil {
		in = os.Stdin
	}
	m, err := cine.Load(path)
	if err != nil {
		return err
	}
	return cine.PlayWith(out, in, m, drawPlayBox)
}

// storyTexts : l'histoire du debut, racontee pendant que soul.cine tourne.
var storyTexts = []string{
	"... . . .. Hey ... Reveille-toi !! ",
	"Ecoute-moi bien, voyageur. Le village de LOREM est en danger.",
	"Les ombres avancent, et il nous faut un heros. Alors... quel est ton nom ?",
}

// ShowStory joue path EN BOUCLE avec l'histoire par-dessus, puis demande
// le pseudo dans une inputbox. Rend le pseudo tape (X).
// ESPACE/ENTREE : skip l'anim / texte suivant / valider le pseudo.
// q/ECHAP (ou EOF) : on saute l'intro, rend "", nil.
func ShowStory(in io.Reader, out io.Writer, path string) (string, error) {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	m, err := cine.Load(path)
	if err != nil {
		return "", err
	}
	if len(m.Frames) == 0 {
		return "", fmt.Errorf("intro: film vide")
	}
	w, h, err := tui.Size()
	if err != nil {
		return "", err
	}
	c := tui.NewCanvas(w, h)
	if err := tui.EnterAltScreen(out); err != nil {
		return "", err
	}
	defer func() {
		_ = tui.ExitAltScreen(out)
	}()

	// Pompe partagee, arretee avant Ask (qui a la sienne).
	keys, stop := tui.PumpKeys(in)
	defer stop()

	frameDur := time.Second / time.Duration(m.FPS)
	if m.FPS <= 0 {
		frameDur = 100 * time.Millisecond
	}
	// Une lettre toutes les ~25ms, une frame par frameDur.
	letterEvery := 1
	if d := int(frameDur / (25 * time.Millisecond)); d > 1 {
		letterEvery = d
	}

	// L'histoire, texte par texte, video en boucle derriere.
	frame := 0
	for _, s := range storyTexts {
		plain, colors := tui.ParseMarkup(s)
		box := tui.LayoutBottomBox(c, "HISTOIRE", plain)
		tw := tui.NewTypewriter(plain)
		step := 0
		finished := false
		for !finished {
			c.Clear()
			cine.DrawFrame(c, m, frame%len(m.Frames))
			frame++
			tui.DrawTextBox(c, box, colors, tui.VisibleLen(tw), "[ESPACE] suite")
			if err := tui.FlushStyled(out, c); err != nil {
				return "", nil
			}
			select {
			case ev, ok := <-keys:
				if !ok || tui.IsQuit(ev) {
					return "", nil
				}
				if tui.IsConfirm(ev) {
					if !tui.IsDone(tw) {
						tui.Skip(tw)
					} else {
						finished = true
					}
				}
			case <-time.After(frameDur):
				step++
				if !tui.IsDone(tw) && step%letterEvery == 0 {
					tui.Tick(tw, 1)
				}
			}
		}
	}
	// L'inputbox du pseudo, video toujours en boucle derriere.
	// On reutilise LE MEME canal (AskKeys), pas une 2e pompe :
	// sinon les touches deja lues par la 1re seraient perdues.
	// (defer stop() en tete de fonction nettoie a la fin.)
	behind := func(c *tui.Canvas) {
		cine.DrawFrame(c, m, frame%len(m.Frames))
		frame++
	}
	name, ok := tui.AskKeys(c, out, keys, "QUI ES-TU ?", "Quel est ton nom, voyageur ?", 12, behind)
	if !ok {
		return "", nil
	}
	return name, nil
}

// drawPlayBox dessine la boite PLAY centree, par-dessus la frame.
// Que PLAY, rien d'autre : pas de titre, pas d'indice.
// Boite opaque : l'interieur est rempli en noir, on ne voit plus
// la video a travers. Seul le texte PLAY reste, blanc sur noir.
func drawPlayBox(c *tui.Canvas) {
	if c == nil {
		return
	}
	bw, bh := 26, 5
	if bw > c.W {
		bw = c.W
	}
	if bh > c.H {
		bh = c.H
	}
	x, y := tui.CenteredBox(c.W, c.H, bw, bh)
	// Fond noir opaque d'abord : cache la video derriere la boite.
	tui.FillStyled(c, x, y, bw, bh, ' ', "", tui.BGBlack)
	tui.DrawBox(c, x, y, bw, bh)
	label := []rune("PLAY")
	lx := x + (bw-len(label))/2
	if lx < x+1 {
		lx = x + 1
	}
	c.WriteStyled(lx, y+bh/2, "PLAY", tui.FGBrightWhite, tui.BGBlack)
}
