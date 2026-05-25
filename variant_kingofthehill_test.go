package chessongo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKingOfTheHillEndsWhenKingReachesCenter(t *testing.T) {
	tests := []struct {
		name   string
		fen    string
		winner Color
	}{
		{"white king on d4", "4k3/8/8/8/3K4/8/8/R7 b - - 0 1", WHITE},
		{"white king on e4", "4k3/8/8/8/4K3/8/8/R7 b - - 0 1", WHITE},
		{"white king on d5", "4k3/8/8/3K4/8/8/8/R7 b - - 0 1", WHITE},
		{"white king on e5", "4k3/8/8/4K3/8/8/8/R7 b - - 0 1", WHITE},
		{"black king on d4", "R7/8/8/8/3k4/8/8/4K3 w - - 0 1", BLACK},
		{"black king on e5", "R7/8/8/4k3/8/8/8/4K3 w - - 0 1", BLACK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := NewGameFromFENWithVariant(tt.fen, VariantKingOfTheHill)
			require.NoError(t, err)

			require.True(t, g.IsTerminal())
			require.Equal(t, GameStatusVariantWin, g.Status())
			require.Equal(t, tt.winner, g.Winner())
		})
	}
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

func TestExportPGNIncludesKingOfTheHillVariantTag(t *testing.T) {
	g, err := NewGameFromFENWithVariant(STARTING_POSITION_FEN, VariantKingOfTheHill)
	require.NoError(t, err)

	require.NoError(t, g.TryMoveUCI("e2e4"))

	tags := g.PGNTags()
	require.Equal(t, "King of the Hill", tags["Variant"])
	require.Equal(t, "1", tags["SetUp"])
	require.Equal(t, STARTING_POSITION_FEN, tags["FEN"])
}

func TestKingOfTheHillMoveIntoCenterWins(t *testing.T) {
	g, err := NewGameFromFENWithVariant("4k3/8/8/8/8/3K4/8/R7 w - - 0 1", VariantKingOfTheHill)
	require.NoError(t, err)
	require.False(t, g.IsTerminal())

	require.NoError(t, g.TryMoveUCI("d3d4"))

	require.True(t, g.IsTerminal())
	require.Equal(t, GameStatusVariantWin, g.Status())
	require.Equal(t, Color(WHITE), g.Winner())
}
