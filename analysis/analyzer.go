package analysis

import (
	"math"
	"sort"

	"coup-game/simulation"
)

// StatisticsResult contains all analyzed data from the simulation
type StatisticsResult struct {
	TotalGames        int
	AverageGameLength float64
	CharacterStats    map[string]*CharacterStatistics
	PlayerCountStats  map[int]*PlayerCountStatistics
	RankedCharacters  []CharacterRanking
	SignificanceLevel float64 // p-value for statistical significance
}

// CharacterStatistics holds detailed statistics for a character
type CharacterStatistics struct {
	Name                 string
	WinRate              float64
	ActionSuccessRate    float64
	SurvivalTime         float64
	BluffSuccessRate     float64
	ChallengeSuccessRate float64
	BlockSuccessRate     float64
	TimesUsed            int
	TimesWon             int
}

// PlayerCountStatistics holds statistics for games with a specific player count
type PlayerCountStatistics struct {
	PlayerCount       int
	GamesPlayed       int
	AverageGameLength float64
	CharacterWinRates map[string]float64
	AverageTurnCount  float64
}

// CharacterRanking represents a character's strength ranking
type CharacterRanking struct {
	Name       string
	PowerScore float64
	WinRate    float64
}

// Analyzer processes simulation results to extract insights
type Analyzer struct {
	Results        simulation.SimulationResults
	Metrics        *simulation.MetricsCollector
	CharacterMap   map[string]*CharacterStatistics
	PlayerCountMap map[int]*PlayerCountStatistics
}

// NewAnalyzer creates a new analyzer with the given simulation results
func NewAnalyzer(results simulation.SimulationResults) *Analyzer {
	return &Analyzer{
		Results:        results,
		Metrics:        simulation.NewMetricsCollector(),
		CharacterMap:   make(map[string]*CharacterStatistics),
		PlayerCountMap: make(map[int]*PlayerCountStatistics),
	}
}

// AnalyzeResults performs comprehensive analysis of simulation results
func (a *Analyzer) AnalyzeResults() *StatisticsResult {
	// Process results with metrics collector
	a.Metrics.ProcessGameResults(a.Results.Results)

	// Extract character statistics
	a.analyzeCharacters()

	// Analyze player count impacts
	a.analyzePlayerCounts()

	// Calculate statistical significance
	significanceLevel := a.calculateSignificance()

	// Build and return complete statistics
	return &StatisticsResult{
		TotalGames:        len(a.Results.Results),
		AverageGameLength: a.Metrics.AverageGameLength,
		CharacterStats:    a.CharacterMap,
		PlayerCountStats:  a.PlayerCountMap,
		RankedCharacters:  a.convertRankings(a.Metrics.RankedCharacters),
		SignificanceLevel: significanceLevel,
	}
}

// analyzeCharacters extracts detailed character statistics
func (a *Analyzer) analyzeCharacters() {
	// Process each character from metrics
	for name, metrics := range a.Metrics.CharacterStats {
		// Calculate action success rate
		actionAttempts := 0
		actionSuccesses := 0
		for action, attempts := range metrics.ActionAttempts {
			actionAttempts += attempts
			actionSuccesses += metrics.ActionSuccesses[action]
		}

		actionRate := float64(0)
		if actionAttempts > 0 {
			actionRate = float64(actionSuccesses) / float64(actionAttempts)
		}

		// Calculate bluff success rate
		bluffRate := float64(0)
		if metrics.Challenges > 0 {
			bluffRate = float64(metrics.ChallengeSuccesses) / float64(metrics.Challenges)
		}

		// Calculate block success rate
		blockRate := float64(0)
		if metrics.Blocks > 0 {
			blockRate = float64(metrics.BlockSuccesses) / float64(metrics.Blocks)
		}

		// Win rate
		winRate := float64(0)
		if metrics.GamesPlayed > 0 {
			winRate = float64(metrics.GamesWon) / float64(metrics.GamesPlayed)
		}

		// Create character statistics
		a.CharacterMap[name] = &CharacterStatistics{
			Name:                 name,
			WinRate:              winRate,
			ActionSuccessRate:    actionRate,
			SurvivalTime:         float64(metrics.TotalSurvivalTurns) / math.Max(1, float64(metrics.GamesPlayed)),
			BluffSuccessRate:     bluffRate,
			ChallengeSuccessRate: 1.0 - bluffRate, // Inverse of bluff success rate
			BlockSuccessRate:     blockRate,
			TimesUsed:            metrics.GamesPlayed,
			TimesWon:             metrics.GamesWon,
		}
	}
}

// analyzePlayerCounts examines the impact of player count on game dynamics
func (a *Analyzer) analyzePlayerCounts() {
	// Get player count statistics from metrics
	pcStats := a.Metrics.GetStatisticsByPlayerCount()

	// Convert to our internal format
	for count, stats := range pcStats {
		a.PlayerCountMap[count] = &PlayerCountStatistics{
			PlayerCount:       count,
			GamesPlayed:       stats.GamesPlayed,
			AverageGameLength: stats.AverageGameLength,
			CharacterWinRates: stats.CharacterWinRates,
			AverageTurnCount:  stats.AverageGameLength,
		}
	}

	// Calculate more detailed metrics by iterating through results
	countTurns := make(map[int]int)
	countGames := make(map[int]int)

	for _, result := range a.Results.Results {
		playerCount := result.PlayerCount
		countTurns[playerCount] += result.TotalTurns
		countGames[playerCount]++
	}

	// Update average turn counts
	for count, stats := range a.PlayerCountMap {
		if countGames[count] > 0 {
			stats.AverageTurnCount = float64(countTurns[count]) / float64(countGames[count])
		}
	}
}

// calculateSignificance determines if character differences are statistically significant
func (a *Analyzer) calculateSignificance() float64 {
	// For a simple approach, we'll check if there's enough data for statistical significance
	// In a real implementation, we would use hypothesis testing (chi-squared, t-test, etc.)

	totalGames := len(a.Results.Results)
	characters := len(a.Metrics.CharacterStats)

	// We want at least 100 games per character for good confidence
	if totalGames < characters*100 {
		return 0.05 // Lower confidence if not enough games
	}

	// Check if win rates have enough spread
	winRates := make([]float64, 0, len(a.CharacterMap))
	for _, stats := range a.CharacterMap {
		winRates = append(winRates, stats.WinRate)
	}

	sort.Float64s(winRates)
	spread := winRates[len(winRates)-1] - winRates[0]

	if spread < 0.05 {
		return 0.10 // Less confidence if win rates are very close
	} else if spread < 0.15 {
		return 0.05 // Moderate confidence
	} else {
		return 0.01 // High confidence with large differences
	}
}

// convertRankings converts the metrics rankings to our format
func (a *Analyzer) convertRankings(rankings []simulation.CharacterRanking) []CharacterRanking {
	result := make([]CharacterRanking, len(rankings))

	for i, rank := range rankings {
		result[i] = CharacterRanking{
			Name:       rank.Name,
			PowerScore: rank.PowerScore,
			WinRate:    rank.WinRate,
		}
	}

	return result
}
