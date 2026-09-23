package menu

import (
	"os"
	"strconv"

	"runa/internal/character"
	"runa/internal/inventory"
	"runa/internal/tui"
)

// RunChestMenu ouvre le coffre par-dessus le rendu du jeu.
// Deux panneaux : sac du joueur (gauche) et coffre (droite).
// ↑/↓ navigue, ←/→ change de panneau, ENTER deplace 1 unite
// (depot sac->coffre, retrait coffre->sac). Sortie avec 'e', ESC ou 'q'.
// chest est un pointeur car le menu modifie son contenu.
func RunChestMenu(in, out *os.File, c *tui.Canvas, ch *character.Character, chest *inventory.Inventory, renderUnder func()) {
	active := 0 // 0 = sac joueur, 1 = coffre
	curs := [2]int{0, 0}
	offs := [2]int{0, 0}
	msg := ""
	msgCol := tui.FGLightGreen

	bw, bh := 76, 16
	bx, by := tui.CenteredBox(c.W, c.H, bw, bh)
	pw := 35
	px := [2]int{bx + 2, bx + 2 + pw + 2}
	maxVisible := 6

	slotsOf := func(panel int) []inventory.Slot {
		if panel == 0 {
			return ch.Inventory.Slots
		}
		return chest.Slots
	}

	for {
		renderUnder()
		tui.FillStyled(c, bx, by, bw, bh, ' ', "", tui.BGBlack)
		tui.DrawBoxWithTitle(c, bx, by, bw, bh, "COFFRE")

		for panel := 0; panel < 2; panel++ {
			title := "SAC (" + strconv.Itoa(len(ch.Inventory.Slots)) + "/" + strconv.Itoa(int(ch.Inventory.Capacity)) + ")"
			if panel == 1 {
				title = "COFFRE (" + strconv.Itoa(len(chest.Slots)) + "/" + strconv.Itoa(int(chest.Capacity)) + ")"
			}
			titleCol := tui.FGGray
			if panel == active {
				titleCol = tui.FGBrightWhite
			}
			c.WriteStyled(px[panel], by+1, title, titleCol, tui.BGBlack)

			slots := slotsOf(panel)
			cur := &curs[panel]
			if *cur >= len(slots) {
				*cur = len(slots) - 1
			}
			if *cur < 0 {
				*cur = 0
			}
			if *cur < offs[panel] {
				offs[panel] = *cur
			}
			if *cur >= offs[panel]+maxVisible {
				offs[panel] = *cur - maxVisible + 1
			}

			if len(slots) == 0 {
				c.WriteStyled(px[panel]+1, by+4, "(vide)", tui.FGGray, tui.BGBlack)
				continue
			}
			for vi := 0; vi < maxVisible; vi++ {
				idx := offs[panel] + vi
				if idx >= len(slots) {
					break
				}
				slot := slots[idx]
				rowY := by + 3 + vi

				prefix := "  "
				fg := tui.FGWhite
				if panel == active && idx == *cur {
					prefix = "> "
					fg = tui.FGBrightWhite
				}

				nameStr := prefix + slot.Item.Name()
				if len([]rune(nameStr)) > pw-8 {
					nameStr = string([]rune(nameStr)[:pw-9]) + "…"
				}
				qtyStr := "x" + strconv.Itoa(int(slot.Quantity))
				c.WriteStyled(px[panel], rowY, nameStr, fg, tui.BGBlack)
				c.WriteStyled(px[panel]+pw-len(qtyStr)-1, rowY, qtyStr, tui.FGLightYellow, tui.BGBlack)
			}
		}

		// Detail de la selection du panneau actif.
		if slots := slotsOf(active); curs[active] >= 0 && curs[active] < len(slots) {
			desc := itemDescription(slots[curs[active]].Item)
			c.WriteStyled(bx+3, by+bh-4, desc, tui.FGWhite, tui.BGBlack)
		}

		if msg != "" {
			c.WriteStyled(bx+3, by+bh-3, msg, msgCol, tui.BGBlack)
		}

		hint := "[↑/↓] Choisir [←/→] Sac/Coffre [ENTER] Déplacer x1 [E/ESC] Sortir"
		c.WriteStyled(bx+(bw-len(hint))/2, by+bh-2, hint, tui.FGBrightWhite, tui.BGBlack)

		_ = tui.FlushStyled(out, c)

		ev, err := tui.ReadKey(in)
		if err != nil {
			return
		}
		if tui.IsQuit(ev) || (ev.K == tui.KeyRune && (ev.R == 'e' || ev.R == 'E')) {
			return
		}
		switch ev.K {
		case tui.KeyUp:
			if curs[active] > 0 {
				curs[active]--
			}
		case tui.KeyDown:
			if curs[active] < len(slotsOf(active))-1 {
				curs[active]++
			}
		case tui.KeyLeft:
			active = 0
		case tui.KeyRight:
			active = 1
		case tui.KeyEnter:
			if active == 0 {
				msg, msgCol = depositOne(ch, chest, curs[0])
			} else {
				msg, msgCol = withdrawOne(ch, chest, curs[1])
			}
		}
	}
}

// depositOne deplace 1 unite du sac vers le coffre.
// Le coffre refuse avant tout retrait : rien n'est jamais perdu.
func depositOne(ch *character.Character, chest *inventory.Inventory, idx int) (string, string) {
	slots := ch.Inventory.Slots
	if idx < 0 || idx >= len(slots) {
		return "", tui.FGLightGreen
	}
	target := slots[idx].Item
	if err := chest.AddItem(target, 1); err != nil {
		return "Coffre plein !", tui.FGLightRed
	}
	_ = ch.Inventory.RemoveItem(target.Name(), 1)
	return "Déposé : " + target.Name(), tui.FGLightGreen
}

// withdrawOne deplace 1 unite du coffre vers le sac.
// Ordre anti-perte : on retire d'abord, et si le sac refuse,
// on repose aussitot dans le coffre (qui vient de liberer la place).
func withdrawOne(ch *character.Character, chest *inventory.Inventory, idx int) (string, string) {
	slots := chest.Slots
	if idx < 0 || idx >= len(slots) {
		return "", tui.FGLightGreen
	}
	target := slots[idx].Item
	if err := chest.RemoveItem(target.Name(), 1); err != nil {
		return "Retrait impossible.", tui.FGLightRed
	}
	if err := ch.Inventory.AddItem(target, 1); err != nil {
		_ = chest.AddItem(target, 1) // repose : la place vient d'etre liberee
		return "Sac plein !", tui.FGLightRed
	}
	return "Retiré : " + target.Name(), tui.FGLightGreen
}
