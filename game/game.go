package game

import (
	"errors"
	"fmt"
	"math/rand"
)

// Game represents the state of a Coup game
type Game struct {
	Players       []Player
	Deck          *Deck
	CurrentPlayer int
	Turn          int
	Finished      bool
	Winner        Player
	ActionLog     []ActionLog
	RNG           *rand.Rand // Random number generator for reproducibility
}

// ActionLog records details of each action taken
type ActionLog struct {
	Turn              int
	PlayerID          int
	Action            string
	Target            int    // -1 if no target
	Success           bool
	Challenged        bool
	Blocker           int    // -1 if not blocked
	BlockingCharacter string // Character claimed by blocker (empty if not blocked)
}

// GameState provides a read-only view of the game state for AI decision making
type GameState struct {
	PlayerID       int            // ID of the player viewing this state
	Players        []PlayerState  // State of all players
	CurrentPlayer  int            // Index of current player
	Turn           int            // Current turn number
	KnownCards     map[int][]Card // Cards known to this player (from challenges)
	RemainingCards []Card         // Cards this player knows are in the deck
}

// PlayerState provides a public view of a player
type PlayerState struct {
	ID         int
	Coins      int
	Influences int // Number of face-down cards
	IsAlive    bool
}

// NewGame creates a new Coup game with the specified number of players
func NewGame(playerCount int, seed int64) (*Game, error) {
	if playerCount < 2 || playerCount > 6 {
		return nil, errors.New("player count must be between 2 and 6")
	}

	rng := rand.New(rand.NewSource(seed))
	deck := NewDeck(rng)
	players := make([]Player, playerCount)

	// Create AI players
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

// Initialize sets up the game, dealing cards and giving starting coins
func (g *Game) Initialize() {
	// Shuffle the deck
	g.Deck.Shuffle()

	// Deal two cards to each player and give 2 coins
	for i := range g.Players {
		cards := g.Deck.Draw(2)
		g.Players[i].AddInfluence(cards)
		g.Players[i].AddCoins(2)
	}
}

// MaxTurns is the maximum number of turns before a game is force-ended.
const MaxTurns = 500

// RunToCompletion runs the game until a winner is determined
func (g *Game) RunToCompletion() Player {
	for !g.Finished {
		g.ExecuteTurn()
		if g.Turn >= MaxTurns {
			g.forceEnd()
			break
		}
	}
	return g.Winner
}

// forceEnd declares the player with the most influence (ties broken by coins) the winner.
func (g *Game) forceEnd() {
	var best Player
	bestScore := -1

	for _, p := range g.Players {
		if !p.IsAlive() {
			continue
		}
		score := p.InfluenceCount()*1000 + p.GetCoins()
		if score > bestScore {
			bestScore = score
			best = p
		}
	}

	g.Finished = true
	g.Winner = best
}

// ExecuteTurn executes a single turn of the game
func (g *Game) ExecuteTurn() {
	// Skip if game is already finished
	if g.Finished {
		return
	}

	// Get current player
	player := g.Players[g.CurrentPlayer]

	// Skip eliminated players
	if !player.IsAlive() {
		g.NextPlayer()
		return
	}

	// Get legal actions for player
	actions := g.GetLegalActions(player)

	// Player chooses action
	gameState := g.GetStateForPlayer(g.CurrentPlayer)
	action := player.ChooseAction(gameState, actions)

	// Log the action
	actionLog := ActionLog{
		Turn:       g.Turn,
		PlayerID:   player.GetID(),
		Action:     action.Name(),
		Target:     -1,   // Will be set if action has target
		Success:    true, // Default to true, will be set to false if blocked or challenged
		Challenged: false,
		Blocker:    -1,
	}

	// Set target if action has one
	if action.RequiresTarget() {
		if targetAction, ok := action.(TargetedAction); ok {
			actionLog.Target = targetAction.GetTarget().GetID()
		}
	}

	// Handle action resolution
	success := g.ResolveAction(player, action, &actionLog)

	// Update the success status
	actionLog.Success = success

	// Add to action log
	g.ActionLog = append(g.ActionLog, actionLog)

	// Check for win condition
	g.CheckWinCondition()

	// If game not finished, move to next player
	if !g.Finished {
		g.NextPlayer()
	}
}

// ResolveAction handles the challenge and block resolution process
func (g *Game) ResolveAction(player Player, action Action, log *ActionLog) bool {
	// Pay upfront costs before challenge/block phase (per Coup rules)
	if action.Name() == "Assassinate" {
		if err := player.RemoveCoins(3); err != nil {
			return false
		}
	}

	// Check if action can be challenged
	if action.CanBeChallenged() {
		// Allow other players to challenge
		for i, opponent := range g.Players {
			if i == player.GetID() || !opponent.IsAlive() {
				continue // Skip self and eliminated players
			}

			// Get state for opponent
			opponentState := g.GetStateForPlayer(i)

			// Opponent decides whether to challenge
			if opponent.ChallengeDecision(opponentState, player, action) {
				log.Challenged = true

				// Resolve the challenge
				challengeSuccess := g.ResolveChallenge(opponent, player, action)

				if challengeSuccess {
					// Challenge succeeded, action fails
					return false
				} else {
					// Challenge failed, continue with action
					break // Only one successful challenge per action
				}
			}
		}
	}

	// Check if action can be blocked
	if action.CanBeBlocked() {
		// Allow other players to block
		for i, opponent := range g.Players {
			if i == player.GetID() || !opponent.IsAlive() {
				continue // Skip self and eliminated players
			}

			// Get state for opponent
			opponentState := g.GetStateForPlayer(i)

			// Opponent decides whether to block
			if opponent.BlockDecision(opponentState, player, action) {
				log.Blocker = opponent.GetID()

				// Get the blocking character
				blockingCharacter := opponent.ChooseBlockingCharacter(action)
				log.BlockingCharacter = blockingCharacter.Name

				// Block can be challenged
				blockChallenged := false
				for j, challenger := range g.Players {
					if j == opponent.GetID() || !challenger.IsAlive() {
						continue // Skip blocker and eliminated players
					}

					// Get state for challenger
					challengerState := g.GetStateForPlayer(j)

					// Create a proper block claim using our BlockClaim struct
					blockClaim := NewBlockClaim(opponent, action, blockingCharacter)

					// Challenger decides whether to challenge the block
					if challenger.ChallengeDecision(challengerState, opponent, blockClaim) {
						blockChallenged = true

						// Resolve the block challenge
						blockChallengeSuccess := g.ResolveBlockChallenge(challenger, opponent, blockingCharacter)

						if blockChallengeSuccess {
							// Block challenge succeeded, block fails, action proceeds
							log.Blocker = -1           // Reset blocker as block failed
							log.BlockingCharacter = "" // Clear blocking character as block failed
							break                      // Only one successful challenge needed
						} else {
							// Block challenge failed, block succeeds, action fails
							return false
						}
					}
				}

				// If block wasn't challenged, action is blocked
				if !blockChallenged {
					return false
				}
			}
		}
	}

	// If we get here, action wasn't successfully blocked or challenged
	// Execute the action
	err := action.Execute(g)
	return err == nil
}

// ResolveChallenge handles the challenge resolution
func (g *Game) ResolveChallenge(challenger Player, challenged Player, action Action) bool {
	// Get required character for this action
	requiredCard := action.GetRequiredCard()

	// Check if challenged player has the claimed character
	if challenged.HasCard(requiredCard) {
		// Challenge fails
		// Challenged player shows and returns the card to deck
		challenged.RevealCard(requiredCard)
		g.Deck.Return([]Card{requiredCard})
		g.Deck.Shuffle()

		// Challenged player draws a new card
		newCard := g.Deck.Draw(1)[0]
		challenged.AddInfluence([]Card{newCard})

		// Challenger loses an influence
		card := challenger.LoseInfluence()
		g.Deck.Return([]Card{card})
		g.Deck.Shuffle()

		return false
	} else {
		// Challenge succeeds
		// Challenged player loses an influence
		card := challenged.LoseInfluence()
		g.Deck.Return([]Card{card})
		g.Deck.Shuffle()

		return true
	}
}

// ResolveBlockChallenge handles challenging a block
func (g *Game) ResolveBlockChallenge(challenger Player, blocker Player, blockingCard Card) bool {
	// Check if blocker has the claimed blocking character
	if blocker.HasCard(blockingCard) {
		// Challenge fails
		// Blocker shows and returns the card to deck
		blocker.RevealCard(blockingCard)
		g.Deck.Return([]Card{blockingCard})
		g.Deck.Shuffle()

		// Blocker draws a new card
		newCard := g.Deck.Draw(1)[0]
		blocker.AddInfluence([]Card{newCard})

		// Challenger loses an influence
		card := challenger.LoseInfluence()
		g.Deck.Return([]Card{card})
		g.Deck.Shuffle()

		return false
	} else {
		// Challenge succeeds
		// Blocker loses an influence
		card := blocker.LoseInfluence()
		g.Deck.Return([]Card{card})
		g.Deck.Shuffle()

		return true
	}
}

// GetLegalActions returns all legal actions for the given player
func (g *Game) GetLegalActions(player Player) []Action {
	actions := make([]Action, 0)

	// Always available: Income
	actions = append(actions, NewIncomeAction(player))

	// Foreign Aid (can be blocked by Duke)
	actions = append(actions, NewForeignAidAction(player))

	// Tax (Duke) - available as character action or bluff
	actions = append(actions, NewTaxAction(player))

	// If player has 3+ coins, they can assassinate
	if player.GetCoins() >= 3 {
		// For each valid target
		for _, target := range g.GetValidTargets(player) {
			actions = append(actions, NewAssassinateAction(player, target))
		}
	}

	// Steal (Captain)
	for _, target := range g.GetValidTargets(player) {
		// Only steal from players with coins
		if target.GetCoins() > 0 {
			actions = append(actions, NewStealAction(player, target))
		}
	}

	// Exchange (Ambassador)
	actions = append(actions, NewExchangeAction(player))

	// Coup - forced if 10+ coins
	if player.GetCoins() >= 7 {
		// For each valid target
		for _, target := range g.GetValidTargets(player) {
			actions = append(actions, NewCoupAction(player, target))
		}

		// If player has 10+ coins, they MUST coup
		if player.GetCoins() >= 10 {
			// Filter to only keep coup actions
			coupActions := make([]Action, 0)
			for _, action := range actions {
				if action.Name() == "Coup" {
					coupActions = append(coupActions, action)
				}
			}
			return coupActions
		}
	}

	return actions
}

// GetValidTargets returns valid target players (alive and not self)
func (g *Game) GetValidTargets(player Player) []Player {
	targets := make([]Player, 0)
	for _, p := range g.Players {
		if p.GetID() != player.GetID() && p.IsAlive() {
			targets = append(targets, p)
		}
	}
	return targets
}

// GetStateForPlayer returns a read-only view of the game state for a player
func (g *Game) GetStateForPlayer(playerID int) GameState {
	playerStates := make([]PlayerState, len(g.Players))
	for i, player := range g.Players {
		playerStates[i] = PlayerState{
			ID:         player.GetID(),
			Coins:      player.GetCoins(),
			Influences: player.InfluenceCount(),
			IsAlive:    player.IsAlive(),
		}
	}

	// In a real implementation, we'd track known cards and remaining deck
	// For simplicity, we'll return empty maps here
	knownCards := make(map[int][]Card)
	remainingCards := make([]Card, 0)

	return GameState{
		PlayerID:       playerID,
		Players:        playerStates,
		CurrentPlayer:  g.CurrentPlayer,
		Turn:           g.Turn,
		KnownCards:     knownCards,
		RemainingCards: remainingCards,
	}
}

// NextPlayer advances to the next player
func (g *Game) NextPlayer() {
	g.CurrentPlayer = (g.CurrentPlayer + 1) % len(g.Players)
	g.Turn++
}

// CheckWinCondition checks if the game is over
func (g *Game) CheckWinCondition() {
	// Count living players
	livingPlayers := 0
	var lastLiving Player

	for _, player := range g.Players {
		if player.IsAlive() {
			livingPlayers++
			lastLiving = player
		}
	}

	// If only one player left, they win
	if livingPlayers == 1 {
		g.Finished = true
		g.Winner = lastLiving
	}
}

// BlockClaim represents a claim made when blocking
type BlockClaim struct {
	Actor  Player
	Action Action
	Card   Card
}

// NewBlockClaim creates a new block claim
func NewBlockClaim(actor Player, action Action, card Card) *BlockClaim {
	return &BlockClaim{
		Actor:  actor,
		Action: action,
		Card:   card,
	}
}

// Name returns the name of the claim
func (c *BlockClaim) Name() string {
	return fmt.Sprintf("Block %s with %s", c.Action.Name(), c.Card.Name)
}

// Execute does nothing for a block claim
func (c *BlockClaim) Execute(g *Game) error {
	return nil
}

// IsLegal always returns nil for a block claim
func (c *BlockClaim) IsLegal(g *Game, p Player) error {
	return nil
}

// RequiresTarget returns false for a block claim
func (c *BlockClaim) RequiresTarget() bool {
	return false
}

// CanBeBlocked returns false for a block claim
func (c *BlockClaim) CanBeBlocked() bool {
	return false
}

// CanBeChallenged returns true for a block claim
func (c *BlockClaim) CanBeChallenged() bool {
	return true
}

// GetRequiredCard returns the card required for this claim
func (c *BlockClaim) GetRequiredCard() Card {
	return c.Card
}
