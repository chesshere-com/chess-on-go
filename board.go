package chessongo

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
