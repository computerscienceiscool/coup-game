package game

import (
	"errors"
	"math/rand"
)

// AIPlayerType defines the type of AI to create
type AIPlayerType int

const (
	RandomAI AIPlayerType = iota
	DukeAI
	AssassinAI
	CaptainAI
	AmbassadorAI
	ContessaAI
)

// NewGameWithAITypes creates a new game with specific AI player types
func NewGameWithAITypes(playerCount int, aiTypes []AIPlayerType, competitiveLevel CompetitiveLevel, seed int64) (*Game, error) {
	if playerCount < 2 || playerCount > 6 {
		return nil, errors.New("player count must be between 2 and 6")
	}

	if len(aiTypes) > 0 && len(aiTypes) != playerCount {
		return nil, errors.New("if AI types are specified, must provide exactly one per player")
	}

	rng := rand.New(rand.NewSource(seed))
	deck := NewDeck(rng)
	players := make([]Player, playerCount)

	// Create AI players based on types
	for i := 0; i < playerCount; i++ {
		var strategy *EnhancedAIStrategy

		if len(aiTypes) > 0 {
			// Create strategy based on specified type
			switch aiTypes[i] {
			case DukeAI:
				strategy = CreateDukeStrategy(competitiveLevel)
			case AssassinAI:
				strategy = CreateAssassinStrategy(competitiveLevel)
			case CaptainAI:
				strategy = CreateCaptainStrategy(competitiveLevel)
			case AmbassadorAI:
				strategy = CreateAmbassadorStrategy(competitiveLevel)
			case ContessaAI:
				strategy = CreateContessaStrategy(competitiveLevel)
			default:
				strategy = CreateRandomStrategy(rng, competitiveLevel)
			}
		} else {
			// Create random strategy
			strategy = CreateRandomStrategy(rng, competitiveLevel)
		}

		players[i] = NewEnhancedAIPlayer(i, strategy, seed+int64(i))
	}

	game := &Game{
		Players:       players,
		Deck:          deck,
		CurrentPlayer: 0,
		Turn:          0,
		Finished:      false,
		ActionLog:     make([]ActionLog, 0),
		RNG:           rng,
	}

	// Deal initial cards and coins
	game.Initialize()
	return game, nil
}

// NewGameWithMixedAIs creates a game with a mix of AI competitive levels
func NewGameWithMixedAIs(playerCount int, seed int64) (*Game, error) {
	if playerCount < 2 || playerCount > 6 {
		return nil, errors.New("player count must be between 2 and 6")
	}

	rng := rand.New(rand.NewSource(seed))
	deck := NewDeck(rng)
	players := make([]Player, playerCount)

	// Create AI players with mixed competitive levels
	for i := 0; i < playerCount; i++ {
		// Randomly select competitive level
		level := CompetitiveLevel(rng.Intn(3)) // 0=Low, 1=Medium, 2=High

		// Randomly select character preference
		aiType := AIPlayerType(rng.Intn(6)) // 0-5

		var strategy *EnhancedAIStrategy

		// Create strategy based on type
		switch aiType {
		case DukeAI:
			strategy = CreateDukeStrategy(level)
		case AssassinAI:
			strategy = CreateAssassinStrategy(level)
		case CaptainAI:
			strategy = CreateCaptainStrategy(level)
		case AmbassadorAI:
			strategy = CreateAmbassadorStrategy(level)
		case ContessaAI:
			strategy = CreateContessaStrategy(level)
		default:
			strategy = CreateRandomStrategy(rng, level)
		}

		players[i] = NewEnhancedAIPlayer(i, strategy, seed+int64(i))
	}

	game := &Game{
		Players:       players,
		Deck:          deck,
		CurrentPlayer: 0,
		Turn:          0,
		Finished:      false,
		ActionLog:     make([]ActionLog, 0),
		RNG:           rng,
	}

	// Deal initial cards and coins
	game.Initialize()
	return game, nil
}

// Compatibility function to create a game with the original AI behavior
func NewGameWithOriginalAI(playerCount int, seed int64) (*Game, error) {
	if playerCount < 2 || playerCount > 6 {
		return nil, errors.New("player count must be between 2 and 6")
	}

	rng := rand.New(rand.NewSource(seed))
	deck := NewDeck(rng)
	players := make([]Player, playerCount)

	// Create original AI players
	for i := 0; i < playerCount; i++ {
		players[i] = NewAIPlayer(i, &AIStrategy{
			BluffRate:     0.3, // 30% chance to bluff
			ChallengeRate: 0.5, // 50% chance to challenge
			AlwaysBlock:   true,
		}, seed+int64(i))
	}

	game := &Game{
		Players:       players,
		Deck:          deck,
		CurrentPlayer: 0,
		Turn:          0,
		Finished:      false,
		ActionLog:     make([]ActionLog, 0),
		RNG:           rng,
	}

	// Deal initial cards and coins
	game.Initialize()
	return game, nil
}
