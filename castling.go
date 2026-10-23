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

type castlingSpec struct {
	color       Color
	right       int
	kingFrom    Square
	kingTo      Square
	rookFrom    Square
	rookTo      Square
	kingPiece   Piece
	rookPiece   Piece
	emptyMask   Bitboard
	safeSquares Bitboard
}

// Standard castling geometry never changes. Keep it out of the per-node path.
var standardCastlingSpecs = [16]castlingSpec{
	CASTLE_WKS: {WHITE, CASTLE_WKS, 60, 62, 63, 61, W_KING, W_ROOK, 0x6000000000000000, 0x7000000000000000},
	CASTLE_WQS: {WHITE, CASTLE_WQS, 60, 58, 56, 59, W_KING, W_ROOK, 0x0e00000000000000, 0x1c00000000000000},
	CASTLE_BKS: {BLACK, CASTLE_BKS, 4, 6, 7, 5, B_KING, B_ROOK, 0x60, 0x70},
	CASTLE_BQS: {BLACK, CASTLE_BQS, 4, 2, 0, 3, B_KING, B_ROOK, 0x0e, 0x1c},
}

func (g *Game) castlingSpec(right int) (castlingSpec, bool) {
	color := castlingRightColor(right)
	if color == NO_COLOR {
		return castlingSpec{}, false
	}
	if g.variant != VariantChess960 {
		return standardCastlingSpecs[right], true
	}

	kingside := castlingRightIsKingside(right)
	kingTo, rookTo := castlingFinalSquares(color, kingside)
	kingFrom := g.kingSquare(color)
	rookFrom := g.castlingRookFrom[right]
	if !kingFrom.Valid() || !rookFrom.Valid() {
		return castlingSpec{}, false
	}

	spec := castlingSpec{
		color:       color,
		right:       right,
		kingFrom:    kingFrom,
		kingTo:      kingTo,
		rookFrom:    rookFrom,
		rookTo:      rookTo,
		kingPiece:   W_KING,
		rookPiece:   W_ROOK,
		safeSquares: castlingKingPathMask(kingFrom, kingTo),
	}
	if color == BLACK {
		spec.kingPiece = B_KING
		spec.rookPiece = B_ROOK
	}
	spec.emptyMask = g.castlingEmptyMask(spec)
	return spec, true
}

func castlingFinalSquares(color Color, kingside bool) (Square, Square) {
	if kingside {
		return squareOnBackRank(color, 6), squareOnBackRank(color, 5)
	}
	return squareOnBackRank(color, 2), squareOnBackRank(color, 3)
}

func (g *Game) castlingEmptyMask(spec castlingSpec) Bitboard {
	occupiedByCastlingPieces := (Bitboard(1) << spec.kingFrom) | (Bitboard(1) << spec.rookFrom)
	return (castlingBetweenInclusiveMask(spec.kingFrom, spec.kingTo) |
		castlingBetweenInclusiveMask(spec.rookFrom, spec.rookTo)) &^ occupiedByCastlingPieces
}

func castlingKingPathMask(from, to Square) Bitboard {
	return castlingBetweenInclusiveMask(from, to)
}

func castlingBetweenInclusiveMask(from, to Square) Bitboard {
	if !from.Valid() || !to.Valid() {
		return 0
	}
	mask := Bitboard(1) << from
	if from == to {
		return mask
	}

	rankDelta := to.Rank() - from.Rank()
	fileDelta := to.File() - from.File()
	rankStep := sign(rankDelta)
	fileStep := sign(fileDelta)
	if rankDelta != 0 && fileDelta != 0 && abs(rankDelta) != abs(fileDelta) {
		return 0
	}

	rank := from.Rank()
	file := from.File()
	for rank != to.Rank() || file != to.File() {
		rank += rankStep
		file += fileStep
		mask |= Bitboard(1) << CoordsToSquare(rank, file)
	}
	return mask
}

func (g *Game) castlingSpecForMove(m Move) (castlingSpec, bool) {
	if !m.IsCastlingMove() {
		return castlingSpec{}, false
	}
	right := castlingRightForMoveDestination(g.turn, m.To())
	spec, ok := g.castlingSpec(right)
	return spec, ok && spec.kingFrom == m.From()
}

func (g *Game) castlingSpecForUndoMove(m Move, color Color) (castlingSpec, bool) {
	if !m.IsCastlingMove() {
		return castlingSpec{}, false
	}
	right := castlingRightForMoveDestination(color, m.To())
	if right == 0 {
		return castlingSpec{}, false
	}
	kingside := castlingRightIsKingside(right)
	kingTo, rookTo := castlingFinalSquares(color, kingside)
	rookFrom := g.castlingRookFrom[right]
	if g.variant != VariantChess960 {
		switch right {
		case CASTLE_WKS:
			rookFrom = WKS_ROOK_ORIGINAL_SQUARE
		case CASTLE_WQS:
			rookFrom = WQS_ROOK_ORIGINAL_SQUARE
		case CASTLE_BKS:
			rookFrom = BKS_ROOK_ORIGINAL_SQUARE
		case CASTLE_BQS:
			rookFrom = BQS_ROOK_ORIGINAL_SQUARE
		}
	}
	if !m.From().Valid() || !rookFrom.Valid() {
		return castlingSpec{}, false
	}
	spec := castlingSpec{
		color:     color,
		right:     right,
		kingFrom:  m.From(),
		kingTo:    kingTo,
		rookFrom:  rookFrom,
		rookTo:    rookTo,
		kingPiece: W_KING,
		rookPiece: W_ROOK,
	}
	if color == BLACK {
		spec.kingPiece = B_KING
		spec.rookPiece = B_ROOK
	}
	return spec, true
}

func castlingRightForMoveDestination(color Color, to Square) int {
	kingsideKingTo, _ := castlingFinalSquares(color, true)
	if to == kingsideKingTo {
		return castlingRightFor(color, true)
	}
	queensideKingTo, _ := castlingFinalSquares(color, false)
	if to == queensideKingTo {
		return castlingRightFor(color, false)
	}
	return 0
}

func (g *Game) castlingKingPathIsSafe(spec castlingSpec, opponent Color) bool {
	squares := spec.safeSquares
	for squares > 0 {
		square := Square(squares.popLSB())
		occupied := g.occupied
		occupied &^= Bitboard(1) << spec.kingFrom
		occupied &^= Bitboard(1) << spec.rookFrom
		occupied |= Bitboard(1) << square
		occupied |= Bitboard(1) << spec.rookTo
		if g.isSquareAttackedByWithOccupied(square, opponent, occupied) {
			return false
		}
	}
	return true
}

func (g *Game) clearCastlingRightsForRookSquare(square Square) {
	for _, right := range [...]int{CASTLE_WKS, CASTLE_WQS, CASTLE_BKS, CASTLE_BQS} {
		if g.castlingRookFrom[right] == square {
			g.castling &^= right
		}
	}
}
