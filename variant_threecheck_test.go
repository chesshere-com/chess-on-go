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

func TestThreeCheckBlackCountsChecksAndEndsOnThird(t *testing.T) {
	g, err := NewGameFromFENWithVariant("4k3/q7/8/8/8/8/8/4K3 b - - 0 1 +0+2", VariantThreeCheck)
	require.NoError(t, err)

	require.NoError(t, g.TryMoveUCI("a7e3"))

	require.True(t, g.IsTerminal())
	require.Equal(t, GameStatusVariantWin, g.Status())
	require.Equal(t, Color(BLACK), g.Winner())
	require.Equal(t, "4k3/8/8/8/8/4q3/8/4K3 w - - 1 2 +0+3", g.FEN())
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

func TestThreeCheckRejectsMalformedCounterField(t *testing.T) {
	tests := []string{
		"4k3/8/8/8/8/8/Q7/4K3 w - - 0 1",
		"4k3/8/8/8/8/8/Q7/4K3 w - - 0 1 +4+0",
		"4k3/8/8/8/8/8/Q7/4K3 w - - 0 1 +0+4",
		"4k3/8/8/8/8/8/Q7/4K3 w - - 0 1 2+0",
	}

	for _, fen := range tests {
		_, err := NewGameFromFENWithVariant(fen, VariantThreeCheck)
		require.Error(t, err, fen)
	}
}

func TestStandardFENRejectsThreeCheckCounterField(t *testing.T) {
	_, err := NewGameFromFEN(STARTING_POSITION_FEN + " +0+0")
	require.Error(t, err)
}

func TestThreeCheckPositionKeyIncludesCounters(t *testing.T) {
	base := "4k3/8/8/8/8/8/Q7/4K3 w - - 0 1"
	twoChecks, err := NewGameFromFENWithVariant(base+" +2+0", VariantThreeCheck)
	require.NoError(t, err)
	oneCheck, err := NewGameFromFENWithVariant(base+" +1+0", VariantThreeCheck)
	require.NoError(t, err)

	require.NotEqual(t, twoChecks.PositionKey(), oneCheck.PositionKey())
}

func TestPGNVariantTagSelectsThreeCheck(t *testing.T) {
	pgn := `[Variant "Three-check"]

1. e4 e5`

	g, err := LoadPGNGame(pgn)

	require.NoError(t, err)
	require.Equal(t, VariantThreeCheck, g.Variant())
}

func TestExportPGNIncludesThreeCheckVariantAndFEN(t *testing.T) {
	g, err := NewGameFromFENWithVariant("4k3/8/8/8/8/8/Q7/4K3 w - - 0 1 +2+0", VariantThreeCheck)
	require.NoError(t, err)

	require.NoError(t, g.TryMoveUCI("a2a3"))

	tags := g.PGNTags()
	require.Equal(t, "Three-check", tags["Variant"])
	require.Equal(t, "1", tags["SetUp"])
	require.Equal(t, "4k3/8/8/8/8/8/Q7/4K3 w - - 0 1 +2+0", tags["FEN"])
}
