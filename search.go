package chessongo

// SearchBoard is a lightweight wrapper around Game for search and perft traversal.
//
// It uses fast make/unmake operations and caller-owned move buffers. Status
// flags such as repetition, checkmate, and draw state are not refreshed during
// traversal.
type SearchBoard struct {
	game Game
}

// NewSearchBoard creates a search board from FEN.
func NewSearchBoard(fen string) (*SearchBoard, error) {
	var game Game
	if err := game.LoadFEN(fen); err != nil {
		return nil, err
	}
	game.pseudoMoves = nil
	game.legalMoves = nil
	game.positionHistory = nil
	game.history = game.history[:0]
	return &SearchBoard{game: game}, nil
}

// FEN returns the current board position as FEN.
func (b *SearchBoard) FEN() string {
	return b.game.ToFEN()
}

// LegalMoves copies legal moves into dst and returns the resulting slice.
func (b *SearchBoard) LegalMoves(dst []Move) []Move {
	return b.game.generateLegalMovesInto(dst)
}

// MakeMove applies a legal move without refreshing game-over status.
func (b *SearchBoard) MakeMove(move Move) {
	b.game.makeMoveNoGenerate(move)
}

// UndoMove reverses a move made by MakeMove.
func (b *SearchBoard) UndoMove(move Move) {
	b.game.undoMoveNoGenerate(move)
}

// Perft returns the legal move tree node count at depth.
func (b *SearchBoard) Perft(depth int) uint64 {
	return b.perft(depth)
}

func (b *SearchBoard) perft(depth int) uint64 {
	if depth == 0 {
		return 1
	}

	var buffer [maxGeneratedMoves]Move
	count := b.game.generateLegalMovesArray(&buffer)
	if depth == 1 {
		return uint64(count)
	}

	var nodes uint64
	for i := 0; i < count; i++ {
		move := buffer[i]
		b.game.makeMoveNoGenerate(move)
		nodes += b.perft(depth - 1)
		b.game.undoMoveNoGenerate(move)
	}
	return nodes
}
