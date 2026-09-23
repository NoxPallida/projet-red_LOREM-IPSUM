package menu

import (
	"os"
	"strconv"

	"runa/assets"
	"runa/internal/character"
	"runa/internal/combat"
	"runa/internal/enemies"
	"runa/internal/tui"
)

type combatScreen int

const (
	screenMain combatScreen = iota
	screenAttack
	screenItem
)

// RunCombat prend la main sur l'écran jusqu'à la fin naturelle du combat
// (victoire, défaite, ou fuite). Réutilise le canvas déjà alloué par
// RunGame plutôt que d'en créer un nouveau. Passe par tui.ReadKey
// (lecture bloquante, séquences ANSI classiques) plutôt que par le
// mécanisme kitty/PollKey de la boucle de déplacement : le combat est
// modal, tour par tour, pas besoin de répétition de touche tenue.
//
// wasKitty doit refléter l'état kitty de l'appelant : si actif, on le
// désactive le temps du combat (le protocole kitty encode les touches
// différemment, ReadKey ne les comprendrait pas), puis on le réactive
// à la sortie pour rendre la main à RunGame dans le même état.
func RunCombat(in *os.File, out *os.File, c *tui.Canvas, ch *character.Character, enemy *enemies.EnemyInstance, wasKitty bool) *combat.Combat {
	if wasKitty {
		tui.PopKitty(out)
		defer tui.PushKitty(out)
	}

	cb := combat.NewCombat(ch, enemy)
	screen := screenMain
	cursor := 0

	draw := func() {
		c.Clear()
		drawCombat(c, cb, screen, cursor)
		_ = tui.FlushStyled(out, c)
	}
	draw()

	for !cb.Over {
		ev, err := tui.ReadKey(in)
		if err != nil {
			cb.Flee() // entrée fermée en plein combat : sortie propre
			break
		}
		handleCombatInput(cb, ev, &screen, &cursor)
		draw()
	}
	return cb
}

func handleCombatInput(cb *combat.Combat, ev tui.Event, screen *combatScreen, cursor *int) {
	switch *screen {
	case screenMain:
		options := 3 // Attaque / Objet / Fuir
		switch ev.K {
		case tui.KeyUp:
			*cursor = (*cursor - 1 + options) % options
		case tui.KeyDown:
			*cursor = (*cursor + 1) % options
		case tui.KeyEnter:
			switch *cursor {
			case 0:
				*screen = screenAttack
				*cursor = 0
			case 1:
				*screen = screenItem
				*cursor = 0
			case 2:
				cb.Flee()
			}
		}

	case screenAttack:
		opts := combat.AttackOptions(cb.Player)
		switch ev.K {
		case tui.KeyUp:
			*cursor = (*cursor - 1 + len(opts)) % len(opts)
		case tui.KeyDown:
			*cursor = (*cursor + 1) % len(opts)
		case tui.KeyEnter:
			cb.PlayerAttack(opts[*cursor].SpellID)
			*screen = screenMain
			*cursor = 0
		case tui.KeyEsc, tui.KeyBackspace:
			*screen = screenMain
			*cursor = 0
		}

	case screenItem:
		items := combat.AvailableConsumables(cb.Player)
		if len(items) == 0 {
			if ev.K == tui.KeyEsc || ev.K == tui.KeyBackspace || ev.K == tui.KeyEnter {
				*screen = screenMain
				*cursor = 0
			}
			return
		}
		switch ev.K {
		case tui.KeyUp:
			*cursor = (*cursor - 1 + len(items)) % len(items)
		case tui.KeyDown:
			*cursor = (*cursor + 1) % len(items)
		case tui.KeyEnter:
			cb.UseItem(items[*cursor])
			*screen = screenMain
			*cursor = 0
		case tui.KeyEsc, tui.KeyBackspace:
			*screen = screenMain
			*cursor = 0
		}
	}
}

// heroArtFor rend l'art du joueur : toujours hero.txt,
// quelle que soit la race. Un seul sprite pour tous les heros.
func heroArtFor(race string) []string {
	return assets.Art("hero")
}

// drawArt dessine un ascii-art remis a l'echelle pour tenir dans
// (maxW x maxH) depuis (x, y), centre dans la zone.
// Si trop grand, on echantillonne (1 rune sur n) au lieu de couper :
// le sprite reste entier et lisible. Retourne la hauteur dessinee.
func drawArt(c *tui.Canvas, lines []string, x, y, maxW, maxH int, fg string) int {
	if maxW <= 0 || maxH <= 0 || len(lines) == 0 {
		return 0
	}
	w := 0
	for _, ln := range lines {
		if n := len([]rune(ln)); n > w {
			w = n
		}
	}
	stepX, stepY := 1, 1
	for w/stepX > maxW {
		stepX++
	}
	for len(lines)/stepY > maxH {
		stepY++
	}
	sw := (w + stepX - 1) / stepX
	sh := (len(lines) + stepY - 1) / stepY
	ox := x + (maxW-sw)/2
	if ox < x {
		ox = x
	}
	drawn := 0
	for sy := 0; sy*stepY < len(lines) && drawn < maxH && drawn < sh; sy++ {
		runes := []rune(lines[sy*stepY])
		col := 0
		for i, r := range runes {
			if i%stepX != 0 {
				continue
			}
			if col >= maxW {
				break
			}
			c.SetStyled(ox+col, y+drawn, r, fg, "")
			col++
		}
		drawn++
	}
	return drawn
}

// artSize rend la taille finale d'un art remis a l'echelle dans
// (maxW x maxH), pour le positionner avant de le dessiner
// (ex : caler le heros en bas de sa zone).
func artSize(lines []string, maxW, maxH int) (sw, sh int) {
	if maxW <= 0 || maxH <= 0 || len(lines) == 0 {
		return 0, 0
	}
	w := 0
	for _, ln := range lines {
		if n := len([]rune(ln)); n > w {
			w = n
		}
	}
	stepX, stepY := 1, 1
	for w/stepX > maxW {
		stepX++
	}
	for len(lines)/stepY > maxH {
		stepY++
	}
	sw = (w + stepX - 1) / stepX
	sh = (len(lines) + stepY - 1) / stepY
	if sw > maxW {
		sw = maxW
	}
	if sh > maxH {
		sh = maxH
	}
	return sw, sh
}

// menuLabels rend les lignes du menu selon l'ecran (sans les dessiner).
func menuLabels(cb *combat.Combat, screen combatScreen) []string {
	switch screen {
	case screenAttack:
		opts := combat.AttackOptions(cb.Player)
		labels := make([]string, len(opts))
		for i, o := range opts {
			labels[i] = o.Name + " (" + strconv.Itoa(int(o.Damage)) + " dmg, " + strconv.Itoa(int(o.ManaCost)) + " mana)"
		}
		return labels
	case screenItem:
		items := combat.AvailableConsumables(cb.Player)
		if len(items) == 0 {
			return []string{"No items. (Press Enter to go back)"}
		}
		labels := make([]string, len(items))
		for i, it := range items {
			labels[i] = it.Name()
		}
		return labels
	default:
		return []string{"Attack", "Item", "Flee"}
	}
}

func drawCombat(c *tui.Canvas, cb *combat.Combat, screen combatScreen, cursor int) {
	tui.DrawBoxWithTitle(c, 0, 0, c.W, c.H, "COMBAT")

	// Disposition facon Pokemon : ennemi en haut a droite, heros en
	// bas a gauche (Y decales), menu en boite en bas a gauche,
	// journal a droite du menu. Les arts sont remis a l'echelle
	// pour tenir dans leur zone, jamais coupes a l'arrache.
	labels := menuLabels(cb, screen)

	// Boite d'action en bas a gauche.
	menuW := 40
	if menuW > c.W-2 {
		menuW = c.W - 2
	}
	if menuW < 10 {
		menuW = 10
	}
	menuH := len(labels) + 2
	menuTop := c.H - menuH
	if menuTop < 2 {
		menuTop = 2
	}
	tui.DrawBoxWithTitle(c, 2, menuTop, menuW, menuH, "ACTION")
	drawMenuList(c, 4, menuTop+1, labels, cursor)

	// Ennemi : nom en haut a droite, art en dessous, PV dessous.
	mid := c.W / 2
	enemyMaxW := c.W - 2 - mid
	if enemyMaxW < 0 {
		enemyMaxW = 0
	}
	enemyMaxH := c.H - 9
	if enemyMaxH < 0 {
		enemyMaxH = 0
	}
	enemyName := cb.Enemy.Template.Name + " (Lvl " + strconv.Itoa(int(cb.Enemy.Level)) + ")"
	c.Write(c.W-2-len([]rune(enemyName)), 1, enemyName)
	enemyH := drawArt(c, assets.Art(cb.Enemy.Template.ID), mid, 2, enemyMaxW, enemyMaxH, tui.FGLightRed)
	ehp := "HP: " + strconv.Itoa(int(cb.Enemy.HP)) + "/" + strconv.Itoa(int(cb.Enemy.MaxHP))
	c.WriteStyled(c.W-2-len([]rune(ehp)), 2+enemyH, ehp, tui.FGLightRed, "")

	// Heros : art cale en bas a gauche (au-dessus du menu),
	// nom au-dessus de l'art, PV/mana en dessous.
	heroMaxW := mid - 4
	if heroMaxW < 0 {
		heroMaxW = 0
	}
	heroBottom := menuTop - 1
	heroMaxH := heroBottom - 5
	if heroMaxH < 0 {
		heroMaxH = 0
	}
	heroLines := heroArtFor(cb.Player.Class.String())
	_, heroH := artSize(heroLines, heroMaxW, heroMaxH)
	heroY := heroBottom - heroH
	drawArt(c, heroLines, 2, heroY, heroMaxW, heroMaxH, tui.FGCyan)
	c.Write(2, heroY-1, cb.Player.Name)
	hp := "HP: " + strconv.Itoa(int(cb.Player.Hp)) + "/" + strconv.Itoa(int(cb.Player.HpMax))
	c.WriteStyled(2, heroBottom, hp, tui.FGLightRed, "")
	mana := "Mana: " + strconv.Itoa(int(cb.Player.Mana)) + "/" + strconv.Itoa(int(cb.Player.ManaMax))
	c.WriteStyled(20, heroBottom, mana, tui.FGCyan, "")

	// Journal de combat : les 4 derniers messages, a droite du menu.
	logX := 2 + menuW + 2
	if c.W-2-logX >= 10 {
		logY := c.H - 5
		start := 0
		if len(cb.Log) > 4 {
			start = len(cb.Log) - 4
		}
		for i, msg := range cb.Log[start:] {
			c.Write(logX, logY+i, msg)
		}
	}
}

func drawMenuList(c *tui.Canvas, x, y int, labels []string, cursor int) {
	for i, label := range labels {
		prefix := "  "
		if i == cursor {
			prefix = "> "
		}
		c.Write(x, y+i, prefix+label)
	}
}
