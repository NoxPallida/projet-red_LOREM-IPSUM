package guild

// Member est l'interface minimale dont la guilde a besoin, côté joueur.
// Évite d'importer character ici (character → guild, jamais l'inverse).
type Member interface {
	Level() uint8
	GainExp(amount uint16)
	EarnMoney(amount uint16)
}

type Rank int

const (
	RankF Rank = iota
	RankE
	RankD
	RankC
	RankB
	RankA
	RankS
)

func (r Rank) String() string {
	switch r {
	case RankF:
		return "F"
	case RankE:
		return "E"
	case RankD:
		return "D"
	case RankC:
		return "C"
	case RankB:
		return "B"
	case RankA:
		return "A"
	case RankS:
		return "S"
	default:
		return "?"
	}
}

var rankOrder = []Rank{RankF, RankE, RankD, RankC, RankB, RankA, RankS}

var RankMinLevel = map[Rank]uint8{
	RankF: 1,
	RankE: 5,
	RankD: 10,
	RankC: 20,
	RankB: 35,
	RankA: 50,
	RankS: 70,
}

func nextRank(r Rank) (Rank, bool) {
	for i, rk := range rankOrder {
		if rk == r && i+1 < len(rankOrder) {
			return rankOrder[i+1], true
		}
	}
	return r, false
}

// Quest décrit une quête de chasse : tuer un certain nombre d'un monstre
// donné, en échange d'exp et d'argent. La difficulté (RequiredKills,
// récompenses) augmente avec le Rank de la quête.
type Quest struct {
	ID            string
	Rank          Rank
	Name          string
	MonsterID     string // identifiant du monstre visé
	RequiredKills uint8
	RewardExp     uint16
	RewardMoney   uint16
}

// registry centralise toutes les quêtes du jeu, indexées par ID.
var registry = make(map[string]Quest)
var registeredQuests []Quest

// NewQuest centralise la création ET l'enregistrement, pour ne jamais
// avoir une quête définie mais introuvable par ID.
func NewQuest(id string, rank Rank, name, monsterID string, requiredKills uint8, rewardExp uint16, rewardMoney uint16) Quest {
	q := Quest{
		ID:            id,
		Rank:          rank,
		Name:          name,
		MonsterID:     monsterID,
		RequiredKills: requiredKills,
		RewardExp:     rewardExp,
		RewardMoney:   rewardMoney,
	}
	registry[id] = q
	registeredQuests = append(registeredQuests, q)
	return q
}

// GetQuest recherche une quête par son ID.
func GetQuest(id string) (Quest, bool) {
	q, ok := registry[id]
	return q, ok
}

// QuestsForRank renvoie UNIQUEMENT les quêtes du rang précis donné
// (pas les rangs inférieurs) : c'est la liste que la guilde doit
// afficher, pour que le joueur ne voie que ce qui correspond à son
// rang actuel plutôt qu'un cumul de toute sa progression passée.
func QuestsForRank(rank Rank) []Quest {
	var out []Quest
	for _, q := range registeredQuests {
		if q.Rank == rank {
			out = append(out, q)
		}
	}
	return out
}

// GuildStatus suit la progression d'UN joueur dans la guilde :
// son rang actuel, ses quêtes en cours et le nombre de fois où
// chacune a été rendue. Les quêtes sont répétables : TimesCompleted
// sert de compteur/historique, PAS de verrou empêchant un nouvel accept.
type GuildStatus struct {
	Rank           Rank
	ActiveQuests   map[string]uint8  // questID -> kills enregistrés pour la run en cours
	TimesCompleted map[string]uint16 // questID -> nombre total de fois rendue
}

// NewGuildStatus crée un statut de guilde neuf : rang F, aucune quête.
func NewGuildStatus() *GuildStatus {
	return &GuildStatus{
		Rank:           RankF,
		ActiveQuests:   make(map[string]uint8),
		TimesCompleted: make(map[string]uint16),
	}
}

type PromoteResult int

const (
	Promoted PromoteResult = iota
	ErrAlreadyMaxRank
	ErrLevelTooLow
)

// TryPromote vérifie si le joueur peut monter au rang suivant, et le
// fait si c'est le cas. La promotion ne dépend QUE du niveau du
// personnage, jamais des quêtes complétées.
func TryPromote(g *GuildStatus, m Member) (PromoteResult, Rank) {
	next, ok := nextRank(g.Rank)
	if !ok {
		return ErrAlreadyMaxRank, g.Rank
	}
	if m.Level() < RankMinLevel[next] {
		return ErrLevelTooLow, next
	}
	g.Rank = next
	return Promoted, g.Rank
}

type AcceptResult int

const (
	Accepted AcceptResult = iota
	ErrQuestNotFound
	ErrRankTooLow
	ErrAlreadyActive
)

// AcceptQuest permet au joueur de prendre une quête, si son rang le
// permet. Les quêtes sont répétables : avoir déjà complété une quête
// par le passé ne bloque plus un nouvel accept, seule une quête DÉJÀ
// EN COURS (ActiveQuests) empêche de la reprendre en double.
func AcceptQuest(g *GuildStatus, questID string) AcceptResult {
	q, ok := GetQuest(questID)
	if !ok {
		return ErrQuestNotFound
	}
	if q.Rank > g.Rank {
		return ErrRankTooLow
	}
	if _, active := g.ActiveQuests[questID]; active {
		return ErrAlreadyActive
	}
	g.ActiveQuests[questID] = 0
	return Accepted
}

// RegisterKill met à jour la progression de toutes les quêtes actives
// qui ciblent monsterID. À appeler depuis le système de combat, à
// chaque monstre tué. Renvoie les IDs des quêtes désormais complétables.
func RegisterKill(g *GuildStatus, monsterID string) []string {
	var readyToTurnIn []string
	for questID, kills := range g.ActiveQuests {
		q, ok := GetQuest(questID)
		if !ok || q.MonsterID != monsterID {
			continue
		}
		kills++
		g.ActiveQuests[questID] = kills
		if kills >= q.RequiredKills {
			readyToTurnIn = append(readyToTurnIn, questID)
		}
	}
	return readyToTurnIn
}

type TurnInResult int

const (
	TurnedIn TurnInResult = iota
	ErrNotActive
	ErrNotYetCompleted
)

// TurnInQuest verse la récompense au joueur si la quête est bien
// terminée, puis la retire des quêtes actives (elle redevient donc
// immédiatement disponible pour être acceptée à nouveau : c'est ce
// qui rend les quêtes répétables). TimesCompleted est incrémenté à
// titre d'historique/affichage, sans jamais bloquer un futur accept.
func TurnInQuest(g *GuildStatus, m Member, questID string) TurnInResult {
	kills, active := g.ActiveQuests[questID]
	if !active {
		return ErrNotActive
	}
	q, ok := GetQuest(questID)
	if !ok || kills < q.RequiredKills {
		return ErrNotYetCompleted
	}

	m.GainExp(q.RewardExp)
	m.EarnMoney(q.RewardMoney)

	delete(g.ActiveQuests, questID)
	g.TimesCompleted[questID]++
	return TurnedIn
}

// --- Quêtes définies ---
// La difficulté monte avec le rang : plus de kills requis, meilleures
// récompenses. À équilibrer plus finement une fois les monstres définis.
var (
	QuestRatsF          = NewQuest("rats_f", RankF, "Rat Infestation", "rat", 5, 20, 10)
	QuestKoboldF        = NewQuest("kobold_f", RankF, "Kobolds Outside Town", "kobold", 3, 30, 15)
	QuestGoblinsF       = NewQuest("goblins_f", RankF, "Goblin Training", "goblin", 3, 25, 12)
	QuestRatsCaveF      = NewQuest("rats_cave_f", RankF, "Cellar Cleanout", "rat", 8, 35, 18)
	QuestGoblinMasteryF = NewQuest("goblin_mastery_f", RankF, "Sand Arena Trial", "goblin", 6, 45, 25)

	QuestBoarsE     = NewQuest("boars_e", RankE, "Pest Boars", "boar", 6, 60, 30)
	QuestWolvesE    = NewQuest("wolves_e", RankE, "Wolf Pack", "wolf", 8, 70, 35)
	QuestBoarHuntE  = NewQuest("boar_hunt_e", RankE, "Great Boar Hunt", "boar", 10, 85, 45)
	QuestWolfAlphaE = NewQuest("wolf_alpha_e", RankE, "Alpha Wolf Hunt", "wolf", 12, 100, 50)

	QuestHobgoblinsD = NewQuest("hobgoblins_d", RankD, "Hobgoblin Raiders", "hobgoblin", 8, 130, 65)
	QuestSkeletonsD  = NewQuest("skeletons_d", RankD, "Restless Bones", "skeleton_warrior", 6, 150, 75)
	QuestHarpiesD    = NewQuest("harpies_d", RankD, "Cliffside Harpies", "harpy", 7, 145, 70)
	QuestKoboldNestD = NewQuest("kobold_nest_d", RankD, "Kobold Nest Purge", "kobold", 12, 120, 60)

	QuestTrollsC     = NewQuest("trolls_c", RankC, "Troll Hunt", "troll", 3, 220, 110)
	QuestOrcsC       = NewQuest("orcs_c", RankC, "Orc Skirmish", "orc", 8, 200, 100)
	QuestHarpyRoostC = NewQuest("harpy_roost_c", RankC, "Harpy Roost Clearing", "harpy", 12, 240, 120)

	QuestOgresB      = NewQuest("ogres_b", RankB, "Ogre Rampage", "ogre", 4, 380, 190)
	QuestMinotaursB  = NewQuest("minotaurs_b", RankB, "Labyrinth Minotaurs", "minotaur", 3, 420, 210)
	QuestOrcWarbandB = NewQuest("orc_warband_b", RankB, "Orc Warband", "orc", 15, 350, 170)

	QuestWyrmsA        = NewQuest("wyrms_a", RankA, "Wyrm Den", "wyrm", 3, 600, 300)
	QuestMinotaurLordA = NewQuest("minotaur_lord_a", RankA, "Minotaur Lord's Guard", "minotaur", 6, 550, 280)

	QuestWyrmNestS = NewQuest("wyrm_nest_s", RankS, "Wyrm Nest Cleansing", "wyrm", 6, 900, 450)
	QuestWyvernS   = NewQuest("wyvern_s", RankS, "The Wyvern's Reign", "wyvern", 1, 1500, 700)
)
