package game

import (
	"fmt"
	"math/rand"
)

// EnhancedAIPlayer implements the Player interface with more sophisticated decisions
type EnhancedAIPlayer struct {
	ID         int
	Coins      int
	Influences []Card
	Strategy   *EnhancedAIStrategy
	RNG        *rand.Rand // For reproducible random decisions
}

// NewEnhancedAIPlayer creates a new enhanced AI player with the given strategy
func NewEnhancedAIPlayer(id int, strategy *EnhancedAIStrategy, seed int64) *EnhancedAIPlayer {
	return &EnhancedAIPlayer{
		ID:         id,
		Coins:      0,
		Influences: make([]Card, 0),
		Strategy:   strategy,
		RNG:        rand.New(rand.NewSource(seed + int64(id))), // Each player has their own RNG
	}
}

// GetID returns the player's ID
func (p *EnhancedAIPlayer) GetID() int {
	return p.ID
}

// GetCoins returns the player's coin count
func (p *EnhancedAIPlayer) GetCoins() int {
	return p.Coins
}

// GetInfluences returns the player's influence cards
func (p *EnhancedAIPlayer) GetInfluences() []Card {
	// Return a copy to prevent modification
	result := make([]Card, len(p.Influences))
	for i, card := range p.Influences {
		result[i] = card.Copy()
	}
	return result
}

// InfluenceCount returns the number of influence cards
func (p *EnhancedAIPlayer) InfluenceCount() int {
	return len(p.Influences)
}

// IsAlive returns true if the player has at least one influence
func (p *EnhancedAIPlayer) IsAlive() bool {
	return len(p.Influences) > 0
}

// AddInfluence adds influence cards to the player's hand
func (p *EnhancedAIPlayer) AddInfluence(cards []Card) {
	p.Influences = append(p.Influences, cards...)
}

// AddCoins adds coins to the player
func (p *EnhancedAIPlayer) AddCoins(amount int) {
	p.Coins += amount
}

// RemoveCoins removes coins from the player
func (p *EnhancedAIPlayer) RemoveCoins(amount int) error {
	if p.Coins < amount {
		return fmt.Errorf("player %d doesn't have enough coins (%d < %d)", p.ID, p.Coins, amount)
	}
	p.Coins -= amount
	return nil
}

// ChooseAction selects an action based on the enhanced strategy
func (p *EnhancedAIPlayer) ChooseAction(state GameState, actions []Action) Action {
	if len(actions) == 0 {
		panic("No legal actions available")
	}

	// Force coup if required by rules
	for _, action := range actions {
		if action.Name() == "Coup" && p.Coins >= 10 {
			return p.selectBestTargetedAction(actions, "Coup", state)
		}
	}

	// Filter actions by character-specific preferences
	characterActions := p.filterPreferredActions(actions)
	if len(characterActions) > 0 && p.RNG.Float64() < p.getActionPreferenceWeight() {
		return p.selectBestActionFromList(characterActions, state)
	}

	// Otherwise, use weighted random selection based on strategy
	return p.weightedActionSelection(actions, state)
}

// filterPreferredActions returns actions that match the AI's character preferences
func (p *EnhancedAIPlayer) filterPreferredActions(actions []Action) []Action {
	result := make([]Action, 0)

	for _, action := range actions {
		// Check if action requires a character that the AI prefers
		for _, preference := range p.Strategy.CharacterPreferences {
			// Some actions like Income don't require cards, so add them if the AI has high preference for low-risk actions
			if action.GetRequiredCard().Name == preference.Character {
				result = append(result, action)
				break
			} else if action.Name() == "Income" && p.Strategy.Level == LowCompetitive {
				// Low competitive AIs prefer safe actions
				result = append(result, action)
				break
			}
		}
	}

	return result
}

// weightedActionSelection chooses an action with weighted probabilities
func (p *EnhancedAIPlayer) weightedActionSelection(actions []Action, state GameState) Action {
	totalWeight := 0.0
	weights := make([]float64, len(actions))

	for i, action := range actions {
		// Base weight
		weight := 1.0

		// Adjust weight based on action preference
		if pref, exists := p.Strategy.ActionPreferences[action.Name()]; exists {
			weight *= (1.0 + pref)
		}

		// Adjust based on coins - prefer income/foreign aid when low on coins
		if p.Coins < 3 {
			if action.Name() == "Income" || action.Name() == "Foreign Aid" || action.Name() == "Tax" {
				weight *= 1.5
			}
		}

		// Prefer Coup when high on coins
		if p.Coins >= 7 && action.Name() == "Coup" {
			weight *= 2.0
		}

		// Prefer Assassinate when can afford it
		if p.Coins >= 3 && action.Name() == "Assassinate" {
			weight *= 1.5
		}

		weights[i] = weight
		totalWeight += weight
	}

	// Normalize weights and select
	r := p.RNG.Float64() * totalWeight
	runningTotal := 0.0

	for i, weight := range weights {
		runningTotal += weight
		if r <= runningTotal {
			return actions[i]
		}
	}

	// Fallback
	return actions[p.RNG.Intn(len(actions))]
}

// selectBestActionFromList chooses the best action from a filtered list
func (p *EnhancedAIPlayer) selectBestActionFromList(actions []Action, state GameState) Action {
	// For targeted actions, select the best target
	for _, action := range actions {
		if action.RequiresTarget() {
			return p.selectBestTargetedAction(actions, action.Name(), state)
		}
	}

	// For non-targeted actions, pick highest preference
	var bestAction Action
	bestScore := -1.0

	for _, action := range actions {
		score := 0.0
		if pref, exists := p.Strategy.ActionPreferences[action.Name()]; exists {
			score = pref
		}

		// Add randomness for variety
		score += p.RNG.Float64() * 0.2

		if score > bestScore {
			bestScore = score
			bestAction = action
		}
	}

	if bestAction != nil {
		return bestAction
	}

	// Fallback to random
	return actions[p.RNG.Intn(len(actions))]
}

// selectBestTargetedAction finds the best target for an action type
func (p *EnhancedAIPlayer) selectBestTargetedAction(actions []Action, actionName string, state GameState) Action {
	// Filter actions of the requested type
	candidates := make([]Action, 0)
	for _, action := range actions {
		if action.Name() == actionName && action.RequiresTarget() {
			candidates = append(candidates, action)
		}
	}

	if len(candidates) == 0 {
		// No matching targeted actions
		return actions[p.RNG.Intn(len(actions))]
	}

	// Get target strategy for this action
	targetStrategy := RandomTarget
	if strategy, exists := p.Strategy.TargetSelection[actionName]; exists {
		targetStrategy = strategy
	}

	// Choose based on targeting strategy
	switch targetStrategy {
	case RichestTarget:
		return p.targetRichestPlayer(candidates)
	case StrongestTarget:
		return p.targetStrongestPlayer(candidates, state)
	case ThreatTarget:
		return p.targetBiggestThreat(candidates, state)
	default:
		// Random target
		return candidates[p.RNG.Intn(len(candidates))]
	}
}

// targetRichestPlayer targets the player with the most coins
func (p *EnhancedAIPlayer) targetRichestPlayer(actions []Action) Action {
	var bestAction Action
	maxCoins := -1

	for _, action := range actions {
		targetAction := action.(TargetedAction)
		target := targetAction.GetTarget()

		if target.GetCoins() > maxCoins {
			maxCoins = target.GetCoins()
			bestAction = action
		}
	}

	if bestAction != nil {
		return bestAction
	}

	// Fallback to random
	return actions[p.RNG.Intn(len(actions))]
}

// targetStrongestPlayer targets the player with the most influence
func (p *EnhancedAIPlayer) targetStrongestPlayer(actions []Action, state GameState) Action {
	var bestAction Action
	maxInfluence := -1

	for _, action := range actions {
		targetAction := action.(TargetedAction)
		target := targetAction.GetTarget()
		targetID := target.GetID()

		// Get influence count from game state
		for _, playerState := range state.Players {
			if playerState.ID == targetID && playerState.Influences > maxInfluence {
				maxInfluence = playerState.Influences
				bestAction = action
				break
			}
		}
	}

	if bestAction != nil {
		return bestAction
	}

	// Fallback to random
	return actions[p.RNG.Intn(len(actions))]
}

// targetBiggestThreat targets the player considered the biggest threat (coins + influence)
func (p *EnhancedAIPlayer) targetBiggestThreat(actions []Action, state GameState) Action {
	var bestAction Action
	maxThreat := -1

	for _, action := range actions {
		targetAction := action.(TargetedAction)
		target := targetAction.GetTarget()
		targetID := target.GetID()

		// Calculate threat score based on coins and influence
		for _, playerState := range state.Players {
			if playerState.ID == targetID {
				threatScore := playerState.Coins + (playerState.Influences * 3)
				if threatScore > maxThreat {
					maxThreat = threatScore
					bestAction = action
					break
				}
			}
		}
	}

	if bestAction != nil {
		return bestAction
	}

	// Fallback to random
	return actions[p.RNG.Intn(len(actions))]
}

// ChallengeDecision determines whether to challenge a claim
func (p *EnhancedAIPlayer) ChallengeDecision(state GameState, claimant Player, claim Action) bool {
	// Base challenge rate
	challengeRate := p.Strategy.ChallengeRate

	// Higher competitive AIs are better at detecting bluffs
	if p.Strategy.Level == HighCompetitive {
		// If AI has the card that's being claimed, definitely challenge
		if claim.CanBeChallenged() && p.hasCardRequiredForAction(claim) {
			return true
		}
	}

	// Special cases based on game state
	// Adjust based on remaining influence
	if p.InfluenceCount() == 1 {
		// More cautious with last influence
		challengeRate *= 0.7
	}

	// Adjust based on claimant's coins
	claimantCoins := 0
	for _, playerState := range state.Players {
		if playerState.ID == claimant.GetID() {
			claimantCoins = playerState.Coins
			break
		}
	}

	if claimantCoins >= 7 {
		// More likely to challenge players close to coup
		challengeRate *= 1.3
	}

	// Finally, make the challenge decision
	return p.RNG.Float64() < challengeRate
}

// hasCardRequiredForAction checks if the player has the card required for an action
func (p *EnhancedAIPlayer) hasCardRequiredForAction(action Action) bool {
	requiredCard := action.GetRequiredCard()
	if requiredCard.Name == "" {
		return false // Action doesn't require a specific card
	}

	return p.HasCard(requiredCard)
}

// BlockDecision determines whether to block an action
func (p *EnhancedAIPlayer) BlockDecision(state GameState, actor Player, action Action) bool {
	// If always block is false, use character-specific block rates
	if !p.Strategy.AlwaysBlock {
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

			// Check character-specific block rates for bluffing
			if blockRate, exists := p.Strategy.CharacterBlockRates[blockChar.Name]; exists {
				return p.RNG.Float64() < blockRate
			}
		}

		// If we don't have a blocking character, decide whether to bluff
		// Base this on the specific character's bluff rate
		for _, blockChar := range blockingChars {
			if bluffRate, exists := p.Strategy.CharacterBluffRates[blockChar.Name]; exists {
				return p.RNG.Float64() < bluffRate
			}
		}

		// Fallback to general bluff rate
		return p.RNG.Float64() < p.Strategy.BluffRate
	}

	// Original behavior: always block if possible
	blockingChars := GetBlockingCharacters(action.Name())
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

// ChooseBlockingCharacter selects a character to claim when blocking
func (p *EnhancedAIPlayer) ChooseBlockingCharacter(action Action) Card {
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

	// If we're bluffing, select based on character preferences
	for _, blockChar := range blockingChars {
		for _, pref := range p.Strategy.CharacterPreferences {
			if pref.Character == blockChar.Name {
				return blockChar // Prefer to bluff with character we like
			}
		}
	}

	// If no preference, randomly select a blocking character
	return blockingChars[p.RNG.Intn(len(blockingChars))]
}

// RevealCard marks a specific card as revealed
func (p *EnhancedAIPlayer) RevealCard(card Card) {
	for i, c := range p.Influences {
		if c.IsEqual(card) && !c.Shown {
			p.Influences[i].Shown = true
			return
		}
	}

	panic(fmt.Sprintf("Player %d doesn't have card %s to reveal", p.ID, card.Name))
}

// LoseInfluence removes an influence card (due to coup or challenge)
func (p *EnhancedAIPlayer) LoseInfluence() Card {
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

	// If no shown cards, lose the least preferred card
	if len(p.Influences) > 1 && p.Strategy.Level != LowCompetitive {
		// Get scores for each card
		scores := make([]float64, len(p.Influences))

		for i, card := range p.Influences {
			scores[i] = 1.0 // Default score

			// Check if card is in preferences
			for _, pref := range p.Strategy.CharacterPreferences {
				if pref.Character == card.Name {
					scores[i] += pref.PreferenceLevel * 2.0
					break
				}
			}

			// Add randomness for variety
			scores[i] += p.RNG.Float64() * 0.2
		}

		// Find lowest score
		lowestIdx := 0
		lowestScore := scores[0]

		for i, score := range scores {
			if score < lowestScore {
				lowestScore = score
				lowestIdx = i
			}
		}

		// Lose the card with lowest score
		lost := p.Influences[lowestIdx]
		p.Influences = append(p.Influences[:lowestIdx], p.Influences[lowestIdx+1:]...)
		return lost
	}

	// Default: randomly lose a card
	idx := p.RNG.Intn(len(p.Influences))
	lost := p.Influences[idx]
	p.Influences = append(p.Influences[:idx], p.Influences[idx+1:]...)
	return lost
}

// HasCard checks if the player has a specific card
func (p *EnhancedAIPlayer) HasCard(card Card) bool {
	for _, c := range p.Influences {
		if c.Name == card.Name && !c.Shown {
			return true
		}
	}
	return false
}

// ChooseExchange handles Ambassador's exchange ability
func (p *EnhancedAIPlayer) ChooseExchange(draw []Card) []Card {
	// Combine current influences and drawn cards
	allCards := append(p.GetInfluences(), draw...)

	// Total number of cards
	totalCards := len(allCards)

	// Number of cards to keep
	keepCount := len(p.Influences)

	// For high competitive, make strategic choices
	if p.Strategy.Level == HighCompetitive {
		// Score cards based on preferences
		scores := make([]float64, totalCards)

		for i, card := range allCards {
			scores[i] = 1.0 // Base score

			// Check if card is in preferences
			for _, pref := range p.Strategy.CharacterPreferences {
				if pref.Character == card.Name {
					scores[i] += pref.PreferenceLevel * 3.0
					break
				}
			}

			// Bonus for certain characters
			if card.Name == Duke {
				scores[i] += 1.0 // Duke is generally useful
			}
			if card.Name == Contessa && p.Coins >= 3 {
				scores[i] += 0.8 // Contessa more useful when you have coins
			}
			if card.Name == Assassin && p.Coins >= 3 {
				scores[i] += 1.2 // Assassin useful when you can afford it
			}

			// Add randomness
			scores[i] += p.RNG.Float64() * 0.2
		}

		// Sort cards by score
		indices := make([]int, totalCards)
		for i := range indices {
			indices[i] = i
		}

		// Bubble sort for simplicity
		for i := 0; i < totalCards; i++ {
			for j := i + 1; j < totalCards; j++ {
				if scores[indices[i]] < scores[indices[j]] {
					indices[i], indices[j] = indices[j], indices[i]
				}
			}
		}

		// Select top cards to keep
		keep := make([]Card, keepCount)
		returnToDeck := make([]Card, totalCards-keepCount)

		for i := 0; i < totalCards; i++ {
			if i < keepCount {
				keep[i] = allCards[indices[i]]
			} else {
				returnToDeck[i-keepCount] = allCards[indices[i]]
			}
		}

		// Update player's influences with kept cards
		p.Influences = keep

		// Return the cards to be put back in the deck
		return returnToDeck
	}

	// For medium and low competitive, use simpler logic
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

// Helper method to get action preference weight based on competitive level
func (p *EnhancedAIPlayer) getActionPreferenceWeight() float64 {
	switch p.Strategy.Level {
	case HighCompetitive:
		return 0.8
	case MediumCompetitive:
		return 0.5
	case LowCompetitive:
		return 0.3
	default:
		return 0.5
	}
}
