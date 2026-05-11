package chessongo

func (g *Game) recordPosition() {
	if g.PositionHistory == nil {
		g.PositionHistory = map[uint64]int{}
	}
	g.ZobristHash = g.computeZobrist()
	g.PositionHistory[g.ZobristHash] = g.PositionHistory[g.ZobristHash] + 1
}

func (g *Game) checkThreefoldRepetition() bool {
	return g.PositionHistory != nil && g.PositionHistory[g.ZobristHash] >= 3
}

func (g *Game) IsFivefoldRepetition() bool {
	return g.PositionHistory != nil && g.PositionHistory[g.ZobristHash] >= 5
}

func (g *Game) checkFiftyMoveRule() bool {
	return g.HalfMoves >= 100
}

func (g *Game) checkSeventyFiveMoveRule() bool {
	return g.HalfMoves >= 150
}

func (g *Game) refreshStatus() {
	g.IsCheckmate = g.IsCheck && !g.hasMoves()
	g.IsStalement = !g.IsCheckmate && !g.hasMoves()
	g.IsMaterialDraw = g.hasInsufficientMaterial()
	g.IsThreefoldRepetition = g.checkThreefoldRepetition()
	g.IsFiftyMoveRule = g.checkFiftyMoveRule()
	g.IsSeventyFiveMoveRule = g.checkSeventyFiveMoveRule()
	g.IsFinished = g.IsCheckmate || g.IsStalement || g.IsMaterialDraw || g.IsFivefoldRepetition() || g.IsSeventyFiveMoveRule
}
