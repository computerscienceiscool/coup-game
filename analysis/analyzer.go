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

// CharacterStatistics holds detailed statistics for a character. See
// simulation.MetricsCollector for the metric definitions.
type CharacterStatistics struct {
	Name              string
	DealtWinRate      float64 // P(win | dealt this character at game start)
	FinalHandWinRate  float64 // share of games whose winner ended holding it
	ActionSuccessRate float64 // signature action success rate
	BlockSuccessRate  float64 // blocks claiming it that stopped the action
	BluffRate         float64 // share of its claims made without the card
	BluffSuccessRate  float64 // bluffed claims that went unchallenged
	ChallengedRate    float64 // share of its claims that were challenged
	SurvivalTime      float64 // average turns survived by players dealt it
	SurvivalRate      float64 // average fraction of the game survived
	TimesDealt        int     // player-slots dealt this character
	WinsWhenDealt     int     // ...that went on to win
}

// PlayerCountStatistics holds statistics for games with a specific player count
type PlayerCountStatistics struct {
	PlayerCount       int
	GamesPlayed       int
	AverageGameLength float64            // measured per player count
	CharacterWinRates map[string]float64 // dealt win rate per character
}

// CharacterRanking represents a character's strength ranking
type CharacterRanking struct {
	Name       string
	PowerScore float64
	WinRate    float64 // dealt win rate
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
	totalGames := len(a.Results.Results)

	for name, metrics := range a.Metrics.CharacterStats {
		// Dealt win rate
		dealtWinRate := float64(0)
		if metrics.TimesDealt > 0 {
			dealtWinRate = float64(metrics.WinsWhenDealt) / float64(metrics.TimesDealt)
		}

		// Final-hand win rate
		finalHandWinRate := float64(0)
		if totalGames > 0 {
			finalHandWinRate = float64(metrics.FinalHandWins) / float64(totalGames)
		}

		// Signature action success rate
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

		// Claim-derived rates, from ground truth
		bluffRate := float64(0)
		challengedRate := float64(0)
		if metrics.Claims > 0 {
			bluffRate = float64(metrics.Bluffs) / float64(metrics.Claims)
			challengedRate = float64(metrics.Challenges) / float64(metrics.Claims)
		}
		bluffSuccessRate := float64(0)
		if metrics.Bluffs > 0 {
			bluffSuccessRate = float64(metrics.Bluffs-metrics.BluffsCaught) / float64(metrics.Bluffs)
		}

		// Block success rate
		blockRate := float64(0)
		if metrics.Blocks > 0 {
			blockRate = float64(metrics.BlockSuccesses) / float64(metrics.Blocks)
		}

		// Survival
		survivalTime := float64(0)
		survivalRate := float64(0)
		if metrics.TimesDealt > 0 {
			survivalTime = float64(metrics.TotalSurvivalTurns) / float64(metrics.TimesDealt)
			survivalRate = metrics.SurvivalFractionSum / float64(metrics.TimesDealt)
		}

		a.CharacterMap[name] = &CharacterStatistics{
			Name:              name,
			DealtWinRate:      dealtWinRate,
			FinalHandWinRate:  finalHandWinRate,
			ActionSuccessRate: actionRate,
			BlockSuccessRate:  blockRate,
			BluffRate:         bluffRate,
			BluffSuccessRate:  bluffSuccessRate,
			ChallengedRate:    challengedRate,
			SurvivalTime:      survivalTime,
			SurvivalRate:      survivalRate,
			TimesDealt:        metrics.TimesDealt,
			WinsWhenDealt:     metrics.WinsWhenDealt,
		}
	}
}

// analyzePlayerCounts examines the impact of player count on game dynamics
func (a *Analyzer) analyzePlayerCounts() {
	for count, stats := range a.Metrics.GetStatisticsByPlayerCount() {
		a.PlayerCountMap[count] = &PlayerCountStatistics{
			PlayerCount:       count,
			GamesPlayed:       stats.GamesPlayed,
			AverageGameLength: stats.AverageGameLength,
			CharacterWinRates: stats.CharacterWinRates,
		}
	}
}

// calculateSignificance performs a chi-squared goodness-of-fit test.
// Null hypothesis: the probability of winning is independent of which
// character a player is dealt — expected wins for each character are
// proportional to how often it was dealt. Returns an approximate p-value.
//
// Caveat: dealt player-slots are not fully independent observations (each
// player holds two characters and each game contributes several slots), so
// treat this as an approximation, not exact inference.
func (a *Analyzer) calculateSignificance() float64 {
	characters := len(a.Metrics.CharacterStats)
	if characters < 2 {
		return 1.0
	}

	totalDealt := 0
	totalWins := 0
	for _, stats := range a.Metrics.CharacterStats {
		totalDealt += stats.TimesDealt
		totalWins += stats.WinsWhenDealt
	}
	if totalDealt == 0 || totalWins == 0 {
		return 1.0
	}

	overallRate := float64(totalWins) / float64(totalDealt)

	// Chi-squared statistic: sum((observed - expected)^2 / expected)
	chiSquared := 0.0
	for _, stats := range a.Metrics.CharacterStats {
		expected := float64(stats.TimesDealt) * overallRate
		if expected == 0 {
			continue
		}
		diff := float64(stats.WinsWhenDealt) - expected
		chiSquared += (diff * diff) / expected
	}

	// Degrees of freedom = characters - 1
	df := float64(characters - 1)

	// Approximate p-value using the Wilson-Hilferty approximation
	// for the chi-squared CDF
	z := math.Pow(chiSquared/df, 1.0/3.0) - (1.0 - 2.0/(9.0*df))
	z = z / math.Sqrt(2.0/(9.0*df))

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
