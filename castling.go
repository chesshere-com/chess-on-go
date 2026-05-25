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

func (g *Game) castlingSpec(right int) (castlingSpec, bool) {
	color := castlingRightColor(right)
	if color == NO_COLOR {
		return castlingSpec{}, false
	}

	kingside := castlingRightIsKingside(right)
	kingTo, rookTo := castlingFinalSquares(color, kingside)
	kingFrom := Square(W_KING_INIT_SQUARE)
	rookFrom := Square(WKS_ROOK_ORIGINAL_SQUARE)

	switch right {
	case CASTLE_WQS:
		rookFrom = WQS_ROOK_ORIGINAL_SQUARE
	case CASTLE_BKS:
		kingFrom = B_KING_INIT_SQUARE
		rookFrom = BKS_ROOK_ORIGINAL_SQUARE
	case CASTLE_BQS:
		kingFrom = B_KING_INIT_SQUARE
		rookFrom = BQS_ROOK_ORIGINAL_SQUARE
	}

	if g.variant == VariantChess960 {
		kingFrom = g.kingSquare(color)
		rookFrom = g.castlingRookFrom[right]
		if !kingFrom.Valid() || !rookFrom.Valid() {
			return castlingSpec{}, false
		}
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
	for _, right := range [...]int{CASTLE_WKS, CASTLE_WQS, CASTLE_BKS, CASTLE_BQS} {
		spec, ok := g.castlingSpec(right)
		if !ok {
			continue
		}
		if spec.kingFrom == m.From() && spec.kingTo == m.To() {
			return spec, true
		}
	}
	return castlingSpec{}, false
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
