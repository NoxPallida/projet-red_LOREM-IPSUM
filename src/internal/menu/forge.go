package menu

import (
	"os"
	"strconv"

	"runa/src/internal/character"
	"runa/src/internal/forge"
	"runa/src/internal/tui"
)

// RunForgeMenu ouvre la forge par-dessus le rendu intérieur
// (ne prend pas tout le canvas). Permet de fabriquer les
// équipements des recettes avec matériaux + pièces.
func RunForgeMenu(in *os.File, out *os.File, c *tui.Canvas, ch *character.Character, renderUnder func()) {
	cursor := 0
	msg := ""
	msgCol := tui.FGLightGreen
	scroll := 0 // la liste depasse la hauteur : on scrolle (cf. shop)

	bw, bh := 54, 15
	bx, by := tui.CenteredBox(c.W, c.H, bw, bh)

	for {
		renderUnder()
		tui.FillStyled(c, bx, by, bw, bh, ' ', "", tui.BGBlack)
		tui.DrawBoxWithTitle(c, bx, by, bw, bh, "FORGE")

		goldStr := "Gold: " + strconv.Itoa(int(ch.Money())) + " G"
		c.WriteStyled(bx+bw-len(goldStr)-3, by+1, goldStr, tui.FGYellow, tui.BGBlack)

		recipes := forge.Recipes
		if cursor >= len(recipes) {
			cursor = len(recipes) - 1
		}
		if cursor < 0 {
			cursor = 0
		}
		if cursor < scroll {
			scroll = cursor
		}
		// Lignes dispo : de by+3 jusqu'a by+bh-5 exclu (detail+msg+hint en bas).
		maxRows := by + bh - 5 - (by + 3)
		if maxRows < 1 {
			maxRows = 1
		}
		if cursor >= scroll+maxRows {
			scroll = cursor - maxRows + 1
		}
		for vi := 0; vi < maxRows; vi++ {
			i := scroll + vi
			if i >= len(recipes) {
				break
			}
			recipe := recipes[i]
			rowY := by + 3 + vi
			priceStr := strconv.Itoa(int(recipe.Price)) + " G"
			line := recipe.Result.Name()
			prefix := "  "
			fg := tui.FGWhite
			if i == cursor {
				prefix = "> "
				fg = tui.FGLightCyan
			}
			c.WriteStyled(bx+3, rowY, prefix+line, fg, tui.BGBlack)
			c.WriteStyled(bx+bw-len(priceStr)-3, rowY, priceStr, tui.FGYellow, tui.BGBlack)
		}

		// Détail de la recette sélectionnée : matériaux possédés / requis.
		if cursor >= 0 && cursor < len(recipes) {
			c.WriteStyled(bx+3, by+bh-5, recipeDetail(ch, recipes[cursor]), tui.FGWhite, tui.BGBlack)
		}

		if msg != "" {
			c.WriteStyled(bx+3, by+bh-3, msg, msgCol, tui.BGBlack)
		}

		hint := "[↑/↓] Choose [ENTER] Craft [ESC] Exit"
		c.WriteStyled(bx+(bw-len(hint))/2, by+bh-2, hint, tui.FGBrightWhite, tui.BGBlack)

		_ = tui.FlushStyled(out, c)

		ev, err := tui.ReadKey(in)
		if err != nil {
			return
		}
		if tui.IsQuit(ev) {
			return
		}
		switch ev.K {
		case tui.KeyUp:
			if cursor > 0 {
				cursor--
			}
		case tui.KeyDown:
			if cursor < len(recipes)-1 {
				cursor++
			}
		case tui.KeyEnter:
			if cursor >= 0 && cursor < len(recipes) {
				res, name := forge.Craft(ch, recipes[cursor].Result)
				switch res {
				case forge.Success:
					msg = "Forged: " + name + "!"
					msgCol = tui.FGLightGreen
				case forge.ErrMissingMaterials:
					msg = "Missing materials!"
					msgCol = tui.FGLightRed
				case forge.ErrInsufficientMoney:
					msg = "Not enough gold!"
					msgCol = tui.FGLightRed
				case forge.ErrInventoryFull:
					msg = "Inventory full!"
					msgCol = tui.FGLightRed
				default:
					msg = "Cannot craft item"
					msgCol = tui.FGLightRed
				}
			}
		}
	}
}

// recipeDetail affiche "nom: possédé/requis ..." pour chaque ingrédient.
func recipeDetail(ch *character.Character, r forge.Recipe) string {
	s := ""
	for i, ing := range r.Ingredients {
		if i > 0 {
			s += " + "
		}
		s += ing.Item.Name() + " " + strconv.Itoa(int(ch.CountItem(ing.Item))) + "/" + strconv.Itoa(int(ing.Quantity))
	}
	return s
}
