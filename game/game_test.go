package game

import (
	"testing"
	"time"
)

// TestNewGame tests creation of a game
func TestNewGame(t *testing.T) {
	seed := time.Now().UnixNano()
	g, err := NewGame(4, seed)

	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	if len(g.Players) != 4 {
		t.Errorf("Expected 4 players, got %d", len(g.Players))
	}

	// Check initial coins
	for i, p := range g.Players {
		if p.GetCoins() != 2 {
			t.Errorf("Player %d should start with 2 coins, has %d", i, p.GetCoins())
		}

		if p.InfluenceCount() != 2 {
			t.Errorf("Player %d should start with 2 influences, has %d", i, p.InfluenceCount())
		}
	}
}

// TestIncomeAction tests the Income action
func TestIncomeAction(t *testing.T) {
	seed := time.Now().UnixNano()
	g, _ := NewGame(2, seed)

	player := g.Players[0]
	startCoins := player.GetCoins()

	// Create and execute Income action
	action := NewIncomeAction(player)
	err := action.Execute(g)

	if err != nil {
		t.Fatalf("Income action failed: %v", err)
	}

	if player.GetCoins() != startCoins+1 {
		t.Errorf("Expected player coins to increase by 1, from %d to %d, got %d",
			startCoins, startCoins+1, player.GetCoins())
	}
}

// TestForcedCoup tests the forced coup rule with 10+ coins
func TestForcedCoup(t *testing.T) {
	seed := time.Now().UnixNano()
	g, _ := NewGame(2, seed)

	player := g.Players[0]

	// Give player 10 coins
	player.AddCoins(8) // Already has 2 coins to start

	// Get legal actions
	actions := g.GetLegalActions(player)

	// Should only have Coup actions
	for _, action := range actions {
		if action.Name() != "Coup" {
			t.Errorf("Player with 10 coins should only have Coup actions, has %s", action.Name())
		}
	}

	if len(actions) == 0 {
		t.Error("No legal actions for player with 10 coins")
	}
}

// TestCardConservation verifies the card economy after every single action
// across many seeds, player counts, and AI modes: 15 cards total, exactly 3
// copies of each character, and no hand ever exceeding 2 influence cards.
// This invariant is what catches challenge-resolution bugs (e.g. a defender
// keeping a revealed card while also drawing a replacement).
func TestCardConservation(t *testing.T) {
	constructors := map[string]func(int, int64) (*Game, error){
		"original": NewGameWithOriginalAI,
		"mixed":    NewGameWithMixedAIs,
		"high": func(pc int, seed int64) (*Game, error) {
			return NewGameWithAITypes(pc, nil, HighCompetitive, seed)
		},
		"low": func(pc int, seed int64) (*Game, error) {
			return NewGameWithAITypes(pc, nil, LowCompetitive, seed)
		},
	}

	for name, newGame := range constructors {
		for players := 2; players <= 6; players++ {
			for seed := int64(0); seed < 50; seed++ {
				g, err := newGame(players, seed)
				if err != nil {
					t.Fatalf("%s %dp seed %d: %v", name, players, seed, err)
				}
				if err := g.ValidateInvariants(); err != nil {
					t.Fatalf("%s %dp seed %d before first turn: %v", name, players, seed, err)
				}
				for !g.Finished && g.Turn < MaxTurns {
					g.ExecuteTurn()
					if err := g.ValidateInvariants(); err != nil {
						t.Fatalf("%s %dp seed %d turn %d: %v", name, players, seed, g.Turn, err)
					}
				}
			}
		}
	}
}

// TestOnlyTargetCanBlock verifies that Steal and Assassinate are only ever
// blocked by the player they target.
func TestOnlyTargetCanBlock(t *testing.T) {
	for seed := int64(0); seed < 300; seed++ {
		g, err := NewGameWithMixedAIs(5, seed)
		if err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		g.RunToCompletion()

		for _, a := range g.ActionLog {
			if (a.Action == "Steal" || a.Action == "Assassinate") && a.Blocker != -1 && a.Blocker != a.Target {
				t.Fatalf("seed %d turn %d: player %d blocked a %s aimed at player %d",
					seed, a.Turn, a.Blocker, a.Action, a.Target)
			}
		}
	}
}

// TestAssassinateRefundOnSuccessfulChallenge verifies the 3-coin cost is
// returned when the assassination claim itself is successfully challenged.
func TestAssassinateRefundOnSuccessfulChallenge(t *testing.T) {
	g, _ := NewGameWithOriginalAI(2, 1)
	actor := g.Players[0].(*AIPlayer)
	target := g.Players[1].(*AIPlayer)

	// Actor bluffs the Assassin; target always challenges
	actor.Influences = []Card{GetCardByName(Duke), GetCardByName(Duke)}
	actor.Coins = 3
	target.Influences = []Card{GetCardByName(Contessa), GetCardByName(Contessa)}
	target.Strategy.ChallengeRate = 1.0

	log := ActionLog{Target: 1, Blocker: -1}
	if ok := g.ResolveAction(actor, NewAssassinateAction(actor, target), &log); ok {
		t.Fatal("bluffed assassination should fail when challenged")
	}
	if actor.GetCoins() != 3 {
		t.Errorf("cost should be refunded after a successful challenge, actor has %d coins", actor.GetCoins())
	}
	if actor.InfluenceCount() != 1 {
		t.Errorf("actor should lose an influence for the failed bluff, has %d", actor.InfluenceCount())
	}
	if target.InfluenceCount() != 2 {
		t.Errorf("challenger should keep both influences, has %d", target.InfluenceCount())
	}
}

// TestAssassinateCostPaidWhenBlocked verifies the 3-coin cost stays paid when
// the assassination is blocked by a Contessa.
func TestAssassinateCostPaidWhenBlocked(t *testing.T) {
	g, _ := NewGameWithOriginalAI(2, 1)
	actor := g.Players[0].(*AIPlayer)
	target := g.Players[1].(*AIPlayer)

	// Actor genuinely holds the Assassin; target holds a Contessa and blocks
	// (AlwaysBlock), and nobody challenges anything
	actor.Influences = []Card{GetCardByName(Assassin), GetCardByName(Duke)}
	actor.Coins = 3
	actor.Strategy.ChallengeRate = 0
	target.Influences = []Card{GetCardByName(Contessa), GetCardByName(Duke)}
	target.Strategy.ChallengeRate = 0

	log := ActionLog{Target: 1, Blocker: -1}
	if ok := g.ResolveAction(actor, NewAssassinateAction(actor, target), &log); ok {
		t.Fatal("assassination should be blocked by the Contessa")
	}
	if actor.GetCoins() != 0 {
		t.Errorf("cost should stay paid when blocked, actor has %d coins", actor.GetCoins())
	}
	if target.InfluenceCount() != 2 {
		t.Errorf("blocked target should keep both influences, has %d", target.InfluenceCount())
	}
}

// TestClaimsTracking verifies public claim history is recorded on character
// claims and reset by a completed Ambassador exchange.
func TestClaimsTracking(t *testing.T) {
	g, _ := NewGameWithOriginalAI(2, 3)
	actor := g.Players[0].(*AIPlayer)
	other := g.Players[1].(*AIPlayer)
	actor.Strategy.ChallengeRate = 0
	other.Strategy.ChallengeRate = 0
	other.Strategy.BluffRate = 0

	log := ActionLog{Target: -1, Blocker: -1}
	if ok := g.ResolveAction(actor, NewTaxAction(actor), &log); !ok {
		t.Fatal("unopposed Tax should succeed")
	}
	state := g.GetStateForPlayer(1)
	if got := state.Claims[0]; len(got) != 1 || got[0] != Duke {
		t.Fatalf("expected claim history [Duke], got %v", got)
	}

	// A completed exchange wipes the history down to the Ambassador claim
	log = ActionLog{Target: -1, Blocker: -1}
	if ok := g.ResolveAction(actor, NewExchangeAction(actor), &log); !ok {
		t.Fatal("unopposed Exchange should succeed")
	}
	state = g.GetStateForPlayer(1)
	if got := state.Claims[0]; len(got) != 1 || got[0] != Ambassador {
		t.Fatalf("expected claim history reset to [Ambassador], got %v", got)
	}
}

// TestImpossibleClaimIsChallenged verifies card-counting AIs always challenge
// a claim whose copies are all visibly accounted for, and that low-competitive
// AIs (no card memory) do not.
func TestImpossibleClaimIsChallenged(t *testing.T) {
	claimant := NewEnhancedAIPlayer(1, NewBasicAIStrategy(0, 0, true), 8)

	medium := NewEnhancedAIPlayer(0, NewBasicAIStrategy(0, 0, true), 7) // Level defaults to Medium
	medium.AddInfluence([]Card{GetCardByName(Duke), GetCardByName(Contessa)})

	state := GameState{
		Players: []PlayerState{
			{ID: 0, Coins: 2, Influences: 2, IsAlive: true},
			{ID: 1, Coins: 2, Influences: 2, IsAlive: true},
		},
		// Two Dukes in the discard pile + one in hand = all three accounted for
		Discarded: []Card{GetCardByName(Duke), GetCardByName(Duke)},
		Claims:    map[int][]string{},
	}

	if !medium.ChallengeDecision(state, claimant, NewTaxAction(claimant)) {
		t.Fatal("card-counting AI should always challenge an impossible Duke claim")
	}

	// With only one Duke visible, the zero challenge rate applies
	possible := state
	possible.Discarded = nil
	if medium.ChallengeDecision(possible, claimant, NewTaxAction(claimant)) {
		t.Fatal("possible claim should not be challenged at challenge rate 0")
	}

	// Low competitive AIs have no card memory
	lowStrategy := NewBasicAIStrategy(0, 0, false)
	lowStrategy.Level = LowCompetitive
	low := NewEnhancedAIPlayer(2, lowStrategy, 9)
	low.AddInfluence([]Card{GetCardByName(Duke), GetCardByName(Contessa)})
	if low.ChallengeDecision(state, claimant, NewTaxAction(claimant)) {
		t.Fatal("low-competitive AI should not count cards (challenge rate 0)")
	}
}

// TestBlockLogRecordsDefeatedBlocks verifies a block defeated by a challenge
// stays in the action log with BlockSucceeded=false and ground truth intact.
func TestBlockLogRecordsDefeatedBlocks(t *testing.T) {
	g, _ := NewGameWithOriginalAI(2, 1)
	actor := g.Players[0].(*AIPlayer)
	target := g.Players[1].(*AIPlayer)

	// Actor genuinely holds the Assassin and always challenges; target holds
	// no Contessa but always bluff-blocks
	actor.Influences = []Card{GetCardByName(Assassin), GetCardByName(Duke)}
	actor.Coins = 3
	actor.Strategy.ChallengeRate = 1.0
	target.Influences = []Card{GetCardByName(Duke), GetCardByName(Duke)}
	target.Strategy.ChallengeRate = 0
	target.Strategy.BluffRate = 1.0

	log := ActionLog{Target: 1, Blocker: -1}
	if ok := g.ResolveAction(actor, NewAssassinateAction(actor, target), &log); !ok {
		t.Fatal("assassination should proceed after the block is defeated")
	}
	if log.Blocker != 1 || log.BlockingCharacter != Contessa {
		t.Fatalf("defeated block should stay in the log, got %+v", log)
	}
	if log.BlockSucceeded {
		t.Fatal("defeated block must not be marked successful")
	}
	if !log.BlockChallenged {
		t.Fatal("the block challenge should be recorded")
	}
	if log.BlockerHadCard {
		t.Fatal("ground truth: the blocker held no Contessa")
	}
	// One influence lost to the failed block challenge, one to the assassination
	if target.InfluenceCount() != 0 {
		t.Fatalf("target should be eliminated, has %d influences", target.InfluenceCount())
	}
}

// TestCompleteGame tests running a complete game
func TestCompleteGame(t *testing.T) {
	seed := time.Now().UnixNano()
	g, _ := NewGame(3, seed)

	// Run the game to completion
	winner := g.RunToCompletion()

	if !g.Finished {
		t.Error("Game should be marked as finished")
	}

	if winner == nil {
		t.Fatal("Game finished without a winner")
	}

	// Count living players
	livingCount := 0
	for _, p := range g.Players {
		if p.IsAlive() {
			livingCount++
		}
	}

	if livingCount != 1 {
		t.Errorf("Game finished with %d living players, should be exactly 1", livingCount)
	}

	if !winner.IsAlive() {
		t.Error("Winner should be alive")
	}
}
