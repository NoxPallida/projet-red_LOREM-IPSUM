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

	bw, bh := 54, 15
	bx, by := tui.CenteredBox(c.W, c.H, bw, bh)

	for {
		renderUnder()
		tui.FillStyled(c, bx, by, bw, bh, ' ', "", tui.BGBlack)
		tui.DrawBoxWithTitle(c, bx, by, bw, bh, "MARCHAND")

		goldStr := "Or : " + strconv.Itoa(int(ch.Money())) + " PO"
		c.WriteStyled(bx+bw-len(goldStr)-3, by+1, goldStr, tui.FGYellow, tui.BGBlack)

		tabBuy := "[ Acheter ]"
		tabSell := "[ Vendre ]"
		if mode == 0 {
			c.WriteStyled(bx+3, by+1, tabBuy, tui.FGLightGreen, tui.BGBlack)
			c.WriteStyled(bx+16, by+1, tabSell, tui.FGGray, tui.BGBlack)
		} else {
			c.WriteStyled(bx+3, by+1, tabBuy, tui.FGGray, tui.BGBlack)
			c.WriteStyled(bx+16, by+1, tabSell, tui.FGLightGreen, tui.BGBlack)
		}

		if mode == 0 {
			catalog := shop.Catalog
			if cursor >= len(catalog) {
				cursor = len(catalog) - 1
			}
			if cursor < 0 {
				cursor = 0
			}
			for i, entry := range catalog {
				rowY := by + 3 + i
				if rowY >= by+bh-3 {
					break
				}
				priceStr := strconv.Itoa(int(entry.Price)) + " PO"
				if entry.Item.Name() == item.HealPotion.Name() && !ch.HasClaimedFreePotion() {
					priceStr = "GRATUIT"
				}
				line := entry.Item.Name()
				prefix := "  "
				fg := tui.FGWhite
				if i == cursor {
					prefix = "> "
					fg = tui.FGLightCyan
				}
				c.WriteStyled(bx+3, rowY, prefix+line, fg, tui.BGBlack)
				c.WriteStyled(bx+bw-len(priceStr)-3, rowY, priceStr, tui.FGYellow, tui.BGBlack)
			}
		} else {
			slots := ch.Inventory.Slots
			if len(slots) == 0 {
				c.WriteStyled(bx+4, by+5, "(Inventaire vide)", tui.FGGray, tui.BGBlack)
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
					priceStr := strconv.Itoa(int(sellPrice)) + " PO"
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

		hint := "[↑/↓] Choisir [ENTRÉE] Action [TAB] Mode [ECHAP] Sortir"
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
			maxLen := len(shop.Catalog)
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
						msg = "Acheté : " + name + " !"
						msgCol = tui.FGLightGreen
					case shop.ErrInsufficientMoney:
						msg = "Pas assez d'argent !"
						msgCol = tui.FGLightRed
					case shop.ErrInventoryFull:
						msg = "Inventaire plein !"
						msgCol = tui.FGLightRed
					default:
						msg = "Achat impossible"
						msgCol = tui.FGLightRed
					}
				}
			} else {
				if cursor >= 0 && cursor < len(ch.Inventory.Slots) {
					target := ch.Inventory.Slots[cursor].Item
					res, name := shop.Sell(ch, target)
					if res == shop.Success {
						msg = "Vendu : " + name + " !"
						msgCol = tui.FGLightGreen
					} else {
						msg = "Vente impossible"
						msgCol = tui.FGLightRed
					}
				}
			}
		}
	}
}
