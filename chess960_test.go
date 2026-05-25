package chessongo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChess960BackRankKnownIDs(t *testing.T) {
	tests := []struct {
		position int
		want     string
	}{
		{position: 0, want: "BBQNNRKR"},
		{position: 518, want: "RNBQKBNR"},
		{position: 959, want: "RKRNNQBB"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got, err := Chess960BackRank(tt.position)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestChess960BackRankRejectsInvalidIDs(t *testing.T) {
	for _, position := range []int{-1, 960} {
		t.Run("", func(t *testing.T) {
			_, err := Chess960BackRank(position)
			require.ErrorIs(t, err, errInvalidChess960Position)
		})
	}
}

func TestChess960StartingFENClassicalPosition(t *testing.T) {
	got, err := Chess960StartingFEN(518)
	require.NoError(t, err)
	require.Equal(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w HAha - 0 1", got)
}

func TestChess960ValidateBackRankRejectsInvalidRanks(t *testing.T) {
	tests := []string{
		"RNBQKBN",  // too short
		"RNBQKBNQ", // wrong pieces
		"RBQBKNNR", // bishops same color
		"KRRNQBBN", // king not between rooks
	}

	for _, rank := range tests {
		t.Run(rank, func(t *testing.T) {
			require.ErrorIs(t, ValidateChess960BackRank(rank), errInvalidChess960Position)
		})
	}
}

func TestChess960ValidateBackRankAcceptsValidRanks(t *testing.T) {
	for _, rank := range []string{"RBQNKNBR", "RKRNQBBN"} {
		t.Run(rank, func(t *testing.T) {
			require.NoError(t, ValidateChess960BackRank(rank))
		})
	}
}
