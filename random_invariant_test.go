package chessongo

import (
	"math/rand"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRandomGamesMaintainInvariantsAndUndo(t *testing.T) {
	games, plies := 8, 80
	if os.Getenv("CHESSONGO_HEAVY_RANDOM") == "1" {
		games, plies = 100, 200
	}

	rng := rand.New(rand.NewSource(1))
	for gameIdx := 0; gameIdx < games; gameIdx++ {
		g := NewGame()
		assertGameInvariants(t, g)

		var movesPlayed []Move
		var fens []string
		var hashes []uint64
		for ply := 0; ply < plies && len(g.legalMoves) > 0 && !g.isFinished; ply++ {
			fens = append(fens, g.ToFEN())
			hashes = append(hashes, g.zobristHash)
			move := g.legalMoves[rng.Intn(len(g.legalMoves))]
			movesPlayed = append(movesPlayed, move)

			g.MakeMove(move)
			assertGameInvariants(t, g)
		}

		for i := len(movesPlayed) - 1; i >= 0; i-- {
			g.UndoMove(movesPlayed[i])
			assertGameInvariants(t, g)
			require.Equal(t, fens[i], g.ToFEN())
			require.Equal(t, hashes[i], g.zobristHash)
		}
		require.Equal(t, STARTING_POSITION_FEN, g.ToFEN())
	}
}

func TestEveryLegalMoveFromKnownPositionsMaintainsInvariantsAndUndoes(t *testing.T) {
	for _, pos := range knownPerftPositions {
		t.Run(pos.name, func(t *testing.T) {
			g := &Game{}
			require.NoError(t, g.LoadFEN(pos.fen))
			startFen := g.ToFEN()
			startHash := g.zobristHash

			moves := g.LegalMovesList()
			for _, move := range moves {
				g.MakeMove(move)
				assertGameInvariants(t, g)
				g.UndoMove(move)
				assertGameInvariants(t, g)
				require.Equal(t, startFen, g.ToFEN())
				require.Equal(t, startHash, g.zobristHash)
			}
		})
	}
}

func TestRandomChess960MakeUndoPreservesFEN(t *testing.T) {
	for _, id := range []int{0, 42, 518, 959} {
		t.Run(strconv.Itoa(id), func(t *testing.T) {
			g, err := NewChess960Game(id)
			require.NoError(t, err)
			start := g.FEN()
			moves := g.LegalMovesList()
			require.NotEmpty(t, moves)

			for _, move := range moves {
				before := g.FEN()
				g.MakeMove(move)
				g.UndoMove(move)
				require.Equal(t, before, g.FEN())
			}
			require.Equal(t, start, g.FEN())
		})
	}
}
