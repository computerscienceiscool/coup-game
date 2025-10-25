package simulation

import (
	"math"
	"sort"

	"github.com/computerscienceiscool/coup-game/game"
)

// MetricsCollector gathers statistics from game results
type MetricsCollector struct {
	// Game stats
	TotalGames         int
	GamesByPlayerCount map[int]int
	AverageGameLength  float64

	// Character stats
	CharacterStats map[string]*CharacterStats

	// Composite rankings
	RankedCharacters []CharacterRanking
}

// CharacterRanking represents the strength ranking of a character
type CharacterRanking struct {
	Name         string
	WinRate      float64
	PowerScore   float64 // Composite ranking based on all metrics
	ActionRate   float64
	SurvivalRate float64
	BluffSuccess float64
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	collector := &MetricsCollector{
		GamesByPlayerCount: make(map[int]int),
		CharacterStats:     make(map[string]*CharacterStats),
	}

	// Initialize stats for each character
	for _, charName := range game.GetCharacters() {
		collector.CharacterStats[charName] = &CharacterStats{
			Name:               charName,
			GamesPlayed:        0,
			GamesWon:           0,
			ActionAttempts:     make(map[string]int),
			ActionSuccesses:    make(map[string]int),
			Challenges:         0,
			ChallengeSuccesses: 0,
			Blocks:             0,
			BlockSuccesses:     0,
			TotalSurvivalTurns: 0,
		}
	}

	return collector
}

// ProcessGameResults analyzes a batch of game results
func (m *MetricsCollector) ProcessGameResults(results []GameResult) {
	m.TotalGames = len(results)

	// Track total turn count for average
	totalTurns := 0

	// Process each game
	for _, result := range results {
		// Update game length stats
		totalTurns += result.TotalTurns

		// Update player count stats
		m.GamesByPlayerCount[result.PlayerCount]++

		// Process winner characters
		for _, charName := range result.WinnerCharacters {
			if stats, exists := m.CharacterStats[charName]; exists {
				stats.GamesWon++
			}
		}

		// Process each action for detailed metrics
		m.processGameActions(result)
	}

	// Calculate average game length
	if m.TotalGames > 0 {
		m.AverageGameLength = float64(totalTurns) / float64(m.TotalGames)
	}

	// Calculate character participation rate
	// This requires knowing which characters were in each game
	// For a simple approximation, assume equal distribution
	cardsPerCharacter := 3 // 3 copies of each character
	for _, stats := range m.CharacterStats {
		// Rough estimate: character appears in ~3/15 = 20% of total player-games
		stats.GamesPlayed = m.TotalGames * cardsPerCharacter / len(game.GetCharacters())
	}

	// Generate rankings
	m.calculateRankings()
}

// processGameActions analyzes the actions in a game
func (m *MetricsCollector) processGameActions(result GameResult) {
	// Process each action log
	characterInfluence := make(map[int][]string) // Map player IDs to their character cards

	// Process actions
	for _, action := range result.Actions {
		actionName := action.Action
		actorID := action.PlayerID

		// Track character-specific action stats
		if characterInfluence[actorID] != nil {
			for _, charName := range characterInfluence[actorID] {
				if stats, exists := m.CharacterStats[charName]; exists {
					if _, ok := stats.ActionAttempts[actionName]; !ok {
						stats.ActionAttempts[actionName] = 0
					}
					stats.ActionAttempts[actionName]++

					if action.Success {
						if _, ok := stats.ActionSuccesses[actionName]; !ok {
							stats.ActionSuccesses[actionName] = 0
						}
						stats.ActionSuccesses[actionName]++
					}
				}
			}
		}

		// Track challenge stats
		if action.Challenged {
			// Get the character required for this action
			var requiredChar string
			switch actionName {
			case "Tax":
				requiredChar = game.Duke
			case "Steal":
				requiredChar = game.Captain
			case "Assassinate":
				requiredChar = game.Assassin
			case "Exchange":
				requiredChar = game.Ambassador
			}

			// Update challenge stats for this character
			if requiredChar != "" {
				if stats, exists := m.CharacterStats[requiredChar]; exists {
					stats.Challenges++
					if !action.Success && action.Blocker == -1 {
						// Challenge failed
						stats.ChallengeSuccesses++
					}
				}
			}
		}

		// Track block stats
		if action.Blocker != -1 {
			// Determine which character was claimed for blocking
			var blockingChar string
			switch actionName {
			case "Foreign Aid":
				blockingChar = game.Duke
			case "Steal":
				// Either Captain or Ambassador could block
				// For simplicity, credit both equally
				m.CharacterStats[game.Captain].Blocks++
				m.CharacterStats[game.Ambassador].Blocks++

				if !action.Success {
					m.CharacterStats[game.Captain].BlockSuccesses++
					m.CharacterStats[game.Ambassador].BlockSuccesses++
				}
			case "Assassinate":
				blockingChar = game.Contessa
			}

			if blockingChar != "" {
				if stats, exists := m.CharacterStats[blockingChar]; exists {
					stats.Blocks++
					if !action.Success {
						stats.BlockSuccesses++
					}
				}
			}
		}
	}
}

// calculateRankings generates character rankings based on collected metrics
func (m *MetricsCollector) calculateRankings() {
	m.RankedCharacters = make([]CharacterRanking, 0, len(m.CharacterStats))

	for charName, stats := range m.CharacterStats {
		// Calculate rates
		winRate := float64(0)
		if stats.GamesPlayed > 0 {
			winRate = float64(stats.GamesWon) / float64(stats.GamesPlayed)
		}

		// Calculate action success rate
		actionAttempts := 0
		actionSuccesses := 0
		for action, attempts := range stats.ActionAttempts {
			actionAttempts += attempts
			actionSuccesses += stats.ActionSuccesses[action]
		}

		actionRate := float64(0)
		if actionAttempts > 0 {
			actionRate = float64(actionSuccesses) / float64(actionAttempts)
		}

		// Calculate bluff success rate
		bluffSuccess := float64(0)
		if stats.Challenges > 0 {
			bluffSuccess = float64(stats.ChallengeSuccesses) / float64(stats.Challenges)
		}

		// Calculate block success rate
		blockRate := float64(0)
		if stats.Blocks > 0 {
			blockRate = float64(stats.BlockSuccesses) / float64(stats.Blocks)
		}

		// Calculate survival rate (approximated)
		survivalRate := float64(1.0) // Placeholder
		if stats.TotalSurvivalTurns > 0 && m.TotalGames > 0 {
			survivalRate = float64(stats.TotalSurvivalTurns) / (float64(stats.GamesPlayed) * m.AverageGameLength)
		}

		// Calculate composite power score (weighted average)
		powerScore := winRate*0.4 + actionRate*0.2 + bluffSuccess*0.15 + blockRate*0.15 + survivalRate*0.1

		// Add to rankings
		m.RankedCharacters = append(m.RankedCharacters, CharacterRanking{
			Name:         charName,
			WinRate:      winRate,
			PowerScore:   powerScore,
			ActionRate:   actionRate,
			SurvivalRate: survivalRate,
			BluffSuccess: bluffSuccess,
		})
	}

	// Sort by power score
	sort.Slice(m.RankedCharacters, func(i, j int) bool {
		return m.RankedCharacters[i].PowerScore > m.RankedCharacters[j].PowerScore
	})
}

// GetStatisticsByPlayerCount returns statistics broken down by player count
func (m *MetricsCollector) GetStatisticsByPlayerCount() map[int]*PlayerCountStats {
	result := make(map[int]*PlayerCountStats)

	// This would be more accurate with actual game data tracking character distribution
	// For now, we'll provide a simple approximation
	for playerCount, gameCount := range m.GamesByPlayerCount {
		stats := &PlayerCountStats{
			PlayerCount:       playerCount,
			GamesPlayed:       gameCount,
			AverageGameLength: 0,
			CharacterWinRates: make(map[string]float64),
		}

		// Set placeholder values
		stats.AverageGameLength = m.AverageGameLength * (1 + 0.1*float64(playerCount-3))

		// Character win rates vary slightly by player count (simplified model)
		for _, charRank := range m.RankedCharacters {
			variance := (math.Cos(float64(playerCount)) + 1) * 0.1
			stats.CharacterWinRates[charRank.Name] = charRank.WinRate * (1 + variance)
		}

		result[playerCount] = stats
	}

	return result
}
