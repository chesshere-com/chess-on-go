package chessongo

import (
	"encoding/binary"
	"errors"
)

const (
	binaryMagic            = "COG1"
	binaryHeaderSize       = 4
	binaryFixedPayloadSize = 79
	binaryHistoryEntrySize = 12
)

// MarshalBinary encodes the board state into a byte slice.
func (g *Game) MarshalBinary() ([]byte, error) {
	if g.variant != VariantStandard {
		return nil, errors.New("binary encoding is not supported for non-standard variants")
	}

	buf := make([]byte, binaryHeaderSize+binaryFixedPayloadSize+len(g.positionHistory)*binaryHistoryEntrySize)

	copy(buf[0:binaryHeaderSize], binaryMagic)
	offset := binaryHeaderSize
	for i, piece := range g.squares {
		buf[offset+i] = uint8(piece)
	}
	offset += 64
	buf[offset] = uint8(g.turn)
	buf[offset+1] = uint8(g.castling)
	buf[offset+2] = uint8(g.enPassant)
	binary.LittleEndian.PutUint32(buf[offset+3:offset+7], uint32(g.halfMoves))
	binary.LittleEndian.PutUint32(buf[offset+7:offset+11], uint32(g.fullMoves))
	binary.LittleEndian.PutUint32(buf[offset+11:offset+15], uint32(len(g.positionHistory)))

	offset += 15
	for hash, count := range g.positionHistory {
		binary.LittleEndian.PutUint64(buf[offset:offset+8], hash)
		binary.LittleEndian.PutUint32(buf[offset+8:offset+12], uint32(count))
		offset += binaryHistoryEntrySize
	}

	return buf, nil
}

// UnmarshalBinary decodes the board state from a byte slice.
func (g *Game) UnmarshalBinary(data []byte) error {
	if len(data) < binaryHeaderSize+binaryFixedPayloadSize {
		return errors.New("insufficient data for board")
	}
	if string(data[0:binaryHeaderSize]) != binaryMagic {
		return errors.New("invalid binary board format")
	}

	decoded := Game{}
	decoded.Reset()

	offset := binaryHeaderSize
	for i := 0; i < 64; i++ {
		piece := Piece(data[offset+i])
		if !isValidPiece(piece) {
			return errors.New("invalid piece in binary board")
		}
		if piece != EMPTY {
			decoded.addPiece(piece, i)
		}
	}
	offset += 64

	decoded.turn = Color(data[offset])
	if decoded.turn != WHITE && decoded.turn != BLACK {
		return errors.New("invalid turn in binary board")
	}
	decoded.castling = int(data[offset+1])
	if decoded.castling < 0 || decoded.castling > 0xF || !decoded.castlingRightsMatchBoard() {
		return errors.New("invalid castling rights in binary board")
	}
	decoded.enPassant = Square(data[offset+2])
	if decoded.enPassant > 63 {
		return errors.New("invalid en-passant square in binary board")
	}
	decoded.halfMoves = int(binary.LittleEndian.Uint32(data[offset+3 : offset+7]))
	decoded.fullMoves = int(binary.LittleEndian.Uint32(data[offset+7 : offset+11]))
	if decoded.fullMoves < 1 {
		return errors.New("invalid fullmove number in binary board")
	}

	count := int(binary.LittleEndian.Uint32(data[offset+11 : offset+15]))
	expectedSize := binaryHeaderSize + binaryFixedPayloadSize + count*binaryHistoryEntrySize
	if len(data) != expectedSize {
		return errors.New("insufficient data for position history")
	}

	if decoded.whites[KING].NumberOfSetBits() != 1 || decoded.blacks[KING].NumberOfSetBits() != 1 {
		return errors.New("invalid king count in binary board")
	}
	if decoded.enPassant != 0 && !decoded.enPassantStateMatchesBoard() {
		return errors.New("invalid en-passant state in binary board")
	}

	decoded.positionHistory = make(map[uint64]int, count)
	offset += 15
	for i := 0; i < count; i++ {
		hash := binary.LittleEndian.Uint64(data[offset : offset+8])
		c := int(binary.LittleEndian.Uint32(data[offset+8 : offset+12]))
		decoded.positionHistory[hash] = c
		offset += binaryHistoryEntrySize
	}

	// Recompute Zobrist hash for current position
	decoded.zobristHash = decoded.computeZobrist()

	// Update legal moves and check status
	decoded.GenerateLegalMoves()
	decoded.refreshStatus()
	*g = decoded

	return nil
}

func isValidPiece(p Piece) bool {
	switch p {
	case EMPTY, W_PAWN, W_KNIGHT, W_BISHOP, W_ROOK, W_QUEEN, W_KING, B_PAWN, B_KNIGHT, B_BISHOP, B_ROOK, B_QUEEN, B_KING:
		return true
	default:
		return false
	}
}
