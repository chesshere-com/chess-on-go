package chessongo

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseEPDWithOperations(t *testing.T) {
	record, err := ParseEPD(`rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - bm e4 d4; id "initial position"; perft 3 8902;`)
	require.NoError(t, err)

	require.Equal(t, STARTING_POSITION_FEN, record.FEN)
	require.Equal(t, []string{"e4", "d4"}, record.Operations["bm"])
	require.Equal(t, []string{"initial position"}, record.Operations["id"])
	require.Equal(t, []string{"3", "8902"}, record.Operations["perft"])

	g, err := record.Game()
	require.NoError(t, err)
	require.Equal(t, STARTING_POSITION_FEN, g.FEN())
}

func TestEPDTypedHelpers(t *testing.T) {
	record, err := ParseEPD(`rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - bm e4 d4; am a3 h3; id "initial"; perft 1 20 2 400; D3 8902;`)
	require.NoError(t, err)

	require.Equal(t, []string{"e4", "d4"}, record.BestMoves())
	require.Equal(t, []string{"a3", "h3"}, record.AvoidMoves())
	id, ok := record.ID()
	require.True(t, ok)
	require.Equal(t, "initial", id)

	perft, err := record.PerftExpectations()
	require.NoError(t, err)
	require.Equal(t, map[int]uint64{1: 20, 2: 400, 3: 8902}, perft)
}

func TestLoadEPDFixtureSuite(t *testing.T) {
	tests := []struct {
		path       string
		wantCount  int
		wantPerft  bool
		wantTactic bool
	}{
		{path: "testdata/epd/perft.epd", wantCount: 3, wantPerft: true},
		{path: "testdata/epd/tactics.epd", wantCount: 2, wantTactic: true},
		{path: "testdata/epd/endgames.epd", wantCount: 2, wantPerft: true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			data, err := os.ReadFile(tt.path)
			require.NoError(t, err)

			records, err := LoadEPDRecords(string(data))
			require.NoError(t, err)
			require.Len(t, records, tt.wantCount)

			for _, record := range records {
				id, ok := record.ID()
				require.True(t, ok)
				require.NotEmpty(t, id)
				_, err := record.Game()
				require.NoError(t, err)
				if tt.wantPerft {
					perft, err := record.PerftExpectations()
					require.NoError(t, err)
					require.NotEmpty(t, perft)
				}
				if tt.wantTactic {
					require.NotEmpty(t, append(record.BestMoves(), record.AvoidMoves()...))
				}
			}
		})
	}
}

func TestParseEPDRejectsInvalidPosition(t *testing.T) {
	_, err := ParseEPD("8/8/8/8/8/8/8/8 w - - id \"bad\";")
	require.ErrorIs(t, err, ErrInvalidFEN)
}

func TestLoadEPDRecordsSkipsBlankAndCommentLines(t *testing.T) {
	records, err := LoadEPDRecords(`
# initial position
rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - id "start";

4k3/8/8/8/8/8/8/4K3 w - - id "kings";
`)
	require.NoError(t, err)
	require.Len(t, records, 2)
	require.Equal(t, "start", records[0].Operation("id")[0])
	require.Equal(t, "kings", records[1].Operation("id")[0])
}
