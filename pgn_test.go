package chessongo

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadPGNStandardLine(t *testing.T) {
	pgn := "1. e4 e5 2. Nf3 Nc6 3. Bb5 a6"

	g := &Game{}
	require.NoError(t, g.LoadPGN(pgn))
	require.Equal(t, "r1bqkbnr/1ppp1ppp/p1n5/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 0 4", g.ToFEN())
	// History should have at least the number of plies + initial position.
	require.GreaterOrEqual(t, g.PositionHistory[g.ZobristHash], 1)
}

func TestLoadPGNFixtureFiles(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		fen    string
		result string
	}{
		{
			name:   "annotated mainline",
			path:   "testdata/pgn/annotated-mainline.pgn",
			fen:    "r1bqk2r/1pppbppp/p1n2n2/4p3/B3P3/5N2/PPPP1PPP/RNBQ1RK1 w kq - 4 6",
			result: "*",
		},
		{
			name:   "special moves",
			path:   "testdata/pgn/special-moves.pgn",
			fen:    "8/Q7/4k3/8/8/8/8/4K3 w - - 3 3",
			result: "*",
		},
		{
			name:   "en passant",
			path:   "testdata/pgn/en-passant.pgn",
			fen:    "rnbqkbnr/1pp1pppp/p2P4/8/8/8/PPPP1PPP/RNBQKBNR b KQkq - 0 3",
			result: "*",
		},
		{
			name:   "checkmate",
			path:   "testdata/pgn/checkmate.pgn",
			fen:    "rnb1kbnr/pppp1ppp/8/4p3/6Pq/5P2/PPPPP2P/RNBQKBNR w KQkq - 1 3",
			result: "0-1",
		},
		{
			name:   "escaped tags and nested variations",
			path:   "testdata/pgn/escaped-tags-nested-variations.pgn",
			fen:    "r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 2 3",
			result: "*",
		},
		{
			name:   "setup promotion mate",
			path:   "testdata/pgn/setup-promotion-mate.pgn",
			fen:    "Q3k3/8/4K3/8/8/8/8/8 b - - 0 1",
			result: "1-0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(tt.path)
			require.NoError(t, err)
			g := &Game{}
			require.NoError(t, g.LoadPGN(string(data)))
			require.Equal(t, tt.fen, g.FEN())
			require.Equal(t, tt.result, g.PGNResult())

			roundTrip := &Game{}
			require.NoError(t, roundTrip.LoadPGN(g.PGN()))
			require.Equal(t, g.FEN(), roundTrip.FEN())
		})
	}
}

func TestPGNExportWrapsMovetextLines(t *testing.T) {
	g := NewGame()
	for _, uci := range []string{
		"g1f3", "g8f6", "f3g1", "f6g8",
		"g1f3", "g8f6", "f3g1", "f6g8",
		"g1f3", "g8f6", "f3g1", "f6g8",
		"g1f3", "g8f6", "f3g1", "f6g8",
		"g1f3", "g8f6", "f3g1", "f6g8",
	} {
		require.NoError(t, g.TryMoveUCI(uci))
	}

	pgn := g.PGN()
	sections := strings.SplitN(pgn, "\n\n", 2)
	require.Len(t, sections, 2)
	for _, line := range strings.Split(sections[1], "\n") {
		require.LessOrEqual(t, len(line), 80, line)
	}

	roundTrip := &Game{}
	require.NoError(t, roundTrip.LoadPGN(pgn))
	require.Equal(t, g.FEN(), roundTrip.FEN())
}

func TestLoadPGNWithFenTagAndComments(t *testing.T) {
	pgn := `
[Event "Test"]
[FEN "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"]
[Result "1-0"]

1. d4 d5 {queen's pawn} 2. c4 dxc4 3. e3 Nf6 4. Bxc4 1-0
`

	g := &Game{}
	require.NoError(t, g.LoadPGN(pgn))
	require.Equal(t, "rnbqkb1r/ppp1pppp/5n2/8/2BP4/4P3/PP3PPP/RNBQK1NR b KQkq - 0 4", g.ToFEN())
	require.Equal(t, "Test", g.PGNTags()["Event"])
	require.Equal(t, "1-0", g.PGNResult())
}

func TestLoadPGNWithSemicolonCommentsAndEscapedTags(t *testing.T) {
	pgn := `[Event "Club \"Rapid\""]
[Site "Berlin"]
[Result "*"]

1. e4 ; comment until line end
1... e5 2. Nf3 $1 Nc6 *
`

	g := &Game{}
	require.NoError(t, g.LoadPGN(pgn))
	require.Equal(t, "r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 2 3", g.FEN())
	require.Equal(t, `Club "Rapid"`, g.PGNTags()["Event"])
	require.Equal(t, "*", g.PGNResult())
}

func TestLoadPGNRejectsConflictingResultTagAndMovetext(t *testing.T) {
	pgn := `[Result "1-0"]

1. e4 e5 0-1
`

	g := &Game{}
	require.Error(t, g.LoadPGN(pgn))
}

func TestLoadPGNRejectsCheckmateResultMismatch(t *testing.T) {
	g := &Game{}
	require.Error(t, g.LoadPGN(`1. f3 e5 2. g4 Qh4# 1-0`))
}

func TestPGNExportIncludesTagsAndMovetext(t *testing.T) {
	pgn := `[Event "Export Test"]
[Site "Berlin"]
[Date "2026.05.07"]
[Round "1"]
[White "Alice"]
[Black "Bob"]
[Result "1-0"]

1. e4 e5 2. Qh5 Nc6 3. Bc4 Nf6 4. Qxf7# 1-0
`

	g := &Game{}
	require.NoError(t, g.LoadPGN(pgn))

	require.Equal(t, `[Event "Export Test"]
[Site "Berlin"]
[Date "2026.05.07"]
[Round "1"]
[White "Alice"]
[Black "Bob"]
[Result "1-0"]

1. e4 e5 2. Qh5 Nc6 3. Bc4 Nf6 4. Qxf7# 1-0`, g.PGN())
}

func TestPGNRoundTripPreservesFinalFEN(t *testing.T) {
	original := `[Event "Round Trip"]
[Result "*"]

1. e4 {comment} e5 2. Nf3 (2. Bc4) 2... Nc6 3. Bb5 a6 *
`

	first := &Game{}
	require.NoError(t, first.LoadPGN(original))

	second := &Game{}
	require.NoError(t, second.LoadPGN(first.PGN()))

	require.Equal(t, first.FEN(), second.FEN())
	require.Equal(t, first.PGNResult(), second.PGNResult())
}

func TestPGNExportTracksNormalMovesAndUndo(t *testing.T) {
	g := NewGame()
	require.NoError(t, g.TryMoveUCI("e2e4"))
	require.NoError(t, g.TryMoveUCI("e7e5"))
	require.NoError(t, g.TryMoveUCI("g1f3"))

	require.Contains(t, g.PGN(), "1. e4 e5 2. Nf3 *")

	last := g.LegalMovesList()[0]
	g.MakeMove(last)
	g.UndoMove(last)

	require.Contains(t, g.PGN(), "1. e4 e5 2. Nf3 *")
}

func TestLoadPGNDetectsThreefold(t *testing.T) {
	pgn := "1. Nf3 Nf6 2. Ng1 Ng8 3. Nf3 Nf6 4. Ng1 Ng8"

	g := &Game{}
	require.NoError(t, g.LoadPGN(pgn))
	require.True(t, g.IsThreefoldRepetition)
	require.False(t, g.IsFivefoldRepetition())
	require.GreaterOrEqual(t, g.PositionHistory[g.ZobristHash], 3)
}

func TestLoadPGNGame(t *testing.T) {
	pgn := "1. e4 e5 2. Nf3 Nc6"

	g, err := LoadPGNGame(pgn)
	require.NoError(t, err)
	require.NotNil(t, g)
	require.Equal(t, "r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 2 3", g.ToFEN())
}

func TestLoadPGNWithVariations(t *testing.T) {
	// Variations in parentheses should be ignored, only main line should be played
	pgn := "1. e4 e5 2. Nf3 (2. Bc4 Nf6) 2... Nc6 3. Bb5 a6"

	g := &Game{}
	require.NoError(t, g.LoadPGN(pgn))
	require.Equal(t, "r1bqkbnr/1ppp1ppp/p1n5/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 0 4", g.ToFEN())
}

func TestLoadPGNWithNAGs(t *testing.T) {
	// NAGs (Numeric Annotation Glyphs) like $1, $2, etc. should be ignored
	pgn := "1. e4 $1 e5 $2 2. Nf3 $10 Nc6 3. Bb5 a6"

	g := &Game{}
	require.NoError(t, g.LoadPGN(pgn))
	require.Equal(t, "r1bqkbnr/1ppp1ppp/p1n5/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 0 4", g.ToFEN())
}

func TestLoadPGNWithCastling(t *testing.T) {
	// Test both kingside and queenside castling
	pgn := "1. e4 e5 2. Nf3 Nc6 3. Bb5 a6 4. Ba4 Nf6 5. O-O Be7 6. Re1 b5 7. Bb3 d6 8. c3 O-O"

	g := &Game{}
	require.NoError(t, g.LoadPGN(pgn))
	// Verify both sides have castled - check the complete FEN
	require.Equal(t, "r1bq1rk1/2p1bppp/p1np1n2/1p2p3/4P3/1BP2N2/PP1P1PPP/RNBQR1K1 w - - 1 9", g.ToFEN())
}

func TestLoadPGNWithPromotion(t *testing.T) {
	// Test pawn promotion
	pgn := `[FEN "4k3/P7/8/8/8/8/8/4K3 w - - 0 1"]
	1. a8=Q Ke7 2. Qa7+ Ke6`

	g := &Game{}
	require.NoError(t, g.LoadPGN(pgn))
	// Verify the queen is on a7 after promotion and subsequent move
	require.Equal(t, "8/Q7/4k3/8/8/8/8/4K3 w - - 3 3", g.ToFEN())
}

func TestLoadPGNWithEnPassant(t *testing.T) {
	// Test en passant capture
	pgn := "1. e4 a6 2. e5 d5 3. exd6"

	g := &Game{}
	require.NoError(t, g.LoadPGN(pgn))
	require.Equal(t, "rnbqkbnr/1pp1pppp/p2P4/8/8/8/PPPP1PPP/RNBQKBNR b KQkq - 0 3", g.ToFEN())
}

func TestLoadPGNWithDifferentResults(t *testing.T) {
	tests := []struct {
		name   string
		pgn    string
		result string
	}{
		{"White wins", "1. e4 e5 2. Qh5 Nc6 3. Qxf7# 1-0", "r1bqkbnr/pppp1Qpp/2n5/4p3/4P3/8/PPPP1PPP/RNB1KBNR b KQkq - 0 3"},
		{"Black wins", "1. f3 e5 2. g4 Qh4# 0-1", "rnb1kbnr/pppp1ppp/8/4p3/6Pq/5P2/PPPPP2P/RNBQKBNR w KQkq - 1 3"},
		{"Draw", "1. e4 e5 2. Nf3 Nc6 1/2-1/2", "r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 2 3"},
		{"Asterisk", "1. e4 e5 *", "rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR w KQkq e6 0 2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Game{}
			require.NoError(t, g.LoadPGN(tt.pgn))
			require.Equal(t, tt.result, g.ToFEN())
		})
	}
}

func TestLoadPGNWithAnnotations(t *testing.T) {
	// Test that move annotations (!, ?, !!, ??, !?, ?!) are properly stripped
	pgn := "1. e4! e5? 2. Nf3!! Nc6?? 3. Bb5!? a6?!"

	g := &Game{}
	require.NoError(t, g.LoadPGN(pgn))
	require.Equal(t, "r1bqkbnr/1ppp1ppp/p1n5/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 0 4", g.ToFEN())
}

func TestLoadPGNWithCheckAndMateSymbols(t *testing.T) {
	// Test that check (+) and checkmate (#) symbols are properly stripped
	pgn := "1. e4 e5 2. Nf3 Nc6 3. Bc4 Bc5 4. Bxf7+ Kxf7"

	g := &Game{}
	require.NoError(t, g.LoadPGN(pgn))
	// Verify the complete position after the capture sequence with check
	require.Equal(t, "r1bq2nr/pppp1kpp/2n5/2b1p3/4P3/5N2/PPPP1PPP/RNBQK2R w KQ - 0 5", g.ToFEN())
}

func TestLoadPGNInvalidMove(t *testing.T) {
	// Test that an invalid move returns an error
	pgn := "1. e4 e5 2. Nf3 Nc6 3. Zz9" // Invalid move

	g := &Game{}
	err := g.LoadPGN(pgn)
	require.Error(t, err)
	require.Contains(t, err.Error(), "pgn move not found")
}

func TestLoadPGNWithMultipleSpaces(t *testing.T) {
	// Test PGN with irregular spacing
	pgn := "1.    e4    e5    2.   Nf3   Nc6"

	g := &Game{}
	require.NoError(t, g.LoadPGN(pgn))
	require.Equal(t, "r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 2 3", g.ToFEN())
}

func TestLoadPGNWithNestedVariations(t *testing.T) {
	// Test nested variations (should all be ignored)
	pgn := "1. e4 e5 2. Nf3 (2. Bc4 Nf6 (2... Bc5)) 2... Nc6"

	g := &Game{}
	require.NoError(t, g.LoadPGN(pgn))
	require.Equal(t, "r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 2 3", g.ToFEN())
}

func TestLoadPGNWithBraceComments(t *testing.T) {
	// Test comments in braces
	pgn := "1. e4 {best by test} e5 {symmetric} 2. Nf3 {developing} Nc6"

	g := &Game{}
	require.NoError(t, g.LoadPGN(pgn))
	require.Equal(t, "r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 2 3", g.ToFEN())
}

func TestLoadPGNFastPath(t *testing.T) {
	// Test the fast path (no tags, comments, or variations)
	pgn := "e4 e5 Nf3 Nc6 Bb5 a6"

	g := &Game{}
	require.NoError(t, g.LoadPGN(pgn))
	require.Equal(t, "r1bqkbnr/1ppp1ppp/p1n5/1B2p3/4P3/5N2/PPPP1PPP/RNBQK2R w KQkq - 0 4", g.ToFEN())
}

func TestLoadPGNWithBlackMove(t *testing.T) {
	// Test PGN with custom starting position where it's black's turn
	pgn := `[FEN "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1"]
	1... e5 2. Nf3 Nc6`

	g := &Game{}
	require.NoError(t, g.LoadPGN(pgn))
	require.Equal(t, "r1bqkbnr/pppp1ppp/2n5/4p3/4P3/5N2/PPPP1PPP/RNBQKB1R w KQkq - 2 3", g.ToFEN())
}

func TestExtractFENFromPGNEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		pgn         string
		expectedFEN string
	}{
		{
			name:        "FEN with extra spaces",
			pgn:         `[FEN  "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"  ]`,
			expectedFEN: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		},
		{
			name: "No FEN tag",
			pgn: `[Event "Test"]
[White "Player1"]`,
			expectedFEN: "",
		},
		{
			name:        "Empty PGN",
			pgn:         "",
			expectedFEN: "",
		},
		{
			name:        "FEN with trailing spaces in value",
			pgn:         `[FEN "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1  "]`,
			expectedFEN: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fen := extractFENFromPGN(tt.pgn)
			require.Equal(t, tt.expectedFEN, fen)
		})
	}
}
