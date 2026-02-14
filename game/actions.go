package game

import (
	"errors"
)

// Action represents any game action
type Action interface {
	Name() string
	Execute(g *Game) error
	IsLegal(g *Game, p Player) error
	RequiresTarget() bool
	CanBeBlocked() bool
	CanBeChallenged() bool
	GetRequiredCard() Card
}

// TargetedAction is an action that has a target player
type TargetedAction interface {
	Action
	GetTarget() Player
}

// BaseAction provides common functionality for all actions
type BaseAction struct {
	Actor        Player
	RequiredChar string // Required character to perform action (empty if any)
}

// IncomeAction represents taking 1 coin from the treasury
type IncomeAction struct {
	BaseAction
}

// ForeignAidAction represents taking 2 coins from the treasury
type ForeignAidAction struct {
	BaseAction
}

// TaxAction represents taking 3 coins (Duke ability)
type TaxAction struct {
	BaseAction
}

// StealAction represents stealing 2 coins from another player
type StealAction struct {
	BaseAction
	Target Player
}

// AssassinateAction represents assassinating another player's influence
type AssassinateAction struct {
	BaseAction
	Target Player
}

// CoupAction represents staging a coup against another player
type CoupAction struct {
	BaseAction
	Target Player
}

// ExchangeAction represents exchanging cards with the court deck
type ExchangeAction struct {
	BaseAction
}

// NewIncomeAction creates a new Income action
func NewIncomeAction(actor Player) *IncomeAction {
	return &IncomeAction{
		BaseAction: BaseAction{
			Actor:        actor,
			RequiredChar: "", // No character required
		},
	}
}

// Name returns the action name
func (a *IncomeAction) Name() string {
	return "Income"
}

// Execute performs the income action
func (a *IncomeAction) Execute(g *Game) error {
	if err := a.IsLegal(g, a.Actor); err != nil {
		return err
	}

	// Take 1 coin
	a.Actor.AddCoins(1)
	return nil
}

// IsLegal checks if the action is legal
func (a *IncomeAction) IsLegal(g *Game, p Player) error {
	if p.GetID() != a.Actor.GetID() {
		return errors.New("player is not the actor")
	}

	if p.GetCoins() >= 10 {
		return errors.New("player must coup with 10+ coins")
	}

	return nil
}

// RequiresTarget returns whether this action requires a target
func (a *IncomeAction) RequiresTarget() bool {
	return false
}

// CanBeBlocked returns whether this action can be blocked
func (a *IncomeAction) CanBeBlocked() bool {
	return false
}

// CanBeChallenged returns whether this action can be challenged
func (a *IncomeAction) CanBeChallenged() bool {
	return false
}

// GetRequiredCard returns the card required for this action
func (a *IncomeAction) GetRequiredCard() Card {
	return Card{} // No card required
}

// NewForeignAidAction creates a new Foreign Aid action
func NewForeignAidAction(actor Player) *ForeignAidAction {
	return &ForeignAidAction{
		BaseAction: BaseAction{
			Actor:        actor,
			RequiredChar: "", // No character required
		},
	}
}

// Name returns the action name
func (a *ForeignAidAction) Name() string {
	return "Foreign Aid"
}

// Execute performs the foreign aid action
func (a *ForeignAidAction) Execute(g *Game) error {
	if err := a.IsLegal(g, a.Actor); err != nil {
		return err
	}

	// Take 2 coins
	a.Actor.AddCoins(2)
	return nil
}

// IsLegal checks if the action is legal
func (a *ForeignAidAction) IsLegal(g *Game, p Player) error {
	if p.GetID() != a.Actor.GetID() {
		return errors.New("player is not the actor")
	}

	if p.GetCoins() >= 10 {
		return errors.New("player must coup with 10+ coins")
	}

	return nil
}

// RequiresTarget returns whether this action requires a target
func (a *ForeignAidAction) RequiresTarget() bool {
	return false
}

// CanBeBlocked returns whether this action can be blocked
func (a *ForeignAidAction) CanBeBlocked() bool {
	return true
}

// CanBeChallenged returns whether this action can be challenged
func (a *ForeignAidAction) CanBeChallenged() bool {
	return false
}

// GetRequiredCard returns the card required for this action
func (a *ForeignAidAction) GetRequiredCard() Card {
	return Card{} // No card required
}

// NewTaxAction creates a new Tax action
func NewTaxAction(actor Player) *TaxAction {
	return &TaxAction{
		BaseAction: BaseAction{
			Actor:        actor,
			RequiredChar: Duke, // Requires Duke
		},
	}
}

// Name returns the action name
func (a *TaxAction) Name() string {
	return "Tax"
}

// Execute performs the tax action
func (a *TaxAction) Execute(g *Game) error {
	if err := a.IsLegal(g, a.Actor); err != nil {
		return err
	}

	// Take 3 coins
	a.Actor.AddCoins(3)
	return nil
}

// IsLegal checks if the action is legal
func (a *TaxAction) IsLegal(g *Game, p Player) error {
	if p.GetID() != a.Actor.GetID() {
		return errors.New("player is not the actor")
	}

	if p.GetCoins() >= 10 {
		return errors.New("player must coup with 10+ coins")
	}

	return nil
}

// RequiresTarget returns whether this action requires a target
func (a *TaxAction) RequiresTarget() bool {
	return false
}

// CanBeBlocked returns whether this action can be blocked
func (a *TaxAction) CanBeBlocked() bool {
	return false
}

// CanBeChallenged returns whether this action can be challenged
func (a *TaxAction) CanBeChallenged() bool {
	return true
}

// GetRequiredCard returns the card required for this action
func (a *TaxAction) GetRequiredCard() Card {
	return GetCardByName(Duke)
}

// NewStealAction creates a new Steal action
func NewStealAction(actor, target Player) *StealAction {
	return &StealAction{
		BaseAction: BaseAction{
			Actor:        actor,
			RequiredChar: Captain, // Requires Captain
		},
		Target: target,
	}
}

// Name returns the action name
func (a *StealAction) Name() string {
	return "Steal"
}

// Execute performs the steal action
func (a *StealAction) Execute(g *Game) error {
	if err := a.IsLegal(g, a.Actor); err != nil {
		return err
	}

	// Amount to steal (up to 2 coins)
	amount := 2
	if a.Target.GetCoins() < amount {
		amount = a.Target.GetCoins()
	}

	// Transfer coins
	if err := a.Target.RemoveCoins(amount); err != nil {
		return err
	}
	a.Actor.AddCoins(amount)

	return nil
}

// IsLegal checks if the action is legal
func (a *StealAction) IsLegal(g *Game, p Player) error {
	if p.GetID() != a.Actor.GetID() {
		return errors.New("player is not the actor")
	}

	if p.GetCoins() >= 10 {
		return errors.New("player must coup with 10+ coins")
	}

	if !a.Target.IsAlive() {
		return errors.New("target is not alive")
	}

	if a.Target.GetCoins() == 0 {
		return errors.New("target has no coins to steal")
	}

	return nil
}

// RequiresTarget returns whether this action requires a target
func (a *StealAction) RequiresTarget() bool {
	return true
}

// GetTarget returns the target player
func (a *StealAction) GetTarget() Player {
	return a.Target
}

// CanBeBlocked returns whether this action can be blocked
func (a *StealAction) CanBeBlocked() bool {
	return true
}

// CanBeChallenged returns whether this action can be challenged
func (a *StealAction) CanBeChallenged() bool {
	return true
}

// GetRequiredCard returns the card required for this action
func (a *StealAction) GetRequiredCard() Card {
	return GetCardByName(Captain)
}

// NewAssassinateAction creates a new Assassinate action
func NewAssassinateAction(actor, target Player) *AssassinateAction {
	return &AssassinateAction{
		BaseAction: BaseAction{
			Actor:        actor,
			RequiredChar: Assassin, // Requires Assassin
		},
		Target: target,
	}
}

// Name returns the action name
func (a *AssassinateAction) Name() string {
	return "Assassinate"
}

// Execute performs the assassinate action
func (a *AssassinateAction) Execute(g *Game) error {
	// Note: 3-coin cost is paid upfront in ResolveAction (per Coup rules,
	// the cost is paid even if the action is blocked or challenged)

	if !a.Target.IsAlive() {
		return errors.New("target is not alive")
	}

	// Target loses an influence
	card := a.Target.LoseInfluence()
	g.Deck.Return([]Card{card})
	g.Deck.Shuffle()

	return nil
}

// IsLegal checks if the action is legal
func (a *AssassinateAction) IsLegal(g *Game, p Player) error {
	if p.GetID() != a.Actor.GetID() {
		return errors.New("player is not the actor")
	}

	if p.GetCoins() < 3 {
		return errors.New("not enough coins to assassinate")
	}

	if p.GetCoins() >= 10 {
		return errors.New("player must coup with 10+ coins")
	}

	if !a.Target.IsAlive() {
		return errors.New("target is not alive")
	}

	return nil
}

// RequiresTarget returns whether this action requires a target
func (a *AssassinateAction) RequiresTarget() bool {
	return true
}

// GetTarget returns the target player
func (a *AssassinateAction) GetTarget() Player {
	return a.Target
}

// CanBeBlocked returns whether this action can be blocked
func (a *AssassinateAction) CanBeBlocked() bool {
	return true
}

// CanBeChallenged returns whether this action can be challenged
func (a *AssassinateAction) CanBeChallenged() bool {
	return true
}

// GetRequiredCard returns the card required for this action
func (a *AssassinateAction) GetRequiredCard() Card {
	return GetCardByName(Assassin)
}

// NewCoupAction creates a new Coup action
func NewCoupAction(actor, target Player) *CoupAction {
	return &CoupAction{
		BaseAction: BaseAction{
			Actor:        actor,
			RequiredChar: "", // No character required
		},
		Target: target,
	}
}

// Name returns the action name
func (a *CoupAction) Name() string {
	return "Coup"
}

// Execute performs the coup action
func (a *CoupAction) Execute(g *Game) error {
	if err := a.IsLegal(g, a.Actor); err != nil {
		return err
	}

	// Pay 7 coins
	if err := a.Actor.RemoveCoins(7); err != nil {
		return err
	}

	// Target loses an influence
	card := a.Target.LoseInfluence()
	g.Deck.Return([]Card{card})
	g.Deck.Shuffle()

	return nil
}

// IsLegal checks if the action is legal
func (a *CoupAction) IsLegal(g *Game, p Player) error {
	if p.GetID() != a.Actor.GetID() {
		return errors.New("player is not the actor")
	}

	if p.GetCoins() < 7 {
		return errors.New("not enough coins for coup")
	}

	if !a.Target.IsAlive() {
		return errors.New("target is not alive")
	}

	return nil
}

// RequiresTarget returns whether this action requires a target
func (a *CoupAction) RequiresTarget() bool {
	return true
}

// GetTarget returns the target player
func (a *CoupAction) GetTarget() Player {
	return a.Target
}

// CanBeBlocked returns whether this action can be blocked
func (a *CoupAction) CanBeBlocked() bool {
	return false
}

// CanBeChallenged returns whether this action can be challenged
func (a *CoupAction) CanBeChallenged() bool {
	return false
}

// GetRequiredCard returns the card required for this action
func (a *CoupAction) GetRequiredCard() Card {
	return Card{} // No card required
}

// NewExchangeAction creates a new Exchange action
func NewExchangeAction(actor Player) *ExchangeAction {
	return &ExchangeAction{
		BaseAction: BaseAction{
			Actor:        actor,
			RequiredChar: Ambassador, // Requires Ambassador
		},
	}
}

// Name returns the action name
func (a *ExchangeAction) Name() string {
	return "Exchange"
}

// Execute performs the exchange action
func (a *ExchangeAction) Execute(g *Game) error {
	if err := a.IsLegal(g, a.Actor); err != nil {
		return err
	}

	// Draw 2 cards (or as many as available)
	drawCount := 2
	if g.Deck.Size() < drawCount {
		drawCount = g.Deck.Size()
	}

	drawn := g.Deck.Draw(drawCount)

	// Let player choose which cards to keep/return
	returned := a.Actor.ChooseExchange(drawn)

	// Return cards to deck and shuffle
	g.Deck.Return(returned)
	g.Deck.Shuffle()

	return nil
}

// IsLegal checks if the action is legal
func (a *ExchangeAction) IsLegal(g *Game, p Player) error {
	if p.GetID() != a.Actor.GetID() {
		return errors.New("player is not the actor")
	}

	if p.GetCoins() >= 10 {
		return errors.New("player must coup with 10+ coins")
	}

	if g.Deck.Size() == 0 {
		return errors.New("deck is empty, cannot exchange")
	}

	return nil
}

// RequiresTarget returns whether this action requires a target
func (a *ExchangeAction) RequiresTarget() bool {
	return false
}

// CanBeBlocked returns whether this action can be blocked
func (a *ExchangeAction) CanBeBlocked() bool {
	return false
}

// CanBeChallenged returns whether this action can be challenged
func (a *ExchangeAction) CanBeChallenged() bool {
	return true
}

// GetRequiredCard returns the card required for this action
func (a *ExchangeAction) GetRequiredCard() Card {
	return GetCardByName(Ambassador)
}
