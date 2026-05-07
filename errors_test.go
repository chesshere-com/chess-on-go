package chessongo

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadFENReturnsStructuredError(t *testing.T) {
	g := &Game{}
	err := g.LoadFEN("not a fen")

	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidFEN)

	var fenErr *FENError
	require.True(t, errors.As(err, &fenErr))
	require.NotEmpty(t, fenErr.Reason)
	require.Equal(t, FENFieldFormat, fenErr.Field)
	require.Contains(t, err.Error(), "format")
}

func TestTryMoveUCIReturnsTypedErrors(t *testing.T) {
	g := NewGame()

	err := g.TryMoveUCI("bad")
	require.ErrorIs(t, err, ErrInvalidMoveNotation)

	err = g.TryMoveUCI("e2e5")
	require.ErrorIs(t, err, ErrIllegalMove)
	var moveErr *IllegalMoveError
	require.True(t, errors.As(err, &moveErr))
	require.Equal(t, "e2e5", moveErr.Notation)
	require.Equal(t, IllegalMoveReasonNotLegal, moveErr.Reason)
	require.Equal(t, Move(NewMove(COORDS_TO_SQUARE["e2"], COORDS_TO_SQUARE["e5"], EMPTY)), moveErr.Move)
}

func TestTryMoveSANReturnsTypedErrors(t *testing.T) {
	g := NewGame()

	err := g.TryMoveSAN("")
	require.ErrorIs(t, err, ErrInvalidMoveNotation)

	err = g.TryMoveSAN("Qh5")
	require.ErrorIs(t, err, ErrIllegalMove)
	var moveErr *IllegalMoveError
	require.True(t, errors.As(err, &moveErr))
	require.Equal(t, "Qh5", moveErr.Notation)
	require.Equal(t, IllegalMoveReasonNoMatchingLegalMove, moveErr.Reason)
}

func TestFENErrorsExposeFieldContext(t *testing.T) {
	tests := []struct {
		name  string
		fen   string
		field FENField
	}{
		{"piece placement", "8/8/8/8/8/8/8/8 w - - 0 1", FENFieldPiecePlacement},
		{"side to move", "4k3/8/8/8/8/8/8/4K3 x - - 0 1", FENFieldSideToMove},
		{"castling", "4k3/8/8/8/8/8/8/4K3 w K - 0 1", FENFieldCastling},
		{"en passant", "4k3/8/8/8/8/8/8/4K3 w - e5 0 1", FENFieldEnPassant},
		{"halfmove", "4k3/8/8/8/8/8/8/4K3 w - - x 1", FENFieldHalfMoveClock},
		{"fullmove", "4k3/8/8/8/8/8/8/4K3 w - - 0 0", FENFieldFullMoveNumber},
		{"legality", "4k3/8/8/8/8/8/4R3/4K3 w - - 0 1", FENFieldLegality},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Game{}
			err := g.LoadFEN(tt.fen)

			require.ErrorIs(t, err, ErrInvalidFEN)
			var fenErr *FENError
			require.True(t, errors.As(err, &fenErr))
			require.Equal(t, tt.field, fenErr.Field)
			require.NotEmpty(t, fenErr.Reason)
		})
	}
}
