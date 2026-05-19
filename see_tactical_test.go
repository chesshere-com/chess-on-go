package chessongo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSEETacticalSuites(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		from, to string
		expected int
	}{
		{
			name:     "Pawn fork exchange",
			fen:      "r1bqk2r/ppp1bppp/2n1pn2/3p4/3PP3/2N2N2/PPP1BPPP/R1BQK2R w KQkq - 4 6",
			from:     "e4",
			to:       "d5",
			expected: 0, // e4xd5 e6xd5 yields net 0 pawn trade
		},
		{
			name:     "X-Ray rook exchange with extra defender",
			fen:      "3r2k1/5ppp/8/3p4/8/8/3R2PP/3R2K1 w - - 0 1",
			from:     "d2",
			to:       "d5",
			expected: 100, // White rook captures d5 pawn (100). If d8 rook recaptures, d1 rook recaptures back, winning a pawn.
		},
		{
			name:     "Invaluable queen capture protected by pinned bishop",
			fen:      "1k1r3q/1ppn3p/p4b2/4p3/8/P2N2P1/1PP1R1BP/2K1Q3 w - - 0 1",
			from:     "d3",
			to:       "e5",
			expected: -220, // Knight captures e5 pawn (100) but is recaptured by black.
		},
		{
			name:     "Promotion capture of rook on h8",
			fen:      "4k2r/6P1/8/8/8/8/8/4K3 w - - 0 1",
			from:     "g7",
			to:       "h8",
			expected: 1300, // g7xh8=Q: Rook (500) + Promoted Queen value (900) - Pawn value (100) = 1300
		},
		{
			name:     "Folk Queen sacrifice termination",
			fen:      "5rk1/5ppp/8/3q4/3Q4/8/8/5RK1 w - - 0 1",
			from:     "d4",
			to:       "d5",
			expected: 900, // Queen swap on d5. White wins a queen (+900), recaptured by rook/king (-900), net 0, but since it's a queen, SEE correctly stops.
		},
		{
			name:     "Complex battery on d4",
			fen:      "k2r4/1p1r4/8/3p4/3R4/3R4/3R4/K7 w - - 0 1",
			from:     "d4",
			to:       "d5",
			expected: 100, // White has a tripled rook battery on d-file, black has a doubled rook battery. White wins the d5 pawn.
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := &Game{}
			require.NoError(t, g.LoadFEN(tc.fen))
			fromSq, err := ParseSquare(tc.from)
			require.NoError(t, err)
			toSq, err := ParseSquare(tc.to)
			require.NoError(t, err)
			require.Equal(t, tc.expected, g.SEE(fromSq, toSq))
		})
	}
}
