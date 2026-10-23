package chessongo

import (
	"encoding/binary"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClonePreservesCachedMovesAndOwnsAllMutableState(t *testing.T) {
	g, err := LoadPGNGame("1. e4 e5 2. Nf3")
	require.NoError(t, err)
	g.GeneratePseudoMoves()
	before := g.Snapshot()
	clone := g.Clone()
	require.Equal(t, before, clone.Snapshot())
	require.Equal(t, g.pseudoMoves, clone.pseudoMoves)
	clone.legalMoves[0] = 0
	clone.pseudoMoves[0] = 0
	clone.history[0].HalfMoves = 123
	clone.positionHistory[g.PositionKey()] = 123
	clone.pgnMoves[0] = "bad"
	clone.pgnTags["Result"] = "bad"
	require.Equal(t, before, g.Snapshot())
	require.NotZero(t, g.pseudoMoves[0])
	require.NotEqual(t, 123, g.history[0].HalfMoves)
	require.NotEqual(t, "bad", g.pgnMoves[0])
	require.NotEqual(t, "bad", g.pgnTags["Result"])
}

func TestRepetitionHashRequiresLegalEnPassant(t *testing.T) {
	for _, fen := range []string{
		"k3r3/8/8/3pP3/8/8/8/4K3 w - d6 0 1",
		"8/8/8/r4pPK/8/8/8/k7 w - f6 0 1",
		"4k3/8/8/8/3Pp3/8/8/K3R3 b - d3 0 1",
	} {
		t.Run(fen, func(t *testing.T) {
			g, err := NewGameFromFEN(fen)
			require.NoError(t, err)
			fields := strings.Fields(fen)
			fields[3] = "-"
			without, err := NewGameFromFEN(strings.Join(fields, " "))
			require.NoError(t, err)
			require.Equal(t, sortedMoveUCIs(g.legalMoves), sortedMoveUCIs(without.legalMoves))
			require.Equal(t, without.PositionKey(), g.PositionKey())
		})
	}
}

func TestMaterialDrawRespectsPossibleWins(t *testing.T) {
	tests := []struct {
		variant Variant
		fen     string
		draw    bool
	}{
		{VariantStandard, "kn6/8/8/8/8/8/8/3NK3 w - - 0 1", false},
		{VariantKingOfTheHill, "k7/8/8/8/8/8/8/4K3 w - - 0 1", false},
		{VariantThreeCheck, "k7/8/8/8/8/8/8/3NK3 w - - 0 1 +0+0", false},
		{VariantThreeCheck, "k7/8/8/8/8/8/8/3BK3 w - - 0 1 +0+0", false},
		{VariantThreeCheck, "k7/8/8/8/8/8/8/4K3 w - - 0 1 +0+0", true},
	}
	for _, tt := range tests {
		g, err := NewGameFromFENWithVariant(tt.fen, tt.variant)
		require.NoError(t, err)
		require.Equal(t, tt.draw, g.DrawStatus().InsufficientMaterial, tt.fen)
		require.Equal(t, tt.draw, g.IsTerminal(), tt.fen)
	}
}

func TestChess960CastlingRequestDistinguishesKingStep(t *testing.T) {
	const fen = "4k3/8/8/8/8/8/8/5K1R w H - 0 1"
	for _, notation := range []string{"O-O", "f1h1"} {
		g, err := NewGameFromFENWithVariant(fen, VariantChess960)
		require.NoError(t, err)
		if notation == "O-O" {
			require.NoError(t, g.TryMoveSAN(notation))
		} else {
			require.NoError(t, g.TryMoveUCI(notation))
		}
		require.Equal(t, "4k3/8/8/8/8/8/8/5RK1 b - - 1 1", g.FEN())
	}
	g, err := NewGameFromFENWithVariant(fen, VariantChess960)
	require.NoError(t, err)
	require.NoError(t, g.TryMoveSAN("Kg1"))
	require.Equal(t, "4k3/8/8/8/8/8/8/6KR b - - 1 1", g.FEN())
}

func TestChess960RejectsKingOffBackRankWithCastlingRights(t *testing.T) {
	_, err := NewGameFromFENWithVariant("4k3/8/8/8/8/8/4K3/7R w H - 0 1", VariantChess960)
	require.Error(t, err)
}

func TestPGNCompactNumbersAndCommentPunctuation(t *testing.T) {
	want, err := LoadPGNGame("1. e4 e5 2. Nf3 Nc6")
	require.NoError(t, err)
	for _, pgn := range []string{
		"1.e4 e5 2.Nf3 Nc6",
		"1.e4 1...e5 2.Nf3 2...Nc6",
		"1. e4 {a comment with an unmatched (} e5 2. Nf3 Nc6",
		"1. e4 {a comment with a ; semicolon} e5 2. Nf3 Nc6",
		"1. e4 {multi-line\n[not a tag] comment} e5 2. Nf3 Nc6",
	} {
		g, err := LoadPGNGame(pgn)
		require.NoError(t, err, pgn)
		require.Equal(t, want.FEN(), g.FEN(), pgn)
	}
}

func TestPGNZeroCastlingAndResultValidation(t *testing.T) {
	g, err := LoadPGNGame("1. e4 e5 2. Nf3 Nc6 3. Bc4 Nf6 4. 0-0")
	require.NoError(t, err)
	require.Contains(t, g.PGN(), "O-O")
	for _, pgn := range []string{
		"1. e4 1-0 e5", "1. e4 1-0 0-1",
		"[Variant \"Atomic\"]\n\n1. e4",
	} {
		_, err := LoadPGNGame(pgn)
		require.Error(t, err, pgn)
	}
}

func TestPGNFromCustomFENRoundTripsAndUndoClearsResult(t *testing.T) {
	g, err := NewGameFromFEN("4k3/8/8/8/8/8/R7/4K3 b - - 0 7")
	require.NoError(t, err)
	require.NoError(t, g.TryMoveUCI("e8f8"))
	replayed, err := LoadPGNGame(g.PGN())
	require.NoError(t, err)
	require.Equal(t, g.FEN(), replayed.FEN())

	g, err = LoadPGNGame("1. e4 1-0")
	require.NoError(t, err)
	move, err := NewMoveFromUCI("e2e4")
	require.NoError(t, err)
	g.UndoMove(move)
	require.Equal(t, "*", g.PGNResult())
	require.Equal(t, "*", g.PGNTags()["Result"])
}

func TestPGNPreservesUnplayedStartingPosition(t *testing.T) {
	for _, variant := range []Variant{VariantStandard, VariantChess960, VariantKingOfTheHill, VariantThreeCheck} {
		fen := "4k3/8/8/8/8/8/R7/4K3 b - - 0 7"
		if variant == VariantThreeCheck {
			fen += " +1+0"
		}
		g, err := NewGameFromFENWithVariant(fen, variant)
		require.NoError(t, err)
		replayed, err := LoadPGNGame(g.PGN())
		require.NoError(t, err)
		require.Equal(t, g.Variant(), replayed.Variant())
		require.Equal(t, g.FEN(), replayed.FEN())
	}
}

func TestEPDWhitespaceQuotedDelimitersAndUnclosedQuotes(t *testing.T) {
	line := "4k3/8/8/8/8/8/8/4K3\tw  -\t-  id \"semi;colon and \\\"quote\\\"\"; bm Kd2;"
	record, err := ParseEPD(line)
	require.NoError(t, err)
	id, ok := record.ID()
	require.True(t, ok)
	require.Equal(t, "semi;colon and \"quote\"", id)
	require.Equal(t, []string{"Kd2"}, record.BestMoves())
	for _, operands := range []string{`id "id`, `id "same" "same`, `id "unfinished;`} {
		_, err := ParseEPD("4k3/8/8/8/8/8/8/4K3 w - - " + operands)
		require.Error(t, err, operands)
	}
}

func TestBinaryRejectsInvalidPositionsTransactionally(t *testing.T) {
	g := NewGame()
	data, err := g.MarshalBinary()
	require.NoError(t, err)
	for name, mutate := range map[string]func([]byte){
		"promotion-rank pawn":   func(b []byte) { b[4] = B_PAWN; b[4+8] = EMPTY; b[69] = 0 },
		"adjacent kings":        func(b []byte) { b[4+4] = EMPTY; b[4+52] = B_KING; b[69] = 0 },
		"wrong ep rank":         func(b []byte) { b[70] = 32; b[4+40] = B_PAWN },
		"wrong rook metadata":   func(b []byte) { b[86+CASTLE_WKS] = 0 },
		"invalid history count": func(b []byte) { binary.LittleEndian.PutUint32(b[110:114], 0) },
	} {
		t.Run(name, func(t *testing.T) {
			bad := append([]byte(nil), data...)
			mutate(bad)
			before := g.Snapshot()
			require.Error(t, g.UnmarshalBinary(bad))
			require.Equal(t, before, g.Snapshot())
		})
	}
}

func TestBinaryRejectsCounterTruncation(t *testing.T) {
	if strconv.IntSize < 64 {
		t.Skip("requires 64-bit int")
	}
	g, err := NewGameFromFEN("4k3/8/8/8/8/8/R7/4K3 w - - 0 4294967296")
	require.NoError(t, err)
	_, err = g.MarshalBinary()
	require.Error(t, err)
}

func TestMoveCountersDoNotOverflow(t *testing.T) {
	fen := "4k3/8/8/8/8/8/R7/4K3 b - - " + strconv.Itoa(math.MaxInt) + " " + strconv.Itoa(math.MaxInt)
	g, err := NewGameFromFEN(fen)
	require.NoError(t, err)
	m, err := NewMoveFromUCI("e8f8")
	require.NoError(t, err)
	g.MakeMoveFast(m)
	require.Equal(t, math.MaxInt, g.HalfMoveClock())
	require.Equal(t, math.MaxInt, g.FullMoveNumber())
	g.UndoMoveFast(m)
	require.Equal(t, fen, g.FEN())
}

func TestSEERejectsDefendedKingRecapture(t *testing.T) {
	// After Rxd7, the rook is defended by a bishop on g4.
	g, err := NewGameFromFEN("4k3/3p4/8/8/6B1/8/8/3RK3 w - - 0 1")
	require.NoError(t, err)
	from, _ := ParseSquare("d1")
	to, _ := ParseSquare("d7")
	require.Equal(t, 100, g.SEE(from, to))
}

func TestSearchBoardPerftNonpositiveDepth(t *testing.T) {
	b, err := NewSearchBoard(STARTING_POSITION_FEN)
	require.NoError(t, err)
	require.EqualValues(t, 0, b.Perft(-1))
	require.EqualValues(t, 1, b.Perft(0))
	require.Equal(t, STARTING_POSITION_FEN, b.FEN())
}

func BenchmarkFENExport(b *testing.B) {
	g := NewGame()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkText = g.FEN()
	}
}

func BenchmarkParseSAN(b *testing.B) {
	g := NewGame()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := g.ParseMoveSAN("Nf3"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFormatSAN(b *testing.B) {
	g := NewGame()
	m, err := g.ParseMoveSAN("Nf3")
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkText = g.GetMoveSan(m)
	}
}

var benchmarkText string
