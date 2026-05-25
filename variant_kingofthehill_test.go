package chessongo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKingOfTheHillEndsWhenKingReachesCenter(t *testing.T) {
	g, err := NewGameFromFENWithVariant("4k3/8/8/8/3K4/8/8/R7 b - - 0 1", VariantKingOfTheHill)
	require.NoError(t, err)

	require.True(t, g.IsTerminal())
	require.Equal(t, GameStatusVariantWin, g.Status())
	require.Equal(t, Color(WHITE), g.Winner())
}

func TestStandardDoesNotUseKingOfTheHillWinCondition(t *testing.T) {
	g, err := NewGameFromFEN("4k3/8/8/8/3K4/8/8/R7 b - - 0 1")
	require.NoError(t, err)

	require.NotEqual(t, GameStatusVariantWin, g.Status())
	require.NotEqual(t, Color(WHITE), g.Winner())
}

func TestPGNVariantTagSelectsKingOfTheHill(t *testing.T) {
	pgn := `[Variant "King of the Hill"]

1. e4 e5`

	g, err := LoadPGNGame(pgn)

	require.NoError(t, err)
	require.Equal(t, VariantKingOfTheHill, g.Variant())
}
