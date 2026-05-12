package chessongo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSEE_EmptyFromReturnsZero(t *testing.T) {
	g := &Game{}
	require.NoError(t, g.LoadFEN("4k3/8/8/8/8/8/8/4K3 w - - 0 1"))
	from, err := ParseSquare("e4")
	require.NoError(t, err)
	to, err := ParseSquare("d5")
	require.NoError(t, err)
	require.Equal(t, 0, g.SEE(from, to))
}

func TestSEE_OutOfRangeReturnsZero(t *testing.T) {
	g := &Game{}
	require.NoError(t, g.LoadFEN("4k3/8/8/8/8/8/8/4K3 w - - 0 1"))
	require.Equal(t, 0, g.SEE(Square(64), Square(0)))
	require.Equal(t, 0, g.SEE(Square(0), Square(64)))
	require.Equal(t, 0, g.SEE(Square(64), Square(64)))
}

func TestSEEComputePins_PinnedKnightAlongFile(t *testing.T) {
	// White rook on b1 pins black knight on b6 against black king on b8.
	g := &Game{}
	require.NoError(t, g.LoadFEN("1k6/8/1n6/8/8/8/8/1R2K3 w - - 0 1"))

	b6, _ := ParseSquare("b6")
	pinned, pinRays := g.seeComputePins(BLACK)

	require.NotZero(t, pinned&(Bitboard(1)<<uint(b6)), "expected b6 to be pinned")
	// Pin ray must cover the entire b-file from king (exclusive) to pinner (inclusive).
	b1, _ := ParseSquare("b1")
	b7, _ := ParseSquare("b7")
	require.NotZero(t, pinRays[b6]&(Bitboard(1)<<uint(b1)), "pin ray should include pinner b1")
	require.NotZero(t, pinRays[b6]&(Bitboard(1)<<uint(b7)), "pin ray should include intermediate b7")
	// A square off the b-file (e.g. d5) must NOT be on the pin ray.
	d5, _ := ParseSquare("d5")
	require.Zero(t, pinRays[b6]&(Bitboard(1)<<uint(d5)), "pin ray should not include d5")
}

func TestSEEComputePins_NoPin(t *testing.T) {
	// No sliders, no pins.
	g := &Game{}
	require.NoError(t, g.LoadFEN("4k3/8/8/8/8/8/8/4K3 w - - 0 1"))

	whitePinned, _ := g.seeComputePins(WHITE)
	blackPinned, _ := g.seeComputePins(BLACK)
	require.Zero(t, whitePinned)
	require.Zero(t, blackPinned)
}

func TestSEEComputePins_MissingKingReturnsZero(t *testing.T) {
	// Defensive: an unusual board with one king missing should not panic.
	// Use a board where black has no king (only valid via direct field manipulation,
	// since LoadFEN rejects this — we set up a synthetic Game).
	g := &Game{}
	g.Whites[KING] = Bitboard(1) << uint(60) // e1
	g.WhitePieces = g.Whites[KING]
	g.Occupied = g.WhitePieces

	pinned, _ := g.seeComputePins(BLACK) // black has no king
	require.Zero(t, pinned)
}

func TestSEELVA_PicksPawnFirst(t *testing.T) {
	// Black pawn on d6 attacks e5; black queen on b8 attacks e5 too.
	// LVA must return the pawn.
	g := &Game{}
	require.NoError(t, g.LoadFEN("1q2k3/8/3p4/4P3/8/8/8/4K3 w - - 0 1"))

	e5, _ := ParseSquare("e5")
	d6, _ := ParseSquare("d6")

	var pinRays [64]Bitboard
	sq, kind, ok := g.seeLeastValuableAttacker(e5, BLACK, g.Occupied, 0, &pinRays)
	require.True(t, ok)
	require.Equal(t, d6, sq)
	require.Equal(t, Piece(PAWN), kind)
}

func TestSEELVA_NoAttackerReturnsFalse(t *testing.T) {
	g := &Game{}
	require.NoError(t, g.LoadFEN("4k3/8/8/8/8/8/8/4K3 w - - 0 1"))

	e4, _ := ParseSquare("e4")
	var pinRays [64]Bitboard
	_, _, ok := g.seeLeastValuableAttacker(e4, BLACK, g.Occupied, 0, &pinRays)
	require.False(t, ok)
}

func TestSEELVA_FiltersPinnedAttacker(t *testing.T) {
	// Black knight on b6 attacks d5 but is pinned along the b-file by white
	// rook b1 against black king b8. With pin filter, no attacker; without
	// pin, knight is the LVA.
	g := &Game{}
	require.NoError(t, g.LoadFEN("1k6/8/1n6/3p4/8/8/8/1R2K3 w - - 0 1"))

	d5, _ := ParseSquare("d5")
	b6, _ := ParseSquare("b6")

	// With pin filter active.
	pinned, pinRays := g.seeComputePins(BLACK)
	require.NotZero(t, pinned&(Bitboard(1)<<uint(b6)))
	_, _, ok := g.seeLeastValuableAttacker(d5, BLACK, g.Occupied, pinned, &pinRays)
	require.False(t, ok, "pinned knight should be filtered out")

	// Without pin filter (zeroed snapshot) — knight is found.
	var empty [64]Bitboard
	sq, kind, ok := g.seeLeastValuableAttacker(d5, BLACK, g.Occupied, 0, &empty)
	require.True(t, ok)
	require.Equal(t, b6, sq)
	require.Equal(t, Piece(KNIGHT), kind)
}

func TestSEELVA_RespectsWorkingOccupancy(t *testing.T) {
	// Black rook on e7 attacks e5 only if the e6 square is empty. We give
	// it a position with a black bishop on e6 blocking, then remove e6 from
	// the working occupancy and verify the rook surfaces as the LVA.
	g := &Game{}
	require.NoError(t, g.LoadFEN("4k3/4r3/4b3/4p3/8/8/8/4K3 w - - 0 1"))

	e5, _ := ParseSquare("e5")
	e6, _ := ParseSquare("e6")
	e7, _ := ParseSquare("e7")

	var pinRays [64]Bitboard

	// Full occupancy: bishop on e6 blocks the rook from attacking e5.
	// Bishop on e6 itself attacks d5/f5 but not e5; black king on e8
	// is not adjacent to e5. LVA should be the bishop only if it
	// attacked e5 — it doesn't. So LVA returns false.
	_, _, ok := g.seeLeastValuableAttacker(e5, BLACK, g.Occupied, 0, &pinRays)
	require.False(t, ok)

	// Remove e6 from working occupancy — rook should now attack e5.
	occ := g.Occupied &^ (Bitboard(1) << uint(e6))
	sq, kind, ok := g.seeLeastValuableAttacker(e5, BLACK, occ, 0, &pinRays)
	require.True(t, ok)
	require.Equal(t, e7, sq)
	require.Equal(t, Piece(ROOK), kind)
}

type seeCase struct {
	name     string
	fen      string
	from, to string
	expected int
}

func runSEECases(t *testing.T, cases []seeCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := &Game{}
			require.NoError(t, g.LoadFEN(tc.fen))
			from, err := ParseSquare(tc.from)
			require.NoError(t, err)
			to, err := ParseSquare(tc.to)
			require.NoError(t, err)
			require.Equal(t, tc.expected, g.SEE(from, to))
		})
	}
}

func TestSEE_BasicCaptures(t *testing.T) {
	runSEECases(t, []seeCase{
		{
			name:     "rook×pawn no defender (spec #1)",
			fen:      "1k1r4/1pp4p/p7/4p3/8/P5P1/1PP4P/2K1R3 w - - 0 1",
			from:     "e1", to: "e5", expected: 100,
		},
		{
			name:     "multi-step swap (spec #2)",
			fen:      "1k1r3q/1ppn3p/p4b2/4p3/8/P2N2P1/1PP1R1BP/2K1Q3 w - - 0 1",
			from:     "d3", to: "e5", expected: -220,
		},
		{
			name:     "plain rook×pawn (spec #3)",
			fen:      "4k3/8/8/4p3/8/8/4R3/4K3 w - - 0 1",
			from:     "e2", to: "e5", expected: 100,
		},
		{
			name:     "rook×pawn then rook recaptures (spec #4)",
			fen:      "4k3/8/4r3/4p3/8/8/4R3/4K3 w - - 0 1",
			from:     "e2", to: "e5", expected: -400,
		},
		{
			name:     "honest x-ray reveal (spec #11)",
			fen:      "4k3/4r3/8/4p3/8/8/4R3/4R2K w - - 0 1",
			from:     "e2", to: "e5", expected: 100,
		},
		{
			name:     "pinned attacker filtered out (spec #12a)",
			fen:      "1k6/8/1n6/3p4/2P5/8/8/1R2K3 w - - 0 1",
			from:     "c4", to: "d5", expected: 100,
		},
		{
			name:     "pin-filter control (spec #12b)",
			fen:      "7k/8/1n6/3p4/2P5/8/8/1R2K3 w - - 0 1",
			from:     "c4", to: "d5", expected: 0,
		},
		{
			name:     "king terminates swap (spec #13)",
			fen:      "4k3/8/8/8/8/8/4p3/4K3 w - - 0 1",
			from:     "e1", to: "e2", expected: 100,
		},
	})
}

func TestSEE_EnPassant(t *testing.T) {
	runSEECases(t, []seeCase{
		{
			name:     "white pawn EP captures black pawn (spec #5)",
			fen:      "4k3/8/8/3pP3/8/8/8/4K3 w - d6 0 1",
			from:     "e5", to: "d6", expected: 100,
		},
	})
}
