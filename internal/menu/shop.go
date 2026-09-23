package menu

import (
	"os"
	"strconv"

	"runa/internal/character"
	"runa/internal/item"
	"runa/internal/shop"
	"runa/internal/tui"
)

// RunShopMenu ouvre la boutique du marchand par-dessus le rendu intérieur
// (ne prend pas tout le canvas). Permet d'acheter et de vendre.
func RunShopMenu(in *os.File, out *os.File, c *tui.Canvas, ch *character.Character, renderUnder func()) {
	mode := 0 // 0 = Acheter, 1 = Vendre
	cursor := 0
	msg := ""
	msgCol := tui.FGLightGreen
	scrollBuy := 0 // la liste d'achat depasse la hauteur : on scrolle

	bw, bh := 54, 15
	bx, by := tui.CenteredBox(c.W, c.H, bw, bh)

	for {
		renderUnder()
		tui.FillStyled(c, bx, by, bw, bh, ' ', "", tui.BGBlack)
		tui.DrawBoxWithTitle(c, bx, by, bw, bh, "SHOP")

		goldStr := "Gold: " + strconv.Itoa(int(ch.Money())) + " G"
		c.WriteStyled(bx+bw-len(goldStr)-3, by+1, goldStr, tui.FGYellow, tui.BGBlack)

		tabBuy := "[ Buy ]"
		tabSell := "[ Sell ]"
		if mode == 0 {
			c.WriteStyled(bx+3, by+1, tabBuy, tui.FGLightGreen, tui.BGBlack)
			c.WriteStyled(bx+16, by+1, tabSell, tui.FGGray, tui.BGBlack)
		} else {
			c.WriteStyled(bx+3, by+1, tabBuy, tui.FGGray, tui.BGBlack)
			c.WriteStyled(bx+16, by+1, tabSell, tui.FGLightGreen, tui.BGBlack)
		}

		if mode == 0 {
			catalog := shop.Catalog
			// Derniere ligne : l'augmentation d'inventaire (pas un item,
			// donc hors catalogue, cf. shop.BuyInventoryUpgrade).
			buyLen := len(catalog) + 1
			if cursor >= buyLen {
				cursor = buyLen - 1
			}
			if cursor < 0 {
				cursor = 0
			}
			if cursor < scrollBuy {
				scrollBuy = cursor
			}
			maxRows := by + bh - 3 - (by + 3)
			if cursor >= scrollBuy+maxRows {
				scrollBuy = cursor - maxRows + 1
			}
			for vi := 0; vi < maxRows; vi++ {
				i := scrollBuy + vi
				if i >= buyLen {
					break
				}
				rowY := by + 3 + vi
				prefix := "  "
				fg := tui.FGWhite
				if i == cursor {
					prefix = "> "
					fg = tui.FGLightCyan
				}
				if i < len(catalog) {
					entry := catalog[i]
					priceStr := strconv.Itoa(int(entry.Price)) + " G"
					if entry.Item.Name() == item.HealPotion.Name() && !ch.HasClaimedFreePotion() {
						priceStr = "FREE"
					}
					c.WriteStyled(bx+3, rowY, prefix+entry.Item.Name(), fg, tui.BGBlack)
					c.WriteStyled(bx+bw-len(priceStr)-3, rowY, priceStr, tui.FGYellow, tui.BGBlack)
				} else {
					priceStr := strconv.Itoa(int(shop.InventoryUpgradePrice)) + " G"
					if !ch.CanUpgradeInventory() {
						priceStr = "MAX"
					}
					c.WriteStyled(bx+3, rowY, prefix+"Inventory Upgrade", fg, tui.BGBlack)
					c.WriteStyled(bx+bw-len(priceStr)-3, rowY, priceStr, tui.FGYellow, tui.BGBlack)
				}
			}
		} else {
			slots := ch.Inventory.Slots
			if len(slots) == 0 {
				c.WriteStyled(bx+4, by+5, "(Inventory empty)", tui.FGGray, tui.BGBlack)
			} else {
				if cursor >= len(slots) {
					cursor = len(slots) - 1
				}
				if cursor < 0 {
					cursor = 0
				}
				for i, slot := range slots {
					rowY := by + 3 + i
					if rowY >= by+bh-3 {
						break
					}
					sellPrice := slot.Item.PriceSell()
					priceStr := strconv.Itoa(int(sellPrice)) + " G"
					qtyStr := "x" + strconv.Itoa(int(slot.Quantity))
					line := slot.Item.Name() + " " + qtyStr
					prefix := "  "
					fg := tui.FGWhite
					if i == cursor {
						prefix = "> "
						fg = tui.FGLightCyan
					}
					c.WriteStyled(bx+3, rowY, prefix+line, fg, tui.BGBlack)
					c.WriteStyled(bx+bw-len(priceStr)-3, rowY, priceStr, tui.FGYellow, tui.BGBlack)
				}
			}
		}

		if msg != "" {
			c.WriteStyled(bx+3, by+bh-3, msg, msgCol, tui.BGBlack)
		}

		hint := "[↑/↓] Choose [ENTER] Action [TAB] Mode [ESC] Exit"
		c.WriteStyled(bx+(bw-len(hint))/2, by+bh-2, hint, tui.FGBrightWhite, tui.BGBlack)

		_ = tui.FlushStyled(out, c)

		ev, err := tui.ReadKey(in)
		if err != nil {
			return
		}
		if ev.K == tui.KeyEsc || (ev.K == tui.KeyRune && (ev.R == 'q' || ev.R == 'Q')) {
			return
		}
		switch ev.K {
		case tui.KeyUp:
			if cursor > 0 {
				cursor--
			}
		case tui.KeyDown:
			maxLen := len(shop.Catalog) + 1 // +1 : ligne upgrade d'inventaire
			if mode == 1 {
				maxLen = len(ch.Inventory.Slots)
			}
			if cursor < maxLen-1 {
				cursor++
			}
		case tui.KeyRune:
			if ev.R == '\t' {
				mode = 1 - mode
				cursor = 0
				msg = ""
			}
		case tui.KeyEnter:
			if mode == 0 {
				if cursor >= 0 && cursor < len(shop.Catalog) {
					entry := shop.Catalog[cursor]
					res, name := shop.Buy(ch, entry.Item)
					switch res {
					case shop.Success:
						msg = "Bought: " + name + "!"
						msgCol = tui.FGLightGreen
					case shop.ErrInsufficientMoney:
						msg = "Not enough gold!"
						msgCol = tui.FGLightRed
					case shop.ErrInventoryFull:
						msg = "Inventory full!"
						msgCol = tui.FGLightRed
					default:
						msg = "Cannot buy item"
						msgCol = tui.FGLightRed
					}
				} else if cursor == len(shop.Catalog) {
					switch shop.BuyInventoryUpgrade(ch) {
					case shop.Success:
						msg = "Inventory upgraded!"
						msgCol = tui.FGLightGreen
					case shop.ErrUpgradeMaxed:
						msg = "Upgrade maxed out!"
						msgCol = tui.FGLightRed
					case shop.ErrInsufficientMoney:
						msg = "Not enough gold!"
						msgCol = tui.FGLightRed
					default:
						msg = "Cannot upgrade"
						msgCol = tui.FGLightRed
					}
				}
			} else {
				if cursor >= 0 && cursor < len(ch.Inventory.Slots) {
					target := ch.Inventory.Slots[cursor].Item
					res, name := shop.Sell(ch, target)
					if res == shop.Success {
						msg = "Sold: " + name + "!"
						msgCol = tui.FGLightGreen
					} else {
						msg = "Cannot sell item"
						msgCol = tui.FGLightRed
					}
				}
			}
		}
	}
}
