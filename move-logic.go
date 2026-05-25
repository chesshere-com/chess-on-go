package chessongo

// Checks whether our king is in check or not
func (g *Game) ComputeIsCheck() bool {
	var kingBB Bitboard
	if g.turn == WHITE {
		kingBB = g.whites[KING]
	} else {
		kingBB = g.blacks[KING]
	}
	if kingBB == 0 {
		return false
	}
	kingIdx := kingBB.lsbIndex()
	return g.isSquareAttackedByWithOccupied(Square(kingIdx), oppositeColor(g.turn), g.occupied)
}

// Checks whether the given move is possible or not
func (g *Game) CanMove(m Move) bool {
	if m.IsCastlingMove() {
		if g.isCheck {
			return false
		}
		spec, ok := g.castlingSpecForMove(m)
		if !ok {
			return false
		}
		return g.castlingKingPathIsSafe(spec, oppositeColor(g.turn))
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
		CapturedPiece:    capturedPiece,
		Castling:         g.castling,
		CastlingRookFrom: g.castlingRookFrom,
		EnPassant:        g.enPassant,
		HalfMoves:        g.halfMoves,
		FullMoves:        g.fullMoves,
		ZobristHash:      g.zobristHash,
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
		g.clearCastlingRightsForRookSquare(m.From())
	}

	g.clearCastlingRightsForRookSquare(m.To())

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
	if m.IsCastlingMove() {
		g.justCastle(m)
		return
	}

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

func (g *Game) justCastle(m Move) {
	spec, ok := g.castlingSpecForMove(m)
	if !ok {
		return
	}
	g.clearSquare(spec.kingFrom)
	g.clearSquare(spec.rookFrom)
	g.addPiece(spec.kingPiece, int(spec.kingTo))
	g.addPiece(spec.rookPiece, int(spec.rookTo))
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

func (g *Game) clearSquare(square Square) {
	if !square.Valid() {
		return
	}
	piece := g.squares[square]
	if piece == EMPTY {
		return
	}
	g.capturePiece(square, piece)
	g.occupied &^= Bitboard(1) << square
	g.squares[square] = EMPTY
}

func (g *Game) unmakeMove(m Move, captured Piece) {
	from := m.From()
	to := m.To()

	if m.IsCastlingMove() {
		g.unmakeCastle(m)
		return
	}

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
}

func (g *Game) unmakeCastle(m Move) {
	spec, ok := g.castlingSpecForUndoMove(m, g.turn)
	if !ok {
		return
	}
	g.clearSquare(spec.kingTo)
	g.clearSquare(spec.rookTo)
	g.addPiece(spec.kingPiece, int(spec.kingFrom))
	g.addPiece(spec.rookPiece, int(spec.rookFrom))
}
