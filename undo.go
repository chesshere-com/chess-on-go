package chessongo

type GameState struct {
	CapturedPiece    Piece
	Castling         int
	CastlingRookFrom [16]Square
	EnPassant        Square
	Ply              int
	HalfMoves        int
	FullMoves        int
	ZobristHash      uint64
}

func (g *Game) UndoMove(m Move) {
	if len(g.pgnMoves) > 0 {
		g.pgnMoves = g.pgnMoves[:len(g.pgnMoves)-1]
	}
	g.undoMove(m, true)
}

// UndoMoveFast reverses a move made by MakeMoveFast.
func (g *Game) UndoMoveFast(m Move) {
	g.undoMove(m, false)
}

func (g *Game) undoMove(m Move, updatePositionHistory bool) {
	g.undoMoveInternal(m, updatePositionHistory, true)
}

func (g *Game) undoMoveNoGenerate(m Move) {
	g.undoMoveInternal(m, false, false)
}

func (g *Game) undoMoveInternal(m Move, updatePositionHistory, generateLegalMoves bool) {
	if len(g.history) == 0 {
		return
	}
	// Decrement history count for current position
	if updatePositionHistory && g.positionHistory != nil {
		g.positionHistory[g.zobristHash]--
		if g.positionHistory[g.zobristHash] <= 0 {
			delete(g.positionHistory, g.zobristHash)
		}
	}

	// Pop state
	state := g.history[len(g.history)-1]
	g.history = g.history[:len(g.history)-1]

	// Restore simple fields
	g.castling = state.Castling
	g.castlingRookFrom = state.CastlingRookFrom
	g.enPassant = state.EnPassant
	g.halfMoves = state.HalfMoves
	g.fullMoves = state.FullMoves
	g.zobristHash = state.ZobristHash

	// Flip Turn back
	if g.turn == WHITE {
		g.turn = BLACK
	} else {
		g.turn = WHITE
	}

	g.unmakeMove(m, state.CapturedPiece)

	// Re-calculate derived state
	if generateLegalMoves {
		g.GenerateLegalMoves()
		if updatePositionHistory {
			g.refreshStatus()
		}
	}
}
