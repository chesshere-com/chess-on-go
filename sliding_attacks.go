package chessongo

import (
	"fmt"
)

type slidingAttackTable struct {
	mask   Bitboard
	magic  uint64
	shift  uint
	offset int
	size   int
}

var rookAttackTables [64]slidingAttackTable
var bishopAttackTables [64]slidingAttackTable
var rookAttackData []Bitboard
var bishopAttackData []Bitboard

func init() {
	initSlidingAttackTables()
}

func initSlidingAttackTables() {
	for square := Square(0); square < 64; square++ {
		table, attacks := newSlidingAttackTable(square, ROOK_DIRECTIONS[:], rookMagics[square], len(rookAttackData))
		rookAttackTables[square] = table
		rookAttackData = append(rookAttackData, attacks...)

		table, attacks = newSlidingAttackTable(square, BISHOP_DIRECTIONS[:], bishopMagics[square], len(bishopAttackData))
		bishopAttackTables[square] = table
		bishopAttackData = append(bishopAttackData, attacks...)
	}
}

func newSlidingAttackTable(square Square, directions []Direction, magic uint64, offset int) (slidingAttackTable, []Bitboard) {
	mask := slidingRelevantOccupancy(square, directions)
	squares := bitboardSquares(mask)
	size := 1 << len(squares)
	shift := uint(64 - len(squares))
	attacks := make([]Bitboard, size)
	used := make([]bool, size)

	for idx := 0; idx < size; idx++ {
		occupied := expandOccupancy(idx, squares)
		attack := rayAttacks(square, occupied, directions)
		magicIndex := int((uint64(occupied&mask) * magic) >> shift)
		if used[magicIndex] && attacks[magicIndex] != attack {
			panic(fmt.Sprintf("invalid sliding magic for %s", square.Coords()))
		}
		used[magicIndex] = true
		attacks[magicIndex] = attack
	}

	return slidingAttackTable{
		mask:   mask,
		magic:  magic,
		shift:  shift,
		offset: offset,
		size:   size,
	}, attacks
}

func slidingRelevantOccupancy(square Square, directions []Direction) Bitboard {
	var mask Bitboard
	for _, direction := range directions {
		ray := Ray{square: square, direction: direction}
		var previous Square
		for {
			onBoard, next := ray.step()
			if !onBoard {
				break
			}
			if previous != 0 || Square(next) != square {
				mask |= Bitboard(1) << Square(next)
			}
			previous = Square(next)
		}
		// Edge blockers do not change attacks beyond the edge square.
		mask &^= Bitboard(1) << previous
	}
	return mask
}

func rookAttacks(square Square, occupied Bitboard) Bitboard {
	table := &rookAttackTables[square]
	return rookAttackData[table.offset+table.index(occupied)]
}

func bishopAttacks(square Square, occupied Bitboard) Bitboard {
	table := &bishopAttackTables[square]
	return bishopAttackData[table.offset+table.index(occupied)]
}

func (t *slidingAttackTable) index(occupied Bitboard) int {
	blockers := occupied & t.mask
	return int((uint64(blockers) * t.magic) >> t.shift)
}

func bitboardSquares(mask Bitboard) []Square {
	squares := make([]Square, 0, mask.NumberOfSetBits())
	for mask > 0 {
		squares = append(squares, Square(mask.popLSB()))
	}
	return squares
}

func expandOccupancy(index int, squares []Square) Bitboard {
	var occupied Bitboard
	for bit, square := range squares {
		if index&(1<<bit) != 0 {
			occupied |= Bitboard(1) << square
		}
	}
	return occupied
}

func rayAttacks(square Square, occupied Bitboard, directions []Direction) Bitboard {
	var attacks Bitboard
	for _, direction := range directions {
		targets := RAY_MASKS[direction][square]
		blockers := targets & occupied
		if blockers > 0 {
			if DIRECTION_LSB_MSP[direction] == LSB {
				targets ^= RAY_MASKS[direction][blockers.lsbIndex()]
			} else {
				targets ^= RAY_MASKS[direction][blockers.msbIndex()]
			}
		}
		attacks |= targets
	}
	return attacks
}

func verifyMagicTable(square Square, table *slidingAttackTable, attacks []Bitboard, directions []Direction) error {
	squares := bitboardSquares(table.mask)
	used := make([]bool, table.size)
	values := make([]Bitboard, table.size)
	for idx := 0; idx < 1<<len(squares); idx++ {
		occupied := expandOccupancy(idx, squares)
		attack := rayAttacks(square, occupied, directions)
		magicIndex := table.index(occupied)
		if magicIndex < 0 || magicIndex >= table.size {
			return fmt.Errorf("magic index out of range")
		}
		if attacks[table.offset+magicIndex] != attack {
			return fmt.Errorf("attack mismatch")
		}
		if used[magicIndex] && values[magicIndex] != attack {
			return fmt.Errorf("constructive collision mismatch")
		}
		used[magicIndex] = true
		values[magicIndex] = attack
	}
	return nil
}

var rookMagics = [64]uint64{
	0x808000234000d880,
	0x12c000100040e000,
	0x0900200040090010,
	0x0500210010000408,
	0x0980180080020400,
	0x0080040002008001,
	0xe20000a811041200,
	0x010000844121000a,
	0x60d8800880400022,
	0x0410400050002004,
	0x0408801000802000,
	0x0041002010000900,
	0x0201000410080100,
	0x1000800400020080,
	0x8200808001000200,
	0x0401003a00418100,
	0x4000848000400020,
	0x0001020040208a00,
	0x0000808010002000,
	0x0402020040082012,
	0xa500818028000400,
	0x1888818006000400,
	0x4000040082215008,
	0x1008020020408114,
	0x6400400080008020,
	0x2100400180200180,
	0x4080100080200080,
	0xc190004040080400,
	0x0290080080040080,
	0x40c4000202001008,
	0x0009000b00060004,
	0x0704802080084500,
	0x0000400086800020,
	0x0030002000c00040,
	0x0a00825002802000,
	0x0080100021000902,
	0x0070800800800402,
	0x002e00042a001158,
	0x0000800200800100,
	0x8040210042000084,
	0x0800804000208000,
	0x0038210082060040,
	0x0010002000808011,
	0x4010100008008080,
	0x8004000408008080,
	0x3c02000805020010,
	0x0440100108040002,
	0x9019000040810002,
	0x0032004024811200,
	0x4450288040010500,
	0x00a0080040100240,
	0x0010001508210100,
	0x000c008208004480,
	0x0803800200040180,
	0x2102010270088400,
	0x2204010400608200,
	0x0203410614228001,
	0x0100801020400109,
	0x0c60010020100841,
	0x2600042010010009,
	0x0201000800308407,
	0x02c2000850014402,
	0x0040010210080084,
	0x8008114c04a10082,
}

var bishopMagics = [64]uint64{
	0x0240042880810308,
	0x0a04500081010d00,
	0x00c8008400940130,
	0x0482208a00a00080,
	0x0041114001106208,
	0x4402020220000004,
	0x0000820813408400,
	0x840b808410020201,
	0x2610082004208220,
	0x0080720a08010900,
	0x0400042104010120,
	0x0802042512000098,
	0x05000410a801a080,
	0x2010009004a02001,
	0x0004008090101210,
	0x0001810048040400,
	0x0a05014204080208,
	0x282011880200aa0a,
	0x24010aa80c430201,
	0x4244009804101224,
	0x1009020820080492,
	0x0000800d08014000,
	0x40440044424a1020,
	0x0020401084040100,
	0x0408044820059011,
	0x5008140603940800,
	0x08080404a0802080,
	0x00b0082004040090,
	0x420100c419004002,
	0x6220860029004200,
	0xc8c80441070c0291,
	0x8500404101011800,
	0x0008020800112014,
	0x00880202a808081a,
	0x8102004040040909,
	0x150c400820020200,
	0x5040080200044104,
	0x0011112100a20040,
	0x0218110c00011880,
	0x0000808100008408,
	0x0508480404101000,
	0x0000440405002012,
	0x280808209000280a,
	0x00000442008a1800,
	0x0000240104003210,
	0x4150103000882040,
	0x00028a420406660a,
	0x4012008501040606,
	0x0000480808080401,
	0xc00a008401080848,
	0x0000084244500840,
	0x000c020042020000,
	0x0000004003820220,
	0x0010042014550001,
	0x2060203102008080,
	0x011f100a08810040,
	0xe010248404200200,
	0x0000044202016000,
	0x0008022b4600b044,
	0x3002a00102104400,
	0x50010880c0028208,
	0x0400000803080200,
	0x0482423208020289,
	0x00410a0a02022102,
}
