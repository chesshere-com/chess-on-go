package chessongo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUCIHelpers(t *testing.T) {
	e2, err := ParseUCISquare("e2")
	require.NoError(t, err)
	require.Equal(t, "e2", e2.UCI())

	move, err := NewMoveFromUCI("e2e4")
	require.NoError(t, err)
	require.Equal(t, "e2e4", move.UCI())

	promotion, err := NewMoveFromUCI("a7a8q")
	require.NoError(t, err)
	require.True(t, promotion.IsPromotionMove())
	require.Equal(t, Piece(QUEEN), promotion.GetPromotionTo())
	require.Equal(t, "a7a8q", promotion.UCI())

	_, err = ParseUCISquare("i9")
	require.ErrorIs(t, err, ErrInvalidMoveNotation)
	_, err = NewMoveFromUCI("e2e")
	require.ErrorIs(t, err, ErrInvalidMoveNotation)
}
