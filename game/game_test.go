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
