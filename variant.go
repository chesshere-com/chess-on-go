package chessongo

// Variant identifies the rule set used by a Game.
type Variant uint8

const (
	VariantStandard Variant = iota
	VariantChess960
	VariantKingOfTheHill
	VariantThreeCheck
	VariantAtomic
	VariantAntichess
	VariantHorde
	VariantCrazyhouse
)

const (
	whiteStateIndex = 0
	blackStateIndex = 1
)

type variantState struct {
	checksGiven [2]uint8
	pockets     [2][7]uint8
	promoted    Bitboard
}

func (v Variant) String() string {
	switch v {
	case VariantStandard:
		return "standard"
	case VariantChess960:
		return "chess960"
	case VariantKingOfTheHill:
		return "kingofthehill"
	case VariantThreeCheck:
		return "threecheck"
	case VariantAtomic:
		return "atomic"
	case VariantAntichess:
		return "antichess"
	case VariantHorde:
		return "horde"
	case VariantCrazyhouse:
		return "crazyhouse"
	default:
		return "unknown"
	}
}

func validVariant(v Variant) bool {
	return rulesForVariant(v).implemented
}

func variantColorIndex(color Color) int {
	if color == BLACK {
		return blackStateIndex
	}
	return whiteStateIndex
}

// Variant returns the rule variant for this game.
func (g *Game) Variant() Variant {
	return g.variant
}

// NewGameFromFENWithVariant creates a game initialized from FEN under a specific variant.
func NewGameFromFENWithVariant(fen string, variant Variant) (*Game, error) {
	g := &Game{}
	if err := g.LoadFENWithVariant(fen, variant); err != nil {
		return nil, err
	}
	return g, nil
}

// NewChess960Game creates a Chess960 game from a position ID in [0, 959].
func NewChess960Game(position int) (*Game, error) {
	fen, err := Chess960StartingFEN(position)
	if err != nil {
		return nil, err
	}
	return NewGameFromFENWithVariant(fen, VariantChess960)
}

// LoadFENWithVariant initializes the game from FEN using variant-specific parsing and rules.
func (g *Game) LoadFENWithVariant(fen string, variant Variant) error {
	if !validVariant(variant) {
		return invalidFEN("invalid variant")
	}
	return g.loadFEN(fen, variant)
}
