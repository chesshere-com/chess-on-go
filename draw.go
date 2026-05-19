package chessongo

func (g *Game) hasInsufficientMaterial() bool {
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

	if g.whites[KNIGHT].NumberOfSetBits() > 1 {
		return false
	}

	if g.blacks[KNIGHT].NumberOfSetBits() > 1 {
		return false
	}

	return true
}

const (
	lightSquares Bitboard = 0xAA55AA55AA55AA55
	darkSquares  Bitboard = ^lightSquares
)

func allBishopsOnSameColor(bishops Bitboard) bool {
	return (bishops & lightSquares == 0) || (bishops & darkSquares == 0)
}

func squareColor(square Square) int {
	return (square.Rank() + square.File()) & 1
}
