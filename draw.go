package chessongo

func (g *Game) hasInsufficientMaterial() bool {
	if g.Whites[QUEEN] > 0 || g.Whites[ROOK] > 0 || g.Whites[PAWN] > 0 {
		return false
	}
	if g.Blacks[QUEEN] > 0 || g.Blacks[ROOK] > 0 || g.Blacks[PAWN] > 0 {
		return false
	}
	if g.Whites[KNIGHT] > 0 && g.Whites[BISHOP] > 0 {
		return false
	}
	if g.Blacks[KNIGHT] > 0 && g.Blacks[BISHOP] > 0 {
		return false
	}

	if (g.Whites[BISHOP] > 0 || g.Blacks[BISHOP] > 0) && (g.Whites[KNIGHT] > 0 || g.Blacks[KNIGHT] > 0) {
		return false
	}

	if g.Whites[BISHOP] > 0 || g.Blacks[BISHOP] > 0 {
		return allBishopsOnSameColor(g.Whites[BISHOP] | g.Blacks[BISHOP])
	}

	if g.Whites[KNIGHT].NumberOfSetBits() > 1 {
		return false
	}

	if g.Blacks[KNIGHT].NumberOfSetBits() > 1 {
		return false
	}

	return true
}

func allBishopsOnSameColor(bishops Bitboard) bool {
	if bishops == 0 {
		return true
	}
	first := Square(bishops.popLSB())
	color := squareColor(first)
	for bishops > 0 {
		if squareColor(Square(bishops.popLSB())) != color {
			return false
		}
	}
	return true
}

func squareColor(square Square) int {
	return (square.Rank() + square.File()) & 1
}
