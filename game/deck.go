package game

import (
	"math/rand"
)

// Deck represents the court deck in Coup
type Deck struct {
	Cards []Card
	RNG   *rand.Rand // For reproducible shuffling
}

// NewDeck creates a new deck with 3 of each character card
func NewDeck(rng *rand.Rand) *Deck {
	characters := GetCharacters()
	cards := make([]Card, 0, len(characters)*3)

	// Add 3 copies of each character
	for _, char := range characters {
		for i := 0; i < 3; i++ {
			cards = append(cards, GetCardByName(char))
		}
	}

	deck := &Deck{
		Cards: cards,
		RNG:   rng,
	}

	return deck
}

// Shuffle randomizes the order of cards in the deck
func (d *Deck) Shuffle() {
	// Fisher-Yates shuffle algorithm
	n := len(d.Cards)
	for i := n - 1; i > 0; i-- {
		j := d.RNG.Intn(i + 1)
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	}
}

// Draw removes and returns n cards from the top of the deck
func (d *Deck) Draw(n int) []Card {
	if n > len(d.Cards) {
		// If not enough cards, draw all remaining
		n = len(d.Cards)
	}

	drawn := d.Cards[:n]
	d.Cards = d.Cards[n:]

	return drawn
}

// Return adds cards back to the deck (doesn't shuffle)
func (d *Deck) Return(cards []Card) {
	d.Cards = append(d.Cards, cards...)
}

// Size returns the number of cards in the deck
func (d *Deck) Size() int {
	return len(d.Cards)
}

// Peek returns but doesn't remove the top n cards
func (d *Deck) Peek(n int) []Card {
	if n > len(d.Cards) {
		n = len(d.Cards)
	}

	// Create copies to avoid modification
	result := make([]Card, n)
	for i := 0; i < n; i++ {
		result[i] = d.Cards[i].Copy()
	}

	return result
}
