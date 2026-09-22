package spawner

import (
	"math/rand"

	"runa/internal/enemies"
	"runa/internal/guild"
	"runa/internal/world"
)

// PoolByRank liste les IDs de monstres qui peuvent apparaître à un rang
// de guilde donné.
var PoolByRank = map[guild.Rank][]string{
	guild.RankF: {"rat", "slime", "goblin", "kobold"},
	guild.RankE: {"rat", "slime", "goblin", "kobold", "hobgoblin", "wolf", "boar"},
	guild.RankD: {"rat", "slime", "goblin", "kobold", "hobgoblin", "wolf", "boar", "skeleton_warrior", "harpy"},
	guild.RankC: {"rat", "slime", "goblin", "kobold", "hobgoblin", "wolf", "boar", "skeleton_warrior", "harpy", "orc", "troll"},
	guild.RankB: {"rat", "slime", "goblin", "kobold", "hobgoblin", "wolf", "boar", "skeleton_warrior", "harpy", "orc", "troll", "ogre", "minotaur"},
	guild.RankA: {"rat", "slime", "goblin", "kobold", "hobgoblin", "wolf", "boar", "skeleton_warrior", "harpy", "orc", "troll", "ogre", "minotaur", "wyrm"},
	guild.RankS: {"rat", "slime", "goblin", "kobold", "hobgoblin", "wolf", "boar", "skeleton_warrior", "harpy", "orc", "troll", "ogre", "minotaur", "wyrm", "wyvern"},
}

// Population : nombre total de monstres présents sur la carte en
// permanence. Chaque mort déclenche un respawn ailleurs
const Population = 30

// spawnableKinds : les seules natures de tile où un monstre peut
// apparaître, indépendamment de leur couleur/rendu.
var spawnableKinds = map[world.TileKind]bool{
	world.TileGrass:  true,
	world.TileForest: true,
	world.TileSand:   true,
}

// Spawner tient la population de monstres statiques de la carte :
// où ils sont, et comment les remplacer à leur mort.
type Spawner struct {
	tiles      [][]world.Tile
	candidates [][2]int // toutes les positions spawnables, précalculées une fois
	mobs       map[[2]int]*enemies.EnemyInstance
	rng        *rand.Rand
}

// NewSpawner scanne toute la carte une fois pour lister les positions
// spawnables (H/C/S uniquement), puis fait apparaître Population
// monstres au hasard parmi elles, selon le rang de départ du joueur.
func NewSpawner(tiles [][]world.Tile, rank guild.Rank, seed int64) *Spawner {
	s := &Spawner{
		tiles: tiles,
		mobs:  make(map[[2]int]*enemies.EnemyInstance),
		rng:   rand.New(rand.NewSource(seed)),
	}
	for y, row := range tiles {
		for x, t := range row {
			if spawnableKinds[t.Kind] {
				if isTrainingGround(x, y) {
					continue
				}
				s.candidates = append(s.candidates, [2]int{x, y})
			}
		}
	}
	s.spawnTrainingGoblin()
	s.Repopulate(rank)
	return s
}

// Repopulate fait apparaître des monstres jusqu'à atteindre Population,
// en piochant dans le pool débloqué par le rang actuel. À appeler au
// démarrage ET après une promotion de guilde, pour que les EMPLACEMENTS
// VIDES se remplissent avec la diversité du nouveau rang (les monstres
// déjà en vie ne changent pas rétroactivement : seuls les futurs
// remplacements profitent du rang supérieur).
func (s *Spawner) Repopulate(rank guild.Rank) {
	for len(s.mobs) < Population && len(s.mobs) < len(s.candidates) {
		pos, ok := s.randomFreePosition()
		if !ok {
			break // plus aucune case libre sur la carte
		}
		s.spawnAt(pos[0], pos[1], rank)
	}
}

// randomFreePosition tire une position spawnable non occupée. Abandonne
// après un nombre raisonnable d'essais pour ne jamais boucler à l'infini
// si la carte est presque saturée de monstres.
func (s *Spawner) randomFreePosition() ([2]int, bool) {
	const maxAttempts = 200
	for i := 0; i < maxAttempts; i++ {
		pos := s.candidates[s.rng.Intn(len(s.candidates))]
		if _, occupied := s.mobs[pos]; !occupied {
			return pos, true
		}
	}
	return [2]int{}, false
}

// nearbyFreePosition cherche une position libre à proximité de (x, y),
// en élargissant progressivement le rayon. C'est ce qui fait qu'un
// monstre tué est remplacé "à côté" plutôt que n'importe où sur la carte.
func (s *Spawner) nearbyFreePosition(x, y int) ([2]int, bool) {
	for radius := 1; radius <= 10; radius++ {
		for _, pos := range s.candidates {
			dx, dy := pos[0]-x, pos[1]-y
			if dx < -radius || dx > radius || dy < -radius || dy > radius {
				continue
			}
			if _, occupied := s.mobs[pos]; !occupied {
				return pos, true
			}
		}
	}
	return s.randomFreePosition() // repli : n'importe où si le voisinage est saturé
}

// spawnAt fait apparaître un monstre choisi au hasard dans le pool du
// rang donné, à la position (x, y). Le niveau du monstre reprend
// directement guild.RankMinLevel : plus le rang est haut, plus les
// monstres sont forts, alignés sur la progression attendue du joueur.
func (s *Spawner) spawnAt(x, y int, rank guild.Rank) {
	pool := PoolByRank[rank]
	if len(pool) == 0 {
		return
	}
	id := pool[s.rng.Intn(len(pool))]
	level := guild.RankMinLevel[rank]
	if level < 1 {
		level = 1
	}
	inst, ok := enemies.NewEnemyInstance(id, level)
	if !ok {
		return
	}
	s.mobs[[2]int{x, y}] = inst
}

// EnemyAt renvoie le monstre vivant présent à (x, y), s'il y en a un.
// À appeler depuis la logique de déplacement : si ok == true, déclencher
// le combat au lieu de laisser le joueur avancer sur la case (le mob
// est statique, donc "marcher dessus" = seule façon de le rencontrer).
func (s *Spawner) EnemyAt(x, y int) (*enemies.EnemyInstance, bool) {
	e, ok := s.mobs[[2]int{x, y}]
	if !ok || !e.IsAlive() {
		return nil, false
	}
	return e, true
}

// Kill retire le monstre mort à (x, y) et fait immédiatement réapparaître
// un remplaçant à proximité, pour garder Population strictement constant
// sur la carte. rank est repassé à chaque appel (plutôt que mémorisé)
// pour que le remplaçant profite d'une éventuelle promotion récente.
func (s *Spawner) Kill(x, y int, rank guild.Rank) {
	delete(s.mobs, [2]int{x, y})
	if isTrainingGround(x, y) {
		s.spawnTrainingGoblin()
		return
	}
	if pos, ok := s.nearbyFreePosition(x, y); ok {
		s.spawnAt(pos[0], pos[1], rank)
	}
}

func isTrainingGround(x, y int) bool {
	return x >= 43 && x <= 51 && y >= 173 && y <= 181
}

func (s *Spawner) spawnTrainingGoblin() {
	if inst, ok := enemies.NewEnemyInstance("goblin", 1); ok {
		s.mobs[[2]int{47, 177}] = inst
	}
}
