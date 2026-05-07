package chessongo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStableAPIAccessorsReturnCurrentState(t *testing.T) {
	g := NewGame()

	require.Equal(t, Color(WHITE), g.SideToMove())
	require.Equal(t, STARTING_POSITION_FEN, g.FEN())
	require.Equal(t, Square(0), g.EnPassantSquare())
	require.Equal(t, CastlingRights(CASTLE_WKS|CASTLE_WQS|CASTLE_BKS|CASTLE_BQS), g.CastlingRights())
	require.Equal(t, 0, g.HalfMoveClock())
	require.Equal(t, 1, g.FullMoveNumber())
	require.Equal(t, GameStatusOngoing, g.Status())
}

func TestPieceAtAndBitboardsAreReadOnlyValues(t *testing.T) {
	g := NewGame()

	piece, ok := g.PieceAt(COORDS_TO_SQUARE["e1"])
	require.True(t, ok)
	require.Equal(t, Piece(W_KING), piece)

	whitePieces := g.Pieces(WHITE)
	require.NotZero(t, g.Pieces(WHITE))
	require.Equal(t, whitePieces, g.Pieces(WHITE))

	_, ok = g.PieceAt(Square(99))
	require.False(t, ok)
}

func TestBoardViewWrapsInternalsWithReadOnlyAccessors(t *testing.T) {
	g := NewGame()
	view := g.BoardView()

	piece, ok := view.PieceAt(COORDS_TO_SQUARE["e1"])
	require.True(t, ok)
	require.Equal(t, Piece(W_KING), piece)
	require.Equal(t, g.Pieces(WHITE), view.Pieces(WHITE))
	require.Equal(t, g.PiecesOfKind(WHITE, KING), view.PiecesOfKind(WHITE, KING))
	require.Equal(t, g.OccupiedSquares(), view.OccupiedSquares())

	squares := view.Squares()
	squares[COORDS_TO_SQUARE["e1"]] = EMPTY

	piece, ok = view.PieceAt(COORDS_TO_SQUARE["e1"])
	require.True(t, ok)
	require.Equal(t, Piece(W_KING), piece)
}

func TestLegalMovesIntoReusesBuffer(t *testing.T) {
	g := NewGame()
	buffer := make([]Move, 0, 64)

	moves := g.LegalMovesInto(buffer)

	require.Len(t, moves, 20)
	require.Equal(t, cap(buffer), cap(moves))
}

func TestNewGameFromFEN(t *testing.T) {
	g, err := NewGameFromFEN("4k3/8/8/8/8/8/8/4K3 w - - 0 1")
	require.NoError(t, err)

	require.Equal(t, "4k3/8/8/8/8/8/8/4K3 w - - 0 1", g.FEN())
	require.Equal(t, Color(WHITE), g.SideToMove())
}

func TestStatusValues(t *testing.T) {
	g := &Game{}
	require.NoError(t, g.LoadFEN("7k/6Q1/6K1/8/8/8/8/8 b - - 0 1"))

	require.Equal(t, GameStatusCheckmate, g.Status())
	require.True(t, g.IsTerminal())
}

func TestBoardReturnsReadOnlySnapshot(t *testing.T) {
	g := NewGame()

	board := g.Board()
	board[COORDS_TO_SQUARE["e1"]] = EMPTY

	piece, ok := g.PieceAt(COORDS_TO_SQUARE["e1"])
	require.True(t, ok)
	require.Equal(t, Piece(W_KING), piece)
}

func TestSnapshotReturnsReadOnlyGameState(t *testing.T) {
	g := NewGame()

	snapshot := g.Snapshot()
	require.Equal(t, STARTING_POSITION_FEN, snapshot.FEN)
	require.Equal(t, Color(WHITE), snapshot.SideToMove)
	require.Equal(t, GameStatusOngoing, snapshot.Status)
	require.False(t, snapshot.Terminal)
	require.Equal(t, g.PositionKey(), snapshot.PositionKey)
	require.Len(t, snapshot.LegalMoves, 20)

	snapshot.Board[COORDS_TO_SQUARE["e1"]] = EMPTY
	snapshot.LegalMoves[0] = NewMove(0, 1, EMPTY)

	piece, ok := g.PieceAt(COORDS_TO_SQUARE["e1"])
	require.True(t, ok)
	require.Equal(t, Piece(W_KING), piece)
	require.NotEqual(t, NewMove(0, 1, EMPTY), g.LegalMovesList()[0])
}

func TestDrawStatusAvoidsDirectRuleFieldAccess(t *testing.T) {
	g := &Game{}
	require.NoError(t, g.LoadFEN("4k3/8/8/8/8/8/6R1/4K3 w - - 99 1"))

	require.NoError(t, g.TryMoveUCI("g2g3"))

	status := g.DrawStatus()
	require.False(t, status.InsufficientMaterial)
	require.False(t, status.FivefoldRepetition)
	require.False(t, status.SeventyFiveMoveRule)
	require.True(t, status.CanClaimFiftyMoveRule)
	require.False(t, status.CanClaimThreefoldRepetition)
	require.True(t, g.CanClaimFiftyMoveRule())
	require.False(t, g.CanClaimThreefoldRepetition())
	require.False(t, g.IsFivefoldRepetitionDraw())
	require.False(t, g.IsSeventyFiveMoveRuleDraw())
}

func TestDrawHelpersExposeAutomaticDraws(t *testing.T) {
	g := &Game{}
	require.NoError(t, g.LoadFEN("4k3/8/8/8/8/8/6R1/4K3 w - - 149 1"))

	require.NoError(t, g.TryMoveUCI("g2g3"))

	require.True(t, g.CanClaimFiftyMoveRule())
	require.True(t, g.IsSeventyFiveMoveRuleDraw())
	require.Equal(t, GameStatusDrawSeventyFiveMoveRule, g.Status())
}

func TestPositionKeyExposesStableHashAccessor(t *testing.T) {
	g := NewGame()
	startKey := g.PositionKey()

	require.NotZero(t, startKey)

	require.NoError(t, g.TryMoveUCI("e2e4"))
	require.NotEqual(t, startKey, g.PositionKey())

	require.NoError(t, g.TryMoveUCI("e7e5"))
	require.NotEqual(t, startKey, g.PositionKey())
}

func TestCloneReturnsIndependentGame(t *testing.T) {
	g := NewGame()
	clone := g.Clone()

	require.NotSame(t, g, clone)
	require.Equal(t, g.FEN(), clone.FEN())

	require.NoError(t, clone.TryMoveUCI("e2e4"))

	require.Equal(t, STARTING_POSITION_FEN, g.FEN())
	require.NotEqual(t, g.FEN(), clone.FEN())
}

func TestCastlingRightsHelpers(t *testing.T) {
	g := NewGame()
	rights := g.CastlingRights()

	require.True(t, rights.Has(CastlingWhiteKingSide))
	require.True(t, rights.Has(CastlingWhiteQueenSide))
	require.True(t, rights.Has(CastlingBlackKingSide))
	require.True(t, rights.Has(CastlingBlackQueenSide))
	require.Equal(t, "KQkq", rights.String())
	require.Equal(t, "-", CastlingRights(0).String())
}

func TestPublicEnumStringers(t *testing.T) {
	require.Equal(t, "white", Color(WHITE).String())
	require.Equal(t, "black", Color(BLACK).String())
	require.Equal(t, "none", Color(NO_COLOR).String())
	require.Equal(t, "unknown", Color(99).String())

	require.Equal(t, "ongoing", GameStatusOngoing.String())
	require.Equal(t, "check", GameStatusCheck.String())
	require.Equal(t, "checkmate", GameStatusCheckmate.String())
	require.Equal(t, "stalemate", GameStatusStalemate.String())
	require.Equal(t, "draw_insufficient_material", GameStatusDrawInsufficientMaterial.String())
	require.Equal(t, "draw_fivefold_repetition", GameStatusDrawFivefoldRepetition.String())
	require.Equal(t, "draw_seventy_five_move_rule", GameStatusDrawSeventyFiveMoveRule.String())
	require.Equal(t, "unknown", GameStatus(99).String())
}

func TestStableConversionHelpers(t *testing.T) {
	square, err := ParseSquare("e4")
	require.NoError(t, err)
	require.True(t, square.Valid())
	require.Equal(t, "e4", square.String())

	invalid := Square(99)
	require.False(t, invalid.Valid())
	require.Equal(t, "-", invalid.String())

	piece, ok := PieceFromRune('n')
	require.True(t, ok)
	require.Equal(t, Piece(B_KNIGHT), piece)

	piece, ok = PieceFromRune('x')
	require.False(t, ok)
	require.Equal(t, Piece(EMPTY), piece)
}
