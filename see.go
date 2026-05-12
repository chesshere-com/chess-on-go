package chessongo

// see.go — static exchange evaluation (SEE).
//
// See docs/superpowers/specs/2026-05-12-game-see-design.md for the design
// contract.

var seePieceValue = [7]int{
	0,     // EMPTY
	100,   // PAWN
	320,   // KNIGHT
	330,   // BISHOP
	500,   // ROOK
	900,   // QUEEN
	20000, // KING — large enough that capturing the king never appears
	//        profitable to the opposing side in negamax.
}

// SEE returns the static exchange evaluation of capturing on `to` with the
// piece currently on `from`, in centipawns from the moving side's perspective.
// Positive values mean the moving side gains material; negative values mean it
// loses material after optimal recaptures.
//
// SEE is square-based, not move-based: it derives the moving side's color from
// the piece on `from`, not from g.SideToMove. Callers passing promotion or
// en-passant moves supply the move's from/to squares; SEE detects both from
// the source piece and engine state (g.EnPassant, the rank of `to`).
//
// SEE assumes both sides recapture with their cheapest legal attacker until
// continuing would lose material. It honors absolute pins (precomputed once
// at entry) and reveals x-ray attackers as front pieces are removed from a
// working occupancy. Mid-sequence pawn captures that reach the promotion rank
// receive the Q-P bonus; promotions are always to queen.
//
// SEE does not validate move legality: it does not check that the initial
// capture is a real legal move, nor does it filter en-passant discovered
// checks. Callers should only pass moves the engine considers legal in the
// current position.
//
// Returns 0 (no error, no panic) when either square is out of range or the
// piece on `from` is empty.
func (g *Game) SEE(from, to Square) int {
	if !from.Valid() || !to.Valid() {
		return 0
	}
	if g.Squares[from] == EMPTY {
		return 0
	}
	return 0 // TODO: implement in later tasks
}

// seeComputePins returns the absolute-pin snapshot for `side`:
//
//   - `pinned`: bitboard of squares holding pieces of `side` that are pinned
//     against their own king.
//   - `pinRays[sq]` (for each pinned square `sq`): the set of squares on the
//     ray from the king through the pinning slider, excluding the king
//     square itself. A pinned piece on `sq` may move only to squares in
//     `pinRays[sq]`.
//
// The snapshot is taken once at SEE entry and is not updated as the swap
// progresses. Returns zero values if `side` has no king on the board.
func (g *Game) seeComputePins(side Color) (Bitboard, [64]Bitboard) {
	var pinned Bitboard
	var pinRays [64]Bitboard

	var ourKing Bitboard
	var ourPieces Bitboard
	var theirRookQueen, theirBishopQueen Bitboard
	if side == WHITE {
		ourKing = g.Whites[KING]
		ourPieces = g.WhitePieces
		theirRookQueen = g.Blacks[ROOK] | g.Blacks[QUEEN]
		theirBishopQueen = g.Blacks[BISHOP] | g.Blacks[QUEEN]
	} else {
		ourKing = g.Blacks[KING]
		ourPieces = g.BlackPieces
		theirRookQueen = g.Whites[ROOK] | g.Whites[QUEEN]
		theirBishopQueen = g.Whites[BISHOP] | g.Whites[QUEEN]
	}
	if ourKing == 0 {
		return pinned, pinRays
	}
	kingSq := Square(ourKing.lsbIndex())

	scan := func(directions []Direction, sliders Bitboard) {
		for _, dir := range directions {
			ray := RAY_MASKS[dir][kingSq]
			if ray&sliders == 0 {
				continue
			}
			firstSq, ok := nearestRayBlocker(kingSq, dir, g.Occupied)
			if !ok {
				continue
			}
			firstBB := Bitboard(1) << uint(firstSq)
			if firstBB&ourPieces == 0 {
				continue
			}
			secondSq, ok := nearestRayBlocker(firstSq, dir, g.Occupied)
			if !ok {
				continue
			}
			secondBB := Bitboard(1) << uint(secondSq)
			if secondBB&sliders == 0 {
				continue
			}
			pinned |= firstBB
			pinRays[firstSq] = ray ^ RAY_MASKS[dir][secondSq]
		}
	}

	scan(ROOK_DIRECTIONS[:], theirRookQueen)
	scan(BISHOP_DIRECTIONS[:], theirBishopQueen)

	return pinned, pinRays
}
