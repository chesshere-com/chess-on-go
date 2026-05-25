package chessongo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadChess960FENWithShredderCastlingRights(t *testing.T) {
	const fen = "bbqnnrkr/pppppppp/8/8/8/8/PPPPPPPP/BBQNNRKR w HFhf - 0 1"

	g, err := NewGameFromFENWithVariant(fen, VariantChess960)

	require.NoError(t, err)
	require.Equal(t, VariantChess960, g.Variant())
	require.Equal(t, fen, g.FEN())
	require.True(t, g.CastlingRights().Has(CastlingWhiteKingSide))
	require.True(t, g.CastlingRights().Has(CastlingWhiteQueenSide))
	require.Equal(t, COORDS_TO_SQUARE["h1"], g.castlingRookFrom[CASTLE_WKS])
	require.Equal(t, COORDS_TO_SQUARE["f1"], g.castlingRookFrom[CASTLE_WQS])
}

func TestLoadChess960FENAcceptsUnambiguousKQCastlingRights(t *testing.T) {
	const input = "bbqnnrkr/pppppppp/8/8/8/8/PPPPPPPP/BBQNNRKR w KQkq - 0 1"
	const output = "bbqnnrkr/pppppppp/8/8/8/8/PPPPPPPP/BBQNNRKR w HFhf - 0 1"

	g, err := NewGameFromFENWithVariant(input, VariantChess960)

	require.NoError(t, err)
	require.Equal(t, output, g.FEN())
}

func TestLoadChess960FENRejectsBadCastlingRights(t *testing.T) {
	tests := []string{
		"bbqnnrkr/pppppppp/8/8/8/8/PPPPPPPP/BBQNNRKR w HHhf - 0 1",
		"bbqnnrkr/pppppppp/8/8/8/8/PPPPPPPP/BBQNNRKR w HFHf - 0 1",
		"bbqnnrkr/pppppppp/8/8/8/8/PPPPPPPP/BBQNNRKR w E - 0 1",
		"bbqnnrkr/pppppppp/8/8/8/8/PPPPPPPP/BBQNNRKR w A - 0 1",
	}

	for _, fen := range tests {
		t.Run(fen, func(t *testing.T) {
			g := &Game{}
			require.Error(t, g.LoadFENWithVariant(fen, VariantChess960))
		})
	}
}

func TestStandardFENStillRejectsChess960CastlingLetters(t *testing.T) {
	g := &Game{}
	require.Error(t, g.LoadFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w HAha - 0 1"))
}

func TestStandardFENRejectsCastlingWhenKingIsOffInitialSquare(t *testing.T) {
	tests := []string{
		"4k3/8/8/8/8/8/8/3K3R w K - 0 1",
		"3k3r/8/8/8/8/8/8/4K3 b k - 0 1",
	}

	for _, fen := range tests {
		t.Run(fen, func(t *testing.T) {
			g := &Game{}
			require.Error(t, g.LoadFEN(fen))
		})
	}
}

func TestChess960StartingFENRoundTripsThroughLoader(t *testing.T) {
	fen, err := Chess960StartingFEN(518)
	require.NoError(t, err)

	g, err := NewGameFromFENWithVariant(fen, VariantChess960)
	require.NoError(t, err)
	require.Equal(t, fen, g.FEN())
}
