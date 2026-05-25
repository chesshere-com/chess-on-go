package chessongo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChess960CastlingGeneratedFromNonStandardKingSquare(t *testing.T) {
	g, err := NewGameFromFENWithVariant("6k1/8/8/8/8/8/8/5RK1 w F - 0 1", VariantChess960)
	require.NoError(t, err)

	moves := movesToStrings(g.LegalMovesList())

	assertContains(t, moves, "g1 c1")
}

func TestChess960CastlingGeneratedWhenKingAlreadyOnFinalSquare(t *testing.T) {
	g, err := NewGameFromFENWithVariant("6k1/8/8/8/8/8/8/6KR w H - 0 1", VariantChess960)
	require.NoError(t, err)

	moves := movesToStrings(g.LegalMovesList())

	assertContains(t, moves, "g1 g1")
}

func TestChess960CastlingBlockedByPiecesBetweenKingAndRookFinalPaths(t *testing.T) {
	g, err := NewGameFromFENWithVariant("6k1/8/8/8/8/8/8/4BRK1 w F - 0 1", VariantChess960)
	require.NoError(t, err)

	assertNotContains(t, movesToStrings(g.LegalMovesList()), "g1 c1")
}

func TestChess960CastlingRequiresSafeKingPath(t *testing.T) {
	g, err := NewGameFromFENWithVariant("6k1/8/8/8/8/2r5/8/5RK1 w F - 0 1", VariantChess960)
	require.NoError(t, err)

	assertNotContains(t, movesToStrings(g.LegalMovesList()), "g1 c1")
}

func TestStandardCastlingGenerationStillWorks(t *testing.T) {
	g := &Game{}
	require.NoError(t, g.LoadFEN("r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1"))

	moves := movesToStrings(g.LegalMovesList())
	assertContains(t, moves, "e1 g1")
	assertContains(t, moves, "e1 c1")
}
