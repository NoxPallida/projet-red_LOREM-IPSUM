package world

// Tile porte aussi une couleur, pour que le rendu tui n'ait qu'à lire
// cette donnée plutôt que de la déduire du Symbol par un switch.
type Tile struct {
	Symbol   rune
	FG       string // un des codes tui.FG*
	Walkable bool
}

const ChunkSize = 16

type Chunk struct {
	X, Y  int
	Tiles [ChunkSize][ChunkSize]Tile
}

type ZoneKind int

const (
	ZoneWild ZoneKind = iota
	ZoneTown
	ZoneShop
	ZoneForge
	ZoneHome
	ZoneGuild
)

type Zone struct {
	Name                   string
	Kind                   ZoneKind
	MinX, MinY, MaxX, MaxY int
}

func (z Zone) Contains(x, y int) bool {
	return x >= z.MinX && x <= z.MaxX && y >= z.MinY && y <= z.MaxY
}

var Zones = []Zone{
	{Name: "Town", Kind: ZoneTown, MinX: 10, MinY: 10, MaxX: 30, MaxY: 30},
	{Name: "Shop", Kind: ZoneShop, MinX: 14, MinY: 12, MaxX: 16, MaxY: 14},
	{Name: "Forge", Kind: ZoneForge, MinX: 20, MinY: 12, MaxX: 22, MaxY: 14},
	{Name: "Home", Kind: ZoneHome, MinX: 14, MinY: 20, MaxX: 16, MaxY: 22},
	{Name: "Guild", Kind: ZoneGuild, MinX: 20, MinY: 20, MaxX: 22, MaxY: 22},
}

func ZoneAt(x, y int) (Zone, bool) {
	for _, z := range Zones {
		if z.Contains(x, y) {
			return z, true
		}
	}
	return Zone{}, false
}

type Player struct{ X, Y int }

type Direction int

const (
	North Direction = iota
	South
	East
	West
)

type World struct {
	loaded map[[2]int]*Chunk
	loader ChunkLoader
}

type ChunkLoader interface {
	Load(chunkX, chunkY int) *Chunk
}

func NewWorld(loader ChunkLoader) *World {
	return &World{loaded: make(map[[2]int]*Chunk), loader: loader}
}

func (w *World) EnsureLoaded(playerX, playerY, radius int) {
	pcx, pcy := chunkCoords(playerX, playerY)
	needed := make(map[[2]int]bool)
	for dx := -radius; dx <= radius; dx++ {
		for dy := -radius; dy <= radius; dy++ {
			key := [2]int{pcx + dx, pcy + dy}
			needed[key] = true
			if _, ok := w.loaded[key]; !ok {
				w.loaded[key] = w.loader.Load(pcx+dx, pcy+dy)
			}
		}
	}
	for key := range w.loaded {
		if !needed[key] {
			delete(w.loaded, key)
		}
	}
}

func (w *World) TileAt(x, y int) (Tile, bool) {
	cx, cy := chunkCoords(x, y)
	chunk, ok := w.loaded[[2]int{cx, cy}]
	if !ok {
		return Tile{}, false
	}
	lx := ((x % ChunkSize) + ChunkSize) % ChunkSize
	ly := ((y % ChunkSize) + ChunkSize) % ChunkSize
	return chunk.Tiles[ly][lx], true
}

func (w *World) TryMove(p *Player, dir Direction) bool {
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

func chunkCoords(x, y int) (int, int) { return floorDiv(x, ChunkSize), floorDiv(y, ChunkSize) }
func floorDiv(a, b int) int {
	if a < 0 {
		return (a - b + 1) / b
	}
	return a / b
}
