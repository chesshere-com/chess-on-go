package chessongo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestThreeCheckCountsChecksAndEndsOnThird(t *testing.T) {
	g, err := NewGameFromFENWithVariant("4k3/8/8/8/8/8/Q7/4K3 w - - 0 1 +2+0", VariantThreeCheck)
	require.NoError(t, err)

	require.NoError(t, g.TryMoveUCI("a2e6"))

	require.True(t, g.IsTerminal())
	require.Equal(t, GameStatusVariantWin, g.Status())
	require.Equal(t, Color(WHITE), g.Winner())
	require.Equal(t, "4k3/8/4Q3/8/8/8/8/4K3 b - - 1 1 +3+0", g.FEN())
}

func TestThreeCheckUndoRestoresCheckCount(t *testing.T) {
	g, err := NewGameFromFENWithVariant("4k3/8/8/8/8/8/Q7/4K3 w - - 0 1 +2+0", VariantThreeCheck)
	require.NoError(t, err)

	var move Move
	for _, candidate := range g.LegalMovesList() {
		if candidate.UCI() == "a2e6" {
			move = candidate
			break
		}
	}
	require.NotZero(t, move)

	g.MakeMove(move)
	g.UndoMove(move)

	require.Equal(t, "4k3/8/8/8/8/8/Q7/4K3 w - - 0 1 +2+0", g.FEN())
	require.Equal(t, uint8(2), g.variantState.checksGiven[whiteStateIndex])
}

func TestThreeCheckRejectsMissingCounterField(t *testing.T) {
	_, err := NewGameFromFENWithVariant(STARTING_POSITION_FEN, VariantThreeCheck)
	require.Error(t, err)
}

func TestPGNVariantTagSelectsThreeCheck(t *testing.T) {
	pgn := `[Variant "Three-check"]

1. e4 e5`

	g, err := LoadPGNGame(pgn)

	require.NoError(t, err)
	require.Equal(t, VariantThreeCheck, g.Variant())
}
