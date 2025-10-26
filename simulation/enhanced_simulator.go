package simulation

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/computerscienceiscool/coup-game/game"
)

// EnhancedConfig extends the basic Config with options for AI behavior
type EnhancedConfig struct {
	TotalGames          int
	GamesPerPlayerCount map[int]int
	Workers             int
	Verbose             bool
	Seed                int64
	OutputDir           string
	AIMode              string
	CompetitiveLevel    game.CompetitiveLevel
	CharacterBalance    bool
}

// EnhancedSimulator extends the basic Simulator with support for different AI types
type EnhancedSimulator struct {
	Config             EnhancedConfig
	Results            []GameResult
	GameChannel        chan int
	ResultChannel      chan GameResult
	WaitGroup          sync.WaitGroup
	Progress           int64
	TotalGames         int
	StartTime          time.Time
	ProgressInterval   time.Duration
	LastProgressUpdate time.Time
}

// NewEnhancedSimulator creates a new simulator with the given configuration
func NewEnhancedSimulator(config EnhancedConfig) *EnhancedSimulator {
	return &EnhancedSimulator{
		Config:           config,
		Results:          make([]GameResult, 0, config.TotalGames),
		GameChannel:      make(chan int, config.Workers*2),
		ResultChannel:    make(chan GameResult, config.Workers*2),
		TotalGames:       config.TotalGames,
		ProgressInterval: 2 * time.Second, // Update progress every 2 seconds
	}
}

// Run executes the simulation and returns the results
func (s *EnhancedSimulator) Run() SimulationResults {
	s.StartTime = time.Now()
	s.LastProgressUpdate = s.StartTime

	// Start worker goroutines
	for i := 0; i < s.Config.Workers; i++ {
		s.WaitGroup.Add(1)
		go s.worker(i)
	}

	// Start result collector
	resultsDone := make(chan bool)
	go s.collectResults(resultsDone)

	// Start game submitter
	go s.submitGames()

	// Wait for all workers to finish
	s.WaitGroup.Wait()
	close(s.ResultChannel)

	// Wait for results to be collected
	<-resultsDone

	// Return simulation results
	endTime := time.Now()
	return SimulationResults{
		Results:   s.Results,
		Duration:  endTime.Sub(s.StartTime),
		StartTime: s.StartTime,
		EndTime:   endTime,
		Config: Config{
			TotalGames:          s.Config.TotalGames,
			GamesPerPlayerCount: s.Config.GamesPerPlayerCount,
			Workers:             s.Config.Workers,
			Verbose:             s.Config.Verbose,
			Seed:                s.Config.Seed,
			OutputDir:           s.Config.OutputDir,
		},
	}
}

// worker processes games from the game channel
func (s *EnhancedSimulator) worker(workerID int) {
	defer s.WaitGroup.Done()

	// Worker-specific RNG derived from the main seed
	workerSeed := s.Config.Seed + int64(workerID)
	rng := rand.New(rand.NewSource(workerSeed))

	for gameID := range s.GameChannel {
		// Determine player count for this game
		playerCount := s.determinePlayerCount(gameID, rng)

		// Generate game-specific seed
		gameSeed := workerSeed + int64(gameID)

		// Create and run game based on AI mode
		var g *game.Game
		var err error

		switch s.Config.AIMode {
		case "original":
			g, err = game.NewGameWithOriginalAI(playerCount, gameSeed)
		case "mixed":
			g, err = game.NewGameWithMixedAIs(playerCount, gameSeed)
		default:
			// Create game with all AIs at the same competitive level
			var aiTypes []game.AIPlayerType

			// If character balancing is enabled, create a balanced distribution of AI types
			if s.Config.CharacterBalance {
				aiTypes = s.generateBalancedAITypes(playerCount, gameID, rng)
			}

			g, err = game.NewGameWithAITypes(playerCount, aiTypes, s.Config.CompetitiveLevel, gameSeed)
		}

		if err != nil {
			fmt.Printf("Error creating game %d: %v\n", gameID, err)
			continue
		}

		startTime := time.Now()
		winner := g.RunToCompletion()
		endTime := time.Now()

		// Collect winner's characters
		winnerChars := make([]string, 0)
		for _, card := range winner.GetInfluences() {
			winnerChars = append(winnerChars, card.Name)
		}

		// Create game result
		result := GameResult{
			ID:               gameID,
			PlayerCount:      playerCount,
			WinnerID:         winner.GetID(),
			WinnerCharacters: winnerChars,
			TotalTurns:       g.Turn,
			Actions:          g.ActionLog,
			StartTime:        startTime,
			EndTime:          endTime,
		}

		// Send result
		s.ResultChannel <- result

		// Update progress
		newProgress := atomic.AddInt64(&s.Progress, 1)
		if s.Config.Verbose && time.Since(s.LastProgressUpdate) > s.ProgressInterval {
			s.updateProgress(newProgress)
		}
	}
}

// generateBalancedAITypes creates a balanced distribution of AI character preferences
func (s *EnhancedSimulator) generateBalancedAITypes(playerCount int, gameID int, rng *rand.Rand) []game.AIPlayerType {
	aiTypes := make([]game.AIPlayerType, playerCount)

	// Ensure each character type appears roughly equally across games
	// We'll use the game ID to cycle through character types
	baseType := (gameID % 5) // 5 character types

	for i := 0; i < playerCount; i++ {
		// Assign character type in a round-robin fashion
		aiTypes[i] = game.AIPlayerType((baseType + i) % 5)
	}

	// Shuffle the assignments to avoid bias
	for i := len(aiTypes) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		aiTypes[i], aiTypes[j] = aiTypes[j], aiTypes[i]
	}

	return aiTypes
}

// determinePlayerCount selects a player count based on the configuration
func (s *EnhancedSimulator) determinePlayerCount(gameID int, rng *rand.Rand) int {
	// If specific counts are configured, use them
	if len(s.Config.GamesPerPlayerCount) > 0 {
		// Create buckets for each player count
		totalAssigned := 0
		for count, games := range s.Config.GamesPerPlayerCount {
			totalAssigned += games
			if gameID < totalAssigned {
				return count
			}
		}
	}

	// Default: random between 2-6 players
	return rng.Intn(5) + 2
}

// submitGames sends game IDs to the game channel
func (s *EnhancedSimulator) submitGames() {
	for i := 0; i < s.TotalGames; i++ {
		s.GameChannel <- i
	}
	close(s.GameChannel)
}

// collectResults gathers results from the result channel
func (s *EnhancedSimulator) collectResults(done chan bool) {
	for result := range s.ResultChannel {
		s.Results = append(s.Results, result)
	}
	close(done)
}

// updateProgress displays the current simulation progress
func (s *EnhancedSimulator) updateProgress(progress int64) {
	s.LastProgressUpdate = time.Now()
	elapsedSeconds := time.Since(s.StartTime).Seconds()
	gamesPerSecond := float64(progress) / elapsedSeconds

	percentComplete := float64(progress) * 100.0 / float64(s.TotalGames)

	// Calculate ETA
	var eta string
	if gamesPerSecond > 0 {
		remainingGames := s.TotalGames - int(progress)
		remainingSeconds := float64(remainingGames) / gamesPerSecond
		eta = formatDuration(time.Duration(remainingSeconds) * time.Second)
	} else {
		eta = "unknown"
	}

	fmt.Printf("\r[%.2f%%] %d/%d games | %.2f games/sec | ETA: %s | AI: %s",
		percentComplete, progress, s.TotalGames, gamesPerSecond, eta, s.Config.AIMode)
}
