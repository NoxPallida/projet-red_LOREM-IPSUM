package menu

import (
	"os"
	"strconv"

	"runa/internal/character"
	"runa/internal/guild"
	"runa/internal/tui"
)

// RunGuildMenu ouvre le panneau de la guilde par-dessus le rendu intérieur
// (ne prend pas tout le canvas). Affiche les quêtes du rang, permet d'accepter,
// de rendre et de demander une promotion de rang.
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
		tui.DrawBoxWithTitle(c, bx, by, bw, bh, "GUILDE DES AVENTURIERS")

		hdr := "Rang : " + gs.Rank.String() + "   |   Niveau : " + strconv.Itoa(int(ch.Level()))
		c.WriteStyled(bx+3, by+1, hdr, tui.FGCyan, tui.BGBlack)

		quests := guild.QuestsAvailable(gs.Rank)
		if len(quests) == 0 {
			c.WriteStyled(bx+4, by+4, "(Aucune quête disponible pour ce rang)", tui.FGGray, tui.BGBlack)
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

				rewStr := "+" + strconv.Itoa(int(q.RewardExp)) + " XP  +" + strconv.Itoa(int(q.RewardMoney)) + " PO"
				qTitle := "[" + q.Rank.String() + "] " + q.Name
				c.WriteStyled(bx+3, rowY, prefix+qTitle, fg, tui.BGBlack)
				c.WriteStyled(bx+bw-len(rewStr)-3, rowY, rewStr, tui.FGYellow, tui.BGBlack)

				statStr := ""
				statCol := tui.FGWhite
				if gs.CompletedQuests[q.ID] {
					statStr = "[Terminée]"
					statCol = tui.FGGray
				} else if kills, active := gs.ActiveQuests[q.ID]; active {
					if kills >= q.RequiredKills {
						statStr = "[Terminée : ENTRÉE pour rendre !]"
						statCol = tui.FGLightGreen
					} else {
						statStr = "[En cours : " + strconv.Itoa(int(kills)) + "/" + strconv.Itoa(int(q.RequiredKills)) + " " + q.MonsterID + "s]"
						statCol = tui.FGLightYellow
					}
				} else {
					statStr = "[Disponible : ENTRÉE pour accepter]"
					statCol = tui.FGLightCyan
				}
				c.WriteStyled(bx+6, rowY+1, statStr, statCol, tui.BGBlack)
			}
		}

		if msg != "" {
			c.WriteStyled(bx+3, by+bh-3, msg, msgCol, tui.BGBlack)
		}

		hint := "[↑/↓] Choisir [ENTRÉE] Action [P] Promotion [ECHAP] Sortir"
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
			if cursor < len(quests)-1 {
				cursor++
			}
		case tui.KeyRune:
			if ev.R == 'p' || ev.R == 'P' {
				res, newRank := guild.TryPromote(gs, ch)
				switch res {
				case guild.Promoted:
					msg = "Félicitations ! Promu au Rang " + newRank.String() + " !"
					msgCol = tui.FGLightGreen
				case guild.ErrLevelTooLow:
					msg = "Niveau insuffisant pour le rang supérieur."
					msgCol = tui.FGLightRed
				case guild.ErrAlreadyMaxRank:
					msg = "Vous êtes déjà au rang maximal (S) !"
					msgCol = tui.FGYellow
				}
			}
		case tui.KeyEnter:
			if cursor >= 0 && cursor < len(quests) {
				q := quests[cursor]
				if gs.CompletedQuests[q.ID] {
					msg = "Quête déjà terminée."
					msgCol = tui.FGGray
				} else if kills, active := gs.ActiveQuests[q.ID]; active {
					if kills >= q.RequiredKills {
						res := guild.TurnInQuest(gs, ch, q.ID)
						if res == guild.TurnedIn {
							msg = "Quête validée ! +" + strconv.Itoa(int(q.RewardExp)) + " XP, +" + strconv.Itoa(int(q.RewardMoney)) + " PO"
							msgCol = tui.FGLightGreen
						}
					} else {
						msg = "Quête en cours : " + strconv.Itoa(int(kills)) + "/" + strconv.Itoa(int(q.RequiredKills)) + " éliminés."
						msgCol = tui.FGLightYellow
					}
				} else {
					res := guild.AcceptQuest(gs, q.ID)
					if res == guild.Accepted {
						msg = "Quête acceptée : " + q.Name
						msgCol = tui.FGLightGreen
					} else {
						msg = "Impossible d'accepter la quête."
						msgCol = tui.FGLightRed
					}
				}
			}
		}
	}
}
