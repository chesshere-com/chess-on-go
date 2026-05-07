package chessongo

import (
	"math/rand"
	"sync"
)

// Castling permissions
const (
	CASTLE_WKS = 1 //White king side castling  0001
	CASTLE_WQS = 2 //White queen side castling 0010
	CASTLE_BKS = 4 //Black king side castling  0100
	CASTLE_BQS = 8 //Black queen side castling 1000
)

// castling squares
const (
	W_KING_INIT_SQUARE       = 60 // e1
	B_KING_INIT_SQUARE       = 4  // e8
	WKS_KING_TO_SQUARE       = 62 // g1
	WQS_KING_TO_SQUARE       = 58 // c1
	BKS_KING_TO_SQUARE       = 6  // g8
	BQS_KING_TO_SQUARE       = 2  // c8
	WKS_ROOK_ORIGINAL_SQUARE = 63 // h1
	WQS_ROOK_ORIGINAL_SQUARE = 56 // a1
	BKS_ROOK_ORIGINAL_SQUARE = 7  // h8
	BQS_ROOK_ORIGINAL_SQUARE = 0  // a8
)

type Game struct {
	// Deprecated: use FEN, ToFEN, or Snapshot. This field is not kept current after moves.
	Fen string
	// Deprecated: use Pieces(WHITE), BoardView, or Snapshot.
	WhitePieces Bitboard
	// Deprecated: use Pieces(BLACK), BoardView, or Snapshot.
	BlackPieces Bitboard
	// _, pawns, knights, bishops, rooks, queens, king
	// Deprecated: use PiecesOfKind(WHITE, kind), BoardView, or Snapshot.
	Whites [7]Bitboard
	// Deprecated: use PiecesOfKind(BLACK, kind), BoardView, or Snapshot.
	Blacks [7]Bitboard
	// Deprecated: use OccupiedSquares, BoardView, or Snapshot.
	Occupied Bitboard
	// Deprecated: use PieceAt, BoardView, Board, or Snapshot.
	Squares [64]Piece
	// Deprecated: use EnPassantSquare or Snapshot.
	EnPassant Square
	// Deprecated: use CastlingRights or Snapshot.
	Castling int
	// Deprecated: retained for compatibility; prefer FullMoveNumber and HalfMoveClock.
	Ply int
	// Deprecated: use HalfMoveClock or Snapshot.
	HalfMoves int
	// Deprecated: use FullMoveNumber or Snapshot.
	FullMoves int
	// Deprecated: use SideToMove or Snapshot.
	Turn Color
	// Deprecated: internal generation buffer; use LegalMovesList or LegalMovesInto.
	PseudoMoves []Move
	// Deprecated: use LegalMovesList, LegalMovesInto, or Snapshot.
	LegalMoves []Move
	// Deprecated: internal repetition state.
	PositionHistory map[uint64]int
	// Deprecated: use PositionKey or Snapshot.
	ZobristHash uint64
	// Deprecated: use Status or Snapshot.
	IsCheck bool
	// Deprecated: use Status or Snapshot.
	IsCheckmate bool
	// Deprecated: misspelled compatibility field; use IsStalemate, Status, or Snapshot.
	IsStalement bool
	// Deprecated: use DrawStatus, Status, or Snapshot.
	IsMaterialDraw bool
	// Deprecated: use CanClaimThreefoldRepetition, DrawStatus, Status, or Snapshot.
	IsThreefoldRepetition bool
	// Deprecated: use CanClaimFiftyMoveRule, DrawStatus, Status, or Snapshot.
	IsFiftyMoveRule bool
	// Deprecated: use IsSeventyFiveMoveRuleDraw, DrawStatus, Status, or Snapshot.
	IsSeventyFiveMoveRule bool
	// Deprecated: use IsTerminal or Snapshot.
	IsFinished bool
	// Deprecated: internal undo stack.
	History []GameState

	pgnTags          map[string]string
	pgnMoves         []string
	pgnResult        string
	pgnStartTurn     Color
	pgnStartFullMove int
	pgnStartFEN      string
}

// IsStalemate reports whether the current side to move has no legal moves and is not in check.
func (g *Game) IsStalemate() bool {
	return g.IsStalement
}

func (g *Game) Reset() {
	g.Fen = ""
	g.WhitePieces = 0
	g.BlackPieces = 0
	for _, kind := range [6]Piece{PAWN, KNIGHT, BISHOP, ROOK, QUEEN, KING} {
		g.Whites[kind] = 0
		g.Blacks[kind] = 0
	}
	g.Occupied = 0
	g.Squares = [64]Piece{}
	g.EnPassant = 0
	g.Castling = 0
	g.Ply = 0
	g.HalfMoves = 0
	g.FullMoves = 0
	g.Turn = WHITE
	g.PseudoMoves = []Move{}
	g.LegalMoves = []Move{}
	g.PositionHistory = map[uint64]int{}
	g.ZobristHash = 0
	g.IsCheck = false
	g.IsCheckmate = false
	g.IsStalement = false
	g.IsMaterialDraw = false
	g.IsThreefoldRepetition = false
	g.IsFiftyMoveRule = false
	g.IsSeventyFiveMoveRule = false
	g.IsFinished = false
	g.History = []GameState{}
	g.pgnTags = nil
	g.pgnMoves = nil
	g.pgnResult = ""
	g.pgnStartTurn = WHITE
	g.pgnStartFullMove = 0
	g.pgnStartFEN = ""
}

func NewGame() *Game {
	g := Game{}
	g.LoadFEN(STARTING_POSITION_FEN)
	//g.LoadFEN("8/PPPPPPPP/8/8/8/8/8 w - - 0 1")
	return &g
}

func CloneGame(g *Game) Game {
	clone := Game{
		Fen:                   g.Fen,
		WhitePieces:           g.WhitePieces,
		BlackPieces:           g.BlackPieces,
		Whites:                [7]Bitboard{},
		Blacks:                [7]Bitboard{},
		Squares:               [64]Piece{},
		Occupied:              g.Occupied,
		EnPassant:             g.EnPassant,
		Castling:              g.Castling,
		Ply:                   g.Ply,
		HalfMoves:             g.HalfMoves,
		FullMoves:             g.FullMoves,
		Turn:                  g.Turn,
		PseudoMoves:           []Move{},
		LegalMoves:            []Move{},
		PositionHistory:       map[uint64]int{},
		ZobristHash:           g.ZobristHash,
		IsCheck:               g.IsCheck,
		IsCheckmate:           g.IsCheckmate,
		IsStalement:           g.IsStalement,
		IsMaterialDraw:        g.IsMaterialDraw,
		IsThreefoldRepetition: g.IsThreefoldRepetition,
		IsFiftyMoveRule:       g.IsFiftyMoveRule,
		IsSeventyFiveMoveRule: g.IsSeventyFiveMoveRule,
		IsFinished:            g.IsFinished,
		History:               make([]GameState, len(g.History)),
		pgnTags:               cloneStringMap(g.pgnTags),
		pgnMoves:              append([]string(nil), g.pgnMoves...),
		pgnResult:             g.pgnResult,
		pgnStartTurn:          g.pgnStartTurn,
		pgnStartFullMove:      g.pgnStartFullMove,
		pgnStartFEN:           g.pgnStartFEN,
	}
	copy(clone.Whites[:], g.Whites[:])
	copy(clone.Blacks[:], g.Blacks[:])
	copy(clone.Squares[:], g.Squares[:])
	copy(clone.History, g.History)
	for k, v := range g.PositionHistory {
		clone.PositionHistory[k] = v
	}
	return clone
}

func (g *Game) addPiece(piece Piece, index int) {
	g.Squares[index] = piece
	if piece == EMPTY {
		return
	}
	bit := Bitboard(0x1 << uint(index))
	kind := piece.Kind()
	switch piece.Color() {
	case WHITE:
		g.Whites[kind] |= bit
		g.WhitePieces |= bit
	case BLACK:
		g.Blacks[kind] |= bit
		g.BlackPieces |= bit
	}
	g.Occupied |= bit
}

// Get our pawns and opponent's
func (g *Game) GetPawns() (Bitboard, Bitboard) {
	if g.Turn == WHITE {
		return g.Whites[PAWN], g.Blacks[PAWN]
	}
	return g.Blacks[PAWN], g.Whites[PAWN]
}

// Get our color and opponent's
func (g *Game) GetColors() (Color, Color) {
	if g.Turn == WHITE {
		return WHITE, BLACK
	}
	return BLACK, WHITE
}

func (g *Game) hasMoves() bool {
	return len(g.LegalMoves) > 0
}

func (g *Game) ShouldIncFullMoves(m Move) bool {
	return g.Squares[m.From()].Color() == BLACK
}

func (g *Game) ShouldResetHalfMoves(m Move) bool {
	return m.GetCapturedPiece() > 0 || g.Squares[m.From()].Kind() == PAWN
}

var (
	zobristOnce       sync.Once
	zobristPiece      [12][64]uint64
	zobristCastling   [16]uint64
	zobristEnPassant  [8]uint64
	zobristTurnToMove uint64
)

func initZobrist() {
	rng := rand.New(rand.NewSource(1))
	for i := 0; i < 12; i++ {
		for j := 0; j < 64; j++ {
			zobristPiece[i][j] = rng.Uint64()
		}
	}
	for i := 0; i < 16; i++ {
		zobristCastling[i] = rng.Uint64()
	}
	for i := 0; i < 8; i++ {
		zobristEnPassant[i] = rng.Uint64()
	}
	zobristTurnToMove = rng.Uint64()
}

func ensureZobrist() {
	zobristOnce.Do(initZobrist)
}

func zobristPieceIndex(p Piece) int {
	switch p {
	case W_PAWN:
		return 0
	case W_KNIGHT:
		return 1
	case W_BISHOP:
		return 2
	case W_ROOK:
		return 3
	case W_QUEEN:
		return 4
	case W_KING:
		return 5
	case B_PAWN:
		return 6
	case B_KNIGHT:
		return 7
	case B_BISHOP:
		return 8
	case B_ROOK:
		return 9
	case B_QUEEN:
		return 10
	case B_KING:
		return 11
	default:
		return -1
	}
}

func (g *Game) computeZobrist() uint64 {
	ensureZobrist()
	h := uint64(0)
	for sq, piece := range g.Squares {
		idx := zobristPieceIndex(piece)
		if idx >= 0 {
			h ^= zobristPiece[idx][sq]
		}
	}

	h ^= zobristCastling[g.Castling&0xF]

	if g.EnPassant != 0 && g.hasLegalEnPassantCapture() {
		file := g.EnPassant.File()
		h ^= zobristEnPassant[file]
	}

	if g.Turn == BLACK {
		h ^= zobristTurnToMove
	}

	return h
}

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

func (g *Game) hasLegalEnPassantCapture() bool {
	if g.EnPassant == 0 {
		return false
	}

	ep := g.EnPassant
	if g.Turn == WHITE {
		if ep.Rank() != 2 || ep+8 > 63 || g.Squares[ep+8] != B_PAWN {
			return false
		}
		if ep.File() > 0 && ep+7 <= 63 && g.Squares[ep+7] == W_PAWN {
			return true
		}
		if ep.File() < 7 && ep+9 <= 63 && g.Squares[ep+9] == W_PAWN {
			return true
		}
		return false
	}

	if ep.Rank() != 5 || ep < 8 || g.Squares[ep-8] != W_PAWN {
		return false
	}
	if ep.File() < 7 && ep >= 7 && g.Squares[ep-7] == B_PAWN {
		return true
	}
	if ep.File() > 0 && ep >= 9 && g.Squares[ep-9] == B_PAWN {
		return true
	}
	return false
}

type GameState struct {
	CapturedPiece Piece
	Castling      int
	EnPassant     Square
	Ply           int
	HalfMoves     int
	FullMoves     int
	ZobristHash   uint64
}
