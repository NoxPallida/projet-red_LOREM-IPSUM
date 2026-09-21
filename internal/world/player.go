package world

// Player c'est qui bouge. Separe de World : World tient la carte,
// Player tient la position + ou il regarde.
// X,Y en coords monde (peuvent etre negatifs, d'ou floorDiv).

type Player struct {
	X, Y int
	Dir  Direction
}

// Direction : 4 cotes, facon vieux Pokemon.
type Direction int

const (
	North Direction = iota
	South
	East
	West
)

// MovePlayer tourne toujours, avance seulement si la case devant
// est chargee et marchable. Rend true si on a bouge.
// A appeler depuis la boucle de jeu avec ZQSD -> Direction.
func MovePlayer(w *World, p *Player, dir Direction) bool {
	if p == nil || w == nil {
		return false
	}
	p.Dir = dir
	nx, ny := p.X, p.Y
	switch dir {
	case North:
		ny--
	case South:
		ny++
	case East:
		nx++
	case West:
		nx--
	}
	tile, loaded := w.TileAt(nx, ny)
	if !loaded || !tile.Walkable {
		return false
	}
	p.X, p.Y = nx, ny
	return true
}

// FrontTile rend la case devant le joueur (dans la direction
// ou il regarde) + si elle est chargee. Pour ESPACE = interagir.
func FrontTile(w *World, p Player) (Tile, bool) {
	x, y := p.X, p.Y
	switch p.Dir {
	case North:
		y--
	case South:
		y++
	case East:
		x++
	case West:
		x--
	}
	return w.TileAt(x, y)
}

// TryMove garde pour compat : meme chose que MovePlayer, tourne
// et avance. Prefere MovePlayer dans le nouveau code.
func TryMove(p *Player, dir Direction) bool {
	// Sans World on ne peut pas verifier la case, on tourne seulement.
	if p == nil {
		return false
	}
	p.Dir = dir
	return false
}

// TryMoveWorld : ancien appel TryMove(w,p,dir) -> MovePlayer(w,p,dir).
func TryMoveWorld(w *World, p *Player, dir Direction) bool {
	return MovePlayer(w, p, dir)
}
