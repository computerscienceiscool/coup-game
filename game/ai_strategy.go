package game

import (
	"math/rand"
)

// CompetitiveLevel represents the skill level of an AI player
type CompetitiveLevel int

const (
	LowCompetitive CompetitiveLevel = iota
	MediumCompetitive
	HighCompetitive
)

// CharacterPreference indicates which character an AI prefers to play as
type CharacterPreference struct {
	Character       string
	PreferenceLevel float64 // 0.0-1.0 scale of how much this AI prefers this character
}

// EnhancedAIStrategy extends the basic AIStrategy with more sophisticated parameters
type EnhancedAIStrategy struct {
	// Basic strategy parameters
	BluffRate     float64 // Probability of bluffing (claiming a character they don't have)
	ChallengeRate float64 // Probability of challenging a claim
	AlwaysBlock   bool    // Whether to always block when possible

	// Advanced strategy parameters
	Level                CompetitiveLevel          // Skill level of this AI
	CharacterPreferences []CharacterPreference     // Which characters this AI prefers
	ActionPreferences    map[string]float64        // Preference for certain actions (0.0-1.0)
	TargetSelection      map[string]TargetStrategy // How to select targets for different actions
	CharacterBluffRates  map[string]float64        // Character-specific bluff rates
	CharacterBlockRates  map[string]float64        // Character-specific block rates
}

// TargetStrategy defines how an AI selects targets for actions
type TargetStrategy int

const (
	RandomTarget    TargetStrategy = iota // Choose targets randomly
	RichestTarget                         // Target player with most coins
	StrongestTarget                       // Target player with most influence
	ThreatTarget                          // Target player considered the biggest threat
)

// NewBasicAIStrategy creates a simple strategy with the given parameters
func NewBasicAIStrategy(bluffRate, challengeRate float64, alwaysBlock bool) *EnhancedAIStrategy {
	return &EnhancedAIStrategy{
		BluffRate:            bluffRate,
		ChallengeRate:        challengeRate,
		AlwaysBlock:          alwaysBlock,
		Level:                MediumCompetitive,
		CharacterPreferences: []CharacterPreference{},
		ActionPreferences:    make(map[string]float64),
		TargetSelection:      make(map[string]TargetStrategy),
		CharacterBluffRates:  make(map[string]float64),
		CharacterBlockRates:  make(map[string]float64),
	}
}

// CreateDukeStrategy creates an AI strategy focused on the Duke character
func CreateDukeStrategy(level CompetitiveLevel) *EnhancedAIStrategy {
	strategy := &EnhancedAIStrategy{
		Level: level,
		CharacterPreferences: []CharacterPreference{
			{Character: Duke, PreferenceLevel: 0.9},
		},
		ActionPreferences:   make(map[string]float64),
		TargetSelection:     make(map[string]TargetStrategy),
		CharacterBluffRates: make(map[string]float64),
		CharacterBlockRates: make(map[string]float64),
	}

	// Set up strategy parameters based on competitive level
	switch level {
	case HighCompetitive:
		strategy.BluffRate = DukeHighBluffRate
		strategy.ChallengeRate = DukeHighChallengeRate
		strategy.AlwaysBlock = true
		strategy.ActionPreferences["Tax"] = TaxHighPreference
		strategy.CharacterBluffRates[Duke] = DukeHighBluffRate
		strategy.CharacterBlockRates[Duke] = DukeHighBlockRate

	case MediumCompetitive:
		strategy.BluffRate = DukeMediumBluffRate
		strategy.ChallengeRate = DukeMediumChallengeRate
		strategy.AlwaysBlock = true
		strategy.ActionPreferences["Tax"] = TaxMediumPreference
		strategy.CharacterBluffRates[Duke] = DukeMediumBluffRate
		strategy.CharacterBlockRates[Duke] = DukeMediumBlockRate

	case LowCompetitive:
		strategy.BluffRate = DukeLowBluffRate
		strategy.ChallengeRate = DukeLowChallengeRate
		strategy.AlwaysBlock = false
		strategy.ActionPreferences["Tax"] = TaxLowPreference
		strategy.CharacterBluffRates[Duke] = DukeLowBluffRate
		strategy.CharacterBlockRates[Duke] = DukeLowBlockRate
	}

	return strategy
}

// CreateAssassinStrategy creates an AI strategy focused on the Assassin character
func CreateAssassinStrategy(level CompetitiveLevel) *EnhancedAIStrategy {
	strategy := &EnhancedAIStrategy{
		Level: level,
		CharacterPreferences: []CharacterPreference{
			{Character: Assassin, PreferenceLevel: 0.9},
		},
		ActionPreferences:   make(map[string]float64),
		TargetSelection:     make(map[string]TargetStrategy),
		CharacterBluffRates: make(map[string]float64),
		CharacterBlockRates: make(map[string]float64),
	}

	// Set up strategy parameters based on competitive level
	switch level {
	case HighCompetitive:
		strategy.BluffRate = AssassinHighBluffRate
		strategy.ChallengeRate = AssassinHighChallengeRate
		strategy.AlwaysBlock = true
		strategy.ActionPreferences["Assassinate"] = AssassinateHighPreference
		strategy.CharacterBluffRates[Assassin] = AssassinHighBluffRate
		strategy.TargetSelection["Assassinate"] = ThreatTarget

	case MediumCompetitive:
		strategy.BluffRate = AssassinMediumBluffRate
		strategy.ChallengeRate = AssassinMediumChallengeRate
		strategy.AlwaysBlock = true
		strategy.ActionPreferences["Assassinate"] = AssassinateMediumPreference
		strategy.CharacterBluffRates[Assassin] = AssassinMediumBluffRate
		strategy.TargetSelection["Assassinate"] = StrongestTarget

	case LowCompetitive:
		strategy.BluffRate = AssassinLowBluffRate
		strategy.ChallengeRate = AssassinLowChallengeRate
		strategy.AlwaysBlock = false
		strategy.ActionPreferences["Assassinate"] = AssassinateLowPreference
		strategy.CharacterBluffRates[Assassin] = AssassinLowBluffRate
		strategy.TargetSelection["Assassinate"] = RandomTarget
	}

	return strategy
}

// CreateCaptainStrategy creates an AI strategy focused on the Captain character
func CreateCaptainStrategy(level CompetitiveLevel) *EnhancedAIStrategy {
	strategy := &EnhancedAIStrategy{
		Level: level,
		CharacterPreferences: []CharacterPreference{
			{Character: Captain, PreferenceLevel: 0.9},
		},
		ActionPreferences:   make(map[string]float64),
		TargetSelection:     make(map[string]TargetStrategy),
		CharacterBluffRates: make(map[string]float64),
		CharacterBlockRates: make(map[string]float64),
	}

	// Set up strategy parameters based on competitive level
	switch level {
	case HighCompetitive:
		strategy.BluffRate = CaptainHighBluffRate
		strategy.ChallengeRate = CaptainHighChallengeRate
		strategy.AlwaysBlock = true
		strategy.ActionPreferences["Steal"] = StealHighPreference
		strategy.CharacterBluffRates[Captain] = CaptainHighBluffRate
		strategy.CharacterBlockRates[Captain] = CaptainHighBlockRate
		strategy.TargetSelection["Steal"] = RichestTarget

	case MediumCompetitive:
		strategy.BluffRate = CaptainMediumBluffRate
		strategy.ChallengeRate = CaptainMediumChallengeRate
		strategy.AlwaysBlock = true
		strategy.ActionPreferences["Steal"] = StealMediumPreference
		strategy.CharacterBluffRates[Captain] = CaptainMediumBluffRate
		strategy.CharacterBlockRates[Captain] = CaptainMediumBlockRate
		strategy.TargetSelection["Steal"] = RichestTarget

	case LowCompetitive:
		strategy.BluffRate = CaptainLowBluffRate
		strategy.ChallengeRate = CaptainLowChallengeRate
		strategy.AlwaysBlock = false
		strategy.ActionPreferences["Steal"] = StealLowPreference
		strategy.CharacterBluffRates[Captain] = CaptainLowBluffRate
		strategy.CharacterBlockRates[Captain] = CaptainLowBlockRate
		strategy.TargetSelection["Steal"] = RandomTarget
	}

	return strategy
}

// CreateAmbassadorStrategy creates an AI strategy focused on the Ambassador character
func CreateAmbassadorStrategy(level CompetitiveLevel) *EnhancedAIStrategy {
	strategy := &EnhancedAIStrategy{
		Level: level,
		CharacterPreferences: []CharacterPreference{
			{Character: Ambassador, PreferenceLevel: 0.9},
		},
		ActionPreferences:   make(map[string]float64),
		TargetSelection:     make(map[string]TargetStrategy),
		CharacterBluffRates: make(map[string]float64),
		CharacterBlockRates: make(map[string]float64),
	}

	// Set up strategy parameters based on competitive level
	switch level {
	case HighCompetitive:
		strategy.BluffRate = AmbassadorHighBluffRate
		strategy.ChallengeRate = AmbassadorHighChallengeRate
		strategy.AlwaysBlock = true
		strategy.ActionPreferences["Exchange"] = ExchangeHighPreference
		strategy.CharacterBluffRates[Ambassador] = AmbassadorHighBluffRate
		strategy.CharacterBlockRates[Ambassador] = AmbassadorHighBlockRate

	case MediumCompetitive:
		strategy.BluffRate = AmbassadorMediumBluffRate
		strategy.ChallengeRate = AmbassadorMediumChallengeRate
		strategy.AlwaysBlock = true
		strategy.ActionPreferences["Exchange"] = ExchangeMediumPreference
		strategy.CharacterBluffRates[Ambassador] = AmbassadorMediumBluffRate
		strategy.CharacterBlockRates[Ambassador] = AmbassadorMediumBlockRate

	case LowCompetitive:
		strategy.BluffRate = AmbassadorLowBluffRate
		strategy.ChallengeRate = AmbassadorLowChallengeRate
		strategy.AlwaysBlock = false
		strategy.ActionPreferences["Exchange"] = ExchangeLowPreference
		strategy.CharacterBluffRates[Ambassador] = AmbassadorLowBluffRate
		strategy.CharacterBlockRates[Ambassador] = AmbassadorLowBlockRate
	}

	return strategy
}

// CreateContessaStrategy creates an AI strategy focused on the Contessa character
func CreateContessaStrategy(level CompetitiveLevel) *EnhancedAIStrategy {
	strategy := &EnhancedAIStrategy{
		Level: level,
		CharacterPreferences: []CharacterPreference{
			{Character: Contessa, PreferenceLevel: 0.9},
		},
		ActionPreferences:   make(map[string]float64),
		TargetSelection:     make(map[string]TargetStrategy),
		CharacterBluffRates: make(map[string]float64),
		CharacterBlockRates: make(map[string]float64),
	}

	// Set up strategy parameters based on competitive level
	switch level {
	case HighCompetitive:
		strategy.BluffRate = ContessaHighBluffRate
		strategy.ChallengeRate = ContessaHighChallengeRate
		strategy.AlwaysBlock = true
		strategy.ActionPreferences["Income"] = IncomeHighPreference
		strategy.CharacterBluffRates[Contessa] = ContessaHighBluffRate
		strategy.CharacterBlockRates[Contessa] = ContessaHighBlockRate

	case MediumCompetitive:
		strategy.BluffRate = ContessaMediumBluffRate
		strategy.ChallengeRate = ContessaMediumChallengeRate
		strategy.AlwaysBlock = true
		strategy.ActionPreferences["Income"] = IncomeMediumPreference
		strategy.CharacterBluffRates[Contessa] = ContessaMediumBluffRate
		strategy.CharacterBlockRates[Contessa] = ContessaMediumBlockRate

	case LowCompetitive:
		strategy.BluffRate = ContessaLowBluffRate
		strategy.ChallengeRate = ContessaLowChallengeRate
		strategy.AlwaysBlock = false
		strategy.ActionPreferences["Income"] = IncomeLowPreference
		strategy.CharacterBluffRates[Contessa] = ContessaLowBluffRate
		strategy.CharacterBlockRates[Contessa] = ContessaLowBlockRate
	}

	return strategy
}

// CreateRandomStrategy creates a strategy with a random character preference
func CreateRandomStrategy(rng *rand.Rand, level CompetitiveLevel) *EnhancedAIStrategy {
	// Choose a random character to focus on
	characters := []string{Duke, Assassin, Captain, Ambassador, Contessa}
	randomChar := characters[rng.Intn(len(characters))]

	switch randomChar {
	case Duke:
		return CreateDukeStrategy(level)
	case Assassin:
		return CreateAssassinStrategy(level)
	case Captain:
		return CreateCaptainStrategy(level)
	case Ambassador:
		return CreateAmbassadorStrategy(level)
	case Contessa:
		return CreateContessaStrategy(level)
	default:
		// Should never happen, but fallback
		return NewBasicAIStrategy(DefaultBluffRate, DefaultChallengeRate, true)
	}
}
