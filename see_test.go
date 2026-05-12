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
