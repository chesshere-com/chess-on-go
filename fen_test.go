package chessongo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_LoadFEN(t *testing.T) {
	g := NewGame()
	fens := []string{
		"rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
		"rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR w KQkq c6 0 2",
		"rnbqkbnr/pp1ppppp/8/2p5/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2",
		"rnbqkbnr/pp1ppppp/8/2p5/4P3/5N2/PPPP1PPP/RNBQKB1R b KQ e3 1 2",
		"rnbqkbnr/pp1ppppp/8/2p5/4P3/5N2/PPPP1PPP/RNBQKB1R b - - 1 2",
		"rnbqkbnr/ppp2ppp/8/3pp3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 0 3",
		"rnb1kbnr/ppp2ppp/8/8/2qp4/5N2/PPP2PPP/RNBQK2R w Qkq - 0 6",
	}
	for _, fen := range fens {
		require.NoError(t, g.LoadFEN(fen))
		require.Equal(t, fen, g.ToFEN())
	}

}

func Test_ToFEN(t *testing.T) {
	g := NewGame()
	require.Equal(t, STARTING_POSITION_FEN, g.ToFEN())
	g.LoadFEN("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1")
	require.Equal(t, "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1", g.ToFEN())
}

func TestLoadFENRejectsInvalidPositions(t *testing.T) {
	tests := []string{
		"8/8/8/8/8/8/8/8 w - - 0 1",                                       // no kings
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNRR w KQkq - 0 1",       // rank too wide
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPP/RNBQKBNR w KQkq - 0 1",         // rank too narrow
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KK - 0 1",          // duplicate castling right
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq e5 0 1",       // invalid en-passant rank
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 0",        // invalid fullmove
		"Pnbqkbnr/pppppppp/8/8/8/8/1PPPPPPP/RNBQKBNR w KQkq - 0 1",        // white pawn on eighth rank
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBN1 w KQkq - 0 1",        // missing white king
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq e3 0 1 extra", // trailing data
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq e3x 0 1",      // malformed en-passant token
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/1NBQKBNR w Qkq - 0 1",         // queenside right without a1 rook
	}

	for _, fen := range tests {
		t.Run(fen, func(t *testing.T) {
			g := &Game{}
			require.Error(t, g.LoadFEN(fen))
		})
	}
}

func TestLoadFENRefreshesLegalMovesAndStatus(t *testing.T) {
	g := &Game{}
	require.NoError(t, g.LoadFEN(STARTING_POSITION_FEN))

	require.Len(t, g.legalMoves, 20)
	require.False(t, g.isCheck)
	require.False(t, g.isCheckmate)
	require.False(t, g.IsStalemate())
}
