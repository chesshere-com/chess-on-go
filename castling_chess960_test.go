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

func TestChess960KingsideCastleWhereKingDoesNotMove(t *testing.T) {
	g, err := NewGameFromFENWithVariant("6k1/8/8/8/8/8/8/6KR w H - 0 1", VariantChess960)
	require.NoError(t, err)
	start := g.FEN()

	var castle Move
	for _, move := range g.LegalMovesList() {
		if move.IsCastlingMove() && move.To() == COORDS_TO_SQUARE["g1"] {
			castle = move
			break
		}
	}
	require.NotZero(t, castle)

	g.MakeMove(castle)
	require.Equal(t, Piece(W_KING), g.squares[COORDS_TO_SQUARE["g1"]])
	require.Equal(t, Piece(W_ROOK), g.squares[COORDS_TO_SQUARE["f1"]])
	require.False(t, g.CastlingRights().Has(CastlingWhiteKingSide))

	g.UndoMove(castle)
	require.Equal(t, start, g.FEN())
}

func TestChess960CanMoveAllowsKingsideCastleWhereKingDoesNotMove(t *testing.T) {
	g, err := NewGameFromFENWithVariant("6k1/8/8/8/8/8/8/6KR w H - 0 1", VariantChess960)
	require.NoError(t, err)

	var castle Move
	for _, move := range g.LegalMovesList() {
		if move.IsCastlingMove() && move.To() == COORDS_TO_SQUARE["g1"] {
			castle = move
			break
		}
	}
	require.NotZero(t, castle)

	require.True(t, g.CanMove(castle))
}

func TestChess960CanMoveAllowsKingsideCastleWhereRookOriginIsAttacked(t *testing.T) {
	g, err := NewGameFromFENWithVariant("k6r/8/8/8/8/8/8/6KR w H - 0 1", VariantChess960)
	require.NoError(t, err)

	var castle Move
	for _, move := range g.LegalMovesList() {
		if move.IsCastlingMove() && move.To() == COORDS_TO_SQUARE["g1"] {
			castle = move
			break
		}
	}
	require.NotZero(t, castle)

	require.True(t, g.CanMove(castle))
}

func TestChess960QueensideCastleInterchangesKingAndRook(t *testing.T) {
	g, err := NewGameFromFENWithVariant("4k3/8/8/8/8/8/8/2RK4 w C - 0 1", VariantChess960)
	require.NoError(t, err)

	require.NoError(t, g.TryMoveUCI("d1c1"))
	require.Equal(t, Piece(W_KING), g.squares[COORDS_TO_SQUARE["c1"]])
	require.Equal(t, Piece(W_ROOK), g.squares[COORDS_TO_SQUARE["d1"]])
}

func TestChess960CanMoveAllowsQueensideCastleInterchange(t *testing.T) {
	g, err := NewGameFromFENWithVariant("4k3/8/8/8/8/8/8/2RK4 w C - 0 1", VariantChess960)
	require.NoError(t, err)

	var castle Move
	for _, move := range g.LegalMovesList() {
		if move.IsCastlingMove() && move.To() == COORDS_TO_SQUARE["c1"] {
			castle = move
			break
		}
	}
	require.NotZero(t, castle)

	require.True(t, g.CanMove(castle))
}

func TestChess960MovingCastlingRookClearsOnlyThatRight(t *testing.T) {
	g, err := NewGameFromFENWithVariant("6k1/8/8/8/8/8/8/4R1KR w HE - 0 1", VariantChess960)
	require.NoError(t, err)

	require.NoError(t, g.TryMoveUCI("e1e2"))

	require.False(t, g.CastlingRights().Has(CastlingWhiteQueenSide))
	require.True(t, g.CastlingRights().Has(CastlingWhiteKingSide))
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
