package chessongo

// Directions of movement
const (
	DIRECTION_N = iota
	DIRECTION_S
	DIRECTION_E
	DIRECTION_W
	DIRECTION_NE
	DIRECTION_NW
	DIRECTION_SE
	DIRECTION_SW
)

// Significant bit type
const (
	LSB = iota
	MSB
)

type Direction uint

// ALL_DIRECTIONS lists all ray directions.
//
// Compatibility: this low-level table is retained for older callers.
var ALL_DIRECTIONS = [8]Direction{
	DIRECTION_N,
	DIRECTION_S,
	DIRECTION_E,
	DIRECTION_W,
	DIRECTION_NE,
	DIRECTION_NW,
	DIRECTION_SE,
	DIRECTION_SW,
}

// ROOK_DIRECTIONS lists the orthogonal ray directions.
//
// Compatibility: this low-level table is retained for older callers.
var ROOK_DIRECTIONS = [4]Direction{
	DIRECTION_N,
	DIRECTION_S,
	DIRECTION_E,
	DIRECTION_W,
}

// BISHOP_DIRECTIONS lists the diagonal ray directions.
//
// Compatibility: this low-level table is retained for older callers.
var BISHOP_DIRECTIONS = [4]Direction{
	DIRECTION_NE,
	DIRECTION_NW,
	DIRECTION_SE,
	DIRECTION_SW,
}

// DIRECTION_SHIFT contains rank/file deltas for each direction.
//
// Compatibility: this low-level table is retained for older callers.
var DIRECTION_SHIFT = [8][2]int{}

// DIRECTION_LSB_MSP identifies the blocking-bit side for each direction.
//
// Compatibility: this low-level table is retained for older callers.
var DIRECTION_LSB_MSP = [8]uint{}

// KING_RANK_FILE_SHIFTS lists king move rank/file offsets.
//
// Compatibility: this low-level table is retained for older callers.
var KING_RANK_FILE_SHIFTS = [8][2]int{
	{-1, -1}, {0, -1}, {1, -1}, {-1, 0},
	{1, 0}, {-1, 1}, {0, 1}, {1, 1},
}

// KNIGHT_RANK_FILE_SHIFTS lists knight move rank/file offsets.
//
// Compatibility: this low-level table is retained for older callers.
var KNIGHT_RANK_FILE_SHIFTS = [8][2]int{
	{-2, -1}, {-2, 1}, {2, -1}, {2, 1},
	{-1, -2}, {-1, 2}, {1, -2}, {1, 2},
}

func init() {

	DIRECTION_LSB_MSP[DIRECTION_E] = MSB
	DIRECTION_LSB_MSP[DIRECTION_W] = LSB
	DIRECTION_LSB_MSP[DIRECTION_N] = LSB
	DIRECTION_LSB_MSP[DIRECTION_S] = MSB
	DIRECTION_LSB_MSP[DIRECTION_NE] = LSB
	DIRECTION_LSB_MSP[DIRECTION_NW] = LSB
	DIRECTION_LSB_MSP[DIRECTION_SE] = MSB
	DIRECTION_LSB_MSP[DIRECTION_SW] = MSB

	DIRECTION_SHIFT[DIRECTION_N] = [2]int{1 /*Rank*/, 0 /*File*/}
	DIRECTION_SHIFT[DIRECTION_S] = [2]int{-1, 0}
	DIRECTION_SHIFT[DIRECTION_E] = [2]int{0, -1}
	DIRECTION_SHIFT[DIRECTION_W] = [2]int{0, 1}
	DIRECTION_SHIFT[DIRECTION_NE] = [2]int{1, -1}
	DIRECTION_SHIFT[DIRECTION_NW] = [2]int{1, 1}
	DIRECTION_SHIFT[DIRECTION_SE] = [2]int{-1, -1}
	DIRECTION_SHIFT[DIRECTION_SW] = [2]int{-1, 1}
}
