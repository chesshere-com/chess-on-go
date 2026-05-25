package chessongo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultGameUsesStandardVariant(t *testing.T) {
	g := NewGame()

	require.Equal(t, VariantStandard, g.Variant())
	require.Equal(t, STARTING_POSITION_FEN, g.FEN())
	require.Len(t, g.LegalMovesList(), 20)
}

func TestNewGameFromFENWithVariantRecordsVariant(t *testing.T) {
	g, err := NewGameFromFENWithVariant(STARTING_POSITION_FEN, VariantStandard)

	require.NoError(t, err)
	require.Equal(t, VariantStandard, g.Variant())
	require.Equal(t, STARTING_POSITION_FEN, g.FEN())
}

func TestNewGameFromFENWithChess960VariantRecordsVariant(t *testing.T) {
	g, err := NewGameFromFENWithVariant(STARTING_POSITION_FEN, VariantChess960)

	require.NoError(t, err)
	require.Equal(t, VariantChess960, g.Variant())
	require.Equal(t, STARTING_POSITION_FEN, g.FEN())
}

func TestClonePreservesVariant(t *testing.T) {
	g, err := NewGameFromFENWithVariant(STARTING_POSITION_FEN, VariantChess960)
	require.NoError(t, err)

	require.Equal(t, VariantChess960, g.Clone().Variant())
}

func TestLoadFENResetsVariantToStandard(t *testing.T) {
	g, err := NewGameFromFENWithVariant(STARTING_POSITION_FEN, VariantChess960)
	require.NoError(t, err)

	require.NoError(t, g.LoadFEN(STARTING_POSITION_FEN))
	require.Equal(t, VariantStandard, g.Variant())
}

func TestLoadFENWithInvalidVariantDoesNotMutateReceiver(t *testing.T) {
	g, err := NewGameFromFENWithVariant(STARTING_POSITION_FEN, VariantChess960)
	require.NoError(t, err)
	beforeFEN := g.FEN()

	err = g.LoadFENWithVariant("4k3/8/8/8/8/8/8/4K3 w - - 0 1", Variant(99))

	require.Error(t, err)
	require.Equal(t, VariantChess960, g.Variant())
	require.Equal(t, beforeFEN, g.FEN())
}

func TestVariantString(t *testing.T) {
	require.Equal(t, "standard", VariantStandard.String())
	require.Equal(t, "chess960", VariantChess960.String())
	require.Equal(t, "unknown", Variant(99).String())
}
