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

// CharacterStrategyConfig holds configuration for a character strategy at a specific level
type CharacterStrategyConfig struct {
	BluffRate           float64
	ChallengeRate       float64
	AlwaysBlock         bool
	ActionName          string // Primary action for this character
	ActionPreference    float64
	CharacterBluffRate  float64
	CharacterBlockRate  float64
	TargetStrategy      TargetStrategy
	UseTargetStrategy   bool // Whether to set target strategy
}

// strategyConfigs maps character and level to configuration
var strategyConfigs = map[string]map[CompetitiveLevel]CharacterStrategyConfig{
	Duke: {
		HighCompetitive: {
			BluffRate: DukeHighBluffRate, ChallengeRate: DukeHighChallengeRate, AlwaysBlock: true,
			ActionName: "Tax", ActionPreference: TaxHighPreference,
			CharacterBluffRate: DukeHighBluffRate, CharacterBlockRate: DukeHighBlockRate,
		},
		MediumCompetitive: {
			BluffRate: DukeMediumBluffRate, ChallengeRate: DukeMediumChallengeRate, AlwaysBlock: true,
			ActionName: "Tax", ActionPreference: TaxMediumPreference,
			CharacterBluffRate: DukeMediumBluffRate, CharacterBlockRate: DukeMediumBlockRate,
		},
		LowCompetitive: {
			BluffRate: DukeLowBluffRate, ChallengeRate: DukeLowChallengeRate, AlwaysBlock: false,
			ActionName: "Tax", ActionPreference: TaxLowPreference,
			CharacterBluffRate: DukeLowBluffRate, CharacterBlockRate: DukeLowBlockRate,
		},
	},
	Assassin: {
		HighCompetitive: {
			BluffRate: AssassinHighBluffRate, ChallengeRate: AssassinHighChallengeRate, AlwaysBlock: true,
			ActionName: "Assassinate", ActionPreference: AssassinateHighPreference,
			CharacterBluffRate: AssassinHighBluffRate,
			TargetStrategy: ThreatTarget, UseTargetStrategy: true,
		},
		MediumCompetitive: {
			BluffRate: AssassinMediumBluffRate, ChallengeRate: AssassinMediumChallengeRate, AlwaysBlock: true,
			ActionName: "Assassinate", ActionPreference: AssassinateMediumPreference,
			CharacterBluffRate: AssassinMediumBluffRate,
			TargetStrategy: StrongestTarget, UseTargetStrategy: true,
		},
		LowCompetitive: {
			BluffRate: AssassinLowBluffRate, ChallengeRate: AssassinLowChallengeRate, AlwaysBlock: false,
			ActionName: "Assassinate", ActionPreference: AssassinateLowPreference,
			CharacterBluffRate: AssassinLowBluffRate,
			TargetStrategy: RandomTarget, UseTargetStrategy: true,
		},
	},
	Captain: {
		HighCompetitive: {
			BluffRate: CaptainHighBluffRate, ChallengeRate: CaptainHighChallengeRate, AlwaysBlock: true,
			ActionName: "Steal", ActionPreference: StealHighPreference,
			CharacterBluffRate: CaptainHighBluffRate, CharacterBlockRate: CaptainHighBlockRate,
			TargetStrategy: RichestTarget, UseTargetStrategy: true,
		},
		MediumCompetitive: {
			BluffRate: CaptainMediumBluffRate, ChallengeRate: CaptainMediumChallengeRate, AlwaysBlock: true,
			ActionName: "Steal", ActionPreference: StealMediumPreference,
			CharacterBluffRate: CaptainMediumBluffRate, CharacterBlockRate: CaptainMediumBlockRate,
			TargetStrategy: RichestTarget, UseTargetStrategy: true,
		},
		LowCompetitive: {
			BluffRate: CaptainLowBluffRate, ChallengeRate: CaptainLowChallengeRate, AlwaysBlock: false,
			ActionName: "Steal", ActionPreference: StealLowPreference,
			CharacterBluffRate: CaptainLowBluffRate, CharacterBlockRate: CaptainLowBlockRate,
			TargetStrategy: RandomTarget, UseTargetStrategy: true,
		},
	},
	Ambassador: {
		HighCompetitive: {
			BluffRate: AmbassadorHighBluffRate, ChallengeRate: AmbassadorHighChallengeRate, AlwaysBlock: true,
			ActionName: "Exchange", ActionPreference: ExchangeHighPreference,
			CharacterBluffRate: AmbassadorHighBluffRate, CharacterBlockRate: AmbassadorHighBlockRate,
		},
		MediumCompetitive: {
			BluffRate: AmbassadorMediumBluffRate, ChallengeRate: AmbassadorMediumChallengeRate, AlwaysBlock: true,
			ActionName: "Exchange", ActionPreference: ExchangeMediumPreference,
			CharacterBluffRate: AmbassadorMediumBluffRate, CharacterBlockRate: AmbassadorMediumBlockRate,
		},
		LowCompetitive: {
			BluffRate: AmbassadorLowBluffRate, ChallengeRate: AmbassadorLowChallengeRate, AlwaysBlock: false,
			ActionName: "Exchange", ActionPreference: ExchangeLowPreference,
			CharacterBluffRate: AmbassadorLowBluffRate, CharacterBlockRate: AmbassadorLowBlockRate,
		},
	},
	Contessa: {
		HighCompetitive: {
			BluffRate: ContessaHighBluffRate, ChallengeRate: ContessaHighChallengeRate, AlwaysBlock: true,
			ActionName: "Income", ActionPreference: IncomeHighPreference,
			CharacterBluffRate: ContessaHighBluffRate, CharacterBlockRate: ContessaHighBlockRate,
		},
		MediumCompetitive: {
			BluffRate: ContessaMediumBluffRate, ChallengeRate: ContessaMediumChallengeRate, AlwaysBlock: true,
			ActionName: "Income", ActionPreference: IncomeMediumPreference,
			CharacterBluffRate: ContessaMediumBluffRate, CharacterBlockRate: ContessaMediumBlockRate,
		},
		LowCompetitive: {
			BluffRate: ContessaLowBluffRate, ChallengeRate: ContessaLowChallengeRate, AlwaysBlock: false,
			ActionName: "Income", ActionPreference: IncomeLowPreference,
			CharacterBluffRate: ContessaLowBluffRate, CharacterBlockRate: ContessaLowBlockRate,
		},
	},
}

// createCharacterStrategy creates a strategy for a specific character and level using config
func createCharacterStrategy(character string, level CompetitiveLevel) *EnhancedAIStrategy {
	config, exists := strategyConfigs[character][level]
	if !exists {
		// Fallback to basic strategy
		return NewBasicAIStrategy(DefaultBluffRate, DefaultChallengeRate, true)
	}

	strategy := &EnhancedAIStrategy{
		Level: level,
		CharacterPreferences: []CharacterPreference{
			{Character: character, PreferenceLevel: 0.9},
		},
		ActionPreferences:   make(map[string]float64),
		TargetSelection:     make(map[string]TargetStrategy),
		CharacterBluffRates: make(map[string]float64),
		CharacterBlockRates: make(map[string]float64),
	}

	// Apply configuration
	strategy.BluffRate = config.BluffRate
	strategy.ChallengeRate = config.ChallengeRate
	strategy.AlwaysBlock = config.AlwaysBlock
	strategy.ActionPreferences[config.ActionName] = config.ActionPreference
	strategy.CharacterBluffRates[character] = config.CharacterBluffRate
	if config.CharacterBlockRate > 0 {
		strategy.CharacterBlockRates[character] = config.CharacterBlockRate
	}
	if config.UseTargetStrategy {
		strategy.TargetSelection[config.ActionName] = config.TargetStrategy
	}

	return strategy
}

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
	return createCharacterStrategy(Duke, level)
}

// CreateAssassinStrategy creates an AI strategy focused on the Assassin character
func CreateAssassinStrategy(level CompetitiveLevel) *EnhancedAIStrategy {
	return createCharacterStrategy(Assassin, level)
}

// CreateCaptainStrategy creates an AI strategy focused on the Captain character
func CreateCaptainStrategy(level CompetitiveLevel) *EnhancedAIStrategy {
	return createCharacterStrategy(Captain, level)
}

// CreateAmbassadorStrategy creates an AI strategy focused on the Ambassador character
func CreateAmbassadorStrategy(level CompetitiveLevel) *EnhancedAIStrategy {
	return createCharacterStrategy(Ambassador, level)
}

// CreateContessaStrategy creates an AI strategy focused on the Contessa character
func CreateContessaStrategy(level CompetitiveLevel) *EnhancedAIStrategy {
	return createCharacterStrategy(Contessa, level)
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
