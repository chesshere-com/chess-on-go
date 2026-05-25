package chessongo

const maxGeneratedMoves = 256

// Generate all peseudo moves
func (g *Game) GeneratePseudoMoves() {
	var ours [7]Bitboard
	var oursAll Bitboard
	if g.turn == WHITE {
		ours = g.whites
		oursAll = g.whitePieces
	} else {
		ours = g.blacks
		oursAll = g.blackPieces
	}
	// Reuse underlying array capacity if available
	if cap(g.pseudoMoves) < maxGeneratedMoves {
		g.pseudoMoves = make([]Move, 0, maxGeneratedMoves)
	} else {
		g.pseudoMoves = g.pseudoMoves[:0]
	}
	g.genPawnOneStep()
	g.genPawnTwoSteps()
	g.genPawnAttacks()
	g.genFromMoves(ours[KING], oursAll, KING_ATTACKS_FROM[:])
	g.genFromMoves(ours[KNIGHT], oursAll, KNIGHT_ATTACKS_FROM[:])
	g.genSlidingMoves(ours[BISHOP]|ours[QUEEN], oursAll, bishopAttacks)
	g.genSlidingMoves(ours[ROOK]|ours[QUEEN], oursAll, rookAttacks)
	g.genCastling()
}

// Generate all legal moves
func (g *Game) GenerateLegalMoves() {
	g.GenerateLegalMovesFast()
}

// Generates King & Knight pseudo-legal moves
func (g *Game) genFromMoves(pieces, ours Bitboard, attackFrom []Bitboard) {
	for pieces > 0 {
		from := pieces.popLSB()
		targets := attackFrom[from] & ^ours
		for targets > 0 {
			to := targets.popLSB()
			g.pseudoMoves = append(g.pseudoMoves, NewMove(Square(from), Square(to), g.squares[to]))
		}
	}

}

func (g *Game) genSlidingMoves(pieces, ours Bitboard, attacksFrom func(Square, Bitboard) Bitboard) {
	for pieces > 0 {
		from := Square(pieces.popLSB())
		targets := attacksFrom(from, g.occupied) & ^ours
		for targets > 0 {
			to := targets.popLSB()
			g.pseudoMoves = append(g.pseudoMoves, NewMove(from, Square(to), g.squares[to]))
		}
	}
}

// Generate castling pseudo-legal moves
func (g *Game) genCastling() {
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
		g.pseudoMoves = append(g.pseudoMoves, NewCastlingMove(spec.kingFrom, spec.kingTo))
	}
}

// Generate Pawn-one-step-forward pseudo-legal moves
func (g *Game) genPawnOneStep() {
	var targets Bitboard
	var shift int = 8
	if g.turn == WHITE {
		targets = (g.whites[PAWN] >> 8) & ^g.occupied
	} else {
		targets = (g.blacks[PAWN] << 8) & ^g.occupied
		shift = -8
	}
	for targets > 0 {
		to := Square(targets.popLSB())
		from := Square(int(to) + shift)
		if g.IsToPromotionRank(to) {
			g.pseudoMoves = append(g.pseudoMoves, NewPromotionMove(from, to, g.squares[to], QUEEN))
			g.pseudoMoves = append(g.pseudoMoves, NewPromotionMove(from, to, g.squares[to], ROOK))
			g.pseudoMoves = append(g.pseudoMoves, NewPromotionMove(from, to, g.squares[to], KNIGHT))
			g.pseudoMoves = append(g.pseudoMoves, NewPromotionMove(from, to, g.squares[to], BISHOP))
		} else {
			g.pseudoMoves = append(g.pseudoMoves, NewMove(from, to, g.squares[to]))
		}
	}
}

// Generate Pawn-two-step-forward pseudo-legal moves
func (g *Game) genPawnTwoSteps() {
	var targets Bitboard
	var shift int
	if g.turn == WHITE {
		rank3filtered := ((g.whites[PAWN] & Bitboard(RANK2_MASK)) >> 8) &^ g.occupied
		targets = ((rank3filtered & Bitboard(RANK3_MASK)) >> 8) &^ g.occupied
		shift = 16
	} else {
		rank6filtered := ((g.blacks[PAWN] & Bitboard(RANK7_MASK)) << 8) &^ g.occupied
		targets = ((rank6filtered & Bitboard(RANK6_MASK)) << 8) &^ g.occupied
		shift = -16
	}
	for targets > 0 {
		to := targets.popLSB()
		from := int(to) + shift
		g.pseudoMoves = append(g.pseudoMoves, NewMove(Square(from), Square(to), g.squares[to]))
	}
}

// Generate pawns left and right attacks
func (g *Game) genPawnAttacks() {
	ours, _ := g.GetPawns()
	var targets Bitboard
	enPassant := Bitboard(0)
	if g.enPassant > 0 {
		enPassant = Bitboard(0x1 << uint(g.enPassant))
	}
	for _, shift := range [2]int{7, 9} {
		if g.turn == WHITE {
			if shift == 7 {
				targets = (ours & ^Bitboard(FILE_H_MASK)) >> uint(shift)
			} else {
				targets = (ours & ^Bitboard(FILE_A_MASK)) >> uint(shift)
			}
			targets &= (g.blackPieces | enPassant)
		} else {
			if shift == 7 {
				targets = (ours & ^Bitboard(FILE_A_MASK)) << uint(shift)
			} else {
				targets = (ours & ^Bitboard(FILE_H_MASK)) << uint(shift)
			}
			targets &= (g.whitePieces | enPassant)
		}
		for targets > 0 {
			to := Square(targets.popLSB())
			fromShift := shift
			if g.turn == BLACK {
				fromShift *= -1
			}
			from := Square(int(to) + fromShift)
			if g.enPassant > 0 && to == g.enPassant {
				var capturedPiece Piece
				if g.turn == WHITE {
					capturedPiece = g.squares[to+8]
				} else {
					capturedPiece = g.squares[to-8]
				}
				g.pseudoMoves = append(g.pseudoMoves, NewEnPassantMove(from, to, capturedPiece))
			} else if g.IsToPromotionRank(to) {
				g.pseudoMoves = append(g.pseudoMoves, NewPromotionMove(from, to, g.squares[to], QUEEN))
				g.pseudoMoves = append(g.pseudoMoves, NewPromotionMove(from, to, g.squares[to], ROOK))
				g.pseudoMoves = append(g.pseudoMoves, NewPromotionMove(from, to, g.squares[to], KNIGHT))
				g.pseudoMoves = append(g.pseudoMoves, NewPromotionMove(from, to, g.squares[to], BISHOP))
			} else {
				g.pseudoMoves = append(g.pseudoMoves, NewMove(from, to, g.squares[to]))
			}
		}
	}
}

func (g *Game) IsToPromotionRank(to Square) bool {
	return (g.turn == WHITE && (Bitboard(0x1<<uint(to))&Bitboard(RANK8_MASK) > 0)) || (g.turn == BLACK && (Bitboard(0x1<<uint(to))&Bitboard(RANK1_MASK) > 0))
}
