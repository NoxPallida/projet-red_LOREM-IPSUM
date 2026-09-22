package menu

import (
	"runa/internal/tui"
	"runa/internal/world"
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

// isNearNPC indique si le joueur est sur une case adjacente au PNJ de la piece.
func (s *InteriorSession) isNearNPC() bool {
	if s == nil {
		return false
	}
	dx := s.PX - s.In.NPC.X
	dy := s.PY - s.In.NPC.Y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx <= 1 && dy <= 1
}

// drawInterior : fond noir explicite sur tout l'ecran, piece centree,
// PNJ avec son nom au-dessus, joueur en bloc blanc (comme dehors).
func drawInterior(c *tui.Canvas, s *InteriorSession) {
	if s == nil {
		return
	}
	tui.FillStyled(c, 0, 0, c.W, c.H, ' ', "", tui.BGBlack)
	ox := (c.W - s.In.W) / 2
	oy := (c.H - s.In.H) / 2
	for y := 0; y < s.In.H; y++ {
		for x := 0; x < s.In.W; x++ {
			t := s.In.At(x, y)
			c.SetStyled(ox+x, oy+y, t.Symbol, t.FG, "")
		}
	}
	npc := s.In.NPC
	name := []rune(npc.Name)
	c.WriteStyled(ox+npc.X-len(name)/2, oy+npc.Y-1, npc.Name, npc.FG, "")
	c.SetStyled(ox+npc.X, oy+npc.Y, npc.Glyph, npc.FG, "")
	c.SetStyled(ox+s.PX, oy+s.PY, '█', tui.FGBrightWhite, "")
	if s.isNearNPC() {
		hint := "[ESPACE] Parler avec " + npc.Name
		c.WriteStyled((c.W-len(hint))/2, oy+s.In.H+1, hint, tui.FGYellow, "")
	}
}
