package analysis

import (
	"math"

	"github.com/computerscienceiscool/coup-game/simulation"
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

// calculateSignificance performs a chi-squared goodness-of-fit test to determine
// if character win rate differences are statistically significant.
// Null hypothesis: all characters win at equal rates.
// Returns an approximate p-value.
func (a *Analyzer) calculateSignificance() float64 {
	// Collect observed wins per character
	characters := len(a.Metrics.CharacterStats)
	if characters == 0 {
		return 1.0
	}

	totalWins := 0
	observedWins := make([]int, 0, characters)
	for _, stats := range a.Metrics.CharacterStats {
		observedWins = append(observedWins, stats.GamesWon)
		totalWins += stats.GamesWon
	}

	if totalWins == 0 {
		return 1.0
	}

	// Expected wins under null hypothesis (equal distribution)
	expectedWins := float64(totalWins) / float64(characters)

	// Calculate chi-squared statistic: sum((observed - expected)^2 / expected)
	chiSquared := 0.0
	for _, observed := range observedWins {
		diff := float64(observed) - expectedWins
		chiSquared += (diff * diff) / expectedWins
	}

	// Degrees of freedom = characters - 1
	df := float64(characters - 1)

	// Approximate p-value using the Wilson-Hilferty approximation
	// for the chi-squared CDF
	if df <= 0 {
		return 1.0
	}

	// Normalize chi-squared to approximate standard normal
	z := math.Pow(chiSquared/df, 1.0/3.0) - (1.0 - 2.0/(9.0*df))
	z = z / math.Sqrt(2.0/(9.0*df))

	// Approximate p-value from standard normal using complementary error function
	// P(Z > z) = 0.5 * erfc(z / sqrt(2))
	pValue := 0.5 * math.Erfc(z/math.Sqrt(2.0))

	// Clamp to [0, 1]
	if pValue < 0 {
		pValue = 0
	}
	if pValue > 1 {
		pValue = 1
	}

	return pValue
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
