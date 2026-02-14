package game

// AI Competitive Level Base Rates
const (
	// Bluff rates by competitive level
	HighCompetitiveBluffRate   = 0.7
	MediumCompetitiveBluffRate = 0.4
	LowCompetitiveBluffRate    = 0.1

	// Challenge rates by competitive level
	HighCompetitiveChallengeRate   = 0.7
	MediumCompetitiveChallengeRate = 0.5
	LowCompetitiveChallengeRate    = 0.3

	// Default/Original AI rates
	DefaultBluffRate     = 0.3
	DefaultChallengeRate = 0.5
)

// Character-Specific Base Rates
const (
	// Duke strategy rates
	DukeHighBluffRate      = 0.7
	DukeHighChallengeRate  = 0.7
	DukeMediumBluffRate    = 0.4
	DukeMediumChallengeRate = 0.5
	DukeLowBluffRate       = 0.1
	DukeLowChallengeRate   = 0.3

	// Assassin strategy rates
	AssassinHighBluffRate      = 0.6
	AssassinHighChallengeRate  = 0.6
	AssassinMediumBluffRate    = 0.35
	AssassinMediumChallengeRate = 0.5
	AssassinLowBluffRate       = 0.15
	AssassinLowChallengeRate   = 0.3

	// Captain strategy rates
	CaptainHighBluffRate      = 0.65
	CaptainHighChallengeRate  = 0.6
	CaptainMediumBluffRate    = 0.4
	CaptainMediumChallengeRate = 0.5
	CaptainLowBluffRate       = 0.1
	CaptainLowChallengeRate   = 0.3

	// Ambassador strategy rates
	AmbassadorHighBluffRate      = 0.5
	AmbassadorHighChallengeRate  = 0.6
	AmbassadorMediumBluffRate    = 0.3
	AmbassadorMediumChallengeRate = 0.5
	AmbassadorLowBluffRate       = 0.05
	AmbassadorLowChallengeRate   = 0.3

	// Contessa strategy rates
	ContessaHighBluffRate      = 0.8
	ContessaHighChallengeRate  = 0.6
	ContessaMediumBluffRate    = 0.5
	ContessaMediumChallengeRate = 0.5
	ContessaLowBluffRate       = 0.15
	ContessaLowChallengeRate   = 0.3
)

// Character-Specific Block Rates
const (
	// Duke block rates
	DukeHighBlockRate   = 0.9
	DukeMediumBlockRate = 0.6
	DukeLowBlockRate    = 0.3

	// Captain block rates
	CaptainHighBlockRate   = 0.8
	CaptainMediumBlockRate = 0.6
	CaptainLowBlockRate    = 0.3

	// Ambassador block rates
	AmbassadorHighBlockRate   = 0.9
	AmbassadorMediumBlockRate = 0.6
	AmbassadorLowBlockRate    = 0.3

	// Contessa block rates (Contessa always blocks Assassinations)
	ContessaHighBlockRate   = 1.0
	ContessaMediumBlockRate = 0.8
	ContessaLowBlockRate    = 0.6
)

// Action Preference Weights
const (
	// High competitive action preferences
	TaxHighPreference         = 0.9
	AssassinateHighPreference = 0.8
	StealHighPreference       = 0.8
	ExchangeHighPreference    = 0.7
	IncomeHighPreference      = 0.6

	// Medium competitive action preferences
	TaxMediumPreference         = 0.6
	AssassinateMediumPreference = 0.5
	StealMediumPreference       = 0.5
	ExchangeMediumPreference    = 0.5
	IncomeMediumPreference      = 0.5

	// Low competitive action preferences
	TaxLowPreference         = 0.3
	AssassinateLowPreference = 0.3
	StealLowPreference       = 0.2
	ExchangeLowPreference    = 0.2
	IncomeLowPreference      = 0.4
)

// Threat Assessment
const (
	// Threat score calculation: Coins + (Influences * ThreatInfluenceMultiplier)
	ThreatInfluenceMultiplier = 3
)
