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

// MS1BTABLE is a historical most-significant-bit lookup table.
//
// Compatibility: retained for older callers; new code should prefer Bitboard
// methods or math/bits helpers.
var MS1BTABLE = [256]int{}

// LS1BTABLE is a historical least-significant-bit lookup table.
//
// Compatibility: retained for older callers; new code should prefer Bitboard
// methods or math/bits helpers.
var LS1BTABLE = [64]uint{
	0, 1, 48, 2, 57, 49, 28, 3,
	61, 58, 50, 42, 38, 29, 17, 4,
	62, 55, 59, 36, 53, 51, 43, 22,
	45, 39, 33, 30, 24, 18, 12, 5,
	63, 47, 56, 27, 60, 41, 37, 16,
	54, 35, 52, 21, 44, 32, 23, 11,
	46, 26, 40, 15, 34, 20, 31, 10,
	25, 14, 19, 9, 13, 8, 7, 6,
}

// initilize bitboards
func init() {
	initMostSignificatBit()
}

// Initialze most significant bit lookup table
func initMostSignificatBit() {
	for i := 0; i < 256; i++ {
		if i > 127 {
			MS1BTABLE[i] = 7
		} else if i > 63 {
			MS1BTABLE[i] = 6
		} else if i > 31 {
			MS1BTABLE[i] = 5
		} else if i > 15 {
			MS1BTABLE[i] = 4
		} else if i > 7 {
			MS1BTABLE[i] = 3
		} else if i > 3 {
			MS1BTABLE[i] = 2
		} else if i > 2 {
			MS1BTABLE[i] = 2
		} else if i > 1 {
			MS1BTABLE[i] = 1
		} else {
			MS1BTABLE[i] = 0
		}
	}
}

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
