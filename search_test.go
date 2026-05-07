package chessongo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSearchBoardPerftMatchesKnownInitialPosition(t *testing.T) {
	board, err := NewSearchBoard(STARTING_POSITION_FEN)
	require.NoError(t, err)

	require.EqualValues(t, 8902, board.Perft(3))
}

func TestSearchBoardMakeUndoPreservesFEN(t *testing.T) {
	board, err := NewSearchBoard(STARTING_POSITION_FEN)
	require.NoError(t, err)
	start := board.FEN()
	moves := board.LegalMoves(nil)

	board.MakeMove(moves[0])
	board.UndoMove(moves[0])

	require.Equal(t, start, board.FEN())
}

func TestSearchBoardMakeUndoDoesNotRegeneratePublicMoveState(t *testing.T) {
	board, err := NewSearchBoard(STARTING_POSITION_FEN)
	require.NoError(t, err)
	moves := board.LegalMoves(nil)
	board.game.LegalMoves = []Move{NewMove(0, 1, EMPTY)}

	board.MakeMove(moves[0])
	require.Equal(t, []Move{NewMove(0, 1, EMPTY)}, board.game.LegalMoves)

	board.UndoMove(moves[0])
	require.Equal(t, []Move{NewMove(0, 1, EMPTY)}, board.game.LegalMoves)
}

func TestSearchBoardUsesCompactSearchState(t *testing.T) {
	board, err := NewSearchBoard(STARTING_POSITION_FEN)
	require.NoError(t, err)

	require.Nil(t, board.game.PseudoMoves)
	require.Nil(t, board.game.LegalMoves)
	require.Nil(t, board.game.PositionHistory)
	require.Empty(t, board.game.Fen)
}
