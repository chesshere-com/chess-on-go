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
	whitePieces Bitboard
	blackPieces Bitboard
	// _, pawns, knights, bishops, rooks, queens, king
	whites                [7]Bitboard
	blacks                [7]Bitboard
	occupied              Bitboard
	squares               [64]Piece
	enPassant             Square
	castling              int
	halfMoves             int
	fullMoves             int
	turn                  Color
	variant               Variant
	castlingRookFrom      [16]Square
	pseudoMoves           []Move
	legalMoves            []Move
	positionHistory       map[uint64]int
	zobristHash           uint64
	isCheck               bool
	isCheckmate           bool
	isStalemate           bool
	isMaterialDraw        bool
	isThreefoldRepetition bool
	isFiftyMoveRule       bool
	isSeventyFiveMoveRule bool
	isFinished            bool
	history               []GameState

	pgnTags          map[string]string
	pgnMoves         []string
	pgnResult        string
	pgnStartTurn     Color
	pgnStartFullMove int
	pgnStartFEN      string
}

// IsStalemate reports whether the current side to move has no legal moves and is not in check.
func (g *Game) IsStalemate() bool {
	return g.isStalemate
}

func (g *Game) Reset() {
	g.whitePieces = 0
	g.blackPieces = 0
	for _, kind := range [6]Piece{PAWN, KNIGHT, BISHOP, ROOK, QUEEN, KING} {
		g.whites[kind] = 0
		g.blacks[kind] = 0
	}
	g.occupied = 0
	g.squares = [64]Piece{}
	g.enPassant = 0
	g.castling = 0
	g.halfMoves = 0
	g.fullMoves = 0
	g.turn = WHITE
	g.variant = VariantStandard
	g.castlingRookFrom = defaultCastlingRookFrom()
	g.pseudoMoves = []Move{}
	g.legalMoves = []Move{}
	g.positionHistory = map[uint64]int{}
	g.zobristHash = 0
	g.isCheck = false
	g.isCheckmate = false
	g.isStalemate = false
	g.isMaterialDraw = false
	g.isThreefoldRepetition = false
	g.isFiftyMoveRule = false
	g.isSeventyFiveMoveRule = false
	g.isFinished = false
	g.history = []GameState{}
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
		whitePieces:           g.whitePieces,
		blackPieces:           g.blackPieces,
		whites:                [7]Bitboard{},
		blacks:                [7]Bitboard{},
		squares:               [64]Piece{},
		occupied:              g.occupied,
		enPassant:             g.enPassant,
		castling:              g.castling,
		halfMoves:             g.halfMoves,
		fullMoves:             g.fullMoves,
		turn:                  g.turn,
		variant:               g.variant,
		castlingRookFrom:      g.castlingRookFrom,
		pseudoMoves:           []Move{},
		legalMoves:            []Move{},
		positionHistory:       map[uint64]int{},
		zobristHash:           g.zobristHash,
		isCheck:               g.isCheck,
		isCheckmate:           g.isCheckmate,
		isStalemate:           g.isStalemate,
		isMaterialDraw:        g.isMaterialDraw,
		isThreefoldRepetition: g.isThreefoldRepetition,
		isFiftyMoveRule:       g.isFiftyMoveRule,
		isSeventyFiveMoveRule: g.isSeventyFiveMoveRule,
		isFinished:            g.isFinished,
		history:               make([]GameState, len(g.history)),
		pgnTags:               cloneStringMap(g.pgnTags),
		pgnMoves:              append([]string(nil), g.pgnMoves...),
		pgnResult:             g.pgnResult,
		pgnStartTurn:          g.pgnStartTurn,
		pgnStartFullMove:      g.pgnStartFullMove,
		pgnStartFEN:           g.pgnStartFEN,
	}
	copy(clone.whites[:], g.whites[:])
	copy(clone.blacks[:], g.blacks[:])
	copy(clone.squares[:], g.squares[:])
	copy(clone.history, g.history)
	for k, v := range g.positionHistory {
		clone.positionHistory[k] = v
	}
	return clone
}

func (g *Game) addPiece(piece Piece, index int) {
	g.squares[index] = piece
	if piece == EMPTY {
		return
	}
	bit := Bitboard(0x1 << uint(index))
	kind := piece.Kind()
	switch piece.Color() {
	case WHITE:
		g.whites[kind] |= bit
		g.whitePieces |= bit
	case BLACK:
		g.blacks[kind] |= bit
		g.blackPieces |= bit
	}
	g.occupied |= bit
}

// Get our pawns and opponent's
func (g *Game) GetPawns() (Bitboard, Bitboard) {
	if g.turn == WHITE {
		return g.whites[PAWN], g.blacks[PAWN]
	}
	return g.blacks[PAWN], g.whites[PAWN]
}

// Get our color and opponent's
func (g *Game) GetColors() (Color, Color) {
	if g.turn == WHITE {
		return WHITE, BLACK
	}
	return BLACK, WHITE
}

func (g *Game) hasMoves() bool {
	return len(g.legalMoves) > 0
}

func (g *Game) ShouldIncFullMoves(m Move) bool {
	return g.squares[m.From()].Color() == BLACK
}

func (g *Game) ShouldResetHalfMoves(m Move) bool {
	return m.GetCapturedPiece() > 0 || g.squares[m.From()].Kind() == PAWN
}
