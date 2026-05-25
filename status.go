package chessongo

func (g *Game) recordPosition() {
	if g.positionHistory == nil {
		g.positionHistory = map[uint64]int{}
	}
	g.zobristHash = g.computeZobrist()
	g.positionHistory[g.zobristHash] = g.positionHistory[g.zobristHash] + 1
}

func (g *Game) checkThreefoldRepetition() bool {
	return g.positionHistory != nil && g.positionHistory[g.zobristHash] >= 3
}

func (g *Game) IsFivefoldRepetition() bool {
	return g.positionHistory != nil && g.positionHistory[g.zobristHash] >= 5
}

func (g *Game) checkFiftyMoveRule() bool {
	return g.halfMoves >= 100
}

func (g *Game) checkSeventyFiveMoveRule() bool {
	return g.halfMoves >= 150
}

// SeedPositionHistory appends each Zobrist key in `keys` to the
// position-history counter, then refreshes the cached game status so any
// repetition draw triggered by the seeded history is reflected in Status
// and IsTerminal.
//
// Use it after constructing a Game from FEN to provide the prior position
// keys needed for threefold and fivefold repetition detection in
// applications that persist position keys per game (instead of replaying
// the full PGN on every load).
//
// The current position's key (PositionKey()) is already recorded during
// the FEN load; pass only the keys of PRIOR positions, in chronological
// order. Passing an empty or nil slice is a no-op. The halfmove clock and
// every other board field are untouched — only positionHistory and the
// derived isThreefoldRepetition / isFinished flags change.
//
// Typical usage:
//
//	g, _ := chessongo.NewGameFromFEN(currentFEN)
//	g.SeedPositionHistory(priorKeysFromDB)
//	// g.CanClaimThreefoldRepetition() / g.Status() now reflect repetitions
//	// across the entire game history, not just positions reached on *g.
func (g *Game) SeedPositionHistory(keys []uint64) {
	if len(keys) == 0 {
		return
	}
	if g.positionHistory == nil {
		g.positionHistory = map[uint64]int{}
	}
	for _, k := range keys {
		g.positionHistory[k]++
	}
	g.refreshStatus()
}

func (g *Game) refreshStatus() {
	status := computedStatus{
		isCheckmate:           g.isCheck && !g.hasMoves(),
		isMaterialDraw:        g.hasInsufficientMaterial(),
		isThreefoldRepetition: g.checkThreefoldRepetition(),
		isFiftyMoveRule:       g.checkFiftyMoveRule(),
		isSeventyFiveMoveRule: g.checkSeventyFiveMoveRule(),
		winner:                NO_COLOR,
		status:                GameStatusOngoing,
	}
	status.isStalemate = !status.isCheckmate && !g.hasMoves()
	switch {
	case status.isCheckmate:
		status.status = GameStatusCheckmate
		status.winner = oppositeColor(g.turn)
	case status.isStalemate:
		status.status = GameStatusStalemate
	case status.isMaterialDraw:
		status.status = GameStatusDrawInsufficientMaterial
	case g.IsFivefoldRepetition():
		status.status = GameStatusDrawFivefoldRepetition
	case status.isSeventyFiveMoveRule:
		status.status = GameStatusDrawSeventyFiveMoveRule
	case g.isCheck:
		status.status = GameStatusCheck
	}
	status.isFinished = status.isCheckmate ||
		status.isStalemate ||
		status.isMaterialDraw ||
		g.IsFivefoldRepetition() ||
		status.isSeventyFiveMoveRule

	if hook := g.rules().overrideStatus; hook != nil {
		hook(g, &status)
	}

	g.isCheckmate = status.isCheckmate
	g.isStalemate = status.isStalemate
	g.isMaterialDraw = status.isMaterialDraw
	g.isThreefoldRepetition = status.isThreefoldRepetition
	g.isFiftyMoveRule = status.isFiftyMoveRule
	g.isSeventyFiveMoveRule = status.isSeventyFiveMoveRule
	g.isFinished = status.isFinished
	g.winner = status.winner
	g.status = status.status
}
