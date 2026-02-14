package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"time"

	"github.com/computerscienceiscool/coup-game/analysis"
	"github.com/computerscienceiscool/coup-game/game"
	"github.com/computerscienceiscool/coup-game/simulation"
)

var (
	games            = flag.Int("games", 1000000, "Total games to simulate")
	workers          = flag.Int("workers", runtime.NumCPU(), "Parallel workers")
	verbose          = flag.Bool("v", false, "Verbose output")
	quiet            = flag.Bool("quiet", false, "Quiet mode - only show progress bar and final summary")
	profile          = flag.String("profile", "", "Enable profiling (cpu|mem)")
	seed             = flag.Int64("seed", 0, "Random seed (0 for random)")
	output           = flag.String("output", "./results", "Output directory")
	aiMode           = flag.String("ai", "mixed", "AI mode: original, mixed, high, medium, low")
	characterBalance = flag.Bool("balance", false, "Enable character balancing (equal distribution)")
	testComp         = flag.Int("test-comp", 0, "Run competitiveness test with specified number of games")
)

func main() {
	flag.Parse()

	// Handle profiling if requested
	if *profile != "" {
		switch *profile {
		case "cpu":
			f, err := os.Create("cpu.pprof")
			if err != nil {
				fmt.Printf("Error creating CPU profile: %v\n", err)
				return
			}
			defer f.Close()
			if err := pprof.StartCPUProfile(f); err != nil {
				fmt.Printf("Error starting CPU profile: %v\n", err)
				return
			}
			defer pprof.StopCPUProfile()
			if !*quiet {
				fmt.Println("CPU profiling enabled, writing to cpu.pprof")
			}
		case "mem":
			defer func() {
				f, err := os.Create("mem.pprof")
				if err != nil {
					fmt.Printf("Error creating memory profile: %v\n", err)
					return
				}
				defer f.Close()
				runtime.GC() // Get up-to-date statistics
				if err := pprof.WriteHeapProfile(f); err != nil {
					fmt.Printf("Error writing memory profile: %v\n", err)
				}
				if !*quiet {
					fmt.Println("Memory profile written to mem.pprof")
				}
			}()
			if !*quiet {
				fmt.Println("Memory profiling enabled, will write to mem.pprof on exit")
			}
		default:
			fmt.Printf("Unknown profile type: %s (use 'cpu' or 'mem')\n", *profile)
			return
		}
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*output, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	// If seed is 0, generate a random seed based on current time
	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}

	// Pass the verbose flag to the game package
	// If quiet mode is enabled, disable verbose output
	game.Verbose = *verbose && !*quiet

	// If test-comp flag is set, run competitiveness test and exit
	if *testComp > 0 {
		testCompetitiveLevels(*testComp, *seed)
		return
	}

	if !*quiet {
		fmt.Printf("Starting Coup simulation with %d games using %d workers\n", *games, *workers)
		fmt.Printf("Random seed: %d\n", *seed)
		fmt.Printf("AI mode: %s\n", *aiMode)
	}

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
		if !*quiet {
			fmt.Println("Using High Competitive AI players")
		}
	case "medium":
		competitiveLevel = game.MediumCompetitive
		if !*quiet {
			fmt.Println("Using Medium Competitive AI players")
		}
	case "low":
		competitiveLevel = game.LowCompetitive
		if !*quiet {
			fmt.Println("Using Low Competitive AI players")
		}
	case "original":
		if !*quiet {
			fmt.Println("Using Original AI behavior (30% bluff, 50% challenge)")
		}
	case "mixed":
		if !*quiet {
			fmt.Println("Using Mixed AI players with varied competitive levels")
		}
	default:
		if !*quiet {
			fmt.Println("Unknown AI mode, defaulting to mixed")
		}
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
	if !*quiet {
		fmt.Println("Running test game...")
	}
	if err := runTestGame(*seed, *aiMode, competitiveLevel, *quiet); err != nil {
		fmt.Printf("Error during test game: %v\n", err)
		return
	}

	// Initialize the enhanced simulator
	if !*quiet {
		fmt.Println("Initializing simulation...")
	}
	simulator := simulation.NewEnhancedSimulator(config)

	// Run the simulation
	if !*quiet {
		fmt.Println("Starting simulation...")
	}
	startTime := time.Now()
	results := simulator.Run()
	duration := time.Since(startTime)

	// Analyze results
	if !*quiet {
		fmt.Println("\nAnalyzing results...")
	}
	analyzer := analysis.NewAnalyzer(results)
	stats := analyzer.AnalyzeResults()

	// Generate reports
	if !*quiet {
		fmt.Println("Generating reports...")
	}
	if err := analysis.GenerateCSVs(stats, &results, *output); err != nil {
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
func runTestGame(seed int64, aiMode string, level game.CompetitiveLevel, quiet bool) error {
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

	// Display AI strategy details
	if !quiet {
		fmt.Println("\nTest Game Details:")
		fmt.Printf("Game Mode: %s, Player Count: %d\n", aiMode, len(g.Players))

		fmt.Println("\nPlayers and Strategies:")
		for i, p := range g.Players {
			// Show player's initial cards
			cards := make([]string, 0)
			for _, card := range p.GetInfluences() {
				cards = append(cards, card.Name)
			}

			// Detect player type and show strategy details
			if ep, ok := p.(*game.EnhancedAIPlayer); ok {
				// Enhanced AI - show competitive level
				levelName := "Unknown"
				switch ep.Strategy.Level {
				case game.LowCompetitive:
					levelName = "Low"
				case game.MediumCompetitive:
					levelName = "Medium"
				case game.HighCompetitive:
					levelName = "High"
				}

				// Get character preference
				charPref := "Random"
				if len(ep.Strategy.CharacterPreferences) > 0 {
					charPref = ep.Strategy.CharacterPreferences[0].Character
				}

				fmt.Printf("  Player %d: %s Competitive AI\n", i+1, levelName)
				fmt.Printf("    Preference: %s character\n", charPref)
				fmt.Printf("    Bluff Rate: %.0f%%\n", ep.Strategy.BluffRate*100)
				fmt.Printf("    Challenge Rate: %.0f%%\n", ep.Strategy.ChallengeRate*100)
				fmt.Printf("    Starting Cards: %v\n", cards)

			} else if _, ok := p.(*game.AIPlayer); ok {
				// Original AI
				fmt.Printf("  Player %d: Original AI\n", i+1)
				fmt.Printf("    Bluff Rate: 30%%\n")
				fmt.Printf("    Challenge Rate: 50%%\n")
				fmt.Printf("    Starting Cards: %v\n", cards)
			} else {
				// Unknown player type
				fmt.Printf("  Player %d: Unknown AI Type\n", i+1)
				fmt.Printf("    Starting Cards: %v\n", cards)
			}
		}
		fmt.Println()
	}

	// Run the game
	winner := g.RunToCompletion()

	// Print some details to verify
	if !quiet {
		fmt.Printf("Test game completed in %d turns\n", g.Turn)
		fmt.Printf("Winner: Player %d with %d coins\n", winner.GetID()+1, winner.GetCoins())

		// Show winner's strategy details
		if ep, ok := winner.(*game.EnhancedAIPlayer); ok {
			levelName := "Unknown"
			switch ep.Strategy.Level {
			case game.LowCompetitive:
				levelName = "Low"
			case game.MediumCompetitive:
				levelName = "Medium"
			case game.HighCompetitive:
				levelName = "High"
			}

			// Get preferred character
			charPref := "Random"
			if len(ep.Strategy.CharacterPreferences) > 0 {
				charPref = ep.Strategy.CharacterPreferences[0].Character
			}

			fmt.Printf("Winner strategy: %s Competitive with %s preference\n",
				levelName, charPref)
		} else if _, ok := winner.(*game.AIPlayer); ok {
			fmt.Printf("Winner strategy: Original AI\n")
		}

		// Display winner's cards
		winnerCards := make([]string, 0)
		for _, card := range winner.GetInfluences() {
			winnerCards = append(winnerCards, card.Name)
		}
		fmt.Printf("Winner's cards: %v\n", winnerCards)
	}

	// Check some basic rules were followed
	for _, action := range g.ActionLog {
		// Check for forced coup at 10+ coins
		if action.Action != "Coup" && g.Players[action.PlayerID].GetCoins() >= 10 {
			return fmt.Errorf("rule violation: player %d did not coup with 10+ coins", action.PlayerID+1)
		}
	}

	if !quiet {
		fmt.Println("Test game rules verified successfully")
	}
	return nil
}

// testCompetitiveLevels runs a series of games with mixed AI levels
// and tracks win rates by competitiveness level
func testCompetitiveLevels(numGames int, seed int64) {
	fmt.Println("\nRunning competitiveness level analysis...")
	fmt.Printf("Testing %d games with mixed AI levels...\n", numGames)

	// Track wins by level
	lowWins := 0
	mediumWins := 0
	highWins := 0
	originalWins := 0

	// Track character preferences by level
	lowCharWins := make(map[string]int)
	mediumCharWins := make(map[string]int)
	highCharWins := make(map[string]int)

	// Track actual cards held by winners
	lowCardWins := make(map[string]int)
	mediumCardWins := make(map[string]int)
	highCardWins := make(map[string]int)

	for i := 0; i < numGames; i++ {
		// Create a game with mixed AI
		g, err := game.NewGameWithMixedAIs(4, seed+int64(i))
		if err != nil {
			fmt.Printf("Error creating game: %v\n", err)
			continue
		}

		// Run the game
		winner := g.RunToCompletion()

		// Get winner's cards
		winnerCards := make([]string, 0)
		for _, card := range winner.GetInfluences() {
			winnerCards = append(winnerCards, card.Name)
		}

		// Record win by level
		if ep, ok := winner.(*game.EnhancedAIPlayer); ok {
			switch ep.Strategy.Level {
			case game.LowCompetitive:
				lowWins++
				// Record character preference
				if len(ep.Strategy.CharacterPreferences) > 0 {
					pref := ep.Strategy.CharacterPreferences[0].Character
					lowCharWins[pref]++
				}

				// Record actual cards
				for _, card := range winnerCards {
					lowCardWins[card]++
				}
			case game.MediumCompetitive:
				mediumWins++
				// Record character preference
				if len(ep.Strategy.CharacterPreferences) > 0 {
					pref := ep.Strategy.CharacterPreferences[0].Character
					mediumCharWins[pref]++
				}

				// Record actual cards
				for _, card := range winnerCards {
					mediumCardWins[card]++
				}
			case game.HighCompetitive:
				highWins++
				// Record character preference
				if len(ep.Strategy.CharacterPreferences) > 0 {
					pref := ep.Strategy.CharacterPreferences[0].Character
					highCharWins[pref]++
				}

				// Record actual cards
				for _, card := range winnerCards {
					highCardWins[card]++
				}
			}
		} else {
			originalWins++
		}

		// Print progress
		if (i+1)%50 == 0 {
			fmt.Printf("Completed %d games...\n", i+1)
		}
	}

	total := numGames

	// Print results
	fmt.Println("\nCOMPETITIVE LEVEL WIN RATES")
	fmt.Println("===========================")
	fmt.Printf("Low Competitive:    %d wins (%.2f%%)\n", lowWins, float64(lowWins)/float64(total)*100)
	fmt.Printf("Medium Competitive: %d wins (%.2f%%)\n", mediumWins, float64(mediumWins)/float64(total)*100)
	fmt.Printf("High Competitive:   %d wins (%.2f%%)\n", highWins, float64(highWins)/float64(total)*100)
	fmt.Printf("Original AI:        %d wins (%.2f%%)\n", originalWins, float64(originalWins)/float64(total)*100)

	// Print character preferences by level
	fmt.Println("\nCHARACTER PREFERENCE WIN RATES BY LEVEL")
	fmt.Println("=====================================")

	fmt.Println("\nLow Competitive Winners' Preferences:")
	printCharacterStats(lowCharWins, lowWins)

	fmt.Println("\nMedium Competitive Winners' Preferences:")
	printCharacterStats(mediumCharWins, mediumWins)

	fmt.Println("\nHigh Competitive Winners' Preferences:")
	printCharacterStats(highCharWins, highWins)

	// Print actual cards by level
	fmt.Println("\nACTUAL CARDS HELD BY WINNERS")
	fmt.Println("===========================")

	fmt.Println("\nLow Competitive Winners' Cards:")
	printCharacterStats(lowCardWins, lowWins*2) // *2 because each winner has 2 cards on average

	fmt.Println("\nMedium Competitive Winners' Cards:")
	printCharacterStats(mediumCardWins, mediumWins*2)

	fmt.Println("\nHigh Competitive Winners' Cards:")
	printCharacterStats(highCardWins, highWins*2)
}

// Helper function to print character stats
func printCharacterStats(charStats map[string]int, total int) {
	// Convert map to slice for sorting
	type charStat struct {
		name string
		wins int
	}

	stats := make([]charStat, 0, len(charStats))
	for name, wins := range charStats {
		stats = append(stats, charStat{name, wins})
	}

	// Sort by wins (descending)
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].wins > stats[j].wins
	})

	// Print stats
	for _, stat := range stats {
		if total > 0 {
			fmt.Printf("  %s: %d occurrences (%.2f%%)\n",
				stat.name, stat.wins, float64(stat.wins)/float64(total)*100)
		} else {
			fmt.Printf("  %s: %d occurrences (0.00%%)\n", stat.name, stat.wins)
		}
	}
}
