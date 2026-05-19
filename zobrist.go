package chessongo

import (
	"math/rand"
)

var (
	zobristPieceIndexTable [24]int
	zobristPiece           [12][64]uint64
	zobristCastling        [16]uint64
	zobristEnPassant       [8]uint64
	zobristTurnToMove      uint64
)

func init() {
	for i := range zobristPieceIndexTable {
		zobristPieceIndexTable[i] = -1
	}
	zobristPieceIndexTable[W_PAWN] = 0
	zobristPieceIndexTable[W_KNIGHT] = 1
	zobristPieceIndexTable[W_BISHOP] = 2
	zobristPieceIndexTable[W_ROOK] = 3
	zobristPieceIndexTable[W_QUEEN] = 4
	zobristPieceIndexTable[W_KING] = 5
	zobristPieceIndexTable[B_PAWN] = 6
	zobristPieceIndexTable[B_KNIGHT] = 7
	zobristPieceIndexTable[B_BISHOP] = 8
	zobristPieceIndexTable[B_ROOK] = 9
	zobristPieceIndexTable[B_QUEEN] = 10
	zobristPieceIndexTable[B_KING] = 11

	rng := rand.New(rand.NewSource(1))
	for i := 0; i < 12; i++ {
		for j := 0; j < 64; j++ {
			zobristPiece[i][j] = rng.Uint64()
		}
	}
	for i := 0; i < 16; i++ {
		zobristCastling[i] = rng.Uint64()
	}
	for i := 0; i < 8; i++ {
		zobristEnPassant[i] = rng.Uint64()
	}
	zobristTurnToMove = rng.Uint64()
}

func zobristPieceIndex(p Piece) int {
	if p >= 24 {
		return -1
	}
	return zobristPieceIndexTable[p]
}

func (g *Game) computeZobrist() uint64 {
	h := uint64(0)
	for sq, piece := range g.squares {
		idx := zobristPieceIndex(piece)
		if idx >= 0 {
			h ^= zobristPiece[idx][sq]
		}
	}

	h ^= zobristCastling[g.castling&0xF]

	if g.enPassant != 0 && g.hasAdjacentPawnForEnPassant() {
		file := g.enPassant.File()
		h ^= zobristEnPassant[file]
	}

	if g.turn == BLACK {
		h ^= zobristTurnToMove
	}

	return h
}

func (g *Game) hasAdjacentPawnForEnPassant() bool {
	if g.enPassant == 0 {
		return false
	}

	ep := g.enPassant
	if g.turn == WHITE {
		if ep.Rank() != 2 || ep+8 > 63 || g.squares[ep+8] != B_PAWN {
			return false
		}
		if ep.File() > 0 && ep+7 <= 63 && g.squares[ep+7] == W_PAWN {
			return true
		}
		if ep.File() < 7 && ep+9 <= 63 && g.squares[ep+9] == W_PAWN {
			return true
		}
		return false
	}

	if ep.Rank() != 5 || ep < 8 || g.squares[ep-8] != W_PAWN {
		return false
	}
	if ep.File() < 7 && ep >= 7 && g.squares[ep-7] == B_PAWN {
		return true
	}
	if ep.File() > 0 && ep >= 9 && g.squares[ep-9] == B_PAWN {
		return true
	}
	return false
}
