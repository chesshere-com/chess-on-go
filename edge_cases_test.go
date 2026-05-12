package chessongo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLegalMovesDoubleCheckOnlyKingMoves(t *testing.T) {
	g := &Game{}
	require.NoError(t, g.LoadFEN("k3r3/8/8/8/1b6/8/8/4K3 w - - 0 1"))

	for _, move := range g.legalMoves {
		require.Equal(t, Piece(KING), g.squares[move.From()].Kind(), "double check should allow only king moves, got %s", move.UCI())
	}
}

func TestPinnedPieceCanMoveOnlyAlongPin(t *testing.T) {
	g := &Game{}
	require.NoError(t, g.LoadFEN("k3r3/8/8/8/8/8/4R3/4K3 w - - 0 1"))

	legal := sortedMoveUCIs(g.legalMoves)
	require.Contains(t, legal, "e2e3")
	require.Contains(t, legal, "e2e8")
	require.NotContains(t, legal, "e2d2")
	require.NotContains(t, legal, "e2f2")
}

func TestPinnedEnPassantIsIllegal(t *testing.T) {
	g := &Game{}
	require.NoError(t, g.LoadFEN("k3r3/8/8/3pP3/8/8/8/4K3 w - d6 0 2"))

	require.NotContains(t, sortedMoveUCIs(g.legalMoves), "e5d6")
}

func TestKingCannotCaptureProtectedPiece(t *testing.T) {
	g := &Game{}
	require.NoError(t, g.LoadFEN("k4r2/8/8/8/8/8/5r2/4K3 w - - 0 1"))

	require.NotContains(t, sortedMoveUCIs(g.legalMoves), "e1f2")
}

func TestCastlingThroughAttackedSquareIsIllegal(t *testing.T) {
	g := &Game{}
	require.NoError(t, g.LoadFEN("r3k2r/8/8/8/8/5r2/8/R3K2R w KQkq - 0 1"))

	require.NotContains(t, sortedMoveUCIs(g.legalMoves), "e1g1")
	require.Contains(t, sortedMoveUCIs(g.legalMoves), "e1c1")
}

func TestPinnedPromotionCaptureThatLeavesCheckIsIllegal(t *testing.T) {
	g := &Game{}
	require.NoError(t, g.LoadFEN("k3r2q/4P3/8/8/8/8/8/4K3 w - - 0 1"))

	for _, move := range sortedMoveUCIs(g.legalMoves) {
		require.NotEqual(t, "e7f8q", move)
		require.NotEqual(t, "e7f8r", move)
		require.NotEqual(t, "e7f8b", move)
		require.NotEqual(t, "e7f8n", move)
	}
}
