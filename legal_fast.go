package chessongo

type legalMoveInfo struct {
	king         Square
	checkers     Bitboard
	checkerCount int
	checkMask    Bitboard
	pinSquares   [8]Square
	pinMasks     [8]Bitboard
	pinCount     int
	pinned       Bitboard
}

var whitePawnAttackersTo [64]Bitboard
var blackPawnAttackersTo [64]Bitboard
var betweenLineMasks [64][64]Bitboard

func init() {
	for square := Square(0); square < 64; square++ {
		file := square.File()
		if file > 0 && square+7 <= 63 {
			whitePawnAttackersTo[square] |= Bitboard(1) << (square + 7)
		}
		if file < 7 && square+9 <= 63 {
			whitePawnAttackersTo[square] |= Bitboard(1) << (square + 9)
		}
		if file < 7 && square >= 7 {
			blackPawnAttackersTo[square] |= Bitboard(1) << (square - 7)
		}
		if file > 0 && square >= 9 {
			blackPawnAttackersTo[square] |= Bitboard(1) << (square - 9)
		}
	}
	for from := Square(0); from < 64; from++ {
		for to := Square(0); to < 64; to++ {
			betweenLineMasks[from][to] = computeBetweenLineMask(from, to)
		}
	}
}

// GenerateLegalMovesFast generates legal moves using check and pin masks.
func (g *Game) GenerateLegalMovesFast() {
	if cap(g.legalMoves) < maxGeneratedMoves {
		g.legalMoves = make([]Move, 0, maxGeneratedMoves)
	} else {
		g.legalMoves = g.legalMoves[:0]
	}
	info := g.buildLegalMoveInfo()
	g.isCheck = info.checkerCount > 0
	g.legalMoves = g.generateLegalMovesIntoWithInfo(g.legalMoves, &info)
}

func (g *Game) generateLegalMovesInto(dst []Move) []Move {
	info := g.buildLegalMoveInfo()
	return g.generateLegalMovesIntoWithInfo(dst, &info)
}

func (g *Game) generateLegalMovesArray(dst *[maxGeneratedMoves]Move) int {
	return len(g.generateLegalMovesInto(dst[:0]))
}

func (g *Game) generateLegalMovesIntoWithInfo(dst []Move, info *legalMoveInfo) []Move {
	dst = dst[:0]

	var ours *[7]Bitboard
	var oursAll, theirsAll Bitboard
	if g.turn == WHITE {
		ours = &g.whites
		oursAll = g.whitePieces
		theirsAll = g.blackPieces
	} else {
		ours = &g.blacks
		oursAll = g.blackPieces
		theirsAll = g.whitePieces
	}

	dst = g.appendLegalKingMoves(dst, ours[KING], oursAll, theirsAll, info)
	if info.checkerCount > 1 {
		return dst
	}
	if info.checkerCount == 0 && info.pinned == 0 {
		dst = g.appendPawnMovesNoFilter(dst, ours[PAWN])
		dst = g.appendJumpMovesNoFilter(dst, ours[KNIGHT], oursAll, theirsAll, KNIGHT_ATTACKS_FROM[:])
		dst = g.appendBishopMovesNoFilter(dst, ours[BISHOP]|ours[QUEEN], oursAll, theirsAll)
		dst = g.appendRookMovesNoFilter(dst, ours[ROOK]|ours[QUEEN], oursAll, theirsAll)
		dst = g.appendLegalCastlingMoves(dst, info)
		return dst
	}

	dst = g.appendLegalPawnMoves(dst, ours[PAWN], info)
	dst = g.appendLegalJumpMoves(dst, ours[KNIGHT], oursAll, theirsAll, KNIGHT_ATTACKS_FROM[:], info)
	dst = g.appendLegalBishopMoves(dst, ours[BISHOP]|ours[QUEEN], oursAll, theirsAll, info)
	dst = g.appendLegalRookMoves(dst, ours[ROOK]|ours[QUEEN], oursAll, theirsAll, info)
	if info.checkerCount == 0 {
		dst = g.appendLegalCastlingMoves(dst, info)
	}
	return dst
}

func (g *Game) buildLegalMoveInfo() legalMoveInfo {
	info := legalMoveInfo{checkMask: ^Bitboard(0)}
	var kingBB Bitboard
	var opponent Color
	if g.turn == WHITE {
		kingBB = g.whites[KING]
		opponent = BLACK
	} else {
		kingBB = g.blacks[KING]
		opponent = WHITE
	}
	if kingBB == 0 {
		return info
	}

	info.king = Square(kingBB.lsbIndex())
	info.checkers = g.attackersTo(info.king, opponent)
	info.checkerCount = info.checkers.NumberOfSetBits()
	if info.checkerCount == 1 {
		checker := Square(info.checkers.lsbIndex())
		info.checkMask = Bitboard(1) << checker
		info.checkMask |= betweenLineMasks[info.king][checker]
	}
	if info.checkerCount < 2 {
		g.fillAbsolutePinInfo(&info, g.turn, opponent)
	}
	return info
}

func (g *Game) appendLegalKingMoves(dst []Move, pieces, ours, theirs Bitboard, info *legalMoveInfo) []Move {
	for pieces > 0 {
		from := Square(pieces.popLSB())
		targets := KING_ATTACKS_FROM[from] & ^g.occupied
		for targets > 0 {
			to := Square(targets.popLSB())
			move := NewMove(from, to, EMPTY)
			if g.canKingMoveWithInfo(move, info) {
				dst = append(dst, move)
			}
		}
		targets = KING_ATTACKS_FROM[from] & theirs
		for targets > 0 {
			to := Square(targets.popLSB())
			move := NewMove(from, to, g.squares[to])
			if g.canKingMoveWithInfo(move, info) {
				dst = append(dst, move)
			}
		}
	}
	return dst
}

func (g *Game) appendLegalJumpMoves(dst []Move, pieces, ours, theirs Bitboard, attackFrom []Bitboard, info *legalMoveInfo) []Move {
	for pieces > 0 {
		from := Square(pieces.popLSB())
		targets := attackFrom[from] & ^g.occupied
		for targets > 0 {
			to := Square(targets.popLSB())
			if !canNonKingMoveTo(from, to, info) {
				continue
			}
			dst = append(dst, NewMove(from, to, EMPTY))
		}
		targets = attackFrom[from] & theirs
		for targets > 0 {
			to := Square(targets.popLSB())
			if !canNonKingMoveTo(from, to, info) {
				continue
			}
			dst = append(dst, NewMove(from, to, g.squares[to]))
		}
	}
	return dst
}

func (g *Game) appendLegalBishopMoves(dst []Move, pieces, ours, theirs Bitboard, info *legalMoveInfo) []Move {
	for pieces > 0 {
		from := Square(pieces.popLSB())
		attacks := bishopAttacks(from, g.occupied)
		targets := attacks & ^g.occupied
		for targets > 0 {
			to := Square(targets.popLSB())
			if !canNonKingMoveTo(from, to, info) {
				continue
			}
			dst = append(dst, NewMove(from, to, EMPTY))
		}
		targets = attacks & theirs
		for targets > 0 {
			to := Square(targets.popLSB())
			if !canNonKingMoveTo(from, to, info) {
				continue
			}
			dst = append(dst, NewMove(from, to, g.squares[to]))
		}
	}
	return dst
}

func (g *Game) appendLegalRookMoves(dst []Move, pieces, ours, theirs Bitboard, info *legalMoveInfo) []Move {
	for pieces > 0 {
		from := Square(pieces.popLSB())
		attacks := rookAttacks(from, g.occupied)
		targets := attacks & ^g.occupied
		for targets > 0 {
			to := Square(targets.popLSB())
			if !canNonKingMoveTo(from, to, info) {
				continue
			}
			dst = append(dst, NewMove(from, to, EMPTY))
		}
		targets = attacks & theirs
		for targets > 0 {
			to := Square(targets.popLSB())
			if !canNonKingMoveTo(from, to, info) {
				continue
			}
			dst = append(dst, NewMove(from, to, g.squares[to]))
		}
	}
	return dst
}

func (g *Game) appendJumpMovesNoFilter(dst []Move, pieces, ours, theirs Bitboard, attackFrom []Bitboard) []Move {
	for pieces > 0 {
		from := Square(pieces.popLSB())
		targets := attackFrom[from] & ^g.occupied
		for targets > 0 {
			to := Square(targets.popLSB())
			dst = append(dst, NewMove(from, to, EMPTY))
		}
		targets = attackFrom[from] & theirs
		for targets > 0 {
			to := Square(targets.popLSB())
			dst = append(dst, NewMove(from, to, g.squares[to]))
		}
	}
	return dst
}

func (g *Game) appendBishopMovesNoFilter(dst []Move, pieces, ours, theirs Bitboard) []Move {
	for pieces > 0 {
		from := Square(pieces.popLSB())
		attacks := bishopAttacks(from, g.occupied)
		targets := attacks & ^g.occupied
		for targets > 0 {
			to := Square(targets.popLSB())
			dst = append(dst, NewMove(from, to, EMPTY))
		}
		targets = attacks & theirs
		for targets > 0 {
			to := Square(targets.popLSB())
			dst = append(dst, NewMove(from, to, g.squares[to]))
		}
	}
	return dst
}

func (g *Game) appendRookMovesNoFilter(dst []Move, pieces, ours, theirs Bitboard) []Move {
	for pieces > 0 {
		from := Square(pieces.popLSB())
		attacks := rookAttacks(from, g.occupied)
		targets := attacks & ^g.occupied
		for targets > 0 {
			to := Square(targets.popLSB())
			dst = append(dst, NewMove(from, to, EMPTY))
		}
		targets = attacks & theirs
		for targets > 0 {
			to := Square(targets.popLSB())
			dst = append(dst, NewMove(from, to, g.squares[to]))
		}
	}
	return dst
}

func (g *Game) appendPawnMovesNoFilter(dst []Move, pawns Bitboard) []Move {
	if g.turn == WHITE {
		oneStep := (pawns >> 8) & ^g.occupied
		for targets := oneStep; targets > 0; {
			to := Square(targets.popLSB())
			from := to + 8
			if isPromotionTarget(WHITE, to) {
				dst = appendPromotions(dst, from, to, EMPTY)
			} else {
				dst = append(dst, NewMove(from, to, EMPTY))
			}
		}
		twoStep := ((oneStep & Bitboard(RANK3_MASK)) >> 8) & ^g.occupied
		for targets := twoStep; targets > 0; {
			to := Square(targets.popLSB())
			dst = append(dst, NewMove(to+16, to, EMPTY))
		}
		leftCaptures := ((pawns & ^Bitboard(FILE_H_MASK)) >> 7) & g.blackPieces
		rightCaptures := ((pawns & ^Bitboard(FILE_A_MASK)) >> 9) & g.blackPieces
		for targets := leftCaptures; targets > 0; {
			to := Square(targets.popLSB())
			dst = g.appendPawnCaptureNoFilter(dst, to+7, to, g.squares[to])
		}
		for targets := rightCaptures; targets > 0; {
			to := Square(targets.popLSB())
			dst = g.appendPawnCaptureNoFilter(dst, to+9, to, g.squares[to])
		}
		return g.appendLegalEnPassantMoves(dst, pawns)
	}

	oneStep := (pawns << 8) & ^g.occupied
	for targets := oneStep; targets > 0; {
		to := Square(targets.popLSB())
		from := to - 8
		if isPromotionTarget(BLACK, to) {
			dst = appendPromotions(dst, from, to, EMPTY)
		} else {
			dst = append(dst, NewMove(from, to, EMPTY))
		}
	}
	twoStep := ((oneStep & Bitboard(RANK6_MASK)) << 8) & ^g.occupied
	for targets := twoStep; targets > 0; {
		to := Square(targets.popLSB())
		dst = append(dst, NewMove(to-16, to, EMPTY))
	}
	leftCaptures := ((pawns & ^Bitboard(FILE_A_MASK)) << 7) & g.whitePieces
	rightCaptures := ((pawns & ^Bitboard(FILE_H_MASK)) << 9) & g.whitePieces
	for targets := leftCaptures; targets > 0; {
		to := Square(targets.popLSB())
		dst = g.appendPawnCaptureNoFilter(dst, to-7, to, g.squares[to])
	}
	for targets := rightCaptures; targets > 0; {
		to := Square(targets.popLSB())
		dst = g.appendPawnCaptureNoFilter(dst, to-9, to, g.squares[to])
	}
	return g.appendLegalEnPassantMoves(dst, pawns)
}

func (g *Game) appendPawnCaptureNoFilter(dst []Move, from, to Square, captured Piece) []Move {
	if isPromotionTarget(g.turn, to) {
		return appendPromotions(dst, from, to, captured)
	}
	return append(dst, NewMove(from, to, captured))
}

func (g *Game) appendLegalPawnMoves(dst []Move, pawns Bitboard, info *legalMoveInfo) []Move {
	if g.turn == WHITE {
		oneStep := (pawns >> 8) & ^g.occupied
		for targets := oneStep; targets > 0; {
			to := Square(targets.popLSB())
			from := to + 8
			dst = g.appendLegalPawnQuiet(dst, from, to, info)
		}

		rank3 := oneStep & Bitboard(RANK3_MASK)
		twoStep := (rank3 >> 8) & ^g.occupied
		for targets := twoStep; targets > 0; {
			to := Square(targets.popLSB())
			from := to + 16
			if canNonKingMoveTo(from, to, info) {
				dst = append(dst, NewMove(from, to, EMPTY))
			}
		}

		leftCaptures := ((pawns & ^Bitboard(FILE_H_MASK)) >> 7) & g.blackPieces
		rightCaptures := ((pawns & ^Bitboard(FILE_A_MASK)) >> 9) & g.blackPieces
		for targets := leftCaptures; targets > 0; {
			to := Square(targets.popLSB())
			dst = g.appendLegalPawnCapture(dst, to+7, to, g.squares[to], info)
		}
		for targets := rightCaptures; targets > 0; {
			to := Square(targets.popLSB())
			dst = g.appendLegalPawnCapture(dst, to+9, to, g.squares[to], info)
		}
		return g.appendLegalEnPassantMoves(dst, pawns)
	}

	oneStep := (pawns << 8) & ^g.occupied
	for targets := oneStep; targets > 0; {
		to := Square(targets.popLSB())
		from := to - 8
		dst = g.appendLegalPawnQuiet(dst, from, to, info)
	}

	rank6 := oneStep & Bitboard(RANK6_MASK)
	twoStep := (rank6 << 8) & ^g.occupied
	for targets := twoStep; targets > 0; {
		to := Square(targets.popLSB())
		from := to - 16
		if canNonKingMoveTo(from, to, info) {
			dst = append(dst, NewMove(from, to, EMPTY))
		}
	}

	leftCaptures := ((pawns & ^Bitboard(FILE_A_MASK)) << 7) & g.whitePieces
	rightCaptures := ((pawns & ^Bitboard(FILE_H_MASK)) << 9) & g.whitePieces
	for targets := leftCaptures; targets > 0; {
		to := Square(targets.popLSB())
		dst = g.appendLegalPawnCapture(dst, to-7, to, g.squares[to], info)
	}
	for targets := rightCaptures; targets > 0; {
		to := Square(targets.popLSB())
		dst = g.appendLegalPawnCapture(dst, to-9, to, g.squares[to], info)
	}
	return g.appendLegalEnPassantMoves(dst, pawns)
}

func (g *Game) appendLegalPawnQuiet(dst []Move, from, to Square, info *legalMoveInfo) []Move {
	if !canNonKingMoveTo(from, to, info) {
		return dst
	}
	if isPromotionTarget(g.turn, to) {
		return appendPromotions(dst, from, to, EMPTY)
	}
	return append(dst, NewMove(from, to, EMPTY))
}

func (g *Game) appendLegalPawnCapture(dst []Move, from, to Square, captured Piece, info *legalMoveInfo) []Move {
	if !canNonKingMoveTo(from, to, info) {
		return dst
	}
	if isPromotionTarget(g.turn, to) {
		return appendPromotions(dst, from, to, captured)
	}
	return append(dst, NewMove(from, to, captured))
}

func appendPromotions(dst []Move, from, to Square, captured Piece) []Move {
	dst = append(dst, NewPromotionMove(from, to, captured, QUEEN))
	dst = append(dst, NewPromotionMove(from, to, captured, ROOK))
	dst = append(dst, NewPromotionMove(from, to, captured, KNIGHT))
	dst = append(dst, NewPromotionMove(from, to, captured, BISHOP))
	return dst
}

func (g *Game) appendLegalEnPassantMoves(dst []Move, pawns Bitboard) []Move {
	if g.enPassant == 0 {
		return dst
	}

	to := g.enPassant
	if g.turn == WHITE {
		if to.File() > 0 {
			dst = g.appendLegalEnPassantFrom(dst, pawns, to+7, to, to+8)
		}
		if to.File() < 7 {
			dst = g.appendLegalEnPassantFrom(dst, pawns, to+9, to, to+8)
		}
		return dst
	}

	if to.File() < 7 {
		dst = g.appendLegalEnPassantFrom(dst, pawns, to-7, to, to-8)
	}
	if to.File() > 0 {
		dst = g.appendLegalEnPassantFrom(dst, pawns, to-9, to, to-8)
	}
	return dst
}

func (g *Game) appendLegalEnPassantFrom(dst []Move, pawns Bitboard, from, to, captured Square) []Move {
	if from > 63 || captured > 63 || (pawns&(Bitboard(1)<<from)) == 0 {
		return dst
	}
	capturedPiece := g.squares[captured]
	if capturedPiece == EMPTY {
		return dst
	}
	if g.canEnPassantMove(from, to, captured) {
		dst = append(dst, NewEnPassantMove(from, to, capturedPiece))
	}
	return dst
}

func (g *Game) canEnPassantMove(from, to, captured Square) bool {
	kingBB := g.whites[KING]
	opponent := Color(BLACK)
	if g.turn == BLACK {
		kingBB = g.blacks[KING]
		opponent = WHITE
	}
	if kingBB == 0 {
		return false
	}

	fromBB := Bitboard(1) << from
	toBB := Bitboard(1) << to
	capturedBB := Bitboard(1) << captured
	occupiedAfter := (g.occupied &^ fromBB &^ capturedBB) | toBB

	var pieces [7]Bitboard
	if opponent == WHITE {
		pieces = g.whites
	} else {
		pieces = g.blacks
	}
	pieces[PAWN] &^= capturedBB
	return !isSquareAttackedByPieces(Square(kingBB.lsbIndex()), opponent, occupiedAfter, &pieces)
}

func (g *Game) appendLegalCastlingMoves(dst []Move, info *legalMoveInfo) []Move {
	if g.castling == 0 {
		return dst
	}
	for _, right := range [...]int{CASTLE_WKS, CASTLE_WQS, CASTLE_BKS, CASTLE_BQS} {
		spec, ok := g.castlingSpec(right)
		if !ok {
			continue
		}
		if spec.color != g.turn || (g.castling&spec.right) == 0 {
			continue
		}
		if g.squares[spec.kingFrom] != spec.kingPiece || g.squares[spec.rookFrom] != spec.rookPiece {
			continue
		}
		if g.occupied&spec.emptyMask != 0 {
			continue
		}
		dst = g.appendLegalCastle(dst, NewCastlingMove(spec.kingFrom, spec.kingTo), info)
	}
	return dst
}

func (g *Game) appendLegalCastle(dst []Move, move Move, info *legalMoveInfo) []Move {
	if g.canKingMoveWithInfo(move, info) {
		dst = append(dst, move)
	}
	return dst
}

func canNonKingMoveTo(from, to Square, info *legalMoveInfo) bool {
	if info.checkerCount == 0 && info.pinned == 0 {
		return true
	}
	if (info.pinned&(Bitboard(1)<<from)) != 0 && (info.pinMaskFor(from)&(Bitboard(1)<<to)) == 0 {
		return false
	}
	return info.checkerCount == 0 || (info.checkMask&(Bitboard(1)<<to)) != 0
}

func (info *legalMoveInfo) pinMaskFor(square Square) Bitboard {
	for idx := 0; idx < info.pinCount; idx++ {
		if info.pinSquares[idx] == square {
			return info.pinMasks[idx]
		}
	}
	return 0
}

func isPromotionTarget(turn Color, to Square) bool {
	if turn == WHITE {
		return to < 8
	}
	return to >= 56
}

func (g *Game) attackersTo(square Square, attacker Color) Bitboard {
	return g.attackersToWithOccupied(square, attacker, g.occupied)
}

func (g *Game) attackersToWithOccupied(square Square, attacker Color, occupied Bitboard) Bitboard {
	var pieces *[7]Bitboard
	if attacker == WHITE {
		pieces = &g.whites
	} else {
		pieces = &g.blacks
	}

	attackers := (KNIGHT_ATTACKS_FROM[square] & pieces[KNIGHT]) | (KING_ATTACKS_FROM[square] & pieces[KING])
	attackers |= pawnAttackersTo(square, attacker, pieces[PAWN])
	attackers |= rookAttacks(square, occupied) & (pieces[ROOK] | pieces[QUEEN])
	attackers |= bishopAttacks(square, occupied) & (pieces[BISHOP] | pieces[QUEEN])

	return attackers
}

func (g *Game) isSquareAttackedByWithOccupied(square Square, attacker Color, occupied Bitboard) bool {
	var pieces *[7]Bitboard
	if attacker == WHITE {
		pieces = &g.whites
	} else {
		pieces = &g.blacks
	}
	return isSquareAttackedByPieces(square, attacker, occupied, pieces)
}

func isSquareAttackedByPieces(square Square, attacker Color, occupied Bitboard, pieces *[7]Bitboard) bool {
	if KNIGHT_ATTACKS_FROM[square]&pieces[KNIGHT] != 0 {
		return true
	}
	if KING_ATTACKS_FROM[square]&pieces[KING] != 0 {
		return true
	}
	if pawnAttackersTo(square, attacker, pieces[PAWN]) != 0 {
		return true
	}
	if rookAttacks(square, occupied)&(pieces[ROOK]|pieces[QUEEN]) != 0 {
		return true
	}
	return bishopAttacks(square, occupied)&(pieces[BISHOP]|pieces[QUEEN]) != 0
}

func (g *Game) canKingMoveWithInfo(m Move, info *legalMoveInfo) bool {
	opponent := oppositeColor(g.turn)
	fromBB := Bitboard(1) << m.From()
	toBB := Bitboard(1) << m.To()

	if m.IsCastlingMove() {
		if info.checkerCount > 0 {
			return false
		}

		spec, ok := g.castlingSpecForMove(m)
		if !ok {
			return false
		}

		return g.castlingKingPathIsSafe(spec, opponent)
	}

	occupiedAfter := (g.occupied &^ fromBB) | toBB
	return !g.isSquareAttackedByWithOccupied(m.To(), opponent, occupiedAfter)
}

func oppositeColor(color Color) Color {
	if color == WHITE {
		return BLACK
	}
	return WHITE
}

func pawnAttackersTo(square Square, attacker Color, pawns Bitboard) Bitboard {
	if attacker == WHITE {
		return whitePawnAttackersTo[square] & pawns
	}
	return blackPawnAttackersTo[square] & pawns
}

func computeBetweenLineMask(from, to Square) Bitboard {
	fromRank, fromFile := from.Rank(), from.File()
	toRank, toFile := to.Rank(), to.File()
	rankDelta := toRank - fromRank
	fileDelta := toFile - fromFile

	rankStep := sign(rankDelta)
	fileStep := sign(fileDelta)
	if rankDelta != 0 && fileDelta != 0 && abs(rankDelta) != abs(fileDelta) {
		return 0
	}
	if rankDelta == 0 && fileDelta == 0 {
		return 0
	}

	rank := fromRank + rankStep
	file := fromFile + fileStep
	var mask Bitboard
	for rank != toRank || file != toFile {
		mask |= Bitboard(1) << Square(rank*8+file)
		rank += rankStep
		file += fileStep
	}
	return mask
}

func sign(v int) int {
	if v < 0 {
		return -1
	}
	if v > 0 {
		return 1
	}
	return 0
}

func (g *Game) fillAbsolutePinInfo(info *legalMoveInfo, us, opponent Color) {
	var ours Bitboard
	var rookPinners, bishopPinners Bitboard
	if us == WHITE {
		ours = g.whitePieces
	} else {
		ours = g.blackPieces
	}
	if opponent == WHITE {
		rookPinners = g.whites[ROOK] | g.whites[QUEEN]
		bishopPinners = g.whites[BISHOP] | g.whites[QUEEN]
	} else {
		rookPinners = g.blacks[ROOK] | g.blacks[QUEEN]
		bishopPinners = g.blacks[BISHOP] | g.blacks[QUEEN]
	}

	for _, direction := range ALL_DIRECTIONS {
		orthogonal := isOrthogonalDirection(direction)
		ray := RAY_MASKS[direction][info.king]
		if orthogonal {
			if ray&rookPinners == 0 {
				continue
			}
		} else if ray&bishopPinners == 0 {
			continue
		}

		pinned, ok := nearestRayBlocker(info.king, direction, g.occupied)
		pinnedBB := Bitboard(1) << pinned
		if !ok || (pinnedBB&ours) == 0 {
			continue
		}

		pinner, ok := nearestRayBlocker(pinned, direction, g.occupied)
		if !ok {
			continue
		}
		pinnerBB := Bitboard(1) << pinner
		if (orthogonal && (pinnerBB&rookPinners) != 0) || (!orthogonal && (pinnerBB&bishopPinners) != 0) {
			info.pinSquares[info.pinCount] = pinned
			info.pinMasks[info.pinCount] = pinnedBB | pinnerBB | squaresBetweenAlong(info.king, pinner, direction)
			info.pinCount++
			info.pinned |= pinnedBB
		}
	}
}

func nearestRayBlocker(square Square, direction Direction, occupied Bitboard) (Square, bool) {
	blockers := RAY_MASKS[direction][square] & occupied
	if blockers == 0 {
		return 0, false
	}
	if DIRECTION_LSB_MSP[direction] == LSB {
		return Square(blockers.lsbIndex()), true
	}
	return Square(blockers.msbIndex()), true
}

func isOrthogonalDirection(direction Direction) bool {
	return direction == DIRECTION_N || direction == DIRECTION_S || direction == DIRECTION_E || direction == DIRECTION_W
}

func squaresBetweenAlong(from, to Square, direction Direction) Bitboard {
	return RAY_MASKS[direction][from] & ^RAY_MASKS[direction][to] & ^(Bitboard(1) << to)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
