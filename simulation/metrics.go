package simulation

import (
	"sort"

	"github.com/computerscienceiscool/coup-game/game"
)

// MetricsCollector gathers statistics from game results.
//
// Metric definitions (also documented in docs/specification.md):
//   - Dealt win rate:      P(a player wins | they were dealt the character at game start)
//   - Final-hand win rate: share of games whose winner ended the game holding the character
//   - Action success rate: successes/attempts of the character's signature action (Tax, Steal, ...)
//   - Block success rate:  blocks claiming the character that actually stopped the action
//     (defeated blocks are counted as attempts — the log keeps them)
//   - Bluff rate:          share of the character's claims made without holding it (ground truth)
//   - Bluff success rate:  bluffed claims that went unchallenged (a challenged bluff is always caught)
type MetricsCollector struct {
	// Game stats
	TotalGames         int
	GamesByPlayerCount map[int]int
	TurnsByPlayerCount map[int]int
	AverageGameLength  float64

	// Character stats
	CharacterStats     map[string]*CharacterStats
	DealtByPlayerCount map[int]map[string]int // player-slots dealt, per character per player count
	WinsByPlayerCount  map[int]map[string]int // dealt-and-won, per character per player count

	// Composite rankings
	RankedCharacters []CharacterRanking
}

// CharacterRanking represents the strength ranking of a character
type CharacterRanking struct {
	Name         string
	WinRate      float64 // Dealt win rate: P(win | dealt the character)
	PowerScore   float64 // Composite ranking based on all metrics
	ActionRate   float64 // Signature action success rate
	SurvivalRate float64 // Average fraction of the game survived when dealt
	BluffSuccess float64 // Bluffed claims that went unchallenged
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	collector := &MetricsCollector{
		GamesByPlayerCount: make(map[int]int),
		TurnsByPlayerCount: make(map[int]int),
		CharacterStats:     make(map[string]*CharacterStats),
		DealtByPlayerCount: make(map[int]map[string]int),
		WinsByPlayerCount:  make(map[int]map[string]int),
	}

	// Initialize stats for each character
	for _, charName := range game.GetCharacters() {
		collector.CharacterStats[charName] = &CharacterStats{
			Name:            charName,
			ActionAttempts:  make(map[string]int),
			ActionSuccesses: make(map[string]int),
		}
	}

	return collector
}

// ProcessGameResults analyzes a batch of game results
func (m *MetricsCollector) ProcessGameResults(results []GameResult) {
	m.TotalGames = len(results)

	totalTurns := 0
	for _, result := range results {
		totalTurns += result.TotalTurns
		m.GamesByPlayerCount[result.PlayerCount]++
		m.TurnsByPlayerCount[result.PlayerCount] += result.TotalTurns

		// Final-hand wins: which characters the winner ended the game
		// holding (deduped — holding two copies still counts one game)
		finalHand := make(map[string]bool)
		for _, charName := range result.WinnerCharacters {
			finalHand[charName] = true
		}
		for charName := range finalHand {
			if stats, exists := m.CharacterStats[charName]; exists {
				stats.FinalHandWins++
			}
		}

		// Dealt characters: per player-slot (win rate, survival) and per
		// game (participation)
		seenInGame := make(map[string]bool)
		for playerID, playerCards := range result.PlayerStartingCards {
			dealt := make(map[string]bool)
			for _, charName := range playerCards {
				dealt[charName] = true
			}

			survivalTurns := result.TotalTurns // Default: survived the entire game
			if elimTurn, eliminated := result.EliminationTurns[playerID]; eliminated {
				survivalTurns = elimTurn
			}
			survivalFraction := 1.0
			if result.TotalTurns > 0 {
				survivalFraction = float64(survivalTurns) / float64(result.TotalTurns)
			}

			for charName := range dealt {
				seenInGame[charName] = true
				stats, exists := m.CharacterStats[charName]
				if !exists {
					continue
				}
				stats.TimesDealt++
				stats.TotalSurvivalTurns += survivalTurns
				stats.SurvivalFractionSum += survivalFraction
				if playerID == result.WinnerID {
					stats.WinsWhenDealt++
				}

				// Per-player-count dealt/win tallies
				if m.DealtByPlayerCount[result.PlayerCount] == nil {
					m.DealtByPlayerCount[result.PlayerCount] = make(map[string]int)
					m.WinsByPlayerCount[result.PlayerCount] = make(map[string]int)
				}
				m.DealtByPlayerCount[result.PlayerCount][charName]++
				if playerID == result.WinnerID {
					m.WinsByPlayerCount[result.PlayerCount][charName]++
				}
			}
		}
		for charName := range seenInGame {
			if stats, exists := m.CharacterStats[charName]; exists {
				stats.GamesDealt++
			}
		}

		// Process each action for claim/bluff/block metrics
		m.processGameActions(result)
	}

	// Calculate average game length
	if m.TotalGames > 0 {
		m.AverageGameLength = float64(totalTurns) / float64(m.TotalGames)
	}

	// Generate rankings
	m.calculateRankings()
}

// requiredCharacterFor maps a character action to the character it claims.
func requiredCharacterFor(action string) string {
	switch action {
	case "Tax":
		return game.Duke
	case "Steal":
		return game.Captain
	case "Assassinate":
		return game.Assassin
	case "Exchange":
		return game.Ambassador
	}
	return ""
}

// processGameActions tallies claim, bluff, action, and block statistics from
// a game's action log, using the log's ground-truth fields.
func (m *MetricsCollector) processGameActions(result GameResult) {
	for _, action := range result.Actions {
		// Character-action claims (Tax claims Duke, Steal claims Captain, ...)
		if claimed := requiredCharacterFor(action.Action); claimed != "" {
			if stats, exists := m.CharacterStats[claimed]; exists {
				stats.ActionAttempts[action.Action]++
				if action.Success {
					stats.ActionSuccesses[action.Action]++
				}

				stats.Claims++
				if action.Challenged {
					stats.Challenges++
				}
				if !action.ActorHadCard {
					stats.Bluffs++
					if action.Challenged {
						stats.BluffsCaught++
					}
				}
			}
		}

		// Block claims (every attempt is in the log, including defeated ones)
		if action.Blocker != -1 && action.BlockingCharacter != "" {
			if stats, exists := m.CharacterStats[action.BlockingCharacter]; exists {
				stats.Blocks++
				if action.BlockSucceeded {
					stats.BlockSuccesses++
				}

				stats.Claims++
				if action.BlockChallenged {
					stats.Challenges++
				}
				if !action.BlockerHadCard {
					stats.Bluffs++
					if action.BlockChallenged {
						stats.BluffsCaught++
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
		// Dealt win rate
		winRate := float64(0)
		if stats.TimesDealt > 0 {
			winRate = float64(stats.WinsWhenDealt) / float64(stats.TimesDealt)
		}

		// Signature action success rate
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

		// Bluff success rate: bluffed claims that went unchallenged
		bluffSuccess := float64(0)
		if stats.Bluffs > 0 {
			bluffSuccess = float64(stats.Bluffs-stats.BluffsCaught) / float64(stats.Bluffs)
		}

		// Block success rate
		blockRate := float64(0)
		if stats.Blocks > 0 {
			blockRate = float64(stats.BlockSuccesses) / float64(stats.Blocks)
		}

		// Average fraction of the game survived when dealt
		survivalRate := float64(0)
		if stats.TimesDealt > 0 {
			survivalRate = stats.SurvivalFractionSum / float64(stats.TimesDealt)
		}

		// Composite power score: winning when dealt dominates; the
		// character's utility (action/block success), bluffability, and
		// survivability are secondary. Weights are a documented judgment
		// call, not a fitted model.
		powerScore := winRate*0.6 + actionRate*0.15 + blockRate*0.1 + bluffSuccess*0.1 + survivalRate*0.05

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

// GetStatisticsByPlayerCount returns statistics broken down by player count,
// all measured from the game data (no approximations).
func (m *MetricsCollector) GetStatisticsByPlayerCount() map[int]*PlayerCountStats {
	result := make(map[int]*PlayerCountStats)

	for playerCount, gameCount := range m.GamesByPlayerCount {
		stats := &PlayerCountStats{
			PlayerCount:       playerCount,
			GamesPlayed:       gameCount,
			CharacterWinRates: make(map[string]float64),
		}

		if gameCount > 0 {
			stats.AverageGameLength = float64(m.TurnsByPlayerCount[playerCount]) / float64(gameCount)
		}

		for charName, dealt := range m.DealtByPlayerCount[playerCount] {
			if dealt > 0 {
				stats.CharacterWinRates[charName] = float64(m.WinsByPlayerCount[playerCount][charName]) / float64(dealt)
			}
		}

		result[playerCount] = stats
	}

	return result
}
