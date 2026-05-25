package chessongo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTryMoveUCI(t *testing.T) {
	g := NewGame()

	require.NoError(t, g.TryMoveUCI("e2e4"))
	require.Equal(t, "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1", g.FEN())
}

func TestTryMoveUCIRejectsInvalidMoveAndPreservesPosition(t *testing.T) {
	g := NewGame()
	start := g.FEN()

	require.Error(t, g.TryMoveUCI("e2e5"))
	require.Equal(t, start, g.FEN())
	require.Error(t, g.TryMoveUCI("bad"))
	require.Equal(t, start, g.FEN())
}

func TestTryMoveUCIPromotion(t *testing.T) {
	g, err := NewGameFromFEN("4k3/P7/8/8/8/8/8/4K3 w - - 0 1")
	require.NoError(t, err)

	require.NoError(t, g.TryMoveUCI("a7a8q"))
	require.Equal(t, Piece(W_QUEEN), g.squares[COORDS_TO_SQUARE["a8"]])
}

func TestTryMoveUCIRequiresPromotionSuffix(t *testing.T) {
	g, err := NewGameFromFEN("4k3/P7/8/8/8/8/8/4K3 w - - 0 1")
	require.NoError(t, err)

	require.Error(t, g.TryMoveUCI("a7a8"))
}

func TestChess960ParseUCIRecognizesCastlingToRookSquare(t *testing.T) {
	g, err := NewGameFromFENWithVariant("6k1/8/8/8/8/8/8/6KR w H - 0 1", VariantChess960)
	require.NoError(t, err)

	move, err := g.ParseMoveUCI("g1h1")
	require.NoError(t, err)
	require.True(t, move.IsCastlingMove())
	require.Equal(t, COORDS_TO_SQUARE["g1"], move.To())

	require.NoError(t, g.TryMoveUCI("g1h1"))
	require.Equal(t, Piece(W_ROOK), g.squares[COORDS_TO_SQUARE["f1"]])
}

func TestTryMoveSAN(t *testing.T) {
	g := NewGame()

	require.NoError(t, g.TryMoveSAN("e4"))
	require.NoError(t, g.TryMoveSAN("e5"))
	require.NoError(t, g.TryMoveSAN("Nf3"))

	require.Equal(t, "rnbqkbnr/pppp1ppp/8/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2", g.FEN())
}

func TestTryMoveSANCastlingAndSuffix(t *testing.T) {
	g := NewGame()
	for _, san := range []string{"e4", "e5", "Nf3", "Nc6", "Bb5", "a6", "O-O"} {
		require.NoError(t, g.TryMoveSAN(san))
	}

	require.Equal(t, Piece(W_KING), g.squares[COORDS_TO_SQUARE["g1"]])
	require.Equal(t, Piece(W_ROOK), g.squares[COORDS_TO_SQUARE["f1"]])

	check, err := NewGameFromFEN("6k1/8/8/8/8/8/8/5RK1 w - - 0 1")
	require.NoError(t, err)
	require.NoError(t, check.TryMoveSAN("Rf8+"))
	require.True(t, check.isCheck)
}

func TestChess960SANForCastling(t *testing.T) {
	g, err := NewGameFromFENWithVariant("6k1/8/8/8/8/8/8/6KR w H - 0 1", VariantChess960)
	require.NoError(t, err)

	for _, move := range g.LegalMovesList() {
		if move.IsCastlingMove() {
			require.Equal(t, "O-O", g.GetMoveSanWithoutSuffix(move))
			return
		}
	}
	t.Fatal("missing castling move")
}

func TestTryMoveSANRejectsInvalidMoveAndPreservesPosition(t *testing.T) {
	g := NewGame()
	start := g.FEN()

	require.Error(t, g.TryMoveSAN("Qh5"))
	require.Equal(t, start, g.FEN())
}
