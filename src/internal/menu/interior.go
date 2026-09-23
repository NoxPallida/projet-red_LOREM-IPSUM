package menu

import (
	"runa/src/internal/tui"
	"runa/src/internal/world"
)

// InteriorSession : le joueur est DANS un batiment.
// PX/PY = position locale dans la piece.
// EntryX/EntryY = la porte dehors, pour y reaparaitre en sortant.
type InteriorSession struct {
	In             world.Interior
	PX, PY         int
	EntryX, EntryY int
}

// enterInterior tente d'entrer : le joueur doit etre sur une porte
// (TileDoor) couverte par une zone batiment (Shop/Forge/Home/Guild).
// Rend nil si ce n'est pas une entree (rue, eau, zone sauvage...).
func enterInterior(w *world.World, p *world.Player) *InteriorSession {
	t, ok := w.TileAt(p.X, p.Y)
	if !ok || t.Kind != world.TileDoor {
		return nil
	}
	z, ok := world.ZoneAt(p.X, p.Y)
	if !ok {
		return nil
	}
	switch z.Kind {
	case world.ZoneShop, world.ZoneForge, world.ZoneHome, world.ZoneGuild:
	default:
		return nil
	}
	in := world.BuildInterior(z.Kind)
	return &InteriorSession{In: in, PX: in.SpawnX(), PY: in.SpawnY(), EntryX: p.X, EntryY: p.Y}
}

// move deplace dans la piece (murs + PNJ bloquent). Rend true si bouge.
func (s *InteriorSession) move(dir world.Direction) bool {
	if s == nil {
		return false
	}
	dx, dy := 0, 0
	switch dir {
	case world.North:
		dy--
	case world.South:
		dy++
	case world.East:
		dx++
	case world.West:
		dx--
	}
	if !s.In.At(s.PX+dx, s.PY+dy).Walkable {
		return false
	}
	s.PX += dx
	s.PY += dy
	return true
}

// onDoor : le joueur est sur la porte interieure -> sortir.
func (s *InteriorSession) onDoor() bool {
	return s != nil && s.PX == s.In.DoorX && s.PY == s.In.DoorY
}

// near dit si (px,py) touche (x,y) : les 8 cases autour incluses.
func near(px, py, x, y int) bool {
	dx := px - x
	if dx < 0 {
		dx = -dx
	}
	dy := py - y
	if dy < 0 {
		dy = -dy
	}
	return dx <= 1 && dy <= 1
}

// isNearChest indique si le joueur est sur une case adjacente au coffre.
// Faux s'il n'y a pas de coffre dans la piece (ChestX < 0).
func (s *InteriorSession) isNearChest() bool {
	if s == nil || s.In.ChestX < 0 {
		return false
	}
	return near(s.PX, s.PY, s.In.ChestX, s.In.ChestY)
}

// isNearNPC indique si le joueur est sur une case adjacente au PNJ de la piece.
func (s *InteriorSession) isNearNPC() bool {
	if s == nil {
		return false
	}
	return near(s.PX, s.PY, s.In.NPC.X, s.In.NPC.Y)
}

// drawInterior : fond noir explicite sur tout l'ecran, piece centree,
// PNJ avec son nom au-dessus, joueur en bloc blanc 2x plus large (comme dehors).
func drawInterior(c *tui.Canvas, s *InteriorSession) {
	if s == nil {
		return
	}
	tui.FillStyled(c, 0, 0, c.W, c.H, ' ', "", tui.BGBlack)
	const scaleX = 2
	roomW := s.In.W * scaleX
	ox := (c.W - roomW) / 2
	oy := (c.H - s.In.H) / 2
	for y := 0; y < s.In.H; y++ {
		for x := 0; x < s.In.W; x++ {
			t := s.In.At(x, y)
			sx := ox + x*scaleX
			sy := oy + y
			c.SetStyled(sx, sy, '█', t.FG, "")
			c.SetStyled(sx+1, sy, '█', t.FG, "")
		}
	}
	// PNJ : seulement s'il y en a un (sinon rien a dessiner).
	npc := s.In.NPC
	if npc.Name != "" {
		name := []rune(npc.Name)
		c.WriteStyled(ox+npc.X*scaleX-len(name)/2+1, oy+npc.Y-1, npc.Name, npc.FG, "")
		c.SetStyled(ox+npc.X*scaleX, oy+npc.Y, npc.Glyph, npc.FG, "")
		c.SetStyled(ox+npc.X*scaleX+1, oy+npc.Y, '█', npc.FG, "")
	}

	// Coffre : seulement s'il y en a un (maisons sans habitant).
	// Glyphe 'C' (deja dans Cells) + nom au-dessus, comme le PNJ.
	if s.In.ChestX >= 0 {
		chestName := []rune("Coffre")
		c.WriteStyled(ox+s.In.ChestX*scaleX-len(chestName)/2+1, oy+s.In.ChestY-1, "Coffre", tui.FGLightYellow, "")
		c.SetStyled(ox+s.In.ChestX*scaleX, oy+s.In.ChestY, 'C', tui.FGLightYellow, "")
		c.SetStyled(ox+s.In.ChestX*scaleX+1, oy+s.In.ChestY, '█', tui.FGLightYellow, "")
	}

	// Joueur : pavé 2x1 blanc éclatant
	c.SetStyled(ox+s.PX*scaleX, oy+s.PY, '█', tui.FGBrightWhite, "")
	c.SetStyled(ox+s.PX*scaleX+1, oy+s.PY, '█', tui.FGBrightWhite, "")

	if s.isNearChest() {
		hint := "[SPACE] Coffre"
		c.WriteStyled((c.W-len(hint))/2, oy+s.In.H+1, hint, tui.FGYellow, "")
	} else if s.isNearNPC() {
		hint := "[SPACE] Talk to " + npc.Name
		c.WriteStyled((c.W-len(hint))/2, oy+s.In.H+1, hint, tui.FGYellow, "")
	}
}
