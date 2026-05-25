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
