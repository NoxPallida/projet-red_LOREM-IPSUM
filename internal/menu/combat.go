package menu

import (
	"os"
	"strconv"

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

func drawCombat(c *tui.Canvas, cb *combat.Combat, screen combatScreen, cursor int) {
	tui.DrawBoxWithTitle(c, 0, 0, c.W, c.H, "Combat")

	// Ligne joueur : nom, HP, mana.
	c.Write(2, 2, cb.Player.Name)
	hp := "HP: " + strconv.Itoa(int(cb.Player.Hp)) + "/" + strconv.Itoa(int(cb.Player.HpMax))
	c.WriteStyled(2, 3, hp, tui.FGLightRed, "")
	mana := "Mana: " + strconv.Itoa(int(cb.Player.Mana)) + "/" + strconv.Itoa(int(cb.Player.ManaMax))
	c.WriteStyled(20, 3, mana, tui.FGCyan, "")

	// Ligne ennemi.
	c.Write(2, 5, cb.Enemy.Template.Name+" (niv. "+strconv.Itoa(int(cb.Enemy.Level))+")")
	ehp := "HP: " + strconv.Itoa(int(cb.Enemy.HP)) + "/" + strconv.Itoa(int(cb.Enemy.MaxHP))
	c.WriteStyled(2, 6, ehp, tui.FGLightRed, "")

	// Menu / sous-menu.
	y := 9
	switch screen {
	case screenMain:
		drawMenuList(c, 2, y, []string{"Attaque", "Objet", "Fuir"}, cursor)
	case screenAttack:
		opts := combat.AttackOptions(cb.Player)
		labels := make([]string, len(opts))
		for i, o := range opts {
			labels[i] = o.Name + " (" + strconv.Itoa(int(o.Damage)) + " dgts, " + strconv.Itoa(int(o.ManaCost)) + " mana)"
		}
		drawMenuList(c, 2, y, labels, cursor)
	case screenItem:
		items := combat.AvailableConsumables(cb.Player)
		if len(items) == 0 {
			c.Write(2, y, "Aucun objet. (Entrée pour revenir)")
		} else {
			labels := make([]string, len(items))
			for i, it := range items {
				labels[i] = it.Name()
			}
			drawMenuList(c, 2, y, labels, cursor)
		}
	}

	// Journal de combat : les 4 derniers messages.
	logY := c.H - 6
	start := 0
	if len(cb.Log) > 4 {
		start = len(cb.Log) - 4
	}
	for i, msg := range cb.Log[start:] {
		c.Write(2, logY+i, msg)
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
