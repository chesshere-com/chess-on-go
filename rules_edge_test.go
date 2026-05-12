package chessongo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadFENRejectsSideNotToMoveInCheck(t *testing.T) {
	tests := []string{
		"4k3/8/8/8/8/8/4R3/4K3 w - - 0 1",
		"4k3/4r3/8/8/8/8/8/4K3 b - - 0 1",
		"8/8/8/8/8/8/4k3/4K3 w - - 0 1",
	}

	for _, fen := range tests {
		t.Run(fen, func(t *testing.T) {
			g := &Game{}
			require.Error(t, g.LoadFEN(fen))
		})
	}
}

func TestLoadFENAllowsSideToMoveInCheck(t *testing.T) {
	g := &Game{}
	require.NoError(t, g.LoadFEN("4k3/4r3/8/8/8/8/8/4K3 w - - 0 1"))

	require.True(t, g.isCheck)
	require.Equal(t, GameStatusCheck, g.Status())
}

func TestThreefoldIsClaimableButFivefoldIsTerminal(t *testing.T) {
	g := NewGame()
	cycle := []string{"g1f3", "g8f6", "f3g1", "f6g8"}

	for i := 0; i < 2; i++ {
		for _, uci := range cycle {
			require.NoError(t, g.TryMoveUCI(uci))
		}
	}
	require.True(t, g.isThreefoldRepetition)
	require.False(t, g.IsFivefoldRepetition())
	require.False(t, g.IsTerminal())
	require.Equal(t, GameStatusOngoing, g.Status())

	for i := 0; i < 2; i++ {
		for _, uci := range cycle {
			require.NoError(t, g.TryMoveUCI(uci))
		}
	}
	require.True(t, g.IsFivefoldRepetition())
	require.True(t, g.IsTerminal())
	require.Equal(t, GameStatusDrawFivefoldRepetition, g.Status())
}

func TestHalfmoveClockResetsOnPawnMoveAndCapture(t *testing.T) {
	pawnMove := NewGame()
	pawnMove.halfMoves = 99
	require.NoError(t, pawnMove.TryMoveUCI("e2e4"))
	require.Equal(t, 0, pawnMove.HalfMoveClock())
	require.False(t, pawnMove.isFiftyMoveRule)

	capture := &Game{}
	require.NoError(t, capture.LoadFEN("4k3/8/8/8/3p4/8/4N3/4K3 w - - 99 1"))
	require.NoError(t, capture.TryMoveUCI("e2d4"))
	require.Equal(t, 0, capture.HalfMoveClock())
	require.False(t, capture.isFiftyMoveRule)
}

func TestSeventyFiveMoveRuleStatus(t *testing.T) {
	g := &Game{}
	require.NoError(t, g.LoadFEN("4k3/8/8/8/8/8/6R1/4K3 w - - 149 1"))

	require.NoError(t, g.TryMoveUCI("g2f2"))

	require.True(t, g.isFiftyMoveRule)
	require.True(t, g.isSeventyFiveMoveRule)
	require.True(t, g.IsTerminal())
	require.Equal(t, GameStatusDrawSeventyFiveMoveRule, g.Status())
}

func TestInsufficientMaterialEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		fen  string
		draw bool
	}{
		{"bare kings", "k7/8/8/8/8/8/8/4K3 w - - 0 1", true},
		{"king and bishop", "k7/8/8/8/8/8/8/3BK3 w - - 0 1", true},
		{"king and knight", "k7/8/8/8/8/8/8/3NK3 w - - 0 1", true},
		{"same color bishops only", "k7/8/8/8/8/2b5/8/4BK2 w - - 0 1", true},
		{"opposite color bishops", "k7/8/8/8/8/3b4/8/4BK2 w - - 0 1", false},
		{"rook is sufficient", "k7/8/8/8/8/8/8/3RK3 w - - 0 1", false},
		{"pawn is sufficient", "k7/8/8/8/8/8/3P4/4K3 w - - 0 1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Game{}
			require.NoError(t, g.LoadFEN(tt.fen))
			require.Equal(t, tt.draw, g.isMaterialDraw)
			if tt.draw {
				require.Equal(t, GameStatusDrawInsufficientMaterial, g.Status())
			}
		})
	}
}

func TestTerminalStatusCheckmateAndStalemate(t *testing.T) {
	checkmate := &Game{}
	require.NoError(t, checkmate.LoadFEN("7k/6Q1/6K1/8/8/8/8/8 b - - 0 1"))
	require.True(t, checkmate.IsTerminal())
	require.Equal(t, GameStatusCheckmate, checkmate.Status())

	stalemate := &Game{}
	require.NoError(t, stalemate.LoadFEN("7k/5Q2/6K1/8/8/8/8/8 b - - 0 1"))
	require.True(t, stalemate.IsTerminal())
	require.Equal(t, GameStatusStalemate, stalemate.Status())
}

func TestEnPassantEdgeCases(t *testing.T) {
	legal := &Game{}
	require.NoError(t, legal.LoadFEN("4k3/8/8/3pP3/8/8/8/4K3 w - d6 0 2"))
	require.Contains(t, sortedMoveUCIs(legal.legalMoves), "e5d6")

	discoveredCheck := &Game{}
	require.NoError(t, discoveredCheck.LoadFEN("k3r3/8/8/3pP3/8/8/8/4K3 w - d6 0 2"))
	require.NotContains(t, sortedMoveUCIs(discoveredCheck.legalMoves), "e5d6")

	blackLegal := &Game{}
	require.NoError(t, blackLegal.LoadFEN("4k3/8/8/8/3Pp3/8/8/4K3 b - d3 0 2"))
	require.Contains(t, sortedMoveUCIs(blackLegal.legalMoves), "e4d3")
}

func TestCastlingEdgeCases(t *testing.T) {
	white := &Game{}
	require.NoError(t, white.LoadFEN("4k3/8/8/8/8/8/8/R3K2R w KQ - 0 1"))
	require.Contains(t, sortedMoveUCIs(white.legalMoves), "e1g1")
	require.Contains(t, sortedMoveUCIs(white.legalMoves), "e1c1")

	attackedDestination := &Game{}
	require.NoError(t, attackedDestination.LoadFEN("4k3/8/8/8/8/6r1/8/R3K2R w KQ - 0 1"))
	require.NotContains(t, sortedMoveUCIs(attackedDestination.legalMoves), "e1g1")
	require.Contains(t, sortedMoveUCIs(attackedDestination.legalMoves), "e1c1")

	black := &Game{}
	require.NoError(t, black.LoadFEN("r3k2r/8/8/8/8/8/8/4K3 b kq - 0 1"))
	require.Contains(t, sortedMoveUCIs(black.legalMoves), "e8g8")
	require.Contains(t, sortedMoveUCIs(black.legalMoves), "e8c8")

	attackedRookSquare := &Game{}
	require.NoError(t, attackedRookSquare.LoadFEN("r3k3/8/8/8/8/8/8/R3K3 w Q - 0 1"))
	require.Contains(t, sortedMoveUCIs(attackedRookSquare.legalMoves), "e1c1")

	attackedB1 := &Game{}
	require.NoError(t, attackedB1.LoadFEN("1r2k3/8/8/8/8/8/8/R3K3 w Q - 0 1"))
	require.Contains(t, sortedMoveUCIs(attackedB1.legalMoves), "e1c1")
}
