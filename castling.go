package chessongo

const noSquare Square = 64

func defaultCastlingRookFrom() [16]Square {
	var rooks [16]Square
	for i := range rooks {
		rooks[i] = noSquare
	}
	rooks[CASTLE_WKS] = WKS_ROOK_ORIGINAL_SQUARE
	rooks[CASTLE_WQS] = WQS_ROOK_ORIGINAL_SQUARE
	rooks[CASTLE_BKS] = BKS_ROOK_ORIGINAL_SQUARE
	rooks[CASTLE_BQS] = BQS_ROOK_ORIGINAL_SQUARE
	return rooks
}

func castlingRightFor(color Color, kingside bool) int {
	if color == WHITE {
		if kingside {
			return CASTLE_WKS
		}
		return CASTLE_WQS
	}
	if kingside {
		return CASTLE_BKS
	}
	return CASTLE_BQS
}

func backRankFor(color Color) int {
	if color == WHITE {
		return 7
	}
	return 0
}

func squareOnBackRank(color Color, file int) Square {
	return CoordsToSquare(backRankFor(color), file)
}

func castlingRightColor(right int) Color {
	switch right {
	case CASTLE_WKS, CASTLE_WQS:
		return WHITE
	case CASTLE_BKS, CASTLE_BQS:
		return BLACK
	default:
		return NO_COLOR
	}
}

func castlingRightIsKingside(right int) bool {
	return right == CASTLE_WKS || right == CASTLE_BKS
}
