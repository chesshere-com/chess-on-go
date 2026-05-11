package chessongo

const maxGeneratedMoves = 256

// Generate all peseudo moves
func (g *Game) GeneratePseudoMoves() {
	var ours [7]Bitboard
	var oursAll Bitboard
	if g.Turn == WHITE {
		ours = g.Whites
		oursAll = g.WhitePieces
	} else {
		ours = g.Blacks
		oursAll = g.BlackPieces
	}
	// Reuse underlying array capacity if available
	if cap(g.PseudoMoves) < maxGeneratedMoves {
		g.PseudoMoves = make([]Move, 0, maxGeneratedMoves)
	} else {
		g.PseudoMoves = g.PseudoMoves[:0]
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
			g.PseudoMoves = append(g.PseudoMoves, NewMove(Square(from), Square(to), g.Squares[to]))
		}
	}

}

func (g *Game) genSlidingMoves(pieces, ours Bitboard, attacksFrom func(Square, Bitboard) Bitboard) {
	for pieces > 0 {
		from := Square(pieces.popLSB())
		targets := attacksFrom(from, g.Occupied) & ^ours
		for targets > 0 {
			to := targets.popLSB()
			g.PseudoMoves = append(g.PseudoMoves, NewMove(from, Square(to), g.Squares[to]))
		}
	}
}

// Generate castling pseudo-legal moves
func (g *Game) genCastling() {
	if g.Turn == WHITE && (g.Castling&CASTLE_WKS) > 0 && g.Squares[W_KING_INIT_SQUARE] == W_KING && g.Squares[WKS_ROOK_ORIGINAL_SQUARE] == W_ROOK && (g.Occupied&(0x3<<61)) == 0 {
		from := Square(g.Whites[KING].lsbIndex())
		to := Square(WKS_KING_TO_SQUARE)
		g.PseudoMoves = append(g.PseudoMoves, NewCastlingMove(from, to))

	}

	if g.Turn == WHITE && (g.Castling&CASTLE_WQS) > 0 && g.Squares[W_KING_INIT_SQUARE] == W_KING && g.Squares[WQS_ROOK_ORIGINAL_SQUARE] == W_ROOK && (g.Occupied&(0x7<<57)) == 0 {
		from := Square(g.Whites[KING].lsbIndex())
		to := Square(WQS_KING_TO_SQUARE)
		g.PseudoMoves = append(g.PseudoMoves, NewCastlingMove(from, to))
	}

	if g.Turn == BLACK && (g.Castling&CASTLE_BKS) > 0 && g.Squares[B_KING_INIT_SQUARE] == B_KING && g.Squares[BKS_ROOK_ORIGINAL_SQUARE] == B_ROOK && (g.Occupied&(0x3<<5)) == 0 {
		from := Square(g.Blacks[KING].lsbIndex())
		to := Square(BKS_KING_TO_SQUARE)
		g.PseudoMoves = append(g.PseudoMoves, NewCastlingMove(from, to))
	}

	if g.Turn == BLACK && (g.Castling&CASTLE_BQS) > 0 && g.Squares[B_KING_INIT_SQUARE] == B_KING && g.Squares[BQS_ROOK_ORIGINAL_SQUARE] == B_ROOK && (g.Occupied&(0x7<<1)) == 0 {
		from := Square(g.Blacks[KING].lsbIndex())
		to := Square(BQS_KING_TO_SQUARE)
		g.PseudoMoves = append(g.PseudoMoves, NewCastlingMove(from, to))
	}
}

// Generate Pawn-one-step-forward pseudo-legal moves
func (g *Game) genPawnOneStep() {
	var targets Bitboard
	var shift int = 8
	if g.Turn == WHITE {
		targets = (g.Whites[PAWN] >> 8) & ^g.Occupied
	} else {
		targets = (g.Blacks[PAWN] << 8) & ^g.Occupied
		shift = -8
	}
	for targets > 0 {
		to := Square(targets.popLSB())
		from := Square(int(to) + shift)
		if g.IsToPromotionRank(to) {
			g.PseudoMoves = append(g.PseudoMoves, NewPromotionMove(from, to, g.Squares[to], QUEEN))
			g.PseudoMoves = append(g.PseudoMoves, NewPromotionMove(from, to, g.Squares[to], ROOK))
			g.PseudoMoves = append(g.PseudoMoves, NewPromotionMove(from, to, g.Squares[to], KNIGHT))
			g.PseudoMoves = append(g.PseudoMoves, NewPromotionMove(from, to, g.Squares[to], BISHOP))
		} else {
			g.PseudoMoves = append(g.PseudoMoves, NewMove(from, to, g.Squares[to]))
		}
	}
}

// Generate Pawn-two-step-forward pseudo-legal moves
func (g *Game) genPawnTwoSteps() {
	var targets Bitboard
	var shift int
	if g.Turn == WHITE {
		rank3filtered := ((g.Whites[PAWN] & Bitboard(RANK2_MASK)) >> 8) &^ g.Occupied
		targets = ((rank3filtered & Bitboard(RANK3_MASK)) >> 8) &^ g.Occupied
		shift = 16
	} else {
		rank6filtered := ((g.Blacks[PAWN] & Bitboard(RANK7_MASK)) << 8) &^ g.Occupied
		targets = ((rank6filtered & Bitboard(RANK6_MASK)) << 8) &^ g.Occupied
		shift = -16
	}
	for targets > 0 {
		to := targets.popLSB()
		from := int(to) + shift
		g.PseudoMoves = append(g.PseudoMoves, NewMove(Square(from), Square(to), g.Squares[to]))
	}
}

// Generate pawns left and right attacks
func (g *Game) genPawnAttacks() {
	ours, _ := g.GetPawns()
	var targets Bitboard
	enPassant := Bitboard(0)
	if g.EnPassant > 0 {
		enPassant = Bitboard(0x1 << uint(g.EnPassant))
	}
	for _, shift := range [2]int{7, 9} {
		if g.Turn == WHITE {
			if shift == 7 {
				targets = (ours & ^Bitboard(FILE_H_MASK)) >> uint(shift)
			} else {
				targets = (ours & ^Bitboard(FILE_A_MASK)) >> uint(shift)
			}
			targets &= (g.BlackPieces | enPassant)
		} else {
			if shift == 7 {
				targets = (ours & ^Bitboard(FILE_A_MASK)) << uint(shift)
			} else {
				targets = (ours & ^Bitboard(FILE_H_MASK)) << uint(shift)
			}
			targets &= (g.WhitePieces | enPassant)
		}
		for targets > 0 {
			to := Square(targets.popLSB())
			fromShift := shift
			if g.Turn == BLACK {
				fromShift *= -1
			}
			from := Square(int(to) + fromShift)
			if g.EnPassant > 0 && to == g.EnPassant {
				var capturedPiece Piece
				if g.Turn == WHITE {
					capturedPiece = g.Squares[to+8]
				} else {
					capturedPiece = g.Squares[to-8]
				}
				g.PseudoMoves = append(g.PseudoMoves, NewEnPassantMove(from, to, capturedPiece))
			} else if g.IsToPromotionRank(to) {
				g.PseudoMoves = append(g.PseudoMoves, NewPromotionMove(from, to, g.Squares[to], QUEEN))
				g.PseudoMoves = append(g.PseudoMoves, NewPromotionMove(from, to, g.Squares[to], ROOK))
				g.PseudoMoves = append(g.PseudoMoves, NewPromotionMove(from, to, g.Squares[to], KNIGHT))
				g.PseudoMoves = append(g.PseudoMoves, NewPromotionMove(from, to, g.Squares[to], BISHOP))
			} else {
				g.PseudoMoves = append(g.PseudoMoves, NewMove(from, to, g.Squares[to]))
			}
		}
	}
}

func (g *Game) IsToPromotionRank(to Square) bool {
	return (g.Turn == WHITE && (Bitboard(0x1<<uint(to))&Bitboard(RANK8_MASK) > 0)) || (g.Turn == BLACK && (Bitboard(0x1<<uint(to))&Bitboard(RANK1_MASK) > 0))
}
