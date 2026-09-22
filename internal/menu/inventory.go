package menu

import (
	"os"
	"strconv"

	"runa/internal/character"
	"runa/internal/item"
	"runa/internal/tui"
)

// RunInventoryMenu ouvre le menu inventaire du joueur par-dessus le rendu du jeu.
// Permet de voir ses objets, leurs quantités, de consommer des potions ou
// d'équiper des armes. Sortie avec 'e', 'E', ECHAP ou 'q'.
func RunInventoryMenu(in *os.File, out *os.File, c *tui.Canvas, ch *character.Character, renderUnder func()) {
	cursor := 0
	msg := ""
	msgCol := tui.FGLightGreen

	bw, bh := 58, 16
	bx, by := tui.CenteredBox(c.W, c.H, bw, bh)
	scrollOffset := 0
	maxVisible := 6

	for {
		renderUnder()
		tui.FillStyled(c, bx, by, bw, bh, ' ', "", tui.BGBlack)
		tui.DrawBoxWithTitle(c, bx, by, bw, bh, "INVENTORY")

		goldStr := "Gold: " + strconv.Itoa(int(ch.Money())) + " G"
		capStr := "Slots: " + strconv.Itoa(len(ch.Inventory.Slots)) + "/" + strconv.Itoa(int(ch.Inventory.Capacity))
		c.WriteStyled(bx+3, by+1, goldStr, tui.FGYellow, tui.BGBlack)
		c.WriteStyled(bx+bw-len(capStr)-3, by+1, capStr, tui.FGCyan, tui.BGBlack)

		slots := ch.Inventory.Slots
		if len(slots) == 0 {
			c.WriteStyled(bx+4, by+4, "(Your inventory is empty)", tui.FGGray, tui.BGBlack)
		} else {
			if cursor >= len(slots) {
				cursor = len(slots) - 1
			}
			if cursor < 0 {
				cursor = 0
			}

			if cursor < scrollOffset {
				scrollOffset = cursor
			}
			if cursor >= scrollOffset+maxVisible {
				scrollOffset = cursor - maxVisible + 1
			}

			for vi := 0; vi < maxVisible; vi++ {
				idx := scrollOffset + vi
				if idx >= len(slots) {
					break
				}
				slot := slots[idx]
				rowY := by + 3 + vi

				prefix := "  "
				fg := tui.FGWhite
				if idx == cursor {
					prefix = "> "
					fg = tui.FGBrightWhite
				}

				nameStr := prefix + slot.Item.Name()
				qtyStr := "x" + strconv.Itoa(int(slot.Quantity))
				typeStr := itemTypeTag(slot.Item)

				c.WriteStyled(bx+3, rowY, nameStr, fg, tui.BGBlack)
				c.WriteStyled(bx+28, rowY, qtyStr, tui.FGLightYellow, tui.BGBlack)
				c.WriteStyled(bx+bw-len(typeStr)-3, rowY, typeStr, tui.FGLightCyan, tui.BGBlack)
			}

			// Detail of selected item
			if cursor >= 0 && cursor < len(slots) {
				curItem := slots[cursor].Item
				desc := itemDescription(curItem)
				c.WriteStyled(bx+3, by+bh-4, desc, tui.FGWhite, tui.BGBlack)
			}
		}

		if msg != "" {
			c.WriteStyled(bx+3, by+bh-3, msg, msgCol, tui.BGBlack)
		}

		hint := "[↑/↓] Choose [ENTER] Use/Equip [E/ESC] Exit"
		c.WriteStyled(bx+(bw-len(hint))/2, by+bh-2, hint, tui.FGBrightWhite, tui.BGBlack)

		_ = tui.FlushStyled(out, c)

		ev, err := tui.ReadKey(in)
		if err != nil {
			return
		}
		if ev.K == tui.KeyEsc || (ev.K == tui.KeyRune && (ev.R == 'q' || ev.R == 'Q' || ev.R == 'e' || ev.R == 'E')) {
			return
		}
		switch ev.K {
		case tui.KeyUp:
			if cursor > 0 {
				cursor--
			}
		case tui.KeyDown:
			if cursor < len(slots)-1 {
				cursor++
			}
		case tui.KeyEnter:
			if cursor >= 0 && cursor < len(slots) {
				target := slots[cursor].Item

				// 1. Spellbook
				if unlock, ok := item.SpellbookUnlock(target); ok {
					if ch.Level() < unlock.RequiredLevel {
						msg = "Level too low (Level " + strconv.Itoa(int(unlock.RequiredLevel)) + " required)."
						msgCol = tui.FGLightRed
					} else if !ch.LearnSpell(unlock.SpellID) {
						msg = "Spell already learned!"
						msgCol = tui.FGYellow
					} else {
						_ = ch.Inventory.RemoveItem(target.Name(), 1)
						msg = "Spell learned: " + unlock.SpellID + "!"
						msgCol = tui.FGLightGreen
					}
					continue
				}

				// 2. Consumable effect (potion, etc.)
				effects := item.EffectsOf(target)
				if len(effects) > 0 {
					healed := false
					for _, eff := range effects {
						if eff.Type == item.EffectHeal {
							heal := uint16(eff.Amount)
							ch.Hp += heal
							if ch.Hp > ch.HpMax {
								ch.Hp = ch.HpMax
							}
							healed = true
							msg = "You drink " + target.Name() + " (+" + strconv.Itoa(int(heal)) + " HP). HP: " + strconv.Itoa(int(ch.Hp)) + "/" + strconv.Itoa(int(ch.HpMax))
							msgCol = tui.FGLightGreen
						}
					}
					if healed {
						_ = ch.Inventory.RemoveItem(target.Name(), 1)
					} else {
						msg = "Item cannot be used directly."
						msgCol = tui.FGGray
					}
					continue
				}

				// 3. Weapon
				if w, ok := target.(item.Weapon); ok {
					ch.EquipWeapon(w)
					msg = "Equipped weapon: " + w.Name() + " (+" + strconv.Itoa(int(w.Damage)) + " ATK)!"
					msgCol = tui.FGLightGreen
					continue
				}

				msg = "Item: " + target.Name() + " (Value: " + strconv.Itoa(int(target.PriceSell())) + " G)"
				msgCol = tui.FGLightCyan
			}
		}
	}
}

func itemTypeTag(it item.Item) string {
	switch it.Type() {
	case item.TypeConsumable:
		return "[Potion/Book]"
	case item.TypeWeapon:
		return "[Weapon]"
	case item.TypeEquipment:
		return "[Equipment]"
	case item.TypeLoot:
		return "[Material]"
	default:
		return "[Item]"
	}
}

func itemDescription(it item.Item) string {
	switch it.Type() {
	case item.TypeConsumable:
		if _, ok := item.SpellbookUnlock(it); ok {
			return "Spellbook: allows learning new magic."
		}
		for _, eff := range item.EffectsOf(it) {
			if eff.Type == item.EffectHeal {
				return "Restores " + strconv.Itoa(int(eff.Amount)) + " hit points."
			}
		}
		return "Consumable item."
	case item.TypeWeapon:
		if w, ok := it.(item.Weapon); ok {
			return "Weapon: deals +" + strconv.Itoa(int(w.Damage)) + " physical damage."
		}
		return "Combat weapon."
	case item.TypeLoot:
		return "Monster material. Sell value: " + strconv.Itoa(int(it.PriceSell())) + " G."
	default:
		return "Adventurer item."
	}
}
