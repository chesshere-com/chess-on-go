package chessongo

var kingOfTheHillVariantRules = variantRules{
	variant:        VariantKingOfTheHill,
	name:           "kingofthehill",
	implemented:    true,
	overrideStatus: kingOfTheHillStatus,
	pgnVariantTag:  "King of the Hill",
	pgnNames:       []string{"King of the Hill", "KingOfTheHill", "kingofthehill"},
}

var kingOfTheHillCenter = (Bitboard(1) << CoordsToSquare(3, 3)) |
	(Bitboard(1) << CoordsToSquare(3, 4)) |
	(Bitboard(1) << CoordsToSquare(4, 3)) |
	(Bitboard(1) << CoordsToSquare(4, 4))

func kingOfTheHillStatus(g *Game, status *computedStatus) {
	if g.whites[KING]&kingOfTheHillCenter != 0 {
		status.isFinished = true
		status.winner = WHITE
		status.status = GameStatusVariantWin
		return
	}
	if g.blacks[KING]&kingOfTheHillCenter != 0 {
		status.isFinished = true
		status.winner = BLACK
		status.status = GameStatusVariantWin
	}
}
