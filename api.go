package chessongo

// CastlingRights is a bitset of available castling rights.
type CastlingRights uint8

// GameStatus describes the current terminal/check state of a game.
type GameStatus uint8

// BoardView wraps board internals with read-only accessor methods.
type BoardView struct {
	squares     [64]Piece
	whitePieces Bitboard
	blackPieces Bitboard
	whites      [7]Bitboard
	blacks      [7]Bitboard
	occupied    Bitboard
}

// DrawStatus describes claimable and automatic draw state.
type DrawStatus struct {
	InsufficientMaterial        bool
	CanClaimThreefoldRepetition bool
	CanClaimFiftyMoveRule       bool
	FivefoldRepetition          bool
	SeventyFiveMoveRule         bool
}

// GameSnapshot is a defensive value copy of the public game state.
type GameSnapshot struct {
	FEN            string
	SideToMove     Color
	Board          [64]Piece
	WhitePieces    Bitboard
	BlackPieces    Bitboard
	WhiteByKind    [7]Bitboard
	BlackByKind    [7]Bitboard
	Occupied       Bitboard
	EnPassant      Square
	Castling       CastlingRights
	HalfMoveClock  int
	FullMoveNumber int
	PositionKey    uint64
	LegalMoves     []Move
	Status         GameStatus
	Terminal       bool
	Draw           DrawStatus
}

const (
	// CastlingWhiteKingSide allows white to castle king side.
	CastlingWhiteKingSide CastlingRights = CASTLE_WKS
	// CastlingWhiteQueenSide allows white to castle queen side.
	CastlingWhiteQueenSide CastlingRights = CASTLE_WQS
	// CastlingBlackKingSide allows black to castle king side.
	CastlingBlackKingSide CastlingRights = CASTLE_BKS
	// CastlingBlackQueenSide allows black to castle queen side.
	CastlingBlackQueenSide CastlingRights = CASTLE_BQS
)

const (
	GameStatusOngoing GameStatus = iota
	GameStatusCheck
	GameStatusCheckmate
	GameStatusStalemate
	GameStatusDrawInsufficientMaterial
	GameStatusDrawFivefoldRepetition
	GameStatusDrawSeventyFiveMoveRule
)

// NewGameFromFEN creates a game initialized from FEN.
func NewGameFromFEN(fen string) (*Game, error) {
	g := &Game{}
	if err := g.LoadFEN(fen); err != nil {
		return nil, err
	}
	return g, nil
}

// Clone returns a deep copy of the game.
func (g *Game) Clone() *Game {
	clone := CloneGame(g)
	return &clone
}

// FEN returns the current FEN string.
func (g *Game) FEN() string {
	return g.ToFEN()
}

// SideToMove returns the side whose turn it is.
func (g *Game) SideToMove() Color {
	return g.Turn
}

// EnPassantSquare returns the current en-passant target square, or 0 if none.
func (g *Game) EnPassantSquare() Square {
	return g.EnPassant
}

// CastlingRights returns the current castling rights bitset.
func (g *Game) CastlingRights() CastlingRights {
	return CastlingRights(g.Castling)
}

// Has reports whether all requested castling rights are present.
func (r CastlingRights) Has(rights CastlingRights) bool {
	return r&rights == rights
}

// String returns the FEN castling-rights field.
func (r CastlingRights) String() string {
	if r == 0 {
		return "-"
	}
	out := make([]byte, 0, 4)
	if r.Has(CastlingWhiteKingSide) {
		out = append(out, 'K')
	}
	if r.Has(CastlingWhiteQueenSide) {
		out = append(out, 'Q')
	}
	if r.Has(CastlingBlackKingSide) {
		out = append(out, 'k')
	}
	if r.Has(CastlingBlackQueenSide) {
		out = append(out, 'q')
	}
	return string(out)
}

// HalfMoveClock returns the halfmove clock used by the fifty/seventy-five move rules.
func (g *Game) HalfMoveClock() int {
	return g.HalfMoves
}

// FullMoveNumber returns the fullmove number.
func (g *Game) FullMoveNumber() int {
	return g.FullMoves
}

// PieceAt returns the piece on square. The bool is false for out-of-range squares.
func (g *Game) PieceAt(square Square) (Piece, bool) {
	if square >= 64 {
		return EMPTY, false
	}
	return g.Squares[square], true
}

// Board returns a copy of the board squares.
func (g *Game) Board() [64]Piece {
	return g.Squares
}

// BoardView returns a read-only view of the current board internals.
func (g *Game) BoardView() BoardView {
	return BoardView{
		squares:     g.Squares,
		whitePieces: g.WhitePieces,
		blackPieces: g.BlackPieces,
		whites:      g.Whites,
		blacks:      g.Blacks,
		occupied:    g.Occupied,
	}
}

// PieceAt returns the piece on square. The bool is false for out-of-range squares.
func (v BoardView) PieceAt(square Square) (Piece, bool) {
	if square >= 64 {
		return EMPTY, false
	}
	return v.squares[square], true
}

// Squares returns a copy of the board squares.
func (v BoardView) Squares() [64]Piece {
	return v.squares
}

// Pieces returns all occupied squares for color.
func (v BoardView) Pieces(color Color) Bitboard {
	switch color {
	case WHITE:
		return v.whitePieces
	case BLACK:
		return v.blackPieces
	default:
		return 0
	}
}

// PiecesOfKind returns all occupied squares for color and piece kind.
func (v BoardView) PiecesOfKind(color Color, kind Piece) Bitboard {
	if kind < PAWN || kind > KING {
		return 0
	}
	switch color {
	case WHITE:
		return v.whites[kind]
	case BLACK:
		return v.blacks[kind]
	default:
		return 0
	}
}

// OccupiedSquares returns the occupied-square bitboard.
func (v BoardView) OccupiedSquares() Bitboard {
	return v.occupied
}

// Pieces returns all occupied squares for color.
func (g *Game) Pieces(color Color) Bitboard {
	switch color {
	case WHITE:
		return g.WhitePieces
	case BLACK:
		return g.BlackPieces
	default:
		return 0
	}
}

// PiecesOfKind returns all occupied squares for color and piece kind.
func (g *Game) PiecesOfKind(color Color, kind Piece) Bitboard {
	if kind < PAWN || kind > KING {
		return 0
	}
	switch color {
	case WHITE:
		return g.Whites[kind]
	case BLACK:
		return g.Blacks[kind]
	default:
		return 0
	}
}

// OccupiedSquares returns the occupied-square bitboard.
func (g *Game) OccupiedSquares() Bitboard {
	return g.Occupied
}

// LegalMovesInto copies legal moves into dst and returns the resulting slice.
func (g *Game) LegalMovesInto(dst []Move) []Move {
	return g.CopyLegalMoves(dst)
}

// PositionKey returns the current Zobrist position key.
//
// The key is useful for hash tables and same-version position comparisons. Use
// FEN when persisting positions across package versions.
func (g *Game) PositionKey() uint64 {
	return g.ZobristHash
}

// CanClaimThreefoldRepetition reports whether the current position has occurred at least three times.
func (g *Game) CanClaimThreefoldRepetition() bool {
	return g.checkThreefoldRepetition()
}

// CanClaimFiftyMoveRule reports whether the halfmove clock allows a 50-move-rule claim.
func (g *Game) CanClaimFiftyMoveRule() bool {
	return g.checkFiftyMoveRule()
}

// IsFivefoldRepetitionDraw reports whether fivefold repetition makes the game automatically drawn.
func (g *Game) IsFivefoldRepetitionDraw() bool {
	return g.IsFivefoldRepetition()
}

// IsSeventyFiveMoveRuleDraw reports whether the 75-move rule makes the game automatically drawn.
func (g *Game) IsSeventyFiveMoveRuleDraw() bool {
	return g.checkSeventyFiveMoveRule()
}

// DrawStatus returns the current draw rule state without exposing mutable fields.
func (g *Game) DrawStatus() DrawStatus {
	return DrawStatus{
		InsufficientMaterial:        g.hasInsufficientMaterial(),
		CanClaimThreefoldRepetition: g.CanClaimThreefoldRepetition(),
		CanClaimFiftyMoveRule:       g.CanClaimFiftyMoveRule(),
		FivefoldRepetition:          g.IsFivefoldRepetitionDraw(),
		SeventyFiveMoveRule:         g.IsSeventyFiveMoveRuleDraw(),
	}
}

// Snapshot returns a defensive value copy of the current game state.
func (g *Game) Snapshot() GameSnapshot {
	return GameSnapshot{
		FEN:            g.FEN(),
		SideToMove:     g.SideToMove(),
		Board:          g.Board(),
		WhitePieces:    g.Pieces(WHITE),
		BlackPieces:    g.Pieces(BLACK),
		WhiteByKind:    g.Whites,
		BlackByKind:    g.Blacks,
		Occupied:       g.OccupiedSquares(),
		EnPassant:      g.EnPassantSquare(),
		Castling:       g.CastlingRights(),
		HalfMoveClock:  g.HalfMoveClock(),
		FullMoveNumber: g.FullMoveNumber(),
		PositionKey:    g.PositionKey(),
		LegalMoves:     g.LegalMovesList(),
		Status:         g.Status(),
		Terminal:       g.IsTerminal(),
		Draw:           g.DrawStatus(),
	}
}

// Status returns the current game status.
func (g *Game) Status() GameStatus {
	switch {
	case g.IsCheckmate:
		return GameStatusCheckmate
	case g.IsStalement:
		return GameStatusStalemate
	case g.IsMaterialDraw:
		return GameStatusDrawInsufficientMaterial
	case g.IsFivefoldRepetition():
		return GameStatusDrawFivefoldRepetition
	case g.IsSeventyFiveMoveRule:
		return GameStatusDrawSeventyFiveMoveRule
	case g.IsCheck:
		return GameStatusCheck
	default:
		return GameStatusOngoing
	}
}

// IsTerminal reports whether the game is finished.
func (g *Game) IsTerminal() bool {
	return g.IsFinished
}

// String returns a stable machine-readable status name.
func (s GameStatus) String() string {
	switch s {
	case GameStatusOngoing:
		return "ongoing"
	case GameStatusCheck:
		return "check"
	case GameStatusCheckmate:
		return "checkmate"
	case GameStatusStalemate:
		return "stalemate"
	case GameStatusDrawInsufficientMaterial:
		return "draw_insufficient_material"
	case GameStatusDrawFivefoldRepetition:
		return "draw_fivefold_repetition"
	case GameStatusDrawSeventyFiveMoveRule:
		return "draw_seventy_five_move_rule"
	default:
		return "unknown"
	}
}
