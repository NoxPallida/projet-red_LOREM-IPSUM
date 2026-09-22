package menu

// Intro : cinematique de lancement avec une boite PLAY par-dessus.
// La video defile derriere, la boite reste fixe au milieu.

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"runa/internal/character"
	"runa/internal/cine"
	"runa/internal/spell"
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

// storyTexts : the beginning story told while soul.cine plays.
var storyTexts = []string{
	"... . . .. Hey ... Wake up!! ",
	"Listen closely, traveler. The village of runa is in danger.",
	"Shadows are approaching, and we need a hero. So... what is your name?",
}

// ShowStory joue path EN BOUCLE avec l'histoire par-dessus, puis demande
// le pseudo dans une inputbox, ajoute le dialogue du destin et affiche
// la machine a sous (race et mana). Rend le Character cree.
// ESPACE/ENTREE : skip l'anim / texte suivant / valider le pseudo.
// q/ECHAP (ou EOF) : on saute l'intro.
func ShowStory(in io.Reader, out io.Writer, path string) (*character.Character, error) {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	m, err := cine.Load(path)
	if err != nil {
		return nil, err
	}
	if len(m.Frames) == 0 {
		return nil, fmt.Errorf("intro: film vide")
	}
	w, h, err := tui.Size()
	if err != nil {
		return nil, err
	}
	c := tui.NewCanvas(w, h)
	if err := tui.EnterAltScreen(out); err != nil {
		return nil, err
	}
	defer func() {
		_ = tui.ExitAltScreen(out)
	}()

	// Pompe partagee, arretee a la fin de l'intro.
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

	// 1. L'histoire, texte par texte, video en boucle derriere.
	frame := 0
	for _, s := range storyTexts {
		plain, colors := tui.ParseMarkup(s)
		box := tui.LayoutBottomBox(c, "STORY", plain)
		tw := tui.NewTypewriter(plain)
		step := 0
		finished := false
		for !finished {
			c.Clear()
			cine.DrawFrame(c, m, frame%len(m.Frames))
			frame++
			tui.DrawTextBox(c, box, colors, tui.VisibleLen(tw), "[SPACE] Next")
			if err := tui.FlushStyled(out, c); err != nil {
				return character.InitCharacter("Traveler", character.Human), nil
			}
			select {
			case ev, ok := <-keys:
				if !ok || tui.IsQuit(ev) {
					return character.InitCharacter("Traveler", character.Human), nil
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

	// 2. L'inputbox du pseudo, video toujours en boucle derriere.
	behind := func(c *tui.Canvas) {
		cine.DrawFrame(c, m, frame%len(m.Frames))
		frame++
	}
	name, ok := tui.AskKeys(c, out, keys, "WHO ARE YOU?", "What is your name, traveler?", 12, behind)
	if !ok || name == "" {
		name = "Traveler"
	}

	// 3. Le joueur choisit sa race, puis sa classe parmi celles de la race.
	chosenClass := chooseRace(c, out, keys, m, &frame)
	chosenSubclass := chooseSubclass(c, out, keys, m, &frame, chosenClass)
	chosenMana := character.Mana(chosenClass).Mana()
	if chosenMana == 0 {
		chosenMana = 25
	}
	ch := character.InitCharacter(name, chosenClass)
	ch.SetSubclass(chosenSubclass)
	ch.Mana = chosenMana
	ch.ManaMax = chosenMana

	// 4. On affiche le resultat tire au sort.
	destinyTexts := []string{
		"Welcome, " + name + " the " + chosenSubclass.String() + " (" + chosenClass.String() + ").",
		"Destiny grants you " + strconv.Itoa(int(chosenMana)) + " mana. Your adventure begins now! [SPACE] Begin",
	}
	for _, s := range destinyTexts {
		plain, colors := tui.ParseMarkup(s)
		box := tui.LayoutBottomBox(c, "DESTINY", plain)
		tw := tui.NewTypewriter(plain)
		step := 0
		finished := false
		for !finished {
			c.Clear()
			cine.DrawFrame(c, m, frame%len(m.Frames))
			frame++
			tui.DrawTextBox(c, box, colors, tui.VisibleLen(tw), "[SPACE] Next")
			if err := tui.FlushStyled(out, c); err != nil {
				break
			}
			select {
			case ev, ok := <-keys:
				if !ok || tui.IsQuit(ev) {
					finished = true
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
	return ch, nil
}

// chooseRace affiche le menu de choix de race (Humain, Elfe, Nain).
// ↑/↓ pour naviguer, ENTREE/ESPACE pour valider. EOF = 1re race.
// La classe n'est plus choisie ni tiree : elle decoule de la race.
func chooseRace(c *tui.Canvas, out io.Writer, keys <-chan tui.Event, m cine.Movie, frame *int) character.Class {
	races := spell.AllClasses
	cursor := 0
	bw, bh := 36, len(races)+5
	bx, by := tui.CenteredBox(c.W, c.H, bw, bh)

	for {
		c.Clear()
		if len(m.Frames) > 0 {
			cine.DrawFrame(c, m, (*frame)%len(m.Frames))
			(*frame)++
		}
		tui.FillStyled(c, bx, by, bw, bh, ' ', "", tui.BGBlack)
		tui.DrawBoxWithTitle(c, bx, by, bw, bh, "CHOOSE YOUR RACE")
		c.WriteStyled(bx+3, by+2, "Choose your race:", tui.FGBrightWhite, tui.BGBlack)
		for i, r := range races {
			prefix := "  [ ] "
			fg := tui.FGWhite
			if i == cursor {
				prefix = "> [X] "
				fg = tui.FGLightGreen
			}
			c.WriteStyled(bx+4, by+4+i, prefix+r.String(), fg, tui.BGBlack)
		}
		_ = tui.FlushStyled(out, c)

		select {
		case ev, ok := <-keys:
			if !ok {
				return races[0]
			}
			switch ev.K {
			case tui.KeyUp:
				cursor = (cursor - 1 + len(races)) % len(races)
			case tui.KeyDown:
				cursor = (cursor + 1) % len(races)
			case tui.KeyEnter:
				return races[cursor]
			case tui.KeyRune:
				if ev.R == ' ' {
					return races[cursor]
				}
			}
		case <-time.After(50 * time.Millisecond):
		}
	}
}

// chooseSubclass affiche le menu de choix de classe pour la race donnee.
// ↑/↓ pour naviguer, ENTREE/ESPACE pour valider. EOF ou aucune
// sous-classe = 1re de la liste (ou Any si la race n'en a pas).
func chooseSubclass(c *tui.Canvas, out io.Writer, keys <-chan tui.Event, m cine.Movie, frame *int, race character.Class) character.Subclass {
	subs := spell.SubclassesForClass(race)
	if len(subs) == 0 {
		return spell.SubclassAny
	}
	cursor := 0
	bw, bh := 36, len(subs)+5
	bx, by := tui.CenteredBox(c.W, c.H, bw, bh)

	for {
		c.Clear()
		if len(m.Frames) > 0 {
			cine.DrawFrame(c, m, (*frame)%len(m.Frames))
			(*frame)++
		}
		tui.FillStyled(c, bx, by, bw, bh, ' ', "", tui.BGBlack)
		tui.DrawBoxWithTitle(c, bx, by, bw, bh, "CHOOSE YOUR CLASS")
		c.WriteStyled(bx+3, by+2, "Choose your class:", tui.FGBrightWhite, tui.BGBlack)
		for i, s := range subs {
			prefix := "  [ ] "
			fg := tui.FGWhite
			if i == cursor {
				prefix = "> [X] "
				fg = tui.FGLightGreen
			}
			c.WriteStyled(bx+4, by+4+i, prefix+s.String(), fg, tui.BGBlack)
		}
		_ = tui.FlushStyled(out, c)

		select {
		case ev, ok := <-keys:
			if !ok {
				return subs[0]
			}
			switch ev.K {
			case tui.KeyUp:
				cursor = (cursor - 1 + len(subs)) % len(subs)
			case tui.KeyDown:
				cursor = (cursor + 1) % len(subs)
			case tui.KeyEnter:
				return subs[cursor]
			case tui.KeyRune:
				if ev.R == ' ' {
					return subs[cursor]
				}
			}
		case <-time.After(50 * time.Millisecond):
		}
	}
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
