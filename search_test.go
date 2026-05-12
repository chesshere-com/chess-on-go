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
	board.game.legalMoves = []Move{NewMove(0, 1, EMPTY)}

	board.MakeMove(moves[0])
	require.Equal(t, []Move{NewMove(0, 1, EMPTY)}, board.game.legalMoves)

	board.UndoMove(moves[0])
	require.Equal(t, []Move{NewMove(0, 1, EMPTY)}, board.game.legalMoves)
}

func TestSearchBoardUsesCompactSearchState(t *testing.T) {
	board, err := NewSearchBoard(STARTING_POSITION_FEN)
	require.NoError(t, err)

	require.Nil(t, board.game.pseudoMoves)
	require.Nil(t, board.game.legalMoves)
	require.Nil(t, board.game.positionHistory)
}
