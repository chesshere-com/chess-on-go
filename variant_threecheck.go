package chessongo

import "fmt"

var threeCheckVariantRules = variantRules{
	variant:        VariantThreeCheck,
	name:           "threecheck",
	implemented:    true,
	afterMove:      threeCheckAfterMove,
	overrideStatus: threeCheckStatus,
	hashExtra:      threeCheckHashExtra,
	pgnVariantTag:  "Three-check",
	pgnNames:       []string{"Three-check", "ThreeCheck", "3-check", "3check"},
}

func threeCheckAfterMove(g *Game, move Move, prev *GameState) {
	if !g.ComputeIsCheck() {
		return
	}
	mover := oppositeColor(g.turn)
	idx := variantColorIndex(mover)
	if g.variantState.checksGiven[idx] < 3 {
		g.variantState.checksGiven[idx]++
	}
}

func threeCheckStatus(g *Game, status *computedStatus) {
	if g.variantState.checksGiven[whiteStateIndex] >= 3 {
		status.isFinished = true
		status.winner = WHITE
		status.status = GameStatusVariantWin
		return
	}
	if g.variantState.checksGiven[blackStateIndex] >= 3 {
		status.isFinished = true
		status.winner = BLACK
		status.status = GameStatusVariantWin
	}
}

func threeCheckHashExtra(g *Game) uint64 {
	return uint64(g.variantState.checksGiven[whiteStateIndex])<<8 |
		uint64(g.variantState.checksGiven[blackStateIndex])
}

func parseThreeCheckCounters(field string) ([2]uint8, error) {
	var counts [2]uint8
	if len(field) != 4 || field[0] != '+' || field[2] != '+' {
		return counts, fmt.Errorf("invalid three-check counter")
	}
	if field[1] < '0' || field[1] > '3' || field[3] < '0' || field[3] > '3' {
		return counts, fmt.Errorf("invalid three-check counter")
	}
	counts[whiteStateIndex] = uint8(field[1] - '0')
	counts[blackStateIndex] = uint8(field[3] - '0')
	return counts, nil
}

func (g *Game) threeCheckFENField() string {
	return fmt.Sprintf("+%d+%d",
		g.variantState.checksGiven[whiteStateIndex],
		g.variantState.checksGiven[blackStateIndex],
	)
}
