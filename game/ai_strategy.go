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
		strategy.BluffRate = 0.7
		strategy.ChallengeRate = 0.7
		strategy.AlwaysBlock = true
		strategy.ActionPreferences["Tax"] = 0.9
		strategy.CharacterBluffRates[Duke] = 0.7
		strategy.CharacterBlockRates[Duke] = 0.9

	case MediumCompetitive:
		strategy.BluffRate = 0.4
		strategy.ChallengeRate = 0.5
		strategy.AlwaysBlock = true
		strategy.ActionPreferences["Tax"] = 0.6
		strategy.CharacterBluffRates[Duke] = 0.4
		strategy.CharacterBlockRates[Duke] = 0.6

	case LowCompetitive:
		strategy.BluffRate = 0.1
		strategy.ChallengeRate = 0.3
		strategy.AlwaysBlock = false
		strategy.ActionPreferences["Tax"] = 0.3
		strategy.CharacterBluffRates[Duke] = 0.1
		strategy.CharacterBlockRates[Duke] = 0.3
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
		strategy.BluffRate = 0.6
		strategy.ChallengeRate = 0.6
		strategy.AlwaysBlock = true
		strategy.ActionPreferences["Assassinate"] = 0.8
		strategy.CharacterBluffRates[Assassin] = 0.6
		strategy.TargetSelection["Assassinate"] = ThreatTarget

	case MediumCompetitive:
		strategy.BluffRate = 0.35
		strategy.ChallengeRate = 0.5
		strategy.AlwaysBlock = true
		strategy.ActionPreferences["Assassinate"] = 0.5
		strategy.CharacterBluffRates[Assassin] = 0.35
		strategy.TargetSelection["Assassinate"] = StrongestTarget

	case LowCompetitive:
		strategy.BluffRate = 0.15
		strategy.ChallengeRate = 0.3
		strategy.AlwaysBlock = false
		strategy.ActionPreferences["Assassinate"] = 0.3
		strategy.CharacterBluffRates[Assassin] = 0.15
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
		strategy.BluffRate = 0.65
		strategy.ChallengeRate = 0.6
		strategy.AlwaysBlock = true
		strategy.ActionPreferences["Steal"] = 0.8
		strategy.CharacterBluffRates[Captain] = 0.65
		strategy.CharacterBlockRates[Captain] = 0.9
		strategy.TargetSelection["Steal"] = RichestTarget

	case MediumCompetitive:
		strategy.BluffRate = 0.4
		strategy.ChallengeRate = 0.5
		strategy.AlwaysBlock = true
		strategy.ActionPreferences["Steal"] = 0.5
		strategy.CharacterBluffRates[Captain] = 0.4
		strategy.CharacterBlockRates[Captain] = 0.6
		strategy.TargetSelection["Steal"] = RichestTarget

	case LowCompetitive:
		strategy.BluffRate = 0.1
		strategy.ChallengeRate = 0.3
		strategy.AlwaysBlock = false
		strategy.ActionPreferences["Steal"] = 0.2
		strategy.CharacterBluffRates[Captain] = 0.1
		strategy.CharacterBlockRates[Captain] = 0.3
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
		strategy.BluffRate = 0.5
		strategy.ChallengeRate = 0.6
		strategy.AlwaysBlock = true
		strategy.ActionPreferences["Exchange"] = 0.7
		strategy.CharacterBluffRates[Ambassador] = 0.5
		strategy.CharacterBlockRates[Ambassador] = 0.9

	case MediumCompetitive:
		strategy.BluffRate = 0.3
		strategy.ChallengeRate = 0.5
		strategy.AlwaysBlock = true
		strategy.ActionPreferences["Exchange"] = 0.5
		strategy.CharacterBluffRates[Ambassador] = 0.3
		strategy.CharacterBlockRates[Ambassador] = 0.6

	case LowCompetitive:
		strategy.BluffRate = 0.05
		strategy.ChallengeRate = 0.3
		strategy.AlwaysBlock = false
		strategy.ActionPreferences["Exchange"] = 0.2
		strategy.CharacterBluffRates[Ambassador] = 0.05
		strategy.CharacterBlockRates[Ambassador] = 0.3
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
		strategy.BluffRate = 0.8
		strategy.ChallengeRate = 0.6
		strategy.AlwaysBlock = true
		strategy.ActionPreferences["Income"] = 0.6
		strategy.CharacterBluffRates[Contessa] = 0.8
		strategy.CharacterBlockRates[Contessa] = 1.0

	case MediumCompetitive:
		strategy.BluffRate = 0.5
		strategy.ChallengeRate = 0.5
		strategy.AlwaysBlock = true
		strategy.ActionPreferences["Income"] = 0.5
		strategy.CharacterBluffRates[Contessa] = 0.5
		strategy.CharacterBlockRates[Contessa] = 0.8

	case LowCompetitive:
		strategy.BluffRate = 0.15
		strategy.ChallengeRate = 0.3
		strategy.AlwaysBlock = false
		strategy.ActionPreferences["Income"] = 0.4
		strategy.CharacterBluffRates[Contessa] = 0.15
		strategy.CharacterBlockRates[Contessa] = 0.6
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
		return NewBasicAIStrategy(0.3, 0.5, true)
	}
}
