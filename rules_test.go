package chessongo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFiftyMoveRule(t *testing.T) {
	b := NewGame()
	// Simulate 100 half moves without pawn move or capture
	// We can manually set HalfMoves for testing since we trust internal increment logic which is tested elsewhere or trivial
	b.halfMoves = 100

	// We need to trigger the update logic which happens in MakeMove or Unmarshal or we can call the check method if public (it's private).
	// But we exposed the field. The field is updated in MakeMove.
	// Let's create a situation where we make a move and it updates.

	b.GenerateLegalMoves()
	// Just pick a move that isn't a pawn move or capture to avoid reset
	// Starting position: 1. Nf3 is a knight move, no capture.
	move := NewMove(Square(6), Square(21), EMPTY) // g1 -> f3

	// Reset HalfMoves to 99 so that after this move it becomes 100
	b.halfMoves = 99
	b.MakeMove(move)

	require.True(t, b.isFiftyMoveRule, "Should be 50 move rule enabled")
	require.False(t, b.isSeventyFiveMoveRule, "Should not be 75 move rule yet")
	require.False(t, b.isFinished, "Game should not be finished by 50 move rule alone")
}

func TestSeventyFiveMoveRule(t *testing.T) {
	b := NewGame()
	b.halfMoves = 149

	b.GenerateLegalMoves()
	move := NewMove(Square(6), Square(21), EMPTY) // g1 -> f3
	b.MakeMove(move)

	require.True(t, b.isFiftyMoveRule, "Should be 50 move rule enabled")
	require.True(t, b.isSeventyFiveMoveRule, "Should be 75 move rule enabled")
	require.True(t, b.isFinished, "Game should be finished by 75 move rule")
}

func TestInsufficientMaterialWithSameColorBishops(t *testing.T) {
	g := &Game{}
	require.NoError(t, g.LoadFEN("k7/8/8/8/8/4B3/8/2B4K w - - 0 1"))

	require.True(t, g.isMaterialDraw)
}

func TestSufficientMaterialWithOppositeColorBishops(t *testing.T) {
	g := &Game{}
	require.NoError(t, g.LoadFEN("k7/8/8/8/8/3B4/8/2B4K w - - 0 1"))

	require.False(t, g.isMaterialDraw)
}
