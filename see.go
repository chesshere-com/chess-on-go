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
