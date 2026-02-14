package simulation

import (
	"testing"
	"time"
)

// TestSimulationBasics verifies the simulator can run games
func TestSimulationBasics(t *testing.T) {
	// Create a small simulation config for testing
	config := Config{
		TotalGames: 10,
		Workers:    2,
		Verbose:    false,
		Seed:       time.Now().UnixNano(),
		GamesPerPlayerCount: map[int]int{
			2: 2,
			3: 2,
			4: 2,
			5: 2,
			6: 2,
		},
		OutputDir: "./test_output",
	}

	// Create and run simulator
	simulator := NewSimulator(config)
	results := simulator.Run()

	// Verify results
	if len(results.Results) != config.TotalGames {
		t.Errorf("Expected %d game results, got %d", config.TotalGames, len(results.Results))
	}

	// Verify game distribution
	countMap := make(map[int]int)
	for _, result := range results.Results {
		countMap[result.PlayerCount]++
	}

	// Check if each player count has games
	for count := 2; count <= 6; count++ {
		if countMap[count] == 0 {
			t.Errorf("No games run with %d players", count)
		}
	}

	// Verify all games have a winner
	for i, result := range results.Results {
		if result.WinnerID < 0 || result.WinnerID >= result.PlayerCount {
			t.Errorf("Game %d has invalid winner ID: %d", i, result.WinnerID)
		}

		if len(result.WinnerCharacters) == 0 {
			t.Errorf("Game %d winner has no characters", i)
		}
	}
}

// TestMetricsCollection verifies that metrics are collected correctly
func TestMetricsCollection(t *testing.T) {
	// Create a metrics collector
	collector := NewMetricsCollector()

	// Create sample game results
	results := []GameResult{
		{
			ID:               1,
			PlayerCount:      2,
			WinnerID:         0,
			WinnerCharacters: []string{"Duke", "Captain"},
			PlayerStartingCards: map[int][]string{
				0: {"Duke", "Captain"},
				1: {"Assassin", "Contessa"},
			},
			TotalTurns: 10,
		},
	}
	// Process the results
	collector.ProcessGameResults(results)

	// Verify character stats exist
	if len(collector.CharacterStats) == 0 {
		t.Error("No character stats collected")
	}

	// Verify Duke and Captain have wins
	if collector.CharacterStats["Duke"].GamesWon == 0 {
		t.Error("Duke should have won games")
	}

	if collector.CharacterStats["Captain"].GamesWon == 0 {
		t.Error("Captain should have won games")
	}

	// Verify rankings exist
	if len(collector.RankedCharacters) == 0 {
		t.Error("No character rankings generated")
	}
}
