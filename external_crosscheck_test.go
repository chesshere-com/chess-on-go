package chessongo

import (
	"sort"
	"testing"

	notnil "github.com/notnil/chess"
	"github.com/stretchr/testify/require"
)

func TestLegalMovesMatchNotnilChess(t *testing.T) {
	fens := []string{
		STARTING_POSITION_FEN,
		"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		"8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
		"r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
		"rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
		"r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
		"k3r3/8/8/8/1b6/8/8/4K3 w - - 0 1",
		"rnbqkbnr/ppp1p1pp/8/3pPp2/8/8/PPPP1PPP/RNBQKBNR w KQkq f6 0 3",
	}

	for _, fen := range fens {
		t.Run(fen, func(t *testing.T) {
			g := &Game{}
			require.NoError(t, g.LoadFEN(fen))
			ours := sortedMoveUCIs(g.legalMoves)

			option, err := notnil.FEN(fen)
			require.NoError(t, err)
			other := notnil.NewGame(option)
			theirs := notnilUCIMoves(other)

			require.Equal(t, theirs, ours)
		})
	}
}

func notnilUCIMoves(game *notnil.Game) []string {
	notation := notnil.UCINotation{}
	moves := game.ValidMoves()
	labels := make([]string, len(moves))
	for i, move := range moves {
		labels[i] = notation.Encode(game.Position(), move)
	}
	sort.Strings(labels)
	return labels
}
