package chessongo

import (
	"errors"
	"strings"
)

// TryMoveUCI parses and applies a legal move in UCI coordinate notation.
//
// Examples: "e2e4", "e1g1", "a7a8q".
func (g *Game) TryMoveUCI(uci string) error {
	move, err := g.ParseMoveUCI(uci)
	if err != nil {
		return err
	}
	if err := g.TryMove(move); err != nil {
		if errors.Is(err, ErrIllegalMove) {
			return illegalMove(move, uci, IllegalMoveReasonNotLegal)
		}
		return err
	}
	return nil
}

// ParseMoveUCI parses a UCI move string into a move request.
func (g *Game) ParseMoveUCI(uci string) (Move, error) {
	return NewMoveFromUCI(uci)
}

// TryMoveSAN parses and applies a legal move in standard algebraic notation.
func (g *Game) TryMoveSAN(san string) error {
	move, err := g.ParseMoveSAN(san)
	if err != nil {
		return err
	}
	return g.TryMove(move)
}

// ParseMoveSAN parses SAN by matching against the current legal move list.
func (g *Game) ParseMoveSAN(san string) (Move, error) {
	target := normalizeSANInput(san)
	if target == "" {
		return 0, ErrInvalidMoveNotation
	}

	g.GenerateLegalMoves()
	for _, move := range g.legalMoves {
		if normalizeSANInput(g.GetMoveSan(move)) == target || normalizeSANInput(g.GetMoveSanWithoutSuffix(move)) == target {
			return move, nil
		}
	}
	return 0, illegalMove(0, san, IllegalMoveReasonNoMatchingLegalMove)
}

func normalizeSANInput(san string) string {
	san = strings.TrimSpace(san)
	san = strings.TrimRightFunc(san, func(r rune) bool {
		return strings.ContainsRune("+#!?", r)
	})
	if strings.HasPrefix(san, "0-0") {
		san = "O-O" + strings.TrimPrefix(san, "0-0")
	}
	return san
}
