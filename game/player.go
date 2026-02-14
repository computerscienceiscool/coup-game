package game

import (
	"fmt"
	"math/rand"
)

// Player represents any player (human or AI)
type Player interface {
	// Decision making
	ChooseAction(state GameState, actions []Action) Action
	ChallengeDecision(state GameState, claimant Player, claim Action) bool
	BlockDecision(state GameState, actor Player, action Action) bool
	ChooseBlockingCharacter(action Action) Card

	// When challenged
	RevealCard(card Card)              // Reveal a specific card
	LoseInfluence() Card               // Choose which card to lose
	HasCard(card Card) bool            // Check if player has a specific card
	ChooseExchange(draw []Card) []Card // Ambassador exchange

	// Information
	GetID() int
	GetCoins() int
	GetInfluences() []Card
	InfluenceCount() int
	IsAlive() bool
	AddInfluence(cards []Card)
	AddCoins(amount int)
	RemoveCoins(amount int) error
}

// AIPlayer implements the Player interface with random decisions
type AIPlayer struct {
	ID         int
	Coins      int
	Influences []Card
	Strategy   *AIStrategy
	RNG        *rand.Rand // For reproducible random decisions
}

// AIStrategy defines the probabilities for AI decisions
type AIStrategy struct {
	BluffRate     float64 // Probability of bluffing (claiming a character they don't have)
	ChallengeRate float64 // Probability of challenging a claim
	AlwaysBlock   bool    // Whether to always block when possible
}

// NewAIPlayer creates a new AI player
func NewAIPlayer(id int, strategy *AIStrategy, seed int64) *AIPlayer {
	return &AIPlayer{
		ID:         id,
		Coins:      0,
		Influences: make([]Card, 0),
		Strategy:   strategy,
		RNG:        rand.New(rand.NewSource(seed + int64(id))), // Each player has their own RNG
	}
}

// GetID returns the player's ID
func (p *AIPlayer) GetID() int {
	return p.ID
}

// GetCoins returns the player's coin count
func (p *AIPlayer) GetCoins() int {
	return p.Coins
}

// GetInfluences returns the player's influence cards
func (p *AIPlayer) GetInfluences() []Card {
	// Return a copy to prevent modification
	result := make([]Card, len(p.Influences))
	for i, card := range p.Influences {
		result[i] = card.Copy()
	}
	return result
}

// InfluenceCount returns the number of influence cards
func (p *AIPlayer) InfluenceCount() int {
	return len(p.Influences)
}

// IsAlive returns true if the player has at least one influence
func (p *AIPlayer) IsAlive() bool {
	return len(p.Influences) > 0
}

// AddInfluence adds influence cards to the player's hand
func (p *AIPlayer) AddInfluence(cards []Card) {
	p.Influences = append(p.Influences, cards...)
}

// AddCoins adds coins to the player
func (p *AIPlayer) AddCoins(amount int) {
	p.Coins += amount
}

// RemoveCoins removes coins from the player
func (p *AIPlayer) RemoveCoins(amount int) error {
	if p.Coins < amount {
		return fmt.Errorf("player %d doesn't have enough coins (%d < %d)", p.ID, p.Coins, amount)
	}
	p.Coins -= amount
	return nil
}

// ChooseAction randomly selects an action from legal actions
func (p *AIPlayer) ChooseAction(state GameState, actions []Action) Action {
	if len(actions) == 0 {
		panic("No legal actions available")
	}

	// Randomly select action
	return actions[p.RNG.Intn(len(actions))]
}

// ChallengeDecision determines whether to challenge a claim
func (p *AIPlayer) ChallengeDecision(state GameState, claimant Player, claim Action) bool {
	// 50% chance to challenge
	return p.RNG.Float64() < p.Strategy.ChallengeRate
}

// BlockDecision determines whether to block an action
func (p *AIPlayer) BlockDecision(state GameState, actor Player, action Action) bool {
	// If always block is true, check if we have a blocking character or want to bluff
	if p.Strategy.AlwaysBlock {
		// Get blocking characters for this action
		blockingChars := GetBlockingCharacters(action.Name())

		// If action can't be blocked
		if len(blockingChars) == 0 {
			return false
		}

		// Check if player has any of the blocking characters
		for _, blockChar := range blockingChars {
			if p.HasCard(blockChar) {
				return true // Block with real character
			}
		}

		// If we don't have a blocking character, decide whether to bluff
		return p.RNG.Float64() < p.Strategy.BluffRate
	}

	return false
}

// ChooseBlockingCharacter selects a character to claim when blocking
func (p *AIPlayer) ChooseBlockingCharacter(action Action) Card {
	blockingChars := GetBlockingCharacters(action.Name())
	if len(blockingChars) == 0 {
		panic(fmt.Sprintf("Action %s cannot be blocked", action.Name()))
	}

	// Check if we have any of the blocking characters
	for _, blockChar := range blockingChars {
		if p.HasCard(blockChar) {
			return blockChar // Block with real character
		}
	}

	// If we're bluffing, randomly select a blocking character
	return blockingChars[p.RNG.Intn(len(blockingChars))]
}

// RevealCard marks a specific card as revealed
func (p *AIPlayer) RevealCard(card Card) {
	for i, c := range p.Influences {
		if c.IsEqual(card) && !c.Shown {
			p.Influences[i].Shown = true
			return
		}
	}

	panic(fmt.Sprintf("Player %d doesn't have card %s to reveal", p.ID, card.Name))
}

// LoseInfluence removes an influence card (due to coup or challenge)
func (p *AIPlayer) LoseInfluence() Card {
	if len(p.Influences) == 0 {
		panic(fmt.Sprintf("Player %d has no influences to lose", p.ID))
	}

	// Try to lose a shown card first
	for i, card := range p.Influences {
		if card.Shown {
			lost := p.Influences[i]
			// Remove card at index i
			p.Influences = append(p.Influences[:i], p.Influences[i+1:]...)
			return lost
		}
	}

	// If no shown cards, randomly lose an unshown card
	idx := p.RNG.Intn(len(p.Influences))
	lost := p.Influences[idx]
	p.Influences = append(p.Influences[:idx], p.Influences[idx+1:]...)
	return lost
}

// HasCard checks if the player has a specific card
func (p *AIPlayer) HasCard(card Card) bool {
	for _, c := range p.Influences {
		if c.Name == card.Name && !c.Shown {
			return true
		}
	}
	return false
}

// ChooseExchange handles Ambassador's exchange ability
func (p *AIPlayer) ChooseExchange(draw []Card) []Card {
	// Combine current influences and drawn cards
	allCards := append(p.GetInfluences(), draw...)

	// Total number of cards
	totalCards := len(allCards)

	// Number of cards to keep
	keepCount := len(p.Influences)

	// Create a shuffled index array
	indices := make([]int, totalCards)
	for i := range indices {
		indices[i] = i
	}
	// Shuffle the indices
	for i := totalCards - 1; i > 0; i-- {
		j := p.RNG.Intn(i + 1)
		indices[i], indices[j] = indices[j], indices[i]
	}

	// Select cards to keep and return
	keep := make([]Card, keepCount)
	returnToDeck := make([]Card, totalCards-keepCount)

	keepIdx, returnIdx := 0, 0
	for i := 0; i < totalCards; i++ {
		if i < keepCount {
			keep[keepIdx] = allCards[indices[i]]
			keepIdx++
		} else {
			returnToDeck[returnIdx] = allCards[indices[i]]
			returnIdx++
		}
	}

	// Update player's influences with kept cards
	p.Influences = keep

	// Return the cards to be put back in the deck
	return returnToDeck
}
