package chessongo

import (
	"math/rand"
	"sync"
)

var (
	zobristOnce       sync.Once
	zobristPiece      [12][64]uint64
	zobristCastling   [16]uint64
	zobristEnPassant  [8]uint64
	zobristTurnToMove uint64
)

func initZobrist() {
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

func ensureZobrist() {
	zobristOnce.Do(initZobrist)
}

func zobristPieceIndex(p Piece) int {
	switch p {
	case W_PAWN:
		return 0
	case W_KNIGHT:
		return 1
	case W_BISHOP:
		return 2
	case W_ROOK:
		return 3
	case W_QUEEN:
		return 4
	case W_KING:
		return 5
	case B_PAWN:
		return 6
	case B_KNIGHT:
		return 7
	case B_BISHOP:
		return 8
	case B_ROOK:
		return 9
	case B_QUEEN:
		return 10
	case B_KING:
		return 11
	default:
		return -1
	}
}

func (g *Game) computeZobrist() uint64 {
	ensureZobrist()
	h := uint64(0)
	for sq, piece := range g.Squares {
		idx := zobristPieceIndex(piece)
		if idx >= 0 {
			h ^= zobristPiece[idx][sq]
		}
	}

	h ^= zobristCastling[g.Castling&0xF]

	if g.EnPassant != 0 && g.hasLegalEnPassantCapture() {
		file := g.EnPassant.File()
		h ^= zobristEnPassant[file]
	}

	if g.Turn == BLACK {
		h ^= zobristTurnToMove
	}

	return h
}

func (g *Game) hasLegalEnPassantCapture() bool {
	if g.EnPassant == 0 {
		return false
	}

	ep := g.EnPassant
	if g.Turn == WHITE {
		if ep.Rank() != 2 || ep+8 > 63 || g.Squares[ep+8] != B_PAWN {
			return false
		}
		if ep.File() > 0 && ep+7 <= 63 && g.Squares[ep+7] == W_PAWN {
			return true
		}
		if ep.File() < 7 && ep+9 <= 63 && g.Squares[ep+9] == W_PAWN {
			return true
		}
		return false
	}

	if ep.Rank() != 5 || ep < 8 || g.Squares[ep-8] != W_PAWN {
		return false
	}
	if ep.File() < 7 && ep >= 7 && g.Squares[ep-7] == B_PAWN {
		return true
	}
	if ep.File() > 0 && ep >= 9 && g.Squares[ep-9] == B_PAWN {
		return true
	}
	return false
}
