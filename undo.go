package chessongo

type GameState struct {
	CapturedPiece Piece
	Castling      int
	EnPassant     Square
	Ply           int
	HalfMoves     int
	FullMoves     int
	ZobristHash   uint64
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
	if len(g.History) == 0 {
		return
	}
	// Decrement history count for current position
	if updatePositionHistory && g.PositionHistory != nil {
		g.PositionHistory[g.ZobristHash]--
		if g.PositionHistory[g.ZobristHash] <= 0 {
			delete(g.PositionHistory, g.ZobristHash)
		}
	}

	// Pop state
	state := g.History[len(g.History)-1]
	g.History = g.History[:len(g.History)-1]

	// Restore simple fields
	g.Castling = state.Castling
	g.EnPassant = state.EnPassant
	g.HalfMoves = state.HalfMoves
	g.FullMoves = state.FullMoves
	g.ZobristHash = state.ZobristHash

	// Flip Turn back
	if g.Turn == WHITE {
		g.Turn = BLACK
	} else {
		g.Turn = WHITE
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
