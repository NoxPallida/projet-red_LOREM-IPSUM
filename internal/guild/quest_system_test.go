package guild

import (
	"testing"
)

type mockMember struct {
	level uint8
	exp   uint16
	money uint16
}

func (m *mockMember) Level() uint8            { return m.level }
func (m *mockMember) GainExp(amount uint16)   { m.exp += amount }
func (m *mockMember) EarnMoney(amount uint16) { m.money += amount }

func TestQuestsAvailableRankF(t *testing.T) {
	quests := QuestsAvailable(RankF)
	if len(quests) < 5 {
		t.Fatalf("expected at least 5 quests for Rank F, got %d", len(quests))
	}

	// Verify deterministic ordering
	questsAgain := QuestsAvailable(RankF)
	for i := range quests {
		if quests[i].ID != questsAgain[i].ID {
			t.Fatalf("quest order mismatch at index %d: %s vs %s", i, quests[i].ID, questsAgain[i].ID)
		}
	}
}

func TestAcceptAndTurnInQuest(t *testing.T) {
	gs := NewGuildStatus()
	mem := &mockMember{level: 1}

	res := AcceptQuest(gs, "goblins_f")
	if res != Accepted {
		t.Fatalf("expected Accepted, got %v", res)
	}

	ready := RegisterKill(gs, "goblin")
	if len(ready) != 0 {
		t.Fatalf("expected 0 ready quests after 1 goblin kill, got %d", len(ready))
	}
	RegisterKill(gs, "goblin")
	ready = RegisterKill(gs, "goblin")
	if len(ready) == 0 {
		t.Fatalf("expected goblin quest to be ready after 3 kills")
	}

	turnRes := TurnInQuest(gs, mem, "goblins_f")
	if turnRes != TurnedIn {
		t.Fatalf("expected TurnedIn, got %v", turnRes)
	}
	if !gs.CompletedQuests["goblins_f"] {
		t.Fatalf("expected goblins_f to be marked completed")
	}
	if mem.exp == 0 || mem.money == 0 {
		t.Fatalf("expected exp and money to be rewarded")
	}
}
