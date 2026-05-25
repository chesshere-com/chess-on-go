package chessongo

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadPGNChess960VariantTag(t *testing.T) {
	pgn := `[Event "Chess960 test"]
[Variant "Chess960"]
[SetUp "1"]
[FEN "6k1/8/8/8/8/8/8/6KR w H - 0 1"]
[Result "*"]

1. O-O *`

	g := &Game{}
	require.NoError(t, g.LoadPGN(pgn))
	require.Equal(t, VariantChess960, g.Variant())
	require.Equal(t, Piece(W_ROOK), g.squares[COORDS_TO_SQUARE["f1"]])
}

func TestExportPGNIncludesChess960VariantTags(t *testing.T) {
	g, err := NewGameFromFENWithVariant("6k1/8/8/8/8/8/8/6KR w H - 0 1", VariantChess960)
	require.NoError(t, err)
	require.NoError(t, g.TryMoveSAN("O-O"))

	pgn := g.PGN()
	require.True(t, strings.Contains(pgn, `[Variant "Chess960"]`))
	require.True(t, strings.Contains(pgn, `[SetUp "1"]`))
	require.True(t, strings.Contains(pgn, `[FEN "6k1/8/8/8/8/8/8/6KR w H - 0 1"]`))
}
