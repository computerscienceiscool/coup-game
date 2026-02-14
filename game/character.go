package game

// Card represents a character card in Coup
type Card struct {
	Name  string
	ID    int
	Shown bool // Whether card has been revealed during a challenge
}

// Character constants
const (
	Duke       = "Duke"
	Assassin   = "Assassin"
	Captain    = "Captain"
	Ambassador = "Ambassador"
	Contessa   = "Contessa"
)

// Character IDs
const (
	DukeID = iota
	AssassinID
	CaptainID
	AmbassadorID
	ContessaID
)

// GetCharacters returns a slice of all character types
func GetCharacters() []string {
	return []string{Duke, Assassin, Captain, Ambassador, Contessa}
}

// GetCardByName returns a card by its name
func GetCardByName(name string) Card {
	switch name {
	case Duke:
		return Card{Name: Duke, ID: DukeID, Shown: false}
	case Assassin:
		return Card{Name: Assassin, ID: AssassinID, Shown: false}
	case Captain:
		return Card{Name: Captain, ID: CaptainID, Shown: false}
	case Ambassador:
		return Card{Name: Ambassador, ID: AmbassadorID, Shown: false}
	case Contessa:
		return Card{Name: Contessa, ID: ContessaID, Shown: false}
	default:
		return Card{Name: "Unknown", ID: -1, Shown: false}
	}
}

// BlocksAction checks if a character can block a given action
func (c Card) BlocksAction(action string) bool {
	switch c.Name {
	case Duke:
		// Duke blocks Foreign Aid
		return action == "Foreign Aid"
	case Captain:
		// Captain blocks Steal
		return action == "Steal"
	case Ambassador:
		// Ambassador blocks Steal
		return action == "Steal"
	case Contessa:
		// Contessa blocks Assassinate
		return action == "Assassinate"
	default:
		return false
	}
}

// GetBlockingCharacters returns all characters that can block a given action
func GetBlockingCharacters(action string) []Card {
	result := make([]Card, 0)

	switch action {
	case "Foreign Aid":
		result = append(result, GetCardByName(Duke))
	case "Steal":
		result = append(result, GetCardByName(Captain))
		result = append(result, GetCardByName(Ambassador))
	case "Assassinate":
		result = append(result, GetCardByName(Contessa))
	}

	return result
}

// IsEqual checks if two cards are the same
func (c Card) IsEqual(other Card) bool {
	return c.Name == other.Name && c.ID == other.ID
}

// Copy returns a copy of the card
func (c Card) Copy() Card {
	return Card{
		Name:  c.Name,
		ID:    c.ID,
		Shown: c.Shown,
	}
}
