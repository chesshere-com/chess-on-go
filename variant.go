package chessongo

// Variant identifies the rule set used by a Game.
type Variant uint8

const (
	VariantStandard Variant = iota
	VariantChess960
)

func (v Variant) String() string {
	switch v {
	case VariantStandard:
		return "standard"
	case VariantChess960:
		return "chess960"
	default:
		return "unknown"
	}
}

func validVariant(v Variant) bool {
	return v == VariantStandard || v == VariantChess960
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
