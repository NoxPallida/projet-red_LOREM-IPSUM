package world

import (
	"bufio"
	_ "embed"
	"strings"

	"runa/internal/tui"
)

//go:embed maps/town.txt
var townMapRaw string

//go:embed maps/map.txt
var worldMapRaw string

// Block est le seul caractère affiché : chaque case de la carte
// est un pavé plein coloré selon son type (H = vert, O = bleu...).
const Block = '█'

// TileLegend associe un symbole du fichier .txt à une Tile réelle.
// Centralisé ici : dessiner la carte ne demande de connaître QUE
// ces caractères, jamais la structure interne de Tile.
// Chaque tile est rendue comme un bloc plein de sa couleur.
var TileLegend = map[rune]Tile{
	// --- carte monde (map.txt) ---
	'H': {Symbol: Block, FG: tui.FGGreen, Walkable: true},       // herbe -> vert
	'O': {Symbol: Block, FG: tui.FGBlue, Walkable: false},       // eau -> bleu
	'C': {Symbol: Block, FG: tui.FGLightGreen, Walkable: true},  // foret/champ -> vert clair
	'X': {Symbol: Block, FG: tui.FGGray, Walkable: false},       // rocher/obstacle -> gris
	'S': {Symbol: Block, FG: tui.FGLightYellow, Walkable: true}, // sable/marchand -> jaune clair
	'B': {Symbol: Block, FG: tui.FGWhite, Walkable: false},      // batiment/mur -> blanc
	'E': {Symbol: Block, FG: tui.FGLightCyan, Walkable: true},   // entree/porte spec -> cyan clair
	'-': {Symbol: Block, FG: tui.FGYellow, Walkable: true},      // chemin -> jaune
	'=': {Symbol: ' ', FG: "", Walkable: false},                 // bordure/vide -> vide
	// --- carte ville (town.txt, compat) ---
	'.': {Symbol: Block, FG: tui.FGGreen, Walkable: true},      // herbe / sol
	'#': {Symbol: Block, FG: tui.FGWhite, Walkable: false},     // mur
	'~': {Symbol: Block, FG: tui.FGBlue, Walkable: false},      // eau
	'+': {Symbol: Block, FG: tui.FGYellow, Walkable: true},     // porte
	'F': {Symbol: Block, FG: tui.FGLightMagenta, Walkable: true}, // forgeron
	'G': {Symbol: Block, FG: tui.FGLightCyan, Walkable: true},  // guilde
	'@': {Symbol: Block, FG: tui.FGGreen, Walkable: true},      // spawn joueur : bloc d'herbe
	' ': {Symbol: ' ', FG: "", Walkable: false},                // hors carte / vide
}

// ParsedMap est le résultat du parsing d'un fichier de carte :
// la grille de tiles ET où le joueur doit apparaître au démarrage.
type ParsedMap struct {
	Tiles  [][]Tile
	StartX int
	StartY int
}

// ParseMap lit un texte ASCII ligne par ligne et le convertit en grille
// de Tile, en repérant au passage la position du symbole '@' (spawn).
// Sans '@' (ex: map.txt), le spawn = premiere case marchable en
// partant du centre, sinon 0,0.
func ParseMap(raw string) ParsedMap {
	var pm ParsedMap
	foundSpawn := false
	scanner := bufio.NewScanner(strings.NewReader(raw))
	// lignes tres longues (map.txt ~200 cols) : augmente le buffer.
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	y := 0
	for scanner.Scan() {
		line := scanner.Text()
		row := make([]Tile, len([]rune(line)))
		for x, r := range []rune(line) {
			tile, ok := TileLegend[r]
			if !ok {
				tile = Tile{Symbol: Block, FG: "", Walkable: false} // symbole inconnu : traité comme un mur, pour ne jamais laisser passer par erreur
			}
			row[x] = tile
			if r == '@' {
				pm.StartX, pm.StartY = x, y
				foundSpawn = true
			}
		}
		pm.Tiles = append(pm.Tiles, row)
		y++
	}
	if !foundSpawn && len(pm.Tiles) > 0 {
		cy := len(pm.Tiles) / 2
		cx := 0
		if len(pm.Tiles[cy]) > 0 {
			cx = len(pm.Tiles[cy]) / 2
		}
		// spirale carree autour du centre : premier marchable gagne.
		best := -1
		bx, by := 0, 0
		for yy := 0; yy < len(pm.Tiles); yy++ {
			for xx := 0; xx < len(pm.Tiles[yy]); xx++ {
				if !pm.Tiles[yy][xx].Walkable {
					continue
				}
				d := abs(xx-cx) + abs(yy-cy)
				if best < 0 || d < best {
					best = d
					bx, by = xx, yy
				}
			}
		}
		if best >= 0 {
			pm.StartX, pm.StartY = bx, by
		}
	}
	return pm
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

// StaticLoader charge des chunks à partir d'une carte fixe et unique,
// pas de génération procédurale.
type StaticLoader struct {
	FullMap [][]Tile
}

func (l *StaticLoader) Load(chunkX, chunkY int) *Chunk {
	chunk := &Chunk{X: chunkX, Y: chunkY}
	for ly := 0; ly < ChunkSize; ly++ {
		for lx := 0; lx < ChunkSize; lx++ {
			gx := chunkX*ChunkSize + lx
			gy := chunkY*ChunkSize + ly
			if gy >= 0 && gy < len(l.FullMap) && gx >= 0 && gx < len(l.FullMap[gy]) {
				chunk.Tiles[ly][lx] = l.FullMap[gy][gx]
			} else {
				chunk.Tiles[ly][lx] = Tile{Symbol: ' ', Walkable: false}
			}
		}
	}
	return chunk
}

// LoadTownMap parse la carte embarquée et renvoie tout ce qu'il faut
// pour construire le World + placer le joueur au bon endroit.
func LoadTownMap() (*StaticLoader, ParsedMap) {
	pm := ParseMap(townMapRaw)
	return &StaticLoader{FullMap: pm.Tiles}, pm
}

// LoadWorldMap parse la grande carte du monde (maps/map.txt) :
// blocs pleins colores plein ecran, spawn auto au centre.
func LoadWorldMap() (*StaticLoader, ParsedMap) {
	pm := ParseMap(worldMapRaw)
	return &StaticLoader{FullMap: pm.Tiles}, pm
}
