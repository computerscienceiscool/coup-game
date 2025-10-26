package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/computerscienceiscool/coup-game/analysis"
	"github.com/computerscienceiscool/coup-game/game"
	"github.com/computerscienceiscool/coup-game/simulation"
)

var (
	games            = flag.Int("games", 1000000, "Total games to simulate")
	workers          = flag.Int("workers", runtime.NumCPU(), "Parallel workers")
	verbose          = flag.Bool("v", false, "Verbose output")
	profile          = flag.String("profile", "", "Enable profiling (cpu|mem)")
	seed             = flag.Int64("seed", 0, "Random seed (0 for random)")
	output           = flag.String("output", "./results", "Output directory")
	aiMode           = flag.String("ai", "mixed", "AI mode: original, mixed, high, medium, low")
	characterBalance = flag.Bool("balance", false, "Enable character balancing (equal distribution)")
)

func main() {
	flag.Parse()

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*output, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	// If seed is 0, generate a random seed based on current time
	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}

	fmt.Printf("Starting Coup simulation with %d games using %d workers\n", *games, *workers)
	fmt.Printf("Random seed: %d\n", *seed)
	fmt.Printf("AI mode: %s\n", *aiMode)

	// Evenly distribute games across player counts (2-6 players)
	gamesPerCount := *games / 5
	gamesMap := map[int]int{
		2: gamesPerCount,
		3: gamesPerCount,
		4: gamesPerCount,
		5: gamesPerCount,
		6: *games - 4*gamesPerCount, // Assign remainder to 6-player games
	}

	// Set the AI competitive level based on command line flag
	var competitiveLevel game.CompetitiveLevel
	switch *aiMode {
	case "high":
		competitiveLevel = game.HighCompetitive
		fmt.Println("Using High Competitive AI players")
	case "medium":
		competitiveLevel = game.MediumCompetitive
		fmt.Println("Using Medium Competitive AI players")
	case "low":
		competitiveLevel = game.LowCompetitive
		fmt.Println("Using Low Competitive AI players")
	case "original":
		fmt.Println("Using Original AI behavior (30% bluff, 50% challenge)")
	case "mixed":
		fmt.Println("Using Mixed AI players with varied competitive levels")
	default:
		fmt.Println("Unknown AI mode, defaulting to mixed")
		*aiMode = "mixed"
	}

	// Configure the enhanced simulation
	config := simulation.EnhancedConfig{
		TotalGames:          *games,
		Workers:             *workers,
		Verbose:             *verbose,
		Seed:                *seed,
		OutputDir:           *output,
		GamesPerPlayerCount: gamesMap,
		AIMode:              *aiMode,
		CompetitiveLevel:    competitiveLevel,
		CharacterBalance:    *characterBalance,
	}

	// Run a single test game first to verify rules
	fmt.Println("Running test game...")
	if err := runTestGame(*seed, *aiMode, competitiveLevel); err != nil {
		fmt.Printf("Error during test game: %v\n", err)
		return
	}

	// Initialize the enhanced simulator
	fmt.Println("Initializing simulation...")
	simulator := simulation.NewEnhancedSimulator(config)

	// Run the simulation
	fmt.Println("Starting simulation...")
	startTime := time.Now()
	results := simulator.Run()
	duration := time.Since(startTime)

	// Analyze results
	fmt.Println("\nAnalyzing results...")
	analyzer := analysis.NewAnalyzer(results)
	stats := analyzer.AnalyzeResults()

	// Generate reports
	fmt.Println("Generating reports...")
	if err := analysis.GenerateCSVs(stats, *output); err != nil {
		fmt.Printf("Error generating CSV reports: %v\n", err)
	}

	// Print summary
	fmt.Println("\nSimulation Complete!")
	fmt.Printf("Processed %d games in %s\n", *games, duration)
	fmt.Printf("Games per second: %.2f\n", float64(*games)/duration.Seconds())

	// Print character rankings
	fmt.Println("\nCharacter Power Rankings:")
	for i, char := range stats.RankedCharacters {
		fmt.Printf("%d. %s - Win Rate: %.2f%%\n", i+1, char.Name, char.WinRate*100)
	}

	fmt.Printf("\nDetailed reports saved to %s\n", *output)
}

// runTestGame runs a single game to verify rules are working correctly
func runTestGame(seed int64, aiMode string, level game.CompetitiveLevel) error {
	var g *game.Game
	var err error

	// Create a game with 4 players based on AI mode
	switch aiMode {
	case "original":
		g, err = game.NewGameWithOriginalAI(4, seed)
	case "mixed":
		g, err = game.NewGameWithMixedAIs(4, seed)
	default:
		// Create game with all AIs at the same competitive level
		g, err = game.NewGameWithAITypes(4, nil, level, seed)
	}

	if err != nil {
		return fmt.Errorf("failed to create test game: %w", err)
	}

	// Run the game
	winner := g.RunToCompletion()

	// Print some details to verify
	fmt.Printf("Test game completed in %d turns\n", g.Turn)
	fmt.Printf("Winner: Player %d with %d coins\n", winner.GetID(), winner.GetCoins())

	// Display winner's cards
	winnerCards := make([]string, 0)
	for _, card := range winner.GetInfluences() {
		winnerCards = append(winnerCards, card.Name)
	}
	fmt.Printf("Winner's cards: %v\n", winnerCards)

	// Check some basic rules were followed
	for _, action := range g.ActionLog {
		// Check for forced coup at 10+ coins
		if action.Action != "Coup" && g.Players[action.PlayerID].GetCoins() >= 10 {
			return fmt.Errorf("rule violation: player %d did not coup with 10+ coins", action.PlayerID)
		}
	}

	fmt.Println("Test game rules verified successfully")
	return nil
}
