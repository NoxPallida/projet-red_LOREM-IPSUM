package menu

import (
	"os"
	"strconv"

	"runa/src/internal/character"
	"runa/src/internal/item"
	"runa/src/internal/tui"
)

// equipLabels dans l'ordre d'affichage : les 4 armures puis l'arme.
var equipLabels = [5]string{"Head", "Chest", "Legs", "Feet", "Weapon"}

// RunInventoryMenu ouvre le menu inventaire du joueur par-dessus le rendu du jeu.
// En haut : les objets (potions a consommer, equipements a enfiler).
// En bas : les 5 slots d'equipement (casque, plastron, jambieres,
// bottes, arme), selectionnables pour desequiper.
// Sortie avec 'e', 'E', ECHAP ou 'q'.
func RunInventoryMenu(in *os.File, out *os.File, c *tui.Canvas, ch *character.Character, renderUnder func()) {
	cursor := 0
	msg := ""
	msgCol := tui.FGLightGreen

	bw, bh := 58, 22
	bx, by := tui.CenteredBox(c.W, c.H, bw, bh)
	scrollOffset := 0
	maxVisible := 6
	equipStart := 10 // premiere ligne des slots d'equipement (relative a by)

	for {
		renderUnder()
		tui.FillStyled(c, bx, by, bw, bh, ' ', "", tui.BGBlack)
		tui.DrawBoxWithTitle(c, bx, by, bw, bh, "INVENTORY")

		goldStr := "Gold: " + strconv.Itoa(int(ch.Money())) + " G"
		capStr := "Slots: " + strconv.Itoa(len(ch.Inventory.Slots)) + "/" + strconv.Itoa(int(ch.Inventory.Capacity))
		c.WriteStyled(bx+3, by+1, goldStr, tui.FGYellow, tui.BGBlack)
		c.WriteStyled(bx+bw-len(capStr)-3, by+1, capStr, tui.FGCyan, tui.BGBlack)

		slots := ch.Inventory.Slots
		total := len(slots) + len(equipLabels)
		if cursor >= total {
			cursor = total - 1
		}
		if cursor < 0 {
			cursor = 0
		}

		if len(slots) == 0 {
			c.WriteStyled(bx+4, by+4, "(Your inventory is empty)", tui.FGGray, tui.BGBlack)
		} else {
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
				c.WriteStyled(bx+3, by+bh-5, desc, tui.FGWhite, tui.BGBlack)
			}
		}

		// Slots d'equipement sous l'inventaire.
		c.WriteStyled(bx+3, by+equipStart-1, "--- Equipment ---", tui.FGGray, tui.BGBlack)
		for i, label := range equipLabels {
			rowY := by + equipStart + i
			name := "(empty)"
			if n := equippedName(ch, i); n != "" {
				name = n
			}
			prefix := "  "
			fg := tui.FGWhite
			if cursor-len(slots) == i {
				prefix = "> "
				fg = tui.FGBrightWhite
			}
			c.WriteStyled(bx+3, rowY, prefix+label+": "+name, fg, tui.BGBlack)
		}
		// Detail d'une piece equipee selectionnee.
		if eqIdx := cursor - len(slots); eqIdx >= 0 && eqIdx < len(equipLabels) {
			if desc := equippedDescription(ch, eqIdx); desc != "" {
				c.WriteStyled(bx+3, by+bh-5, desc, tui.FGWhite, tui.BGBlack)
			}
		}

		if msg != "" {
			c.WriteStyled(bx+3, by+bh-3, msg, msgCol, tui.BGBlack)
		}

		hint := "[↑/↓] Choose [ENTER] Action [C] Credits [E] Exit"
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
		case tui.KeyRune:
			if ev.R == 'c' || ev.R == 'C' {
				showCredits(in, out, c)
			}
		case tui.KeyUp:
			if cursor > 0 {
				cursor--
			}
		case tui.KeyDown:
			if cursor < len(ch.Inventory.Slots)+len(equipLabels)-1 {
				cursor++
			}
		case tui.KeyEnter:
			slots := ch.Inventory.Slots
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

				// 4. Armor (casque, plastron, jambieres, bottes)
				if e, ok := target.(item.Equipment); ok {
					ch.EquipArmor(e)
					msg = "Equipped: " + e.Name() + " (+" + strconv.Itoa(int(e.HPBonus)) + " Max HP)!"
					msgCol = tui.FGLightGreen
					continue
				}

				msg = "Item: " + target.Name() + " (Value: " + strconv.Itoa(int(target.PriceSell())) + " G)"
				msgCol = tui.FGLightCyan
			} else if eqIdx := cursor - len(slots); eqIdx >= 0 && eqIdx < len(equipLabels) {
				// Ligne d'equipement : desequipe vers l'inventaire.
				if eqIdx < 4 {
					if ch.UnequipArmor(item.EquipmentSlot(eqIdx)) {
						msg = "Unequipped " + equipLabels[eqIdx] + "."
						msgCol = tui.FGLightGreen
					} else {
						msg = "Nothing to unequip (or inventory full)."
						msgCol = tui.FGLightRed
					}
				} else {
					if ch.UnequipWeapon() {
						msg = "Unequipped weapon."
						msgCol = tui.FGLightGreen
					} else {
						msg = "Nothing to unequip (or inventory full)."
						msgCol = tui.FGLightRed
					}
				}
			}
		}
	}
}

// showCredits affiche la boite easter-egg des credits par-dessus
// l'inventaire. N'importe quelle touche la referme et rend la main.
func showCredits(in *os.File, out *os.File, c *tui.Canvas) {
	bw, bh := 44, 7
	bx, by := tui.CenteredBox(c.W, c.H, bw, bh)
	tui.FillStyled(c, bx, by, bw, bh, ' ', "", tui.BGBlack)
	tui.DrawBoxWithTitle(c, bx, by, bw, bh, "CREDITS")
	line := "credit : ABBA et Steven Spielberg"
	c.WriteStyled(bx+(bw-len(line))/2, by+2, line, tui.FGLightGreen, tui.BGBlack)
	hint := "[touche] Retour"
	c.WriteStyled(bx+(bw-len(hint))/2, by+4, hint, tui.FGGray, tui.BGBlack)
	_ = tui.FlushStyled(out, c)
	_, _ = tui.ReadKey(in)
}

// equippedName rend le nom porte sur la ligne idx ("" si vide).
// idx 0..3 = armures (Head/Chest/Legs/Feet), 4 = arme.
func equippedName(ch *character.Character, idx int) string {
	if idx < 4 {
		if e := ch.EquippedArmor[idx]; e != nil {
			return e.Name()
		}
		return ""
	}
	if ch.EquippedWeapon != nil {
		return ch.EquippedWeapon.Name()
	}
	return ""
}

// equippedDescription decrit la piece equipee selectionnee.
func equippedDescription(ch *character.Character, idx int) string {
	if idx < 4 {
		if e := ch.EquippedArmor[idx]; e != nil {
			return "Armor: +" + strconv.Itoa(int(e.HPBonus)) + " Max HP (" + equipLabels[idx] + ")."
		}
		return ""
	}
	if ch.EquippedWeapon != nil {
		return itemDescription(*ch.EquippedWeapon)
	}
	return ""
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
