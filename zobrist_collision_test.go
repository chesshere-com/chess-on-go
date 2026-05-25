package chessongo

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestZobristMutations ensures that every component of the board state
// (turn, castling rights, en passant, and pieces) is correctly integrated
// into the Zobrist hash, such that mutating any component yields a different hash.
func TestZobristMutations(t *testing.T) {
	g := NewGame()

	// 1. Turn mutation must change the hash
	hStart := g.zobristHash
	g.turn = BLACK
	hBlack := g.computeZobrist()
	require.NotEqual(t, hStart, hBlack, "Toggling turn must change Zobrist hash")

	// 2. Castling mutation must change the hash
	g = NewGame()
	seenCastlingHashes := make(map[uint64]int)
	for castlingMask := 0; castlingMask < 16; castlingMask++ {
		g.castling = castlingMask
		hash := g.computeZobrist()
		require.NotContains(t, seenCastlingHashes, hash, "Each castling state must have a unique Zobrist contribution")
		seenCastlingHashes[hash] = castlingMask
	}

	// 3. En Passant file mutation must change the hash (when a capturer pawn is present)
	// Setup a custom state where white has a pawn on e5, and black has pawns on d5 and f5.
	g = &Game{}
	require.NoError(t, g.LoadFEN("4k3/8/8/3pPp2/8/8/8/4K3 w - - 0 1"))
	seenEpHashes := make(map[uint64]Square)
	// No en passant:
	g.enPassant = 0
	hNoEp := g.computeZobrist()

	// EP on d6 (adjacent file d):
	d6, _ := ParseSquare("d6")
	g.enPassant = d6
	hD6 := g.computeZobrist()
	require.NotEqual(t, hNoEp, hD6, "Setting en passant on d6 must change hash")
	seenEpHashes[hD6] = d6

	// EP on f6 (adjacent file f):
	f6, _ := ParseSquare("f6")
	g.enPassant = f6
	hF6 := g.computeZobrist()
	require.NotEqual(t, hNoEp, hF6, "Setting en passant on f6 must change hash")
	require.NotContains(t, seenEpHashes, hF6, "Different en passant files must yield different hashes")
	seenEpHashes[hF6] = f6

	// 4. Piece mutation must change the hash
	g = NewGame()
	hEmptySq := g.computeZobrist()
	// Place a white rook on e4 (which is currently empty)
	e4, _ := ParseSquare("e4")
	g.squares[e4] = W_ROOK
	g.whites[ROOK] |= Bitboard(1) << e4
	g.whitePieces |= Bitboard(1) << e4
	g.occupied |= Bitboard(1) << e4
	hWithRook := g.computeZobrist()
	require.NotEqual(t, hEmptySq, hWithRook, "Placing a piece must change the hash")
}

func TestZobristDistinguishesVariantCastlingMetadata(t *testing.T) {
	standard := NewGame()
	chess960, err := NewChess960Game(518)
	require.NoError(t, err)

	require.NotEqual(t, standard.PositionKey(), chess960.PositionKey())
}

func TestZobristCastlingCollisionDistinguishesChess960RookOrigin(t *testing.T) {
	g, err := NewGameFromFENWithVariant("6k1/8/8/8/8/8/8/6KR w H - 0 1", VariantChess960)
	require.NoError(t, err)

	mutated := g.Clone()
	// Same board and castling bitmask, but different Chess960 rook-origin metadata.
	mutated.castlingRookFrom[CASTLE_WKS] = COORDS_TO_SQUARE["f1"]

	require.Equal(t, g.squares, mutated.squares)
	require.Equal(t, g.castling, mutated.castling)
	require.NotEqual(t, g.computeZobrist(), mutated.computeZobrist())
}

// TestZobristCollisions generates a large set of unique positions via random game
// playouts and asserts that no two unique positions share the same Zobrist hash.
func TestZobristCollisions(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	seenPositions := make(map[uint64]string) // Zobrist hash -> FEN representation (first 4 fields)

	cleanFEN := func(fen string) string {
		parts := strings.Split(fen, " ")
		if len(parts) >= 4 {
			return strings.Join(parts[:4], " ")
		}
		return fen
	}

	for game := 0; game < 100; game++ {
		g := NewGame()
		for ply := 0; ply < 80 && len(g.legalMoves) > 0 && !g.isFinished; ply++ {
			// Save current state
			fen := cleanFEN(g.ToFEN())
			hash := g.zobristHash

			// Verify that the hash matches the computeZobrist check
			require.Equal(t, g.computeZobrist(), hash, "Incremental hash must match scratch hash")

			// Check for collisions
			if existingFen, exists := seenPositions[hash]; exists {
				require.Equal(t, existingFen, fen, "Zobrist collision detected between %q and %q for hash %x", existingFen, fen, hash)
			}
			seenPositions[hash] = fen

			// Play a random move
			move := g.legalMoves[rng.Intn(len(g.legalMoves))]
			g.MakeMove(move)
		}
	}
}
