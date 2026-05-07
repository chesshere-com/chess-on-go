package chessongo

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func perft(g *Game, depth int) uint64 {
	if depth == 0 {
		return 1
	}

	var moves [maxGeneratedMoves]Move
	count := g.generateLegalMovesArray(&moves)
	if depth == 1 {
		return uint64(count)
	}

	var nodes uint64
	for i := 0; i < count; i++ {
		m := moves[i]
		g.makeMoveNoGenerate(m)
		nodes += perft(g, depth-1)
		g.undoMoveNoGenerate(m)
	}
	return nodes
}

type perftPosition struct {
	name  string
	fen   string
	depth map[int]uint64
}

var knownPerftPositions = []perftPosition{
	{
		name: "initial",
		fen:  STARTING_POSITION_FEN,
		depth: map[int]uint64{
			1: 20,
			2: 400,
			3: 8902,
			4: 197281,
			5: 4865609,
		},
	},
	{
		name: "kiwipete",
		fen:  "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		depth: map[int]uint64{
			1: 48,
			2: 2039,
			3: 97862,
			4: 4085603,
		},
	},
	{
		name: "position3",
		fen:  "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1",
		depth: map[int]uint64{
			1: 14,
			2: 191,
			3: 2812,
			4: 43238,
			5: 674624,
		},
	},
	{
		name: "position4",
		fen:  "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1",
		depth: map[int]uint64{
			1: 6,
			2: 264,
			3: 9467,
			4: 422333,
		},
	},
	{
		name: "position5",
		fen:  "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8",
		depth: map[int]uint64{
			1: 44,
			2: 1486,
			3: 62379,
			4: 2103487,
		},
	},
	{
		name: "position6",
		fen:  "r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10",
		depth: map[int]uint64{
			1: 46,
			2: 2079,
			3: 89890,
			4: 3894594,
		},
	},
}

func TestPerftKnownPositionsShallow(t *testing.T) {
	for _, pos := range knownPerftPositions {
		pos := pos
		for depth, expected := range pos.depth {
			if depth > 2 {
				continue
			}
			t.Run(pos.name, func(t *testing.T) {
				g := &Game{}
				require.NoError(t, g.LoadFEN(pos.fen))
				nodes := perft(g, depth)
				require.Equalf(t, expected, nodes, "Perft(%s, %d)", pos.name, depth)
			})
		}
	}
}

func TestPerftKnownPositionsDeep(t *testing.T) {
	if os.Getenv("CHESSONGO_DEEP_PERFT") != "1" {
		t.Skip("set CHESSONGO_DEEP_PERFT=1 to run deep perft")
	}
	for _, pos := range knownPerftPositions {
		pos := pos
		for depth, expected := range pos.depth {
			if depth <= 2 {
				continue
			}
			t.Run(pos.name, func(t *testing.T) {
				g := &Game{}
				require.NoError(t, g.LoadFEN(pos.fen))
				nodes := perft(g, depth)
				require.Equalf(t, expected, nodes, "Perft(%s, %d)", pos.name, depth)
			})
		}
	}
}
