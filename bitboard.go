package chessongo

import "math/bits"

/*************************************************
*	Bitboard representation
*
*	    H G F E D C B A
*                                                     first bit in rank
*	1   0 0 0 0 0 0 0 0                               <=== 56
*	2   0 0 0 0 0 0 0 0   <== WHITE                   <=== 48
*	3   0 0 0 0 0 0 0 0                               <=== 40
*	4   0 0 0 0 0 0 0 0                               <=== 32
*	5   0 0 0 0 0 0 0 0                               <=== 24
*	6   0 0 0 0 0 0 0 0                               <=== 16
*	7   0 0 0 0 0 0 0 0   <== BLACK                   <=== 8
*	8   0 0 0 0 0 0 0 0                               <=== 0
*
*	bit 0   => A8
*	bit 63  => H1
*

***************************************************/


// Bitboard
type Bitboard uint64

// Get least significant bit
func (bb Bitboard) lsb() Bitboard {
	return bb & (-bb)
}

// Get index of least significant(of Martin Läuter)
func (bb Bitboard) lsbIndex() uint {
	return uint(bits.TrailingZeros64(uint64(bb)))
}

// Get index of most significant bit(of Eugene Nalimov)
func (bb Bitboard) msbIndex() int {
	if bb == 0 {
		return -1
	}
	return 63 - bits.LeadingZeros64(uint64(bb))
}

// Pop least significant bit and return it's index
func (bb *Bitboard) popLSB() uint {
	idx := (*bb).lsbIndex()
	*bb &= *bb - 1
	return idx
}

func (b Bitboard) NumberOfSetBits() int {
	return bits.OnesCount64(uint64(b))
}
