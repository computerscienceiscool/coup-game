package game

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// Game represents the state of a Coup game
type Game struct {
	Players          []Player
	Deck             *Deck
	Discarded        []Card         // Influence cards lost and removed from play (face up)
	CurrentPlayer    int
	Turn             int
	Finished         bool
	Winner           Player
	ActionLog        []ActionLog
	EliminationTurns map[int]int    // playerID -> turn eliminated
	RNG              *rand.Rand     // Random number generator for reproducibility
	claims           map[int]map[string]bool // public claim history per player
}

// MixSeed derives an independent RNG seed from a base seed and an index using
// the SplitMix64 finalizer. Distinct indexes never share an RNG stream, no
// matter how close the inputs are — unlike naive base+index arithmetic, which
// collides whenever two (base, index) pairs have the same sum.
func MixSeed(base, index int64) int64 {
	z := uint64(base) + uint64(index)*0x9E3779B97F4A7C15
	z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
	z = (z ^ (z >> 27)) * 0x94D049BB133111EB
	return int64(z ^ (z >> 31))
}

// ActionLog records details of each action taken. Ground-truth fields
// (ActorHadCard, BlockerHadCard) record hidden information for analysis —
// they are never exposed to AI players.
type ActionLog struct {
	Turn              int
	PlayerID          int
	Action            string
	Target            int    // -1 if no target
	Success           bool   // The action ultimately took effect
	Challenged        bool   // The action's character claim was challenged
	ActorHadCard      bool   // Ground truth: actor held the claimed card (character actions only)
	Blocker           int    // Player who attempted a block, -1 if none (kept even if the block was defeated)
	BlockingCharacter string // Character claimed by the blocker (empty if no block attempt)
	BlockerHadCard    bool   // Ground truth: blocker held the claimed card
	BlockChallenged   bool   // The block claim was challenged
	BlockSucceeded    bool   // The block stopped the action
}

// GameState provides a read-only view of the game state for AI decision
// making. It contains only information a real player could know: public
// player state, the face-up discard pile, and the claim history.
type GameState struct {
	PlayerID      int              // ID of the player viewing this state
	Players       []PlayerState    // State of all players
	CurrentPlayer int              // Index of current player
	Turn          int              // Current turn number
	Discarded     []Card           // Lost influence, face up and out of play
	Claims        map[int][]string // Characters each player has publicly claimed (cleared on Exchange)
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

	// Create AI players, each with an independent RNG stream
	for i := 0; i < playerCount; i++ {
		players[i] = NewAIPlayer(i, &AIStrategy{
			BluffRate:     DefaultBluffRate,
			ChallengeRate: DefaultChallengeRate,
			AlwaysBlock:   true,
		}, MixSeed(seed, int64(i)+1))
	}

	game := &Game{
		Players:          players,
		Deck:             deck,
		CurrentPlayer:    0,
		Turn:             0,
		Finished:         false,
		ActionLog:        make([]ActionLog, 0),
		EliminationTurns: make(map[int]int),
		RNG:              rng,
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

	// One action taken = one turn
	g.Turn++

	// Check for win condition
	g.CheckWinCondition()

	// If game not finished, move to next player
	if !g.Finished {
		g.NextPlayer()
	}
}

// ResolveAction handles the challenge and block resolution process
func (g *Game) ResolveAction(player Player, action Action, log *ActionLog) bool {
	// Pay upfront costs before the challenge/block phase. Per the Coup rules
	// the cost stays paid if the action is blocked, but is returned if the
	// action itself is successfully challenged.
	cost := 0
	if action.Name() == "Assassinate" {
		cost = 3
	}
	if cost > 0 {
		if err := player.RemoveCoins(cost); err != nil {
			return false
		}
	}

	// Check if action can be challenged
	if action.CanBeChallenged() {
		// The action is a public character claim: record it, and record the
		// hidden truth of it for post-game analysis
		g.recordClaim(player.GetID(), action.GetRequiredCard().Name)
		log.ActorHadCard = player.HasCard(action.GetRequiredCard())

		// Allow other players to challenge, starting clockwise from acting player
		numPlayers := len(g.Players)
		for offset := 1; offset < numPlayers; offset++ {
			i := (g.CurrentPlayer + offset) % numPlayers
			opponent := g.Players[i]
			if i == player.GetID() || !opponent.IsAlive() {
				continue // Skip self and eliminated players
			}

			// Get state for opponent
			opponentState := g.GetStateForPlayer(i)

			// Opponent decides whether to challenge
			if opponent.ChallengeDecision(opponentState, player, action) {
				log.Challenged = true

				if g.ResolveChallenge(opponent, player, action) {
					// Challenge succeeded: the action fails and its cost is returned
					player.AddCoins(cost)
					return false
				}
				// Challenge failed, continue with action
				break // Only one challenge per action
			}
		}
	}

	// Check if action can be blocked
	if action.CanBeBlocked() && g.resolveBlocks(player, action, log) {
		return false
	}

	// If we get here, action wasn't successfully blocked or challenged
	// Execute the action
	err := action.Execute(g)
	return err == nil
}

// potentialBlockers returns the players who may block the action, in
// clockwise order from the acting player. Steal and Assassinate can only be
// blocked by their target; Foreign Aid can be blocked by any other player.
func (g *Game) potentialBlockers(player Player, action Action) []Player {
	if targeted, ok := action.(TargetedAction); ok {
		target := targeted.GetTarget()
		if target.IsAlive() && target.GetID() != player.GetID() {
			return []Player{target}
		}
		return nil
	}

	numPlayers := len(g.Players)
	blockers := make([]Player, 0, numPlayers-1)
	for offset := 1; offset < numPlayers; offset++ {
		i := (g.CurrentPlayer + offset) % numPlayers
		opponent := g.Players[i]
		if i == player.GetID() || !opponent.IsAlive() {
			continue
		}
		blockers = append(blockers, opponent)
	}
	return blockers
}

// resolveBlocks offers eligible players the chance to block and returns true
// if the action ends up blocked. An action can be blocked at most once: if
// the block is defeated by a challenge, the action proceeds.
func (g *Game) resolveBlocks(player Player, action Action, log *ActionLog) bool {
	for _, opponent := range g.potentialBlockers(player, action) {
		// Get state for opponent
		opponentState := g.GetStateForPlayer(opponent.GetID())

		// Opponent decides whether to block
		if !opponent.BlockDecision(opponentState, player, action) {
			continue
		}

		// The block is a public character claim: record it, and record the
		// hidden truth of it for post-game analysis. Defeated blocks stay in
		// the log (Blocker is not reset) so block statistics see every attempt.
		log.Blocker = opponent.GetID()
		blockingCharacter := opponent.ChooseBlockingCharacter(action)
		log.BlockingCharacter = blockingCharacter.Name
		log.BlockerHadCard = opponent.HasCard(blockingCharacter)
		g.recordClaim(opponent.GetID(), blockingCharacter.Name)

		// The block claim can be challenged, starting clockwise from the blocker
		numPlayers := len(g.Players)
		for bOffset := 1; bOffset < numPlayers; bOffset++ {
			j := (opponent.GetID() + bOffset) % numPlayers
			challenger := g.Players[j]
			if !challenger.IsAlive() {
				continue // Skip eliminated players
			}

			// Get state for challenger
			challengerState := g.GetStateForPlayer(j)

			// Create a proper block claim using our BlockClaim struct
			blockClaim := NewBlockClaim(opponent, action, blockingCharacter)

			// Challenger decides whether to challenge the block
			if challenger.ChallengeDecision(challengerState, opponent, blockClaim) {
				log.BlockChallenged = true
				if g.ResolveBlockChallenge(challenger, opponent, blockingCharacter) {
					// Blocker was bluffing: the block fails, the action proceeds
					log.BlockSucceeded = false
					return false
				}
				// Blocker proved the claim: the action is blocked
				log.BlockSucceeded = true
				return true
			}
		}

		// Block wasn't challenged: the action is blocked
		log.BlockSucceeded = true
		return true
	}

	return false
}

// recordClaim adds a character to a player's public claim history.
func (g *Game) recordClaim(playerID int, character string) {
	if character == "" {
		return
	}
	if g.claims == nil {
		g.claims = make(map[int]map[string]bool)
	}
	if g.claims[playerID] == nil {
		g.claims[playerID] = make(map[string]bool)
	}
	g.claims[playerID][character] = true
}

// ResetClaimsAfterExchange clears a player's claim history after a completed
// Ambassador exchange — their cards may have changed, so earlier claims no
// longer say anything about their hand. The Ambassador claim that enabled the
// exchange is kept.
func (g *Game) ResetClaimsAfterExchange(playerID int) {
	if g.claims != nil {
		delete(g.claims, playerID)
	}
	g.recordClaim(playerID, Ambassador)
}

// claimsFor returns a player's claimed characters, sorted for determinism.
func (g *Game) claimsFor(playerID int) []string {
	claimed := make([]string, 0, len(g.claims[playerID]))
	for character := range g.claims[playerID] {
		claimed = append(claimed, character)
	}
	sort.Strings(claimed)
	return claimed
}

// ResolveChallenge handles the challenge resolution. It returns true if the
// challenge succeeds (the claimant was bluffing).
func (g *Game) ResolveChallenge(challenger Player, challenged Player, action Action) bool {
	// Get required character for this action
	requiredCard := action.GetRequiredCard()

	// Check if challenged player has the claimed character
	if challenged.HasCard(requiredCard) {
		// Challenge fails: the claimant reveals the card, shuffles that same
		// card back into the deck, and draws a replacement, so their hand
		// size is unchanged
		revealed := challenged.RevealCard(requiredCard)
		g.Deck.Return([]Card{revealed})
		g.Deck.Shuffle()
		challenged.AddInfluence(g.Deck.Draw(1))

		// Challenger loses an influence, which is removed from play
		g.Discard(challenger.LoseInfluence())

		return false
	}

	// Challenge succeeds: the claimant loses an influence, removed from play
	g.Discard(challenged.LoseInfluence())

	return true
}

// ResolveBlockChallenge handles challenging a block. It returns true if the
// challenge succeeds (the blocker was bluffing).
func (g *Game) ResolveBlockChallenge(challenger Player, blocker Player, blockingCard Card) bool {
	// Check if blocker has the claimed blocking character
	if blocker.HasCard(blockingCard) {
		// Challenge fails: the blocker reveals the card, shuffles that same
		// card back into the deck, and draws a replacement, so their hand
		// size is unchanged
		revealed := blocker.RevealCard(blockingCard)
		g.Deck.Return([]Card{revealed})
		g.Deck.Shuffle()
		blocker.AddInfluence(g.Deck.Draw(1))

		// Challenger loses an influence, which is removed from play
		g.Discard(challenger.LoseInfluence())

		return false
	}

	// Challenge succeeds: the blocker loses an influence, removed from play
	g.Discard(blocker.LoseInfluence())

	return true
}

// Discard removes a lost influence card from play. Per the official rules a
// lost influence is revealed face up and never returns to the deck.
func (g *Game) Discard(card Card) {
	g.Discarded = append(g.Discarded, card)
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

	// Public information every real player can see: the face-up discard
	// pile and the claim history
	discarded := make([]Card, len(g.Discarded))
	copy(discarded, g.Discarded)

	claims := make(map[int][]string, len(g.claims))
	for id := range g.claims {
		claims[id] = g.claimsFor(id)
	}

	return GameState{
		PlayerID:      playerID,
		Players:       playerStates,
		CurrentPlayer: g.CurrentPlayer,
		Turn:          g.Turn,
		Discarded:     discarded,
		Claims:        claims,
	}
}

// NextPlayer advances to the next living player. The turn counter is not
// incremented here — turns are counted in ExecuteTurn, one per action taken,
// so skipping eliminated seats doesn't inflate game-length statistics.
func (g *Game) NextPlayer() {
	for i := 0; i < len(g.Players); i++ {
		g.CurrentPlayer = (g.CurrentPlayer + 1) % len(g.Players)
		if g.Players[g.CurrentPlayer].IsAlive() {
			return
		}
	}
}

// ValidateInvariants checks the card economy: every character has exactly
// CopiesPerCharacter copies across the deck, hands, and discard pile, and no
// player holds more than 2 influence cards. It returns an error describing
// the first violation found.
func (g *Game) ValidateInvariants() error {
	counts := make(map[string]int)
	total := 0
	count := func(cards []Card) {
		for _, c := range cards {
			counts[c.Name]++
			total++
		}
	}

	count(g.Deck.Cards)
	count(g.Discarded)
	for _, p := range g.Players {
		hand := p.GetInfluences()
		if len(hand) > 2 {
			return fmt.Errorf("player %d holds %d influence cards (max 2)", p.GetID(), len(hand))
		}
		count(hand)
	}

	expected := len(GetCharacters()) * CopiesPerCharacter
	if total != expected {
		return fmt.Errorf("%d cards in circulation, expected %d", total, expected)
	}
	for _, name := range GetCharacters() {
		if counts[name] != CopiesPerCharacter {
			return fmt.Errorf("%d copies of %s in circulation, expected %d", counts[name], name, CopiesPerCharacter)
		}
	}
	return nil
}

// CheckWinCondition checks if the game is over
func (g *Game) CheckWinCondition() {
	// Count living players and track eliminations
	livingPlayers := 0
	var lastLiving Player

	for _, player := range g.Players {
		if player.IsAlive() {
			livingPlayers++
			lastLiving = player
		} else {
			// Record elimination turn if not already recorded
			if _, recorded := g.EliminationTurns[player.GetID()]; !recorded {
				g.EliminationTurns[player.GetID()] = g.Turn
			}
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
