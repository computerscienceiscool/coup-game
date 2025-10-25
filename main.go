package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"coup-game/analysis"
	"coup-game/game"
	"coup-game/simulation"
)

var (
	games   = flag.Int("games", 1000000, "Total games to simulate")
	workers = flag.Int("workers", runtime.NumCPU(), "Parallel workers")
	verbose = flag.Bool("v", false, "Verbose output")
	profile = flag.String("profile", "", "Enable profiling (cpu|mem)")
	seed    = flag.Int64("seed", 0, "Random seed (0 for random)")
	output  = flag.String("output", "./results", "Output directory")
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

	// Evenly distribute games across player counts (2-6 players)
	gamesPerCount := *games / 5
	gamesMap := map[int]int{
		2: gamesPerCount,
		3: gamesPerCount,
		4: gamesPerCount,
		5: gamesPerCount,
		6: *games - 4*gamesPerCount, // Assign remainder to 6-player games
	}

	// Configure the simulation
	config := simulation.Config{
		TotalGames:          *games,
		Workers:             *workers,
		Verbose:             *verbose,
		Seed:                *seed,
		OutputDir:           *output,
		GamesPerPlayerCount: gamesMap,
	}

	// Run a single test game first to verify rules
	fmt.Println("Running test game...")
	if err := runTestGame(*seed); err != nil {
		fmt.Printf("Error during test game: %v\n", err)
		return
	}

	// Initialize the simulator
	fmt.Println("Initializing simulation...")
	simulator := simulation.NewSimulator(config)

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
func runTestGame(seed int64) error {
	// Create a game with 4 players
	g, err := game.NewGame(4, seed)
	if err != nil {
		return fmt.Errorf("failed to create test game: %w", err)
	}

	// Run the game
	winner := g.RunToCompletion()

	// Print some details to verify
	fmt.Printf("Test game completed in %d turns\n", g.Turn)
	fmt.Printf("Winner: Player %d with %d coins\n", winner.GetID(), winner.GetCoins())

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
