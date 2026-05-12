package chessongo

// Checks whether our king is in check or not
func (g *Game) ComputeIsCheck() bool {
	var kingBB, theirsAll, attackers Bitboard
	var theirs []Bitboard
	if g.turn == WHITE {
		kingBB, theirs, theirsAll = g.whites[KING], g.blacks[:], g.blackPieces
	} else {
		kingBB, theirs, theirsAll = g.blacks[KING], g.whites[:], g.whitePieces
	}
	if kingBB == 0 {
		return false
	}
	kingIdx := kingBB.lsbIndex()
	possibleAttackers := theirsAll & ATTACKS_TO[kingIdx]

	attackers = (theirs[ROOK] | theirs[QUEEN]) & possibleAttackers
	if attackers > 0 && rookAttacks(Square(kingIdx), g.occupied)&attackers > 0 {
		return true
	}

	attackers = (theirs[BISHOP] | theirs[QUEEN]) & possibleAttackers
	if attackers > 0 && bishopAttacks(Square(kingIdx), g.occupied)&attackers > 0 {
		return true
	}

	attackers = theirs[KNIGHT] & possibleAttackers
	for attackers > 0 {
		from := attackers.popLSB()
		if KNIGHT_ATTACKS_FROM[from]&kingBB > 0 {
			return true
		}
	}

	if g.turn == WHITE {
		// Black pawns attack “down” the board (towards higher square indices).
		if ((g.blacks[PAWN]&^Bitboard(FILE_A_MASK))<<7)&kingBB > 0 ||
			((g.blacks[PAWN]&^Bitboard(FILE_H_MASK))<<9)&kingBB > 0 {
			return true
		}
	} else {
		// White pawns attack “up” the board (towards lower square indices).
		if ((g.whites[PAWN]&^Bitboard(FILE_H_MASK))>>7)&kingBB > 0 ||
			((g.whites[PAWN]&^Bitboard(FILE_A_MASK))>>9)&kingBB > 0 {
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
		if g.isCheck || g.WillMoveCauseCheck(inBetweenMove) {
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
	moves := make([]Move, len(g.legalMoves))
	copy(moves, g.legalMoves)
	return moves
}

// CopyLegalMoves copies the current legal moves into dst and returns the resulting slice.
//
// It lets search code reuse a caller-owned buffer instead of allocating a fresh
// slice at each node.
func (g *Game) CopyLegalMoves(dst []Move) []Move {
	if cap(dst) < len(g.legalMoves) {
		dst = make([]Move, len(g.legalMoves))
	} else {
		dst = dst[:len(g.legalMoves)]
	}
	copy(dst, g.legalMoves)
	return dst
}

// TryMove applies a move only if it matches one of the current legal moves.
func (g *Game) TryMove(m Move) error {
	g.GenerateLegalMoves()
	for _, legal := range g.legalMoves {
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
	capturedPiece := g.squares[m.To()]
	if m.IsEnPassant() {
		if g.turn == WHITE {
			capturedPiece = g.squares[m.To()+8]
		} else {
			capturedPiece = g.squares[m.To()-8]
		}
	}
	g.history = append(g.history, GameState{
		CapturedPiece: capturedPiece,
		Castling:      g.castling,
		EnPassant:     g.enPassant,
		HalfMoves:     g.halfMoves,
		FullMoves:     g.fullMoves,
		ZobristHash:   g.zobristHash,
	})

	if g.ShouldResetHalfMoves(m) {
		g.halfMoves = 0
	} else {
		g.halfMoves++
	}

	if g.ShouldIncFullMoves(m) {
		g.fullMoves++
	}

	g.justMove(m)
	kind := g.squares[m.To()].Kind()
	if kind == KING {
		if g.turn == WHITE {
			g.castling &= ^(CASTLE_WKS | CASTLE_WQS)
		} else {
			g.castling &= ^(CASTLE_BKS | CASTLE_BQS)
		}
	}
	if kind == ROOK {
		switch m.From() {
		case WKS_ROOK_ORIGINAL_SQUARE:
			g.castling &= ^CASTLE_WKS
		case WQS_ROOK_ORIGINAL_SQUARE:
			g.castling &= ^CASTLE_WQS
		case BKS_ROOK_ORIGINAL_SQUARE:
			g.castling &= ^CASTLE_BKS
		case BQS_ROOK_ORIGINAL_SQUARE:
			g.castling &= ^CASTLE_BQS
		}
	}

	switch m.To() {
	case WKS_ROOK_ORIGINAL_SQUARE:
		g.castling &= ^CASTLE_WKS
	case WQS_ROOK_ORIGINAL_SQUARE:
		g.castling &= ^CASTLE_WQS
	case BKS_ROOK_ORIGINAL_SQUARE:
		g.castling &= ^CASTLE_BKS
	case BQS_ROOK_ORIGINAL_SQUARE:
		g.castling &= ^CASTLE_BQS
	}
	// enPassant target
	g.enPassant = 0
	if kind == PAWN && g.turn == WHITE {
		if m.From().Rank() == 6 && m.To().Rank() == 4 {
			g.enPassant = m.From() - 8
		}
	}
	if kind == PAWN && g.turn == BLACK {
		if m.From().Rank() == 1 && m.To().Rank() == 3 {
			g.enPassant = m.From() + 8
		}
	}

	if g.turn == WHITE {
		g.turn = BLACK
	} else {
		g.turn = WHITE
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
	capturedPiece := g.squares[to]
	if m.IsEnPassant() && g.turn == WHITE {
		capturedPiece = g.squares[to+8]
	} else if m.IsEnPassant() && g.turn == BLACK {
		capturedPiece = g.squares[to-8]
	}
	fromBBNeg := ^Bitboard(0x1 << from)
	toBB := Bitboard(0x1 << to)
	movingPiece := g.squares[from]
	movingPieceKind := movingPiece.Kind()
	switch movingPiece.Color() {
	case WHITE:
		// update bitmap of moving piece kind, unset bit of source square
		g.whites[movingPieceKind] &= fromBBNeg
		// update bitmap of moving piece kind, set bit of source square
		g.whites[movingPieceKind] |= toBB
		// update white pieces bitboard - unset old square
		g.whitePieces &= fromBBNeg
		// update white pieces bitboard - set new square
		g.whitePieces |= toBB
	case BLACK:
		g.blacks[movingPieceKind] &= fromBBNeg
		g.blacks[movingPieceKind] |= toBB
		g.blackPieces &= fromBBNeg
		g.blackPieces |= toBB
	}

	g.occupied &= fromBBNeg
	g.occupied |= toBB

	g.squares[m.To()] = g.squares[m.From()]
	g.squares[m.From()] = EMPTY
	if capturedPiece != EMPTY {
		if !m.IsEnPassant() {
			g.capturePiece(to, capturedPiece)
		} else {
			if g.turn == WHITE {
				capSq := to + 8
				g.capturePiece(capSq, g.squares[capSq])
				g.occupied &= ^Bitboard(0x1 << capSq)
				g.squares[to+8] = EMPTY
			} else {
				capSq := to - 8
				g.capturePiece(capSq, g.squares[capSq])
				g.occupied &= ^Bitboard(0x1 << capSq)
				g.squares[to-8] = EMPTY
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
		switch g.squares[to].Color() {
		case WHITE:
			// remove advanced pawn from boards
			g.whites[PAWN] &= ^toBB
			// add promotePiece to board
			g.whites[promoteTo] |= toBB
			g.whitePieces |= toBB
		case BLACK:
			// remove advanced pawn from boards
			g.blacks[PAWN] &= ^toBB
			// add promotePiece to board
			g.blacks[promoteTo] |= toBB
			g.blackPieces |= toBB
		}
		g.squares[m.To()] = Piece(uint(promoteTo) | uint(g.turn))
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
		g.whites[kind] &= ^sqBB
		g.whitePieces &= ^sqBB
	case BLACK:
		g.blacks[kind] &= ^sqBB
		g.blackPieces &= ^sqBB
	}
}

func (g *Game) unmakeMove(m Move, captured Piece) {
	from := m.From()
	to := m.To()

	movingPieceKind := g.squares[to].Kind()
	movingColor := g.turn

	if m.IsPromotionMove() {
		// The piece at `to` is the promoted piece.
		promotedKind := m.GetPromotionTo()

		toBB := Bitboard(0x1 << to)
		fromBB := Bitboard(0x1 << from)

		if movingColor == WHITE {
			g.whites[promotedKind] &= ^toBB
			g.whitePieces &= ^toBB
			// Restore Pawn at `from`
			g.whites[PAWN] |= fromBB
			g.whitePieces |= fromBB
		} else {
			g.blacks[promotedKind] &= ^toBB
			g.blackPieces &= ^toBB
			g.blacks[PAWN] |= fromBB
			g.blackPieces |= fromBB
		}
		g.squares[to] = EMPTY
		g.squares[from] = Piece(uint(PAWN) | uint(movingColor))
		g.occupied &= ^toBB
		g.occupied |= fromBB

	} else {
		toBB := Bitboard(0x1 << to)
		fromBB := Bitboard(0x1 << from)

		if movingColor == WHITE {
			g.whites[movingPieceKind] &= ^toBB
			g.whites[movingPieceKind] |= fromBB
			g.whitePieces &= ^toBB
			g.whitePieces |= fromBB
		} else {
			g.blacks[movingPieceKind] &= ^toBB
			g.blacks[movingPieceKind] |= fromBB
			g.blackPieces &= ^toBB
			g.blackPieces |= fromBB
		}

		g.occupied &= ^toBB
		g.occupied |= fromBB

		g.squares[from] = g.squares[to]
		g.squares[to] = EMPTY
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
			g.whites[ROOK] &= ^rFromBB
			g.whites[ROOK] |= rToBB
			g.whitePieces &= ^rFromBB
			g.whitePieces |= rToBB
		} else {
			g.blacks[ROOK] &= ^rFromBB
			g.blacks[ROOK] |= rToBB
			g.blackPieces &= ^rFromBB
			g.blackPieces |= rToBB
		}
		g.occupied &= ^rFromBB
		g.occupied |= rToBB

		g.squares[rookTo] = g.squares[rookFrom]
		g.squares[rookFrom] = EMPTY
	}
}
