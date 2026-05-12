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

	mover := g.Squares[from]
	moverColor := mover.Color()
	moverKind := mover.Kind()

	toBB := Bitboard(1) << uint(to)
	fromBB := Bitboard(1) << uint(from)
	promoMask := RANK1_MASK | RANK8_MASK

	var capturedValue int
	occ := g.Occupied
	isEP := moverKind == PAWN && to == g.EnPassant && g.EnPassant != 0
	if isEP {
		capturedValue = seePieceValue[PAWN]
		var capSq Square
		if moverColor == WHITE {
			capSq = to + 8
		} else {
			capSq = to - 8
		}
		occ &^= Bitboard(1) << uint(capSq)
	} else {
		capturedValue = seePieceValue[g.Squares[to].Kind()]
	}
	moverValueOnTo := seePieceValue[moverKind]
	isPromo := moverKind == PAWN && (toBB&promoMask) != 0
	if isPromo {
		capturedValue += seePieceValue[QUEEN] - seePieceValue[PAWN]
		moverValueOnTo = seePieceValue[QUEEN]
	}

	occ &^= fromBB

	whitePinned, whitePinRays := g.seeComputePins(WHITE)
	blackPinned, blackPinRays := g.seeComputePins(BLACK)

	var gain [32]int
	gain[0] = capturedValue
	d := 0
	pieceOnTo := moverValueOnTo
	side := oppositeColor(moverColor)

	for {
		var pinned Bitboard
		var pinRays *[64]Bitboard
		if side == WHITE {
			pinned = whitePinned
			pinRays = &whitePinRays
		} else {
			pinned = blackPinned
			pinRays = &blackPinRays
		}
		attackerSq, attackerKind, ok := g.seeLeastValuableAttacker(to, side, occ, pinned, pinRays)
		if !ok {
			break
		}

		d++
		gain[d] = pieceOnTo - gain[d-1]
		nextValue := seePieceValue[attackerKind]
		if attackerKind == PAWN && (toBB&promoMask) != 0 {
			gain[d] += seePieceValue[QUEEN] - seePieceValue[PAWN]
			nextValue = seePieceValue[QUEEN]
		}

		occ &^= Bitboard(1) << uint(attackerSq)

		if attackerKind == KING {
			break
		}

		pieceOnTo = nextValue
		side = oppositeColor(side)

		if d >= len(gain)-1 {
			break
		}
	}

	for d > 0 {
		if -gain[d] < gain[d-1] {
			gain[d-1] = -gain[d]
		}
		d--
	}
	return gain[0]
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

// seeLeastValuableAttacker returns the lowest-valued piece of `side` that
// attacks `to`, given the current working occupancy `occ` and a pin
// snapshot for `side`. Returns the attacker square, its piece kind, and
// true; or zero values and false if no legal attacker exists.
//
// Scans in fixed order PAWN, KNIGHT, BISHOP, ROOK, QUEEN, KING and returns
// as soon as the first non-empty (post-pin-filter) candidate set is found.
// Sliding attackers are computed against `occ` using the package's
// magic-bitboard helpers, which is what produces correct x-ray reveals as
// captured pieces are removed from `occ`.
func (g *Game) seeLeastValuableAttacker(
	to Square, side Color, occ Bitboard,
	pinned Bitboard, pinRays *[64]Bitboard,
) (Square, Piece, bool) {
	var pieces *[7]Bitboard
	if side == WHITE {
		pieces = &g.Whites
	} else {
		pieces = &g.Blacks
	}

	pawnAtts := pawnAttackersTo(to, side, pieces[PAWN]) & occ
	if sq, ok := seePickFiltered(pawnAtts, to, pinned, pinRays); ok {
		return sq, PAWN, true
	}

	knightAtts := KNIGHT_ATTACKS_FROM[to] & pieces[KNIGHT] & occ
	if sq, ok := seePickFiltered(knightAtts, to, pinned, pinRays); ok {
		return sq, KNIGHT, true
	}

	bishopAtts := bishopAttacks(to, occ) & pieces[BISHOP] & occ
	if sq, ok := seePickFiltered(bishopAtts, to, pinned, pinRays); ok {
		return sq, BISHOP, true
	}

	rookAtts := rookAttacks(to, occ) & pieces[ROOK] & occ
	if sq, ok := seePickFiltered(rookAtts, to, pinned, pinRays); ok {
		return sq, ROOK, true
	}

	queenAtts := (rookAttacks(to, occ) | bishopAttacks(to, occ)) & pieces[QUEEN] & occ
	if sq, ok := seePickFiltered(queenAtts, to, pinned, pinRays); ok {
		return sq, QUEEN, true
	}

	kingAtts := KING_ATTACKS_FROM[to] & pieces[KING] & occ
	if kingAtts != 0 {
		return Square(kingAtts.lsbIndex()), KING, true
	}

	return 0, EMPTY, false
}

// seePickFiltered iterates the candidate attacker squares and returns the
// first one that is either not pinned, or pinned along a ray that contains
// `to` (in which case the pinned piece may legally move onto `to`).
func seePickFiltered(candidates Bitboard, to Square, pinned Bitboard, pinRays *[64]Bitboard) (Square, bool) {
	toBB := Bitboard(1) << uint(to)
	for candidates != 0 {
		sq := Square(candidates.popLSB())
		bb := Bitboard(1) << uint(sq)
		if pinned&bb == 0 {
			return sq, true
		}
		if pinRays[sq]&toBB != 0 {
			return sq, true
		}
	}
	return 0, false
}
