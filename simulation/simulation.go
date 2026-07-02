package simulation

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/computerscienceiscool/coup-game/game"
)

// Config holds simulation configuration
type Config struct {
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

// GameResult stores the outcome of a single game
type GameResult struct {
	ID                  int
	PlayerCount         int
	WinnerID            int
	WinnerCharacters    []string
	PlayerStartingCards map[int][]string // Starting characters for each player
	EliminationTurns    map[int]int      // playerID -> turn eliminated
	TotalTurns          int
	Actions             []game.ActionLog
	StartTime           time.Time
	EndTime             time.Time
}

// SimulationResults aggregates results from multiple games
type SimulationResults struct {
	Results          []GameResult
	Duration         time.Duration
	StartTime        time.Time
	EndTime          time.Time
	Config           Config
	ErrorCount       int // Number of games that failed to create or run
	CharacterStats   map[string]*CharacterStats
	PlayerCountStats map[int]*PlayerCountStats
}

// CharacterStats tracks metrics for a single character. "Dealt" figures are
// per player-slot (a player who was dealt the character at game start);
// claim figures cover both action claims and block claims of the character,
// with bluff counts taken from the ground-truth fields in the ActionLog.
type CharacterStats struct {
	Name string

	// Dealing and winning
	GamesDealt    int // games where the character appeared in any starting hand
	TimesDealt    int // player-slots dealt the character at game start
	WinsWhenDealt int // ...where that player went on to win the game
	FinalHandWins int // games whose winner ended the game holding the character

	// Claims (actions and blocks) with ground truth
	Claims       int // times the character was publicly claimed
	Challenges   int // ...and that claim was challenged
	Bluffs       int // claims made without holding the character
	BluffsCaught int // bluffed claims that were challenged (a challenge always catches a bluff)

	// Signature action (the action this character enables)
	ActionAttempts  map[string]int
	ActionSuccesses map[string]int

	// Blocks claiming this character
	Blocks         int // block attempts
	BlockSuccesses int // blocks that stopped the action

	// Survival of players dealt this character
	TotalSurvivalTurns  int     // total turns survived
	SurvivalFractionSum float64 // sum of (turns survived / game length) per dealt slot
}

// PlayerCountStats tracks metrics for games with specific player counts.
// CharacterWinRates holds each character's dealt win rate — P(win | dealt) —
// for games with this player count.
type PlayerCountStats struct {
	PlayerCount       int
	GamesPlayed       int
	AverageGameLength float64
	CharacterWinRates map[string]float64
}

// Simulator runs game simulations
type Simulator struct {
	Config             Config
	Results            []GameResult
	GameChannel        chan int
	ResultChannel      chan GameResult
	WaitGroup          sync.WaitGroup
	Progress           int64
	ErrorCount         int64
	TotalGames         int
	StartTime          time.Time
	ProgressInterval   time.Duration
	LastProgressUpdate time.Time
}

// NewSimulator creates a new simulator with the given configuration
func NewSimulator(config Config) *Simulator {
	return &Simulator{
		Config:           config,
		Results:          make([]GameResult, 0, config.TotalGames),
		GameChannel:      make(chan int, config.Workers*2),
		ResultChannel:    make(chan GameResult, config.Workers*2),
		TotalGames:       config.TotalGames,
		ProgressInterval: 2 * time.Second, // Update progress every 2 seconds
	}
}

// Run executes the simulation and returns the results
func (s *Simulator) Run() SimulationResults {
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
		Results:    s.Results,
		Duration:   endTime.Sub(s.StartTime),
		StartTime:  s.StartTime,
		EndTime:    endTime,
		Config:     s.Config,
		ErrorCount: int(atomic.LoadInt64(&s.ErrorCount)),
	}
}

// worker processes games from the game channel
func (s *Simulator) worker(workerID int) {
	defer s.WaitGroup.Done()

	for gameID := range s.GameChannel {
		// Per-game seed derived only from the base seed and game ID, so runs
		// are reproducible and no two games share an RNG stream regardless
		// of which worker picks them up
		gameSeed := game.MixSeed(s.Config.Seed, int64(gameID))
		gameRNG := rand.New(rand.NewSource(game.MixSeed(gameSeed, -1)))

		// Determine player count for this game
		playerCount := s.determinePlayerCount(gameID, gameRNG)

		// Create game based on AI mode
		var g *game.Game
		var err error

		switch s.Config.AIMode {
		case "original":
			g, err = game.NewGameWithOriginalAI(playerCount, gameSeed)
		case "mixed":
			g, err = game.NewGameWithMixedAIs(playerCount, gameSeed)
		default:
			var aiTypes []game.AIPlayerType
			if s.Config.CharacterBalance {
				aiTypes = s.generateBalancedAITypes(playerCount, gameID, gameRNG)
			}
			g, err = game.NewGameWithAITypes(playerCount, aiTypes, s.Config.CompetitiveLevel, gameSeed)
		}

		if err != nil {
			atomic.AddInt64(&s.ErrorCount, 1)
			continue
		}

		// Collect all players' starting cards before the game runs
		playerStartingCards := make(map[int][]string)
		for _, player := range g.Players {
			cards := make([]string, 0)
			for _, card := range player.GetInfluences() {
				cards = append(cards, card.Name)
			}
			playerStartingCards[player.GetID()] = cards
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
			ID:                  gameID,
			PlayerCount:         playerCount,
			WinnerID:            winner.GetID(),
			WinnerCharacters:    winnerChars,
			PlayerStartingCards: playerStartingCards,
			EliminationTurns:    g.EliminationTurns,
			TotalTurns:          g.Turn,
			Actions:             g.ActionLog,
			StartTime:           startTime,
			EndTime:             endTime,
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

// determinePlayerCount selects a player count based on the configuration
func (s *Simulator) determinePlayerCount(gameID int, rng *rand.Rand) int {
	// If specific counts are configured, use them
	if len(s.Config.GamesPerPlayerCount) > 0 {
		// Sort keys for deterministic iteration order
		counts := make([]int, 0, len(s.Config.GamesPerPlayerCount))
		for count := range s.Config.GamesPerPlayerCount {
			counts = append(counts, count)
		}
		sort.Ints(counts)

		// Create buckets for each player count
		totalAssigned := 0
		for _, count := range counts {
			totalAssigned += s.Config.GamesPerPlayerCount[count]
			if gameID < totalAssigned {
				return count
			}
		}
	}

	// Default: random between 2-6 players
	return rng.Intn(5) + 2
}

// submitGames sends game IDs to the game channel
func (s *Simulator) submitGames() {
	for i := 0; i < s.TotalGames; i++ {
		s.GameChannel <- i
	}
	close(s.GameChannel)
}

// collectResults gathers results from the result channel
func (s *Simulator) collectResults(done chan bool) {
	for result := range s.ResultChannel {
		s.Results = append(s.Results, result)
	}
	close(done)
}

// updateProgress displays the current simulation progress
func (s *Simulator) updateProgress(progress int64) {
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

// generateBalancedAITypes creates a balanced distribution of AI character preferences
func (s *Simulator) generateBalancedAITypes(playerCount int, gameID int, rng *rand.Rand) []game.AIPlayerType {
	aiTypes := make([]game.AIPlayerType, playerCount)

	// Ensure each character type appears roughly equally across games
	baseType := (gameID % 5) // 5 character types

	for i := 0; i < playerCount; i++ {
		aiTypes[i] = game.AIPlayerType((baseType + i) % 5)
	}

	// Shuffle the assignments to avoid bias
	for i := len(aiTypes) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		aiTypes[i], aiTypes[j] = aiTypes[j], aiTypes[i]
	}

	return aiTypes
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		return fmt.Sprintf("%ds", seconds)
	}
}
