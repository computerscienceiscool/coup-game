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
		RNG:        rand.New(rand.NewSource(seed)), // Callers derive a distinct seed per player via MixSeed
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

	// Card counting (Medium and High): never claim a character when every
	// copy is visibly accounted for — such a bluff loses to a challenge for
	// free. Income/Foreign Aid/Coup require no card, so this never empties
	// the action list.
	if p.Strategy.Level != LowCompetitive {
		actions = p.filterImpossibleBluffs(state, actions)
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
		targetAction, ok := action.(TargetedAction)
		if !ok {
			continue
		}
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
		targetAction, ok := action.(TargetedAction)
		if !ok {
			continue
		}
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
		targetAction, ok := action.(TargetedAction)
		if !ok {
			continue
		}
		target := targetAction.GetTarget()
		targetID := target.GetID()

		// Calculate threat score based on coins and influence
		for _, playerState := range state.Players {
			if playerState.ID == targetID {
				threatScore := playerState.Coins + (playerState.Influences * ThreatInfluenceMultiplier)
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

// ChallengeDecision determines whether to challenge a claim, using the
// public information a real player has: their own hand, the face-up discard
// pile, and the claimant's public claim history.
func (p *EnhancedAIPlayer) ChallengeDecision(state GameState, claimant Player, claim Action) bool {
	requiredCard := claim.GetRequiredCard()
	if requiredCard.Name == "" {
		return false // Nothing claimed, nothing to challenge
	}

	// Card counting (Medium and High): if every copy of the claimed
	// character is visible to us, the claim is impossible and challenging
	// is a guaranteed, risk-free win.
	visible := p.visibleCopies(state, requiredCard.Name)
	if p.Strategy.Level != LowCompetitive && visible >= CopiesPerCharacter {
		return true
	}

	// Base challenge rate
	challengeRate := p.Strategy.ChallengeRate

	switch p.Strategy.Level {
	case HighCompetitive:
		// Scale by how many copies we can account for: the more we see, the
		// fewer remain for the claimant to plausibly hold
		switch visible {
		case 1:
			challengeRate *= 1.2
		case 2:
			challengeRate *= 1.8
		}

		// Claim history: a player who has claimed many distinct characters
		// is bluffing somewhere; a player with one consistent claim is
		// probably telling the truth
		switch claimed := len(state.Claims[claimant.GetID()]); {
		case claimed >= 4:
			challengeRate *= 2.0
		case claimed == 3:
			challengeRate *= 1.5
		case claimed <= 1:
			challengeRate *= 0.8
		}
	case MediumCompetitive:
		if visible == 2 {
			challengeRate *= 1.4
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

// visibleCopies counts how many copies of a character this player can see
// with certainty: cards in their own hand plus the face-up discard pile.
func (p *EnhancedAIPlayer) visibleCopies(state GameState, name string) int {
	n := 0
	for _, c := range p.Influences {
		if c.Name == name {
			n++
		}
	}
	for _, c := range state.Discarded {
		if c.Name == name {
			n++
		}
	}
	return n
}

// filterImpossibleBluffs removes actions that would claim a character the
// player doesn't hold when every copy is visibly out of play or in their own
// hand — a claim any card-counting opponent challenges for free.
func (p *EnhancedAIPlayer) filterImpossibleBluffs(state GameState, actions []Action) []Action {
	filtered := make([]Action, 0, len(actions))
	for _, action := range actions {
		required := action.GetRequiredCard()
		if required.Name != "" && !p.HasCard(required) &&
			p.visibleCopies(state, required.Name) >= CopiesPerCharacter {
			continue
		}
		filtered = append(filtered, action)
	}
	if len(filtered) == 0 {
		return actions
	}
	return filtered
}

// BlockDecision determines whether to block an action
func (p *EnhancedAIPlayer) BlockDecision(state GameState, actor Player, action Action) bool {
	// Get blocking characters for this action
	blockingChars := GetBlockingCharacters(action.Name())

	// If action can't be blocked
	if len(blockingChars) == 0 {
		return false
	}

	// Holding a real blocking character: block. Non-AlwaysBlock strategies
	// apply their character-specific block rate instead of always blocking.
	for _, blockChar := range blockingChars {
		if !p.HasCard(blockChar) {
			continue
		}
		if !p.Strategy.AlwaysBlock {
			if blockRate, exists := p.Strategy.CharacterBlockRates[blockChar.Name]; exists {
				return p.RNG.Float64() < blockRate
			}
		}
		return true
	}

	// No blocking character in hand: decide whether to bluff a block. Card
	// counters (Medium and High) never bluff a character whose copies are
	// all visibly accounted for.
	plausible := blockingChars
	if p.Strategy.Level != LowCompetitive {
		plausible = plausible[:0:0]
		for _, blockChar := range blockingChars {
			if p.visibleCopies(state, blockChar.Name) < CopiesPerCharacter {
				plausible = append(plausible, blockChar)
			}
		}
		if len(plausible) == 0 {
			return false
		}
	}

	// Use the character-specific bluff rate when one is configured
	for _, blockChar := range plausible {
		if bluffRate, exists := p.Strategy.CharacterBluffRates[blockChar.Name]; exists {
			return p.RNG.Float64() < bluffRate
		}
	}

	// Fallback to general bluff rate
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

// RevealCard removes and returns the specified card from the player's hand,
// proving a challenged claim. The game shuffles the revealed card back into
// the deck and deals a replacement, keeping the hand size unchanged.
func (p *EnhancedAIPlayer) RevealCard(card Card) Card {
	for i, c := range p.Influences {
		if c.IsEqual(card) {
			revealed := p.Influences[i]
			p.Influences = append(p.Influences[:i], p.Influences[i+1:]...)
			return revealed
		}
	}

	panic(fmt.Sprintf("Player %d doesn't have card %s to reveal", p.ID, card.Name))
}

// LoseInfluence removes an influence card (due to coup, assassination, or a
// lost challenge). The lost card is removed from play by the caller.
func (p *EnhancedAIPlayer) LoseInfluence() Card {
	if len(p.Influences) == 0 {
		panic(fmt.Sprintf("Player %d has no influences to lose", p.ID))
	}

	// Lose the least preferred card
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
		if c.Name == card.Name {
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

	// For medium competitive, use simplified strategic logic
	if p.Strategy.Level == MediumCompetitive {
		// Score cards with simpler heuristics
		scores := make([]float64, totalCards)

		for i, card := range allCards {
			scores[i] = 1.0 // Base score

			// Simple bonuses for generally strong cards
			switch card.Name {
			case Duke:
				scores[i] += 1.5 // Duke is very useful (Tax action)
			case Assassin:
				scores[i] += 1.2 // Assassin is powerful
			case Captain:
				scores[i] += 0.8 // Captain is decent (Steal)
			case Contessa:
				scores[i] += 0.7 // Contessa blocks assassinations
			case Ambassador:
				scores[i] += 0.5 // Ambassador is situational
			}

			// Add more randomness than high competitive
			scores[i] += p.RNG.Float64() * 0.8
		}

		// Sort cards by score
		indices := make([]int, totalCards)
		for i := range indices {
			indices[i] = i
		}

		// Simple bubble sort
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

	// For low competitive, use fully random logic
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
