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

func (g *Game) refreshStatus() {
	g.isCheckmate = g.isCheck && !g.hasMoves()
	g.isStalemate = !g.isCheckmate && !g.hasMoves()
	g.isMaterialDraw = g.hasInsufficientMaterial()
	g.isThreefoldRepetition = g.checkThreefoldRepetition()
	g.isFiftyMoveRule = g.checkFiftyMoveRule()
	g.isSeventyFiveMoveRule = g.checkSeventyFiveMoveRule()
	g.isFinished = g.isCheckmate || g.isStalemate || g.isMaterialDraw || g.IsFivefoldRepetition() || g.isSeventyFiveMoveRule
}
