package chessongo

import "strings"

/*
 1. moving piece letter(exclude pawn)
    1.1 originating file letter of the moving piece
    1.2 OR: the originating rank digit of the moving piece
    1.3 OR: originating square
 2. if capturing pawn -> include originating file
 3. x for caputures
 4. destination square
 5. PawnPromotion -> "="  followed by promoted piece rune in uppercase
*/
func (g *Game) GetMoveSan(m Move) string {
	pgn := g.GetMoveSanWithoutSuffix(m)

	clone := *g
	clone.legalMoves = nil
	clone.pseudoMoves = nil
	clone.history = nil
	clone.makeMoveInternal(m, false, true)
	isCheckmate := clone.isCheck && !clone.hasMoves()
	isCheck := clone.isCheck
	if isCheckmate {
		pgn += "#"
	} else if isCheck {
		pgn += "+"
	}

	return pgn
}

// GetMoveSanWithoutSuffix returns the SAN notation for a move without checking for check (+) or checkmate (#).
// This prevents expensive board cloning and is sufficient for PGN parsing matching.
func (g *Game) GetMoveSanWithoutSuffix(m Move) string {
	var sb strings.Builder
	from := m.From()
	to := m.To()

	if m.IsCastlingMove() {
		if m.To() == WKS_KING_TO_SQUARE || m.To() == BKS_KING_TO_SQUARE {
			return "O-O"
		}
		if m.To() == WQS_KING_TO_SQUARE || m.To() == BQS_KING_TO_SQUARE {
			return "O-O-O"
		}
	} else {
		movingKind := g.squares[from].Kind()
		if movingKind != PAWN { // -------> 1.
			sb.WriteString(strings.ToUpper(string(g.squares[from].ToRune())))
		}

		// Disambiguation:
		othersOfSameKind, onSameFileCount, onSameRankCount := g.GetOthersOfSameKindMovingToSameTargetCounts(m)
		if othersOfSameKind > 0 && movingKind != PAWN {
			if onSameFileCount == 0 {
				sb.WriteString(m.From().FileLetter()) // -------> 1.1
			} else if onSameRankCount == 0 {
				sb.WriteString(m.From().RankDigit()) // -------> 1.2
			} else {
				sb.WriteString(m.From().Coords()) // -------> 1.3
			}
		}

		isCapturing := m.GetCapturedPiece() != EMPTY
		if isCapturing && movingKind == PAWN {
			sb.WriteString(from.FileLetter()) // -------> 2.
		}
		if isCapturing {
			sb.WriteString("x") // -------> 3.
		}

		sb.WriteString(to.Coords()) // -------> 4.

		if m.IsPromotionMove() {
			sb.WriteString("=")
			sb.WriteString(strings.ToUpper(string((m.GetPromotionTo() | WHITE).ToRune()))) // -------> 5.
		}
	}
	return sb.String()
}

func (g *Game) GetOthersOfSameKindMovingToSameTargetCounts(themove Move) (otherOfSameKind int, onSameFileCount int, onSameRankCount int) {
	movingPiece := g.squares[themove.From()]
	from := themove.From()
	to := themove.To()
	for _, m := range g.legalMoves {
		if m == themove || m.To() != to || g.squares[m.From()].Kind() != movingPiece.Kind() {
			continue
		}
		otherOfSameKind += 1
		if m.From().File() == from.File() {
			onSameFileCount += 1
		}
		if m.From().Rank() == from.Rank() {
			onSameRankCount += 1
		}
	}
	return
}
