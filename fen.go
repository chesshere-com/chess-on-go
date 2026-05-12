package chessongo

import (
	"fmt"
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

// RUNE_TO_FILE maps file runes to zero-based file indexes.
//
// Compatibility: prefer ParseSquare or Square.File in new code.
var RUNE_TO_FILE = map[rune]int{'a': 0, 'b': 1, 'c': 2, 'd': 3, 'e': 4, 'f': 5, 'g': 6, 'h': 7}

// RUNE_TO_RANK maps rank runes to zero-based rank indexes.
//
// Compatibility: prefer ParseSquare or Square.Rank in new code.
var RUNE_TO_RANK = map[rune]int{'1': 7, '2': 6, '3': 5, '4': 4, '5': 3, '6': 2, '7': 1, '8': 0}

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
	parts, ok := splitFENFields(fen)
	if !ok {
		return invalidFEN("expected six fields")
	}

	parsed := Game{}
	parsed.Reset()

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

	if parts[2] == "-" {
		parsed.castling = 0
	} else {
		seen := [256]bool{}
		for i := 0; i < len(parts[2]); i++ {
			c := parts[2][i]
			if seen[c] {
				return invalidFENField(FENFieldCastling, "duplicate castling right")
			}
			seen[c] = true
			switch c {
			case 'K':
				parsed.castling |= CASTLE_WKS
			case 'Q':
				parsed.castling |= CASTLE_WQS
			case 'k':
				parsed.castling |= CASTLE_BKS
			case 'q':
				parsed.castling |= CASTLE_BQS
			default:
				return invalidFENField(FENFieldCastling, "invalid castling right")
			}
		}
		if !parsed.castlingRightsMatchBoard() {
			return invalidFENField(FENFieldCastling, "castling rights do not match board")
		}
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
	if parsed.sideNotToMoveInCheck() {
		return invalidFENField(FENFieldLegality, "side not to move is in check")
	}

	parsed.recordPosition()
	parsed.GenerateLegalMoves()
	parsed.refreshStatus()
	*g = parsed

	return nil
}

func splitFENFields(fen string) ([6]string, bool) {
	var fields [6]string
	field := 0
	i := 0
	for i < len(fen) && fen[i] == ' ' {
		i++
	}
	for i < len(fen) {
		if field >= len(fields) {
			return fields, false
		}
		start := i
		for i < len(fen) && fen[i] != ' ' {
			i++
		}
		if start == i {
			return fields, false
		}
		fields[field] = fen[start:i]
		field++
		for i < len(fen) && fen[i] == ' ' {
			i++
		}
	}
	return fields, field == len(fields)
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
	var i, emptyCount int = 0, 0
	for rank := 0; rank < 8; rank++ {
		for file := 0; file < 8; file++ {
			if g.squares[i] != EMPTY {
				pieces += string(PIECE_TO_RUNE[g.squares[i]])
				i++
				continue
			}
			for emptyCount = 0; file < 8 && g.squares[i] == EMPTY; {
				emptyCount++
				i++
				file++
			}
			if emptyCount > 0 {
				pieces += strconv.Itoa(emptyCount)
				file--
			}
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

	castling = ""
	if (g.castling & CASTLE_WKS) > 0 {
		castling += "K"
	}
	if (g.castling & CASTLE_WQS) > 0 {
		castling += "Q"
	}
	if (g.castling & CASTLE_BKS) > 0 {
		castling += "k"
	}
	if (g.castling & CASTLE_BQS) > 0 {
		castling += "q"
	}
	if len(castling) == 0 {
		castling = "-"
	}

	if g.enPassant == 0 {
		enPassant = "-"
	} else {
		rank, file := squareCoords(g.enPassant)
		enPassant = FILE_TO_STRING[file] + RANK_TO_STRING[rank]
	}

	return fmt.Sprintf("%s %s %s %s %d %d", pieces, turn, castling, enPassant, g.halfMoves, g.fullMoves)
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
		value = value*10 + int(d-'0')
	}
	return value, nil
}

func (g *Game) castlingRightsMatchBoard() bool {
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
