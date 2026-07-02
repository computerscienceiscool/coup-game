package simulation

import (
	"fmt"
	"testing"
	"time"

	"github.com/computerscienceiscool/coup-game/game"
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

// TestSeedingReproducibleAndDistinct verifies that per-game seeds derive only
// from the base seed and game ID: two runs with the same seed produce
// identical games no matter how work is scheduled across workers, and
// different games don't share RNG streams (no mass duplicates, which the old
// seed+workerID+gameID scheme produced for ~2/3 of all games).
func TestSeedingReproducibleAndDistinct(t *testing.T) {
	config := Config{
		TotalGames:          300,
		Workers:             4,
		Seed:                424242,
		GamesPerPlayerCount: map[int]int{4: 300},
		AIMode:              "mixed",
	}

	fingerprint := func(r GameResult) string {
		s := fmt.Sprintf("pc%d w%d t%d", r.PlayerCount, r.WinnerID, r.TotalTurns)
		for _, a := range r.Actions {
			s += fmt.Sprintf("|%d %s %d %v %v %d", a.PlayerID, a.Action, a.Target, a.Success, a.Challenged, a.Blocker)
		}
		return s
	}

	run := func() map[int]string {
		results := NewSimulator(config).Run()
		out := make(map[int]string, len(results.Results))
		for _, r := range results.Results {
			out[r.ID] = fingerprint(r)
		}
		return out
	}

	first, second := run(), run()
	if len(first) != config.TotalGames || len(second) != config.TotalGames {
		t.Fatalf("expected %d games per run, got %d and %d", config.TotalGames, len(first), len(second))
	}

	for id, fp := range first {
		if second[id] != fp {
			t.Fatalf("game %d differs between two runs with the same seed", id)
		}
	}

	distinct := make(map[string]bool, len(first))
	for _, fp := range first {
		distinct[fp] = true
	}
	if len(distinct) < config.TotalGames*9/10 {
		t.Errorf("games heavily duplicated: only %d distinct of %d", len(distinct), config.TotalGames)
	}
}

// TestMetricsCollection verifies that metrics are collected correctly from
// the ground-truth action log.
func TestMetricsCollection(t *testing.T) {
	// Create a metrics collector
	collector := NewMetricsCollector()

	// One 2-player game: player 0 (dealt Duke+Captain) wins on turn 10,
	// player 1 (dealt Assassin+Contessa) is eliminated on turn 8.
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
			EliminationTurns: map[int]int{
				1: 8, // Player 1 eliminated on turn 8
			},
			TotalTurns: 10,
			Actions: []game.ActionLog{
				// Player 0 taxes truthfully and survives a challenge
				{Turn: 0, PlayerID: 0, Action: "Tax", Target: -1, Success: true,
					Challenged: true, ActorHadCard: true, Blocker: -1},
				// Player 1 bluffs a Steal (unchallenged) but player 0 blocks
				// with a real Captain
				{Turn: 1, PlayerID: 1, Action: "Steal", Target: 0, Success: false,
					Challenged: false, ActorHadCard: false,
					Blocker: 0, BlockingCharacter: "Captain", BlockerHadCard: true,
					BlockChallenged: false, BlockSucceeded: true},
			},
		},
	}
	collector.ProcessGameResults(results)

	duke := collector.CharacterStats["Duke"]
	if duke.TimesDealt != 1 || duke.WinsWhenDealt != 1 || duke.GamesDealt != 1 {
		t.Errorf("Duke dealt tracking wrong: %+v", duke)
	}
	if duke.FinalHandWins != 1 {
		t.Errorf("Duke should have 1 final-hand win, has %d", duke.FinalHandWins)
	}
	if duke.Claims != 1 || duke.Challenges != 1 || duke.Bluffs != 0 {
		t.Errorf("Duke claim tracking wrong: claims %d, challenges %d, bluffs %d",
			duke.Claims, duke.Challenges, duke.Bluffs)
	}
	if duke.ActionAttempts["Tax"] != 1 || duke.ActionSuccesses["Tax"] != 1 {
		t.Errorf("Duke Tax tracking wrong: %v / %v", duke.ActionAttempts, duke.ActionSuccesses)
	}

	captain := collector.CharacterStats["Captain"]
	// One action claim (the bluffed Steal) plus one block claim (the real block)
	if captain.Claims != 2 || captain.Bluffs != 1 || captain.BluffsCaught != 0 {
		t.Errorf("Captain claim tracking wrong: claims %d, bluffs %d, caught %d",
			captain.Claims, captain.Bluffs, captain.BluffsCaught)
	}
	if captain.Blocks != 1 || captain.BlockSuccesses != 1 {
		t.Errorf("Captain block tracking wrong: blocks %d, successes %d",
			captain.Blocks, captain.BlockSuccesses)
	}
	if captain.ActionAttempts["Steal"] != 1 || captain.ActionSuccesses["Steal"] != 0 {
		t.Errorf("Captain Steal tracking wrong: %v / %v", captain.ActionAttempts, captain.ActionSuccesses)
	}

	assassin := collector.CharacterStats["Assassin"]
	if assassin.TimesDealt != 1 || assassin.WinsWhenDealt != 0 {
		t.Errorf("Assassin dealt tracking wrong: %+v", assassin)
	}
	if assassin.TotalSurvivalTurns != 8 {
		t.Errorf("Assassin survival should be 8 turns, got %d", assassin.TotalSurvivalTurns)
	}

	// Verify rankings and per-player-count stats
	if len(collector.RankedCharacters) == 0 {
		t.Error("No character rankings generated")
	}
	pcStats := collector.GetStatisticsByPlayerCount()
	if pcStats[2] == nil || pcStats[2].AverageGameLength != 10 {
		t.Errorf("2-player average game length should be 10, got %+v", pcStats[2])
	}
	if pcStats[2].CharacterWinRates["Duke"] != 1.0 {
		t.Errorf("Duke 2-player dealt win rate should be 1.0, got %v", pcStats[2].CharacterWinRates["Duke"])
	}
}
