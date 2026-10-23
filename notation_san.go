package chessongo

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
	clone.justMove(m)
	clone.turn = oppositeColor(g.turn)
	if !clone.ComputeIsCheck() {
		return pgn
	}

	if !g.moveHasLegalReply(m) {
		return pgn + "#"
	}
	return pgn + "+"
}

func (g *Game) moveHasLegalReply(m Move) bool {
	// Keep this allocation-capable path separate so formatting an ordinary
	// non-checking move needs only a stack copy of the board.
	clone := *g
	var history [1]GameState
	clone.history = history[:0]
	clone.makeMoveNoGenerate(m)
	var moves [maxGeneratedMoves]Move
	return clone.generateLegalMovesArray(&moves) != 0
}

// GetMoveSanWithoutSuffix returns the SAN notation for a move without checking for check (+) or checkmate (#).
// This prevents expensive board cloning and is sufficient for PGN parsing matching.
func (g *Game) GetMoveSanWithoutSuffix(m Move) string {
	var buffer [16]byte
	return string(g.appendMoveSAN(buffer[:0], m))
}

func (g *Game) matchesSAN(m Move, san string) bool {
	var buffer [16]byte
	return string(g.appendMoveSAN(buffer[:0], m)) == san
}

func (g *Game) appendMoveSAN(dst []byte, m Move) []byte {
	from, to := m.From(), m.To()
	if m.IsCastlingMove() {
		if to == WKS_KING_TO_SQUARE || to == BKS_KING_TO_SQUARE {
			return append(dst, "O-O"...)
		}
		return append(dst, "O-O-O"...)
	}

	kind := g.squares[from].Kind()
	if kind != PAWN {
		dst = append(dst, " PNBRQK"[kind])
		others, sameFile, sameRank := g.GetOthersOfSameKindMovingToSameTargetCounts(m)
		if others > 0 {
			switch {
			case sameFile == 0:
				dst = append(dst, byte('a'+from.File()))
			case sameRank == 0:
				dst = append(dst, byte('8'-from.Rank()))
			default:
				dst = append(dst, from.Coords()...)
			}
		}
	}
	if m.GetCapturedPiece() != EMPTY {
		if kind == PAWN {
			dst = append(dst, byte('a'+from.File()))
		}
		dst = append(dst, 'x')
	}
	dst = append(dst, to.Coords()...)
	if m.IsPromotionMove() {
		dst = append(dst, '=', " PNBRQK"[m.GetPromotionTo()])
	}
	return dst
}

func (g *Game) GetOthersOfSameKindMovingToSameTargetCounts(themove Move) (otherOfSameKind int, onSameFileCount int, onSameRankCount int) {
	movingPiece := g.squares[themove.From()]
	from := themove.From()
	to := themove.To()
	for _, m := range g.legalMoves {
		if m.From() == from || m.IsCastlingMove() || m.To() != to || g.squares[m.From()].Kind() != movingPiece.Kind() {
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
