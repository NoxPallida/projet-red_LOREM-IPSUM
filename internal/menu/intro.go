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
	"Listen closely, traveler. The village of LOREM is in danger.",
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

	// 3. Choix de la race et de la classe
	chosenClass := chooseRace(c, out, keys, m, &frame)
	chosenSubclass := chooseSubclass(c, out, keys, m, &frame, chosenClass)

	// 4. Dialogue d'attribution du mana
	destinyTexts := []string{
		"Welcome, " + name + " the " + chosenSubclass.String() + " (" + chosenClass.String() + ").",
		"Now, let us see the mana pool that destiny grants you...",
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

	// 5. Machine a sous (slot machine) pour le Mana uniquement
	ch := runSlotMachine(c, out, keys, m, &frame, name, chosenClass, chosenSubclass)
	return ch, nil
}

// chooseRace affiche une petite boîte pour choisir sa race (Humain, Elfe, Nain).
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

// chooseSubclass affiche une petite boîte pour choisir sa classe selon la race.
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

// runSlotMachine fait defiler une roulette (Mana) facon machine a sous.
// Sans defer ni module externe, base uniquement sur le tui existant.
func runSlotMachine(c *tui.Canvas, out io.Writer, keys <-chan tui.Event, m cine.Movie, frame *int, name string, chosenClass character.Class, chosenSubclass character.Subclass) *character.Character {
	chosenMana := character.Mana(chosenClass).Mana()
	if chosenMana == 0 {
		chosenMana = 25
	}

	manaPool := []string{
		"10", "25", "35", "50", "65", "80", "15", "42", "70", "30",
	}
	targetManaIdx := len(manaPool)
	manaPool = append(manaPool, strconv.Itoa(int(chosenMana)))

	delays := []time.Duration{
		40 * time.Millisecond, 40 * time.Millisecond, 40 * time.Millisecond, 50 * time.Millisecond,
		50 * time.Millisecond, 60 * time.Millisecond, 70 * time.Millisecond, 90 * time.Millisecond,
		120 * time.Millisecond, 160 * time.Millisecond, 220 * time.Millisecond, 300 * time.Millisecond,
		420 * time.Millisecond,
	}

	curManaIdx := (targetManaIdx - len(delays) + len(manaPool)*100) % len(manaPool)
	manaDone := false
	skip := false

	// Hauteur reduite de 1 (11 au lieu de 12)
	bw, bh := 42, 11
	bx, by := tui.CenteredBox(c.W, c.H, bw, bh)

	drawMachine := func() {
		c.Clear()
		if len(m.Frames) > 0 {
			cine.DrawFrame(c, m, (*frame)%len(m.Frames))
			(*frame)++
		}
		tui.FillStyled(c, bx, by, bw, bh, ' ', "", tui.BGBlack)
		tui.DrawBoxWithTitle(c, bx, by, bw, bh, "WHEEL OF DESTINY")

		titleText := "Mana Roll: " + name
		c.WriteStyled(bx+(bw-len(titleText))/2, by+2, titleText, tui.FGBrightWhite, tui.BGBlack)

		// Roulette MANA centree
		rw, rh := 20, 4
		rx := bx + (bw-rw)/2
		tui.DrawBoxWithTitle(c, rx, by+4, rw, rh, "MANA")
		mText := manaPool[curManaIdx%len(manaPool)]
		mCol := tui.FGWhite
		if manaDone {
			mCol = tui.FGLightCyan
		}
		c.WriteStyled(rx+(rw-len(mText)-4)/2, by+6, "> "+mText+" <", mCol, tui.BGBlack)

		// Statut / indice en bas (retire le 'Destin : ...' moche)
		if !manaDone {
			msg := "Rolling for mana..."
			c.WriteStyled(bx+(bw-len(msg))/2, by+9, msg, tui.FGYellow, tui.BGBlack)
		} else {
			hint := "[SPACE] Begin Adventure"
			c.WriteStyled(bx+(bw-len(hint))/2, by+9, hint, tui.FGBrightWhite, tui.BGBlack)
		}
	}

	// Tour de roulette pour le Mana
	for step := 0; step < len(delays) && !skip; step++ {
		curManaIdx = (targetManaIdx - (len(delays) - 1 - step) + len(manaPool)*100) % len(manaPool)
		drawMachine()
		_ = tui.FlushStyled(out, c)
		select {
		case ev, ok := <-keys:
			if !ok || tui.IsQuit(ev) {
				skip = true
			}
			if tui.IsConfirm(ev) {
				skip = true
			}
		case <-time.After(delays[step]):
		}
	}
	manaDone = true
	curManaIdx = targetManaIdx

	// Attente confirmation joueur
	for {
		drawMachine()
		_ = tui.FlushStyled(out, c)
		select {
		case ev, ok := <-keys:
			if !ok || tui.IsQuit(ev) || tui.IsConfirm(ev) {
				ch := character.InitCharacter(name, chosenClass)
				ch.SetSubclass(chosenSubclass)
				ch.Mana = chosenMana
				ch.ManaMax = chosenMana
				return ch
			}
		case <-time.After(100 * time.Millisecond):
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
