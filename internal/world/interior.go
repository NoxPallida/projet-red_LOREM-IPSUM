package world

import "runa/internal/tui"

// NPC : un habitant d'interieur. Il ne bouge pas, on ne peut pas
// marcher dessus, son nom s'affiche au-dessus de lui.
type NPC struct {
	Name  string
	X, Y  int // coords locales dans la piece
	Glyph rune
	FG    string // un des codes tui.FG*
}

// Interior : une piece generee (meme taille pour tous les batiments,
// c'est plus grand dedans que dehors, classique).
// Murs infranchissables, sol bois, porte au sud, PNJ au centre.
type Interior struct {
	Kind  ZoneKind
	W, H  int
	Cells [][]Tile // [y][x], origine en haut a gauche de la piece
	DoorX int
	DoorY int
	NPC   NPC
	// Coffre : case de stockage, infranchissable comme le PNJ.
	// Le CONTENU vit sur le perso (Character.Chest, un seul coffre
	// partage) ; ici ce n'est que la position du meuble.
	ChestX int
	ChestY int
}

const (
	interiorW = 27
	interiorH = 15
)

// npcByKind : un PNJ par batiment.
func npcByKind(kind ZoneKind) NPC {
	switch kind {
	case ZoneShop:
		return NPC{Name: "Merchant", X: interiorW / 2, Y: 5, Glyph: 'M', FG: tui.FGLightGreen}
	case ZoneForge:
		return NPC{Name: "Blacksmith", X: interiorW / 2, Y: 5, Glyph: 'B', FG: tui.FGLightRed}
	case ZoneGuild:
		return NPC{Name: "Guild Master", X: interiorW / 2, Y: 5, Glyph: 'G', FG: tui.FGLightCyan}
	default:
		return NPC{}
	}
}

// BuildInterior fabrique la piece d'un batiment : anneau de murs,
// sol bois (autre couleur que l'herbe dehors), porte au sud,
// PNJ au centre. Le joueur spawne sur la porte.
func BuildInterior(kind ZoneKind) Interior {
	wall := Tile{Symbol: '█', FG: tui.FGWhite, Walkable: false, Kind: TileBuilding}
	floor := Tile{Symbol: '█', FG: tui.FGYellow, Walkable: true, Kind: TileFloor}
	in := Interior{Kind: kind, W: interiorW, H: interiorH}
	in.Cells = make([][]Tile, in.H)
	for y := 0; y < in.H; y++ {
		in.Cells[y] = make([]Tile, in.W)
		for x := 0; x < in.W; x++ {
			if x == 0 || y == 0 || x == in.W-1 || y == in.H-1 {
				in.Cells[y][x] = wall
			} else {
				in.Cells[y][x] = floor
			}
		}
	}
	// Porte plein sud, au milieu du mur.
	in.DoorX, in.DoorY = in.W/2, in.H-1
	in.Cells[in.DoorY][in.DoorX] = Tile{Symbol: '█', FG: tui.FGLightCyan, Walkable: true, Kind: TileDoor}
	// PNJ au centre, sur une case devenue infranchissable.
	// Sans habitant (Name vide) : on ne touche a aucune case.
	in.NPC = npcByKind(kind)
	if in.NPC.Name != "" {
		in.Cells[in.NPC.Y][in.NPC.X] = Tile{Symbol: in.NPC.Glyph, FG: in.NPC.FG, Walkable: false, Kind: TileVoid}
	}
	// Coffre UNIQUEMENT dans les maisons sans habitant : celles qui ont
	// un PNJ n'en ont pas besoin. En haut a gauche (sol libre : loin
	// PNJ, porte et spawn). Meme motif que le PNJ : glyphe + case
	// infranchissable. Sans coffre : ChestX/ChestY = -1 (pas de case).
	if in.NPC.Name == "" {
		in.ChestX, in.ChestY = 3, 3
		in.Cells[in.ChestY][in.ChestX] = Tile{Symbol: 'C', FG: tui.FGLightYellow, Walkable: false, Kind: TileVoid}
	} else {
		in.ChestX, in.ChestY = -1, -1
	}
	return in
}

// At rend la tile locale (x, y). Hors piece = mur vide infranchissable.
func (in *Interior) At(x, y int) Tile {
	if in == nil || y < 0 || y >= in.H || x < 0 || x >= len(in.Cells[y]) {
		return Tile{Symbol: ' ', Walkable: false}
	}
	return in.Cells[y][x]
}

// SpawnX/SpawnY : le joueur apparait devant la porte en entrant.
func (in *Interior) SpawnX() int { return in.DoorX }
func (in *Interior) SpawnY() int { return in.DoorY - 1 }
