package menu

import (
	"os"
	"strconv"

	"runa/internal/character"
	"runa/internal/guild"
	"runa/internal/tui"
)

// RunGuildMenu ouvre le panneau de la guilde par-dessus le rendu intérieur
// (ne prend pas tout le canvas). Affiche les quêtes du rang courant
// UNIQUEMENT (pas les rangs inférieurs déjà dépassés), permet d'accepter,
// de rendre, de recommencer une quête déjà complétée (répétable), et de
// demander une promotion de rang.
func RunGuildMenu(in *os.File, out *os.File, c *tui.Canvas, ch *character.Character, gs *guild.GuildStatus, renderUnder func()) {
	cursor := 0
	msg := ""
	msgCol := tui.FGLightGreen

	bw, bh := 62, 17
	bx, by := tui.CenteredBox(c.W, c.H, bw, bh)
	scrollOffset := 0
	maxVisible := 5

	for {
		renderUnder()
		tui.FillStyled(c, bx, by, bw, bh, ' ', "", tui.BGBlack)
		tui.DrawBoxWithTitle(c, bx, by, bw, bh, "ADVENTURERS GUILD")

		hdr := "Rank: " + gs.Rank.String() + "   |   Level: " + strconv.Itoa(int(ch.Level()))
		c.WriteStyled(bx+3, by+1, hdr, tui.FGCyan, tui.BGBlack)

		quests := guild.QuestsForRank(gs.Rank)
		if len(quests) == 0 {
			c.WriteStyled(bx+4, by+4, "(No quests available for this rank)", tui.FGGray, tui.BGBlack)
		} else {
			if cursor >= len(quests) {
				cursor = len(quests) - 1
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
				if idx >= len(quests) {
					break
				}
				q := quests[idx]
				rowY := by + 3 + vi*2

				prefix := "  "
				fg := tui.FGWhite
				if idx == cursor {
					prefix = "> "
					fg = tui.FGBrightWhite
				}

				rewStr := "+" + strconv.Itoa(int(q.RewardExp)) + " XP  +" + strconv.Itoa(int(q.RewardMoney)) + " G"
				qTitle := "[" + q.Rank.String() + "] " + q.Name
				c.WriteStyled(bx+3, rowY, prefix+qTitle, fg, tui.BGBlack)
				c.WriteStyled(bx+bw-len(rewStr)-3, rowY, rewStr, tui.FGYellow, tui.BGBlack)

				statStr := ""
				statCol := tui.FGWhite
				if kills, active := gs.ActiveQuests[q.ID]; active {
					if kills >= q.RequiredKills {
						statStr = "[Ready: ENTER to turn in!]"
						statCol = tui.FGLightGreen
					} else {
						statStr = "[In progress: " + strconv.Itoa(int(kills)) + "/" + strconv.Itoa(int(q.RequiredKills)) + " " + q.MonsterID + "s]"
						statCol = tui.FGLightYellow
					}
				} else if times := gs.TimesCompleted[q.ID]; times > 0 {
					// Répétable : complétée par le passé, mais reste acceptable.
					statStr = "[Available: ENTER to accept] (completed x" + strconv.Itoa(int(times)) + ")"
					statCol = tui.FGLightCyan
				} else {
					statStr = "[Available: ENTER to accept]"
					statCol = tui.FGLightCyan
				}
				c.WriteStyled(bx+6, rowY+1, statStr, statCol, tui.BGBlack)
			}
		}

		if msg != "" {
			c.WriteStyled(bx+3, by+bh-3, msg, msgCol, tui.BGBlack)
		}

		hint := "[↑/↓] Choose [ENTER] Action [P] Promotion [ESC] Exit"
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
			if cursor < len(quests)-1 {
				cursor++
			}
		case tui.KeyRune:
			if ev.R == 'p' || ev.R == 'P' {
				res, newRank := guild.TryPromote(gs, ch)
				switch res {
				case guild.Promoted:
					msg = "Congratulations! Promoted to Rank " + newRank.String() + "!"
					msgCol = tui.FGLightGreen
					cursor = 0 // le rang a changé : la liste affichée change entièrement
				case guild.ErrLevelTooLow:
					msg = "Level too low for next rank."
					msgCol = tui.FGLightRed
				case guild.ErrAlreadyMaxRank:
					msg = "You are already at maximum rank (S)!"
					msgCol = tui.FGYellow
				}
			}
		case tui.KeyEnter:
			if cursor >= 0 && cursor < len(quests) {
				q := quests[cursor]
				if kills, active := gs.ActiveQuests[q.ID]; active {
					if kills >= q.RequiredKills {
						res := guild.TurnInQuest(gs, ch, q.ID)
						if res == guild.TurnedIn {
							msg = "Quest completed! +" + strconv.Itoa(int(q.RewardExp)) + " XP, +" + strconv.Itoa(int(q.RewardMoney)) + " G"
							msgCol = tui.FGLightGreen
						}
					} else {
						msg = "Quest in progress: " + strconv.Itoa(int(kills)) + "/" + strconv.Itoa(int(q.RequiredKills)) + " defeated."
						msgCol = tui.FGLightYellow
					}
				} else {
					res := guild.AcceptQuest(gs, q.ID)
					if res == guild.Accepted {
						msg = "Quest accepted: " + q.Name
						msgCol = tui.FGLightGreen
					} else {
						msg = "Cannot accept quest."
						msgCol = tui.FGLightRed
					}
				}
			}
		}
	}
}
