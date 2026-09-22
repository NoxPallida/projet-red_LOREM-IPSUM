package world

// TileKind classe une tile par nature (herbe, eau, forêt...), indépendamment
// de son rendu (Symbol/FG). Sert à des règles de jeu qui ne doivent PAS
// dépendre de l'apparence (ex: où les monstres peuvent spawn).
// TileVoid = valeur zéro : une Tile{} non renseignée (chunk non chargé,
// case hors carte) est toujours TileVoid, donc jamais spawnable par défaut.
type TileKind int

const (
	TileVoid TileKind = iota
	TileGrass
	TileWater
	TileForest
	TileRock
	TileSand
	TileBuilding
	TileDoor
	TileFloor
)

// Tile porte aussi une couleur, pour que le rendu tui n'ait qu'à lire
// cette donnée plutôt que de la déduire du Symbol par un switch.
type Tile struct {
	Symbol   rune
	FG       string // un des codes tui.FG*
	Walkable bool
	Kind     TileKind
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

// Zones calees sur les vrais batiments de map.txt (bornes incluses).
// Chaque porte 'E' doit tomber dans le rect de son batiment : c'est
// comme ca qu'entrer/sortir retrouve le bon interieur.
var Zones = []Zone{
	{Name: "Town", Kind: ZoneTown, MinX: 10, MinY: 10, MaxX: 30, MaxY: 30},
	{Name: "Shop", Kind: ZoneShop, MinX: 55, MinY: 172, MaxX: 67, MaxY: 176},
	{Name: "Guild", Kind: ZoneGuild, MinX: 70, MinY: 172, MaxX: 82, MaxY: 176},
	{Name: "Forge", Kind: ZoneForge, MinX: 55, MinY: 179, MaxX: 67, MaxY: 183},
	{Name: "Home", Kind: ZoneHome, MinX: 70, MinY: 179, MaxX: 82, MaxY: 183},
}

func ZoneAt(x, y int) (Zone, bool) {
	for _, z := range Zones {
		if z.Contains(x, y) {
			return z, true
		}
	}
	return Zone{}, false
}

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

// LoadRadius rend le rayon de chunks à charger pour couvrir une vue
// de halfW x halfH cases + 2 chunks de marge, quel que soit l'écran.
// Sans ça, sur grand terminal les bords de la vue restent vides.
func LoadRadius(halfW, halfH int) int {
	half := halfW
	if halfH > half {
		half = halfH
	}
	return half/ChunkSize + 2
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

func chunkCoords(x, y int) (int, int) { return floorDiv(x, ChunkSize), floorDiv(y, ChunkSize) }
func floorDiv(a, b int) int {
	if a < 0 {
		return (a - b + 1) / b
	}
	return a / b
}
