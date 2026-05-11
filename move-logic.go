package chessongo

// Checks whether our king is in check or not
func (g *Game) ComputeIsCheck() bool {
	var kingBB, theirsAll, attackers Bitboard
	var theirs []Bitboard
	if g.Turn == WHITE {
		kingBB, theirs, theirsAll = g.Whites[KING], g.Blacks[:], g.BlackPieces
	} else {
		kingBB, theirs, theirsAll = g.Blacks[KING], g.Whites[:], g.WhitePieces
	}
	if kingBB == 0 {
		return false
	}
	kingIdx := kingBB.lsbIndex()
	possibleAttackers := theirsAll & ATTACKS_TO[kingIdx]

	attackers = (theirs[ROOK] | theirs[QUEEN]) & possibleAttackers
	if attackers > 0 && rookAttacks(Square(kingIdx), g.Occupied)&attackers > 0 {
		return true
	}

	attackers = (theirs[BISHOP] | theirs[QUEEN]) & possibleAttackers
	if attackers > 0 && bishopAttacks(Square(kingIdx), g.Occupied)&attackers > 0 {
		return true
	}

	attackers = theirs[KNIGHT] & possibleAttackers
	for attackers > 0 {
		from := attackers.popLSB()
		if KNIGHT_ATTACKS_FROM[from]&kingBB > 0 {
			return true
		}
	}

	if g.Turn == WHITE {
		// Black pawns attack “down” the board (towards higher square indices).
		if ((g.Blacks[PAWN]&^Bitboard(FILE_A_MASK))<<7)&kingBB > 0 ||
			((g.Blacks[PAWN]&^Bitboard(FILE_H_MASK))<<9)&kingBB > 0 {
			return true
		}
	} else {
		// White pawns attack “up” the board (towards lower square indices).
		if ((g.Whites[PAWN]&^Bitboard(FILE_H_MASK))>>7)&kingBB > 0 ||
			((g.Whites[PAWN]&^Bitboard(FILE_A_MASK))>>9)&kingBB > 0 {
			return true
		}
	}

	attackers = theirs[KING] & possibleAttackers
	if attackers > 0 {
		from := attackers.popLSB()
		if KING_ATTACKS_FROM[from]&kingBB > 0 {
			return true
		}
	}
	return false
}

// Checks whether the given move is possible or not
func (g *Game) CanMove(m Move) bool {
	if m.IsCastlingMove() {
		var inBetweenSq Square
		if m.To() == WKS_KING_TO_SQUARE || m.To() == BKS_KING_TO_SQUARE {
			inBetweenSq = m.From() + 1
		} else if m.To() == WQS_KING_TO_SQUARE || m.To() == BQS_KING_TO_SQUARE {
			inBetweenSq = m.From() - 1
		}
		inBetweenMove := NewMove(m.From(), inBetweenSq, EMPTY)
		if g.IsCheck || g.WillMoveCauseCheck(inBetweenMove) {
			return false
		}
	}
	return !g.WillMoveCauseCheck(m)
}

func (g *Game) WillMoveCauseCheck(m Move) bool {
	// Optimization: Stack-copy the board. accessing underlying arrays by value.
	// Since justMove/ComputeIsCheck don't modify maps/slices (only arrays/primitives), this is safe and allocation-free.
	clone := *g
	clone.justMove(m)
	return clone.ComputeIsCheck()
}

// LegalMovesList returns a copy of the currently legal moves.
func (g *Game) LegalMovesList() []Move {
	moves := make([]Move, len(g.LegalMoves))
	copy(moves, g.LegalMoves)
	return moves
}

// CopyLegalMoves copies the current legal moves into dst and returns the resulting slice.
//
// It lets search code reuse a caller-owned buffer instead of allocating a fresh
// slice at each node.
func (g *Game) CopyLegalMoves(dst []Move) []Move {
	if cap(dst) < len(g.LegalMoves) {
		dst = make([]Move, len(g.LegalMoves))
	} else {
		dst = dst[:len(g.LegalMoves)]
	}
	copy(dst, g.LegalMoves)
	return dst
}

// TryMove applies a move only if it matches one of the current legal moves.
func (g *Game) TryMove(m Move) error {
	g.GenerateLegalMoves()
	for _, legal := range g.LegalMoves {
		if moveMatchesRequest(legal, m) {
			g.MakeMove(legal)
			return nil
		}
	}
	return illegalMove(m, "", IllegalMoveReasonNotLegal)
}

// TryMoveFromCoords applies a legal move from algebraic coordinates like "e2", "e4".
func (g *Game) TryMoveFromCoords(from, to string, promotion ...Piece) error {
	fromSq, ok := COORDS_TO_SQUARE[from]
	if !ok {
		return ErrInvalidMoveNotation
	}
	toSq, ok := COORDS_TO_SQUARE[to]
	if !ok {
		return ErrInvalidMoveNotation
	}

	requested := NewMove(fromSq, toSq, EMPTY)
	if len(promotion) > 0 {
		requested = NewPromotionMove(fromSq, toSq, EMPTY, promotion[0])
	}
	return g.TryMove(requested)
}

func moveMatchesRequest(legal, requested Move) bool {
	if legal.From() != requested.From() || legal.To() != requested.To() {
		return false
	}
	if legal.IsPromotionMove() || requested.IsPromotionMove() {
		return legal.GetPromotionTo() == requested.GetPromotionTo()
	}
	return true
}

func (g *Game) MakeMove(m Move) {
	g.recordPGNMove(m)
	g.makeMove(m, true)
}

// MakeMoveFast applies a legal move and regenerates legal moves without updating repetition or game-over status.
//
// It is intended for search/perft-style traversal where callers only need the
// next legal move list and will undo the move before observing public status
// flags. Use MakeMove for normal game play.
func (g *Game) MakeMoveFast(m Move) {
	g.makeMove(m, false)
}

func (g *Game) makeMove(m Move, updatePositionHistory bool) {
	g.makeMoveInternal(m, updatePositionHistory, true)
}

func (g *Game) makeMoveNoGenerate(m Move) {
	g.makeMoveInternal(m, false, false)
}

func (g *Game) makeMoveInternal(m Move, updatePositionHistory, generateLegalMoves bool) {
	// Capture state for UndoMove
	capturedPiece := g.Squares[m.To()]
	if m.IsEnPassant() {
		if g.Turn == WHITE {
			capturedPiece = g.Squares[m.To()+8]
		} else {
			capturedPiece = g.Squares[m.To()-8]
		}
	}
	g.History = append(g.History, GameState{
		CapturedPiece: capturedPiece,
		Castling:      g.Castling,
		EnPassant:     g.EnPassant,
		HalfMoves:     g.HalfMoves,
		FullMoves:     g.FullMoves,
		ZobristHash:   g.ZobristHash,
	})

	if g.ShouldResetHalfMoves(m) {
		g.HalfMoves = 0
	} else {
		g.HalfMoves++
	}

	if g.ShouldIncFullMoves(m) {
		g.FullMoves++
	}

	g.justMove(m)
	kind := g.Squares[m.To()].Kind()
	if kind == KING {
		if g.Turn == WHITE {
			g.Castling &= ^(CASTLE_WKS | CASTLE_WQS)
		} else {
			g.Castling &= ^(CASTLE_BKS | CASTLE_BQS)
		}
	}
	if kind == ROOK {
		switch m.From() {
		case WKS_ROOK_ORIGINAL_SQUARE:
			g.Castling &= ^CASTLE_WKS
		case WQS_ROOK_ORIGINAL_SQUARE:
			g.Castling &= ^CASTLE_WQS
		case BKS_ROOK_ORIGINAL_SQUARE:
			g.Castling &= ^CASTLE_BKS
		case BQS_ROOK_ORIGINAL_SQUARE:
			g.Castling &= ^CASTLE_BQS
		}
	}

	switch m.To() {
	case WKS_ROOK_ORIGINAL_SQUARE:
		g.Castling &= ^CASTLE_WKS
	case WQS_ROOK_ORIGINAL_SQUARE:
		g.Castling &= ^CASTLE_WQS
	case BKS_ROOK_ORIGINAL_SQUARE:
		g.Castling &= ^CASTLE_BKS
	case BQS_ROOK_ORIGINAL_SQUARE:
		g.Castling &= ^CASTLE_BQS
	}
	// enPassant target
	g.EnPassant = 0
	if kind == PAWN && g.Turn == WHITE {
		if m.From().Rank() == 6 && m.To().Rank() == 4 {
			g.EnPassant = m.From() - 8
		}
	}
	if kind == PAWN && g.Turn == BLACK {
		if m.From().Rank() == 1 && m.To().Rank() == 3 {
			g.EnPassant = m.From() + 8
		}
	}

	if g.Turn == WHITE {
		g.Turn = BLACK
	} else {
		g.Turn = WHITE
	}

	if updatePositionHistory {
		g.recordPosition()
	}

	if generateLegalMoves {
		g.GenerateLegalMoves()
		if updatePositionHistory {
			g.refreshStatus()
		}
	}
}

func (g *Game) justMove(m Move) {
	from := m.From()
	to := m.To()

	//capturedPiece := m.captured()
	capturedPiece := g.Squares[to]
	if m.IsEnPassant() && g.Turn == WHITE {
		capturedPiece = g.Squares[to+8]
	} else if m.IsEnPassant() && g.Turn == BLACK {
		capturedPiece = g.Squares[to-8]
	}
	fromBBNeg := ^Bitboard(0x1 << from)
	toBB := Bitboard(0x1 << to)
	movingPiece := g.Squares[from]
	movingPieceKind := movingPiece.Kind()
	switch movingPiece.Color() {
	case WHITE:
		// update bitmap of moving piece kind, unset bit of source square
		g.Whites[movingPieceKind] &= fromBBNeg
		// update bitmap of moving piece kind, set bit of source square
		g.Whites[movingPieceKind] |= toBB
		// update white pieces bitboard - unset old square
		g.WhitePieces &= fromBBNeg
		// update white pieces bitboard - set new square
		g.WhitePieces |= toBB
	case BLACK:
		g.Blacks[movingPieceKind] &= fromBBNeg
		g.Blacks[movingPieceKind] |= toBB
		g.BlackPieces &= fromBBNeg
		g.BlackPieces |= toBB
	}

	g.Occupied &= fromBBNeg
	g.Occupied |= toBB

	g.Squares[m.To()] = g.Squares[m.From()]
	g.Squares[m.From()] = EMPTY
	if capturedPiece != EMPTY {
		if !m.IsEnPassant() {
			g.capturePiece(to, capturedPiece)
		} else {
			if g.Turn == WHITE {
				capSq := to + 8
				g.capturePiece(capSq, g.Squares[capSq])
				g.Occupied &= ^Bitboard(0x1 << capSq)
				g.Squares[to+8] = EMPTY
			} else {
				capSq := to - 8
				g.capturePiece(capSq, g.Squares[capSq])
				g.Occupied &= ^Bitboard(0x1 << capSq)
				g.Squares[to-8] = EMPTY
			}
		}
	}
	if m.IsCastlingMove() {
		var rookMove Move
		if m.To() == WKS_KING_TO_SQUARE || m.To() == BKS_KING_TO_SQUARE {
			rookMove = NewMove(m.To()+1, m.To()-1, 0)
		} else if m.To() == WQS_KING_TO_SQUARE || m.To() == BQS_KING_TO_SQUARE {
			rookMove = NewMove(m.To()-2, m.To()+1, 0)
		}
		g.justMove(rookMove)
	}
	var promoteTo Piece = m.GetPromotionTo()
	if promoteTo > 0 {
		switch g.Squares[to].Color() {
		case WHITE:
			// remove advanced pawn from boards
			g.Whites[PAWN] &= ^toBB
			// add promotePiece to board
			g.Whites[promoteTo] |= toBB
			g.WhitePieces |= toBB
		case BLACK:
			// remove advanced pawn from boards
			g.Blacks[PAWN] &= ^toBB
			// add promotePiece to board
			g.Blacks[promoteTo] |= toBB
			g.BlackPieces |= toBB
		}
		g.Squares[m.To()] = Piece(uint(promoteTo) | uint(g.Turn))
	}
}

// Remove captured piece from opponent's pieces
func (g *Game) capturePiece(sq Square, captured Piece) {
	if captured == EMPTY {
		return
	}
	sqBB := Bitboard(0x1 << sq)
	kind := captured.Kind()
	switch captured.Color() {
	case WHITE:
		g.Whites[kind] &= ^sqBB
		g.WhitePieces &= ^sqBB
	case BLACK:
		g.Blacks[kind] &= ^sqBB
		g.BlackPieces &= ^sqBB
	}
}

func (g *Game) unmakeMove(m Move, captured Piece) {
	from := m.From()
	to := m.To()

	movingPieceKind := g.Squares[to].Kind()
	movingColor := g.Turn

	if m.IsPromotionMove() {
		// The piece at `to` is the promoted piece.
		promotedKind := m.GetPromotionTo()

		toBB := Bitboard(0x1 << to)
		fromBB := Bitboard(0x1 << from)

		if movingColor == WHITE {
			g.Whites[promotedKind] &= ^toBB
			g.WhitePieces &= ^toBB
			// Restore Pawn at `from`
			g.Whites[PAWN] |= fromBB
			g.WhitePieces |= fromBB
		} else {
			g.Blacks[promotedKind] &= ^toBB
			g.BlackPieces &= ^toBB
			g.Blacks[PAWN] |= fromBB
			g.BlackPieces |= fromBB
		}
		g.Squares[to] = EMPTY
		g.Squares[from] = Piece(uint(PAWN) | uint(movingColor))
		g.Occupied &= ^toBB
		g.Occupied |= fromBB

	} else {
		toBB := Bitboard(0x1 << to)
		fromBB := Bitboard(0x1 << from)

		if movingColor == WHITE {
			g.Whites[movingPieceKind] &= ^toBB
			g.Whites[movingPieceKind] |= fromBB
			g.WhitePieces &= ^toBB
			g.WhitePieces |= fromBB
		} else {
			g.Blacks[movingPieceKind] &= ^toBB
			g.Blacks[movingPieceKind] |= fromBB
			g.BlackPieces &= ^toBB
			g.BlackPieces |= fromBB
		}

		g.Occupied &= ^toBB
		g.Occupied |= fromBB

		g.Squares[from] = g.Squares[to]
		g.Squares[to] = EMPTY
	}

	if captured != EMPTY {
		if m.IsEnPassant() {
			var capSq Square
			if movingColor == WHITE {
				capSq = to + 8
			} else {
				capSq = to - 8
			}
			g.addPiece(captured, int(capSq))
		} else {
			g.addPiece(captured, int(to))
		}
	}

	if m.IsCastlingMove() {
		var rookFrom, rookTo Square

		if m.To() == WKS_KING_TO_SQUARE { // g1
			rookFrom = 61 // f1
			rookTo = 63   // h1
		} else if m.To() == WQS_KING_TO_SQUARE { // c1
			rookFrom = 59 // d1
			rookTo = 56   // a1
		} else if m.To() == BKS_KING_TO_SQUARE { // g8
			rookFrom = 5 // f8
			rookTo = 7   // h8
		} else if m.To() == BQS_KING_TO_SQUARE { // c8
			rookFrom = 3 // d8
			rookTo = 0   // a8
		}

		// Move rook back
		rFromBB := Bitboard(0x1 << rookFrom)
		rToBB := Bitboard(0x1 << rookTo)

		if movingColor == WHITE {
			g.Whites[ROOK] &= ^rFromBB
			g.Whites[ROOK] |= rToBB
			g.WhitePieces &= ^rFromBB
			g.WhitePieces |= rToBB
		} else {
			g.Blacks[ROOK] &= ^rFromBB
			g.Blacks[ROOK] |= rToBB
			g.BlackPieces &= ^rFromBB
			g.BlackPieces |= rToBB
		}
		g.Occupied &= ^rFromBB
		g.Occupied |= rToBB

		g.Squares[rookTo] = g.Squares[rookFrom]
		g.Squares[rookFrom] = EMPTY
	}
}
