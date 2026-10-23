package chessongo

func (g *Game) hasInsufficientMaterial() bool {
	if g.variant == VariantKingOfTheHill {
		return false // Even a bare king can reach the center.
	}
	if g.variant == VariantThreeCheck {
		return g.occupied == g.whites[KING]|g.blacks[KING]
	}
	if g.whites[QUEEN] > 0 || g.whites[ROOK] > 0 || g.whites[PAWN] > 0 {
		return false
	}
	if g.blacks[QUEEN] > 0 || g.blacks[ROOK] > 0 || g.blacks[PAWN] > 0 {
		return false
	}
	if g.whites[KNIGHT] > 0 && g.whites[BISHOP] > 0 {
		return false
	}
	if g.blacks[KNIGHT] > 0 && g.blacks[BISHOP] > 0 {
		return false
	}

	if (g.whites[BISHOP] > 0 || g.blacks[BISHOP] > 0) && (g.whites[KNIGHT] > 0 || g.blacks[KNIGHT] > 0) {
		return false
	}

	if g.whites[BISHOP] > 0 || g.blacks[BISHOP] > 0 {
		return allBishopsOnSameColor(g.whites[BISHOP] | g.blacks[BISHOP])
	}

	// Two knights, including one per side, can participate in a mating position.
	return (g.whites[KNIGHT] | g.blacks[KNIGHT]).NumberOfSetBits() <= 1
}

const (
	lightSquares Bitboard = 0xAA55AA55AA55AA55
	darkSquares  Bitboard = ^lightSquares
)

func allBishopsOnSameColor(bishops Bitboard) bool {
	return (bishops&lightSquares == 0) || (bishops&darkSquares == 0)
}
