package chessongo

import (
	"fmt"
	"math"
	"strconv"
)

// STARTING_POSITION_FEN is the standard chess initial position.
const STARTING_POSITION_FEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

// RUNE_TO_PIECE maps FEN piece runes to pieces.
//
// Compatibility: prefer PieceFromRune in new code.
var RUNE_TO_PIECE = map[rune]Piece{
	'P': W_PAWN, 'N': W_KNIGHT, 'B': W_BISHOP, 'R': W_ROOK, 'Q': W_QUEEN, 'K': W_KING,
	'p': B_PAWN, 'n': B_KNIGHT, 'b': B_BISHOP, 'r': B_ROOK, 'q': B_QUEEN, 'k': B_KING,
}

// PIECE_TO_RUNE maps pieces to FEN runes.
//
// Compatibility: prefer Piece.ToRune in new code.
var PIECE_TO_RUNE = map[Piece]rune{
	W_PAWN: 'P', W_KNIGHT: 'N', W_BISHOP: 'B', W_ROOK: 'R', W_QUEEN: 'Q', W_KING: 'K',
	B_PAWN: 'p', B_KNIGHT: 'n', B_BISHOP: 'b', B_ROOK: 'r', B_QUEEN: 'q', B_KING: 'k',
}

// STRING_TO_KIND maps piece letters to piece kinds.
//
// Compatibility: this low-level lookup table is retained for older callers.
var STRING_TO_KIND = map[string]uint{
	"P": PAWN, "N": KNIGHT, "B": BISHOP, "R": ROOK, "Q": QUEEN, "K": KING,
	"p": PAWN, "n": KNIGHT, "b": BISHOP, "r": ROOK, "q": QUEEN, "k": KING,
}

// FILE_TO_STRING maps zero-based file indexes to file strings.
//
// Compatibility: prefer Square.FileLetter in new code.
var FILE_TO_STRING = map[int]string{0: "a", 1: "b", 2: "c", 3: "d", 4: "e", 5: "f", 6: "g", 7: "h"}

// RANK_TO_STRING maps zero-based rank indexes to rank strings.
//
// Compatibility: prefer Square.RankDigit in new code.
var RANK_TO_STRING = map[int]string{0: "8", 1: "7", 2: "6", 3: "5", 4: "4", 5: "3", 6: "2", 7: "1"}

// LoadFEN initializes the game from a FEN string and refreshes legal moves and status flags.
func (g *Game) LoadFEN(fen string) error {
	return g.loadFEN(fen, VariantStandard)
}

// loadFEN initializes the game from a FEN string and refreshes legal moves and status flags.
func (g *Game) loadFEN(fen string, variant Variant) error {
	parts, ok := splitFENFieldsForVariant(fen, variant)
	if !ok {
		if variant == VariantThreeCheck {
			return invalidFEN("expected seven fields")
		}
		return invalidFEN("expected six fields")
	}

	parsed := Game{}
	parsed.Reset()
	parsed.variant = variant
	parsed.castlingRookFrom = defaultCastlingRookFrom()

	idx, rankIdx := 0, 0
	whiteKings, blackKings := 0, 0
	pieces := parts[0]
	for start := 0; start <= len(pieces); {
		if rankIdx >= 8 {
			return invalidFENField(FENFieldPiecePlacement, "too many ranks")
		}
		end := start
		for end < len(pieces) && pieces[end] != '/' {
			end++
		}
		rankText := pieces[start:end]
		if rankText == "" {
			return invalidFENField(FENFieldPiecePlacement, "empty rank")
		}
		fileCount := 0
		previousWasDigit := false
		for i := 0; i < len(rankText); i++ {
			c := rankText[i]
			if c >= '1' && c <= '8' {
				if previousWasDigit {
					return invalidFENField(FENFieldPiecePlacement, "consecutive empty-square digits")
				}
				empty := int(c - '0')
				fileCount += empty
				idx += empty
				previousWasDigit = true
				continue
			}
			piece, ok := fenPiece(c)
			if !ok || fileCount >= 8 || idx >= 64 {
				return invalidFENField(FENFieldPiecePlacement, "invalid piece placement")
			}
			if piece.Kind() == PAWN && (rankIdx == 0 || rankIdx == 7) {
				return invalidFENField(FENFieldPiecePlacement, "pawn on promotion rank")
			}
			if piece == W_KING {
				whiteKings++
			}
			if piece == B_KING {
				blackKings++
			}
			parsed.addPiece(piece, idx)
			idx++
			fileCount++
			previousWasDigit = false
		}
		if fileCount != 8 {
			return invalidFENField(FENFieldPiecePlacement, "rank does not contain eight files")
		}
		rankIdx++
		start = end + 1
		if end == len(pieces) {
			break
		}
	}
	if rankIdx != 8 || idx != 64 || whiteKings != 1 || blackKings != 1 {
		return invalidFENField(FENFieldPiecePlacement, "expected eight ranks and exactly one king per side")
	}

	switch parts[1] {
	case "w":
		parsed.turn = WHITE
	case "b":
		parsed.turn = BLACK
	default:
		return invalidFENField(FENFieldSideToMove, "invalid side to move")
	}

	if err := parsed.parseFENCastlingField(parts[2]); err != nil {
		return err
	}

	if parts[3] == "-" {
		parsed.enPassant = 0
	} else if len(parts[3]) == 2 {
		fileChar, rankChar := parts[3][0], parts[3][1]
		if fileChar < 'a' || fileChar > 'h' || rankChar < '1' || rankChar > '8' {
			return invalidFENField(FENFieldEnPassant, "invalid en-passant square")
		}
		if (parsed.turn == WHITE && rankChar != '6') || (parsed.turn == BLACK && rankChar != '3') {
			return invalidFENField(FENFieldEnPassant, "invalid en-passant rank")
		}
		parsed.enPassant = CoordsToSquare(8-int(rankChar-'0'), int(fileChar-'a'))
		if !parsed.enPassantStateMatchesBoard() {
			return invalidFENField(FENFieldEnPassant, "en-passant state does not match board")
		}
	} else {
		return invalidFENField(FENFieldEnPassant, "invalid en-passant field")
	}

	halfMoves, err := parseFENNumber(parts[4])
	if err != nil {
		return invalidFENField(FENFieldHalfMoveClock, "invalid halfmove clock")
	}
	parsed.halfMoves = halfMoves

	fullMoves, err := parseFENNumber(parts[5])
	if err != nil || fullMoves < 1 {
		return invalidFENField(FENFieldFullMoveNumber, "invalid fullmove number")
	}
	parsed.fullMoves = fullMoves

	if variant == VariantThreeCheck {
		counts, err := parseThreeCheckCounters(parts[6])
		if err != nil {
			return invalidFENField(FENFieldVariantState, err.Error())
		}
		parsed.variantState.checksGiven = counts
	}

	if parsed.sideNotToMoveInCheck() {
		return invalidFENField(FENFieldLegality, "side not to move is in check")
	}

	parsed.recordPosition()
	parsed.GenerateLegalMoves()
	parsed.refreshStatus()
	*g = parsed

	return nil
}

func splitFENFieldsForVariant(fen string, variant Variant) ([]string, bool) {
	expected := 6
	if variant == VariantThreeCheck {
		expected = 7
	}
	fields := make([]string, 0, expected)
	i := 0
	for i < len(fen) && fen[i] == ' ' {
		i++
	}
	for i < len(fen) {
		if len(fields) >= expected {
			return fields, false
		}
		start := i
		for i < len(fen) && fen[i] != ' ' {
			i++
		}
		if start == i {
			return fields, false
		}
		fields = append(fields, fen[start:i])
		for i < len(fen) && fen[i] == ' ' {
			i++
		}
	}
	return fields, len(fields) == expected
}

func fenPiece(c byte) (Piece, bool) {
	switch c {
	case 'P':
		return W_PAWN, true
	case 'N':
		return W_KNIGHT, true
	case 'B':
		return W_BISHOP, true
	case 'R':
		return W_ROOK, true
	case 'Q':
		return W_QUEEN, true
	case 'K':
		return W_KING, true
	case 'p':
		return B_PAWN, true
	case 'n':
		return B_KNIGHT, true
	case 'b':
		return B_BISHOP, true
	case 'r':
		return B_ROOK, true
	case 'q':
		return B_QUEEN, true
	case 'k':
		return B_KING, true
	default:
		return EMPTY, false
	}
}

// ToFEN returns the FEN representation of the current game state.
func (g *Game) ToFEN() string {
	var pieces, turn, castling, enPassant string

	pieces = ""
	for rank := 0; rank < 8; rank++ {
		emptyCount := 0
		for file := 0; file < 8; file++ {
			idx := rank*8 + file
			if g.squares[idx] != EMPTY {
				if emptyCount > 0 {
					pieces += strconv.Itoa(emptyCount)
					emptyCount = 0
				}
				pieces += string(PIECE_TO_RUNE[g.squares[idx]])
			} else {
				emptyCount++
			}
		}
		if emptyCount > 0 {
			pieces += strconv.Itoa(emptyCount)
		}
		if rank < 7 {
			pieces += "/"
		}
	}

	if g.turn == WHITE {
		turn = "w"
	} else {
		turn = "b"
	}

	castling = g.fenCastlingField()

	if g.enPassant == 0 {
		enPassant = "-"
	} else {
		rank, file := squareCoords(g.enPassant)
		enPassant = FILE_TO_STRING[file] + RANK_TO_STRING[rank]
	}

	fen := fmt.Sprintf("%s %s %s %s %d %d", pieces, turn, castling, enPassant, g.halfMoves, g.fullMoves)
	if g.variant == VariantThreeCheck {
		fen += " " + g.threeCheckFENField()
	}
	return fen
}

func parseFENNumber(token string) (int, error) {
	if token == "" {
		return 0, invalidFEN("invalid format")
	}
	value := 0
	for i := 0; i < len(token); i++ {
		d := token[i]
		if d < '0' || d > '9' {
			return 0, invalidFEN("invalid format")
		}
		digit := int(d - '0')
		if value > (math.MaxInt-digit)/10 {
			return 0, invalidFEN("invalid format")
		}
		value = value*10 + digit
	}
	return value, nil
}

func (g *Game) parseFENCastlingField(field string) error {
	g.castling = 0
	if field == "-" {
		return nil
	}
	if field == "" {
		return invalidFENField(FENFieldCastling, "invalid castling right")
	}

	seenChars := [256]bool{}
	seenRights := [16]bool{}
	for i := 0; i < len(field); i++ {
		c := field[i]
		if seenChars[c] {
			return invalidFENField(FENFieldCastling, "duplicate castling right")
		}
		seenChars[c] = true

		right, rookFrom, err := g.parseFENCastlingRight(c)
		if err != nil {
			return err
		}
		if right <= 0 || right >= len(seenRights) || seenRights[right] {
			return invalidFENField(FENFieldCastling, "duplicate castling right")
		}
		seenRights[right] = true
		g.castling |= right
		g.castlingRookFrom[right] = rookFrom
	}

	if !g.castlingRightsMatchBoard() {
		return invalidFENField(FENFieldCastling, "castling rights do not match board")
	}
	return nil
}

func (g *Game) parseFENCastlingRight(c byte) (int, Square, error) {
	switch c {
	case 'K':
		if g.variant == VariantChess960 {
			return g.parseChess960KQCastlingRight(c)
		}
		return CASTLE_WKS, WKS_ROOK_ORIGINAL_SQUARE, nil
	case 'Q':
		if g.variant == VariantChess960 {
			return g.parseChess960KQCastlingRight(c)
		}
		return CASTLE_WQS, WQS_ROOK_ORIGINAL_SQUARE, nil
	case 'k':
		if g.variant == VariantChess960 {
			return g.parseChess960KQCastlingRight(c)
		}
		return CASTLE_BKS, BKS_ROOK_ORIGINAL_SQUARE, nil
	case 'q':
		if g.variant == VariantChess960 {
			return g.parseChess960KQCastlingRight(c)
		}
		return CASTLE_BQS, BQS_ROOK_ORIGINAL_SQUARE, nil
	}

	if g.variant != VariantChess960 {
		return 0, noSquare, invalidFENField(FENFieldCastling, "invalid castling right")
	}
	if c >= 'A' && c <= 'H' {
		return g.parseChess960FileCastlingRight(WHITE, int(c-'A'))
	}
	if c >= 'a' && c <= 'h' {
		return g.parseChess960FileCastlingRight(BLACK, int(c-'a'))
	}
	return 0, noSquare, invalidFENField(FENFieldCastling, "invalid castling right")
}

func (g *Game) parseChess960FileCastlingRight(color Color, rookFile int) (int, Square, error) {
	king := g.kingSquare(color)
	if !king.Valid() {
		return 0, noSquare, invalidFENField(FENFieldCastling, "castling rights do not match board")
	}

	rookFrom := squareOnBackRank(color, rookFile)
	rookPiece := Piece(W_ROOK)
	if color == BLACK {
		rookPiece = B_ROOK
	}
	if g.squares[rookFrom] != rookPiece {
		return 0, noSquare, invalidFENField(FENFieldCastling, "castling rights do not match board")
	}
	if rookFile == king.File() {
		return 0, noSquare, invalidFENField(FENFieldCastling, "invalid castling right")
	}

	return castlingRightFor(color, rookFile > king.File()), rookFrom, nil
}

func (g *Game) parseChess960KQCastlingRight(c byte) (int, Square, error) {
	color := Color(WHITE)
	if c == 'k' || c == 'q' {
		color = BLACK
	}
	kingside := c == 'K' || c == 'k'

	king := g.kingSquare(color)
	if !king.Valid() {
		return 0, noSquare, invalidFENField(FENFieldCastling, "castling rights do not match board")
	}
	rookFrom, ok := g.unambiguousChess960RookForSide(color, king.File(), kingside)
	if !ok {
		return 0, noSquare, invalidFENField(FENFieldCastling, "invalid castling right")
	}
	return castlingRightFor(color, kingside), rookFrom, nil
}

func (g *Game) unambiguousChess960RookForSide(color Color, kingFile int, kingside bool) (Square, bool) {
	rookPiece := Piece(W_ROOK)
	if color == BLACK {
		rookPiece = B_ROOK
	}

	found := noSquare
	for file := 0; file < 8; file++ {
		if kingside && file <= kingFile {
			continue
		}
		if !kingside && file >= kingFile {
			continue
		}
		sq := squareOnBackRank(color, file)
		if g.squares[sq] != rookPiece {
			continue
		}
		if found.Valid() {
			return noSquare, false
		}
		found = sq
	}
	return found, found.Valid()
}

func (g *Game) kingSquare(color Color) Square {
	kingPiece := Piece(W_KING)
	if color == BLACK {
		kingPiece = B_KING
	}
	rank := backRankFor(color)
	for file := 0; file < 8; file++ {
		sq := squareOnBackRank(color, file)
		if g.squares[sq] == kingPiece && sq.Rank() == rank {
			return sq
		}
	}
	return noSquare
}

func (g *Game) fenCastlingField() string {
	if g.castling == 0 {
		return "-"
	}

	if g.variant != VariantChess960 {
		out := make([]byte, 0, 4)
		if (g.castling & CASTLE_WKS) > 0 {
			out = append(out, 'K')
		}
		if (g.castling & CASTLE_WQS) > 0 {
			out = append(out, 'Q')
		}
		if (g.castling & CASTLE_BKS) > 0 {
			out = append(out, 'k')
		}
		if (g.castling & CASTLE_BQS) > 0 {
			out = append(out, 'q')
		}
		return string(out)
	}

	out := make([]byte, 0, 4)
	for _, right := range []int{CASTLE_WKS, CASTLE_WQS, CASTLE_BKS, CASTLE_BQS} {
		if (g.castling & right) == 0 {
			continue
		}
		rookFrom := g.castlingRookFrom[right]
		if !rookFrom.Valid() {
			continue
		}
		letter := byte('A' + rookFrom.File())
		if castlingRightColor(right) == BLACK {
			letter += 'a' - 'A'
		}
		out = append(out, letter)
	}
	if len(out) == 0 {
		return "-"
	}
	return string(out)
}

func (g *Game) castlingRightsMatchBoard() bool {
	if g.variant == VariantStandard {
		if (g.castling&CASTLE_WKS) > 0 && (g.squares[W_KING_INIT_SQUARE] != W_KING || g.squares[WKS_ROOK_ORIGINAL_SQUARE] != W_ROOK) {
			return false
		}
		if (g.castling&CASTLE_WQS) > 0 && (g.squares[W_KING_INIT_SQUARE] != W_KING || g.squares[WQS_ROOK_ORIGINAL_SQUARE] != W_ROOK) {
			return false
		}
		if (g.castling&CASTLE_BKS) > 0 && (g.squares[B_KING_INIT_SQUARE] != B_KING || g.squares[BKS_ROOK_ORIGINAL_SQUARE] != B_ROOK) {
			return false
		}
		if (g.castling&CASTLE_BQS) > 0 && (g.squares[B_KING_INIT_SQUARE] != B_KING || g.squares[BQS_ROOK_ORIGINAL_SQUARE] != B_ROOK) {
			return false
		}
		return true
	}

	for _, right := range []int{CASTLE_WKS, CASTLE_WQS, CASTLE_BKS, CASTLE_BQS} {
		if (g.castling & right) == 0 {
			continue
		}
		color := castlingRightColor(right)
		king := g.kingSquare(color)
		rookFrom := g.castlingRookFrom[right]
		if !king.Valid() || !rookFrom.Valid() || rookFrom.Rank() != backRankFor(color) {
			return false
		}
		kingPiece, rookPiece := Piece(W_KING), Piece(W_ROOK)
		if color == BLACK {
			kingPiece, rookPiece = B_KING, B_ROOK
		}
		if g.squares[king] != kingPiece {
			return false
		}
		if g.squares[rookFrom] != rookPiece {
			return false
		}
	}
	return true
}

func (g *Game) enPassantStateMatchesBoard() bool {
	ep := g.enPassant
	if g.turn == WHITE {
		return ep+8 <= 63 && g.squares[ep] == EMPTY && g.squares[ep+8] == B_PAWN
	}
	return ep >= 8 && g.squares[ep] == EMPTY && g.squares[ep-8] == W_PAWN
}

func (g *Game) sideNotToMoveInCheck() bool {
	if g.turn == WHITE {
		king := g.blacks[KING]
		return king != 0 && g.isSquareAttackedByWithOccupied(Square(king.lsbIndex()), WHITE, g.occupied)
	}
	king := g.whites[KING]
	return king != 0 && g.isSquareAttackedByWithOccupied(Square(king.lsbIndex()), BLACK, g.occupied)
}
