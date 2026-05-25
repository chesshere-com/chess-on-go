package chessongo

import (
	"encoding/binary"
	"errors"
)

const (
	binaryMagicV1            = "COG1"
	binaryMagicV2            = "COG2"
	binaryHeaderSize         = 4
	binaryFixedPayloadSizeV1 = 79
	binaryVariantPayloadSize = 19
	binaryFixedPayloadSizeV2 = binaryFixedPayloadSizeV1 + binaryVariantPayloadSize
	binaryHistoryEntrySize   = 12
)

// MarshalBinary encodes the board state into a byte slice.
func (g *Game) MarshalBinary() ([]byte, error) {
	buf := make([]byte, binaryHeaderSize+binaryFixedPayloadSizeV2+len(g.positionHistory)*binaryHistoryEntrySize)

	copy(buf[0:binaryHeaderSize], binaryMagicV2)
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
	buf[offset] = uint8(g.variant)
	buf[offset+1] = g.variantState.checksGiven[whiteStateIndex]
	buf[offset+2] = g.variantState.checksGiven[blackStateIndex]
	offset += 3
	for i := range g.castlingRookFrom {
		buf[offset+i] = uint8(g.castlingRookFrom[i])
	}
	offset += len(g.castlingRookFrom)

	for hash, count := range g.positionHistory {
		binary.LittleEndian.PutUint64(buf[offset:offset+8], hash)
		binary.LittleEndian.PutUint32(buf[offset+8:offset+12], uint32(count))
		offset += binaryHistoryEntrySize
	}

	return buf, nil
}

// UnmarshalBinary decodes the board state from a byte slice.
func (g *Game) UnmarshalBinary(data []byte) error {
	if len(data) < binaryHeaderSize+binaryFixedPayloadSizeV1 {
		return errors.New("insufficient data for board")
	}

	magic := string(data[0:binaryHeaderSize])
	if magic != binaryMagicV1 && magic != binaryMagicV2 {
		return errors.New("invalid binary board format")
	}
	fixedPayloadSize := binaryFixedPayloadSizeV1
	if magic == binaryMagicV2 {
		fixedPayloadSize = binaryFixedPayloadSizeV2
		if len(data) < binaryHeaderSize+fixedPayloadSize {
			return errors.New("insufficient data for board")
		}
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
	if decoded.castling < 0 || decoded.castling > 0xF {
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
	expectedSize := binaryHeaderSize + fixedPayloadSize + count*binaryHistoryEntrySize
	if len(data) != expectedSize {
		return errors.New("insufficient data for position history")
	}

	offset += 15
	if magic == binaryMagicV2 {
		decoded.variant = Variant(data[offset])
		if !validVariant(decoded.variant) {
			return errors.New("unsupported variant in binary board")
		}
		decoded.variantState.checksGiven[whiteStateIndex] = data[offset+1]
		decoded.variantState.checksGiven[blackStateIndex] = data[offset+2]
		offset += 3
		for i := range decoded.castlingRookFrom {
			decoded.castlingRookFrom[i] = Square(data[offset+i])
		}
		offset += len(decoded.castlingRookFrom)
	} else {
		decoded.variant = VariantStandard
		decoded.variantState = variantState{}
		decoded.castlingRookFrom = defaultCastlingRookFrom()
	}

	if !decoded.castlingRightsMatchBoard() {
		return errors.New("invalid castling rights in binary board")
	}

	if decoded.whites[KING].NumberOfSetBits() != 1 || decoded.blacks[KING].NumberOfSetBits() != 1 {
		return errors.New("invalid king count in binary board")
	}
	if decoded.enPassant != 0 && !decoded.enPassantStateMatchesBoard() {
		return errors.New("invalid en-passant state in binary board")
	}

	decoded.positionHistory = make(map[uint64]int, count)
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
