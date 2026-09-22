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

// QuestsForRank renvoie toutes les quêtes d'un rang précis.
func QuestsForRank(rank Rank) []Quest {
	var out []Quest
	for _, q := range registeredQuests {
		if q.Rank == rank {
			out = append(out, q)
		}
	}
	return out
}

// QuestsAvailable renvoie toutes les quêtes accessibles jusqu'au rang donné.
func QuestsAvailable(rank Rank) []Quest {
	var out []Quest
	for _, q := range registeredQuests {
		if q.Rank <= rank {
			out = append(out, q)
		}
	}
	return out
}

// --- Quêtes définies ---
// La difficulté monte avec le rang : plus de kills requis, meilleures
// récompenses. À équilibrer plus finement une fois les monstres définis.

var (
	QuestRatsF          = NewQuest("rats_f", RankF, "Infestation de rats", "rat", 5, 20, 10)
	QuestWolvesF        = NewQuest("wolves_f", RankF, "Loups aux abords de la ville", "wolf", 3, 30, 15)
	QuestGoblinsF       = NewQuest("goblins_f", RankF, "Entraînement aux gobelins", "goblin", 3, 25, 12)
	QuestRatsCaveF      = NewQuest("rats_cave_f", RankF, "Nettoyage des caves", "rat", 8, 35, 18)
	QuestGoblinMasteryF = NewQuest("goblin_mastery_f", RankF, "Épreuve du terrain de sable", "goblin", 6, 45, 25)

	QuestBoarsE     = NewQuest("boars_e", RankE, "Sangliers ravageurs", "boar", 6, 60, 30)
	QuestWolvesE    = NewQuest("wolves_e", RankE, "Meute de loups", "wolf", 8, 70, 35)
	QuestBoarHuntE  = NewQuest("boar_hunt_e", RankE, "Chasse aux grands sangliers", "boar", 10, 85, 45)
	QuestWolfAlphaE = NewQuest("wolf_alpha_e", RankE, "Traque du loup dominant", "wolf", 12, 100, 50)

	QuestTrollsD = NewQuest("trolls_d", RankD, "Chasse au troll", "troll", 3, 150, 80)
)

// GuildStatus suit la progression d'UN joueur dans la guilde :
// son rang actuel, ses quêtes en cours et leur avancement.
type GuildStatus struct {
	Rank            Rank
	ActiveQuests    map[string]uint8 // questID -> kills enregistrés
	CompletedQuests map[string]bool
}

// NewGuildStatus crée un statut de guilde neuf : rang F, aucune quête.
func NewGuildStatus() *GuildStatus {
	return &GuildStatus{
		Rank:            RankF,
		ActiveQuests:    make(map[string]uint8),
		CompletedQuests: make(map[string]bool),
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
	ErrAlreadyCompleted
)

// AcceptQuest permet au joueur de prendre une quête, si son rang le
// permet. Une quête d'un rang supérieur au sien reste inaccessible.
func AcceptQuest(g *GuildStatus, questID string) AcceptResult {
	q, ok := GetQuest(questID)
	if !ok {
		return ErrQuestNotFound
	}
	if q.Rank > g.Rank {
		return ErrRankTooLow
	}
	if g.CompletedQuests[questID] {
		return ErrAlreadyCompleted
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
// terminée, puis la retire des quêtes actives. La quête doit être
// rendue explicitement (pas de récompense automatique dès le dernier kill),
// pour laisser la main au menu/guilde pour l'interaction avec le joueur.
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
	g.CompletedQuests[questID] = true
	return TurnedIn
}
