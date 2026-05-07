package chessongo

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSlidingAttackTablesMatchRayLogic(t *testing.T) {
	occupancies := []Bitboard{
		0,
		^Bitboard(0),
		Bitboard(0x0000001818000000),
		Bitboard(0x8142241818244281),
		Bitboard(0x00ff00000000ff00),
	}

	rng := rand.New(rand.NewSource(1))
	for i := 0; i < 128; i++ {
		occupancies = append(occupancies, Bitboard(rng.Uint64()))
	}

	for square := Square(0); square < 64; square++ {
		for _, occupied := range occupancies {
			require.Equalf(t, rayAttacksForTest(square, occupied, ROOK_DIRECTIONS[:]), rookAttacks(square, occupied), "rook %s occupied %064b", square.Coords(), occupied)
			require.Equalf(t, rayAttacksForTest(square, occupied, BISHOP_DIRECTIONS[:]), bishopAttacks(square, occupied), "bishop %s occupied %064b", square.Coords(), occupied)
		}
	}
}

func TestSlidingAttackMagicIndexesAreValid(t *testing.T) {
	for square := Square(0); square < 64; square++ {
		require.NoError(t, verifyMagicTable(square, &rookAttackTables[square], rookAttackData, ROOK_DIRECTIONS[:]), "rook %s", square.Coords())
		require.NoError(t, verifyMagicTable(square, &bishopAttackTables[square], bishopAttackData, BISHOP_DIRECTIONS[:]), "bishop %s", square.Coords())
	}
}

func rayAttacksForTest(square Square, occupied Bitboard, directions []Direction) Bitboard {
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
