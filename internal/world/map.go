package world

import (
	"bufio"
	_ "embed"
	"strings"
)

go:embed maps/town.txt
var townMapRaw string

// TileLegend associe un symbole du fichier .txt à une Tile réelle.
// Centralisé ici : dessiner la carte ne demande de connaître QUE
// ces caractères, jamais la structure interne de Tile.
var TileLegend = map[rune]Tile{
	'H': {Symbol: 'H', FG: "", Walkable: true},  // herbe
	'#': {Symbol: '#', FG: "", Walkable: false}, // mur
	'O': {Symbol: 'O', FG: "", Walkable: false}, // eau
	'+': {Symbol: '+', FG: "", Walkable: true},  // porte
	'S': {Symbol: 'S', FG: "", Walkable: true},  // marchand
	'F': {Symbol: 'F', FG: "", Walkable: true},  // forgeron
	'H': {Symbol: 'H', FG: "", Walkable: true},  // maison
	'G': {Symbol: 'G', FG: "", Walkable: true},  // guilde
	'@': {Symbol: '.', FG: "", Walkable: true},  // spawn joueur : rendu comme de l'herbe
	' ': {Symbol: ' ', FG: "", Walkable: false}, // hors carte / vide
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
func ParseMap(raw string) ParsedMap {
	var pm ParsedMap
	scanner := bufio.NewScanner(strings.NewReader(raw))
	y := 0
	for scanner.Scan() {
		line := scanner.Text()
		row := make([]Tile, len([]rune(line)))
		for x, r := range []rune(line) {
			tile, ok := TileLegend[r]
			if !ok {
				tile = Tile{Symbol: r, Walkable: false} // symbole inconnu : traité comme un mur, pour ne jamais laisser passer par erreur
			}
			row[x] = tile
			if r == '@' {
				pm.StartX, pm.StartY = x, y
			}
		}
		pm.Tiles = append(pm.Tiles, row)
		y++
	}
	return pm
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
