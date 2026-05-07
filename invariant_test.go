package chessongo

import "testing"

func assertGameInvariants(t *testing.T, g *Game) {
	t.Helper()

	var whitePieces, blackPieces, occupied Bitboard
	var whites [7]Bitboard
	var blacks [7]Bitboard
	whiteKings, blackKings := 0, 0

	for square, piece := range g.Squares {
		if piece == EMPTY {
			continue
		}
		if !isValidPiece(piece) {
			t.Fatalf("invalid piece %d on %s", piece, Square(square).Coords())
		}

		bit := Bitboard(1) << Square(square)
		kind := piece.Kind()
		switch piece.Color() {
		case WHITE:
			whitePieces |= bit
			whites[kind] |= bit
			if kind == KING {
				whiteKings++
			}
		case BLACK:
			blackPieces |= bit
			blacks[kind] |= bit
			if kind == KING {
				blackKings++
			}
		default:
			t.Fatalf("piece %d on %s has no color", piece, Square(square).Coords())
		}
		occupied |= bit
	}

	if whiteKings != 1 || blackKings != 1 {
		t.Fatalf("king count mismatch: white=%d black=%d", whiteKings, blackKings)
	}
	if whitePieces&blackPieces != 0 {
		t.Fatalf("white and black bitboards overlap: %064b", whitePieces&blackPieces)
	}
	if g.WhitePieces != whitePieces {
		t.Fatalf("white pieces mismatch: got %064b want %064b", g.WhitePieces, whitePieces)
	}
	if g.BlackPieces != blackPieces {
		t.Fatalf("black pieces mismatch: got %064b want %064b", g.BlackPieces, blackPieces)
	}
	if g.Occupied != occupied {
		t.Fatalf("occupied mismatch: got %064b want %064b", g.Occupied, occupied)
	}
	if g.Occupied != g.WhitePieces|g.BlackPieces {
		t.Fatalf("occupied is not union of colors")
	}
	for kind := PAWN; kind <= KING; kind++ {
		if g.Whites[kind] != whites[kind] {
			t.Fatalf("white piece kind %d bitboard mismatch", kind)
		}
		if g.Blacks[kind] != blacks[kind] {
			t.Fatalf("black piece kind %d bitboard mismatch", kind)
		}
	}
	if g.ZobristHash != g.computeZobrist() {
		t.Fatalf("zobrist mismatch: got %x want %x", g.ZobristHash, g.computeZobrist())
	}
}
