package chessongo

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

const pgnMovetextLineWidth = 80

// LoadPGN loads a PGN string into the board, playing all main-line moves and
// recording position history via Zobrist hashing. Variations and comments are
// ignored; only the main line is applied.
func (g *Game) LoadPGN(pgn string) error {
	tags := parsePGNTags(pgn)
	startFEN := STARTING_POSITION_FEN
	if fen := tags["FEN"]; fen != "" {
		startFEN = strings.TrimSpace(fen)
	}
	variant := pgnVariant(tags["Variant"])

	loaded := &Game{}
	if err := loaded.LoadFENWithVariant(startFEN, variant); err != nil {
		return err
	}
	loaded.pgnTags = cloneStringMap(tags)
	loaded.pgnStartFEN = startFEN
	loaded.pgnStartTurn = loaded.turn
	loaded.pgnStartFullMove = loaded.fullMoves

	fastPath := !strings.ContainsAny(pgn, "[{(;")

	var tokens []string
	if fastPath {
		tokens = fastTokenizeMoves(pgn)
	} else {
		tokens = tokenizePGNMoves(pgn)
	}

	movetextResult := ""
	for _, tok := range tokens {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		if isPGNResult(tok) {
			movetextResult = tok
			continue
		}
		// Skip move numbers and NAGs.
		if strings.Contains(tok, "..") || strings.Contains(tok, ".") || strings.HasPrefix(tok, "$") {
			continue
		}

		tok = trimSANAnnotations(tok)
		if tok == "" {
			continue
		}

		target := getTargetSquare(tok)

		// b.GenerateLegalMoves() is already done by LoadFEN (initially) and MakeMove (subsequently).
		matched := false
		for _, mv := range loaded.legalMoves {
			if target != -1 && int(mv.To()) != target {
				continue
			}

			// Optimization: GetMoveSanWithoutSuffix avoids cloning the board (to check for check/mate)
			// which is very expensive. We strip annotations from the token anyway.
			san := trimSANAnnotations(loaded.GetMoveSanWithoutSuffix(mv))
			if san == tok {
				loaded.MakeMove(mv)
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("pgn move not found: %s", tok)
		}
	}

	result, err := resolvePGNResult(tags["Result"], movetextResult, loaded)
	if err != nil {
		return err
	}
	loaded.pgnResult = result
	if loaded.pgnTags == nil {
		loaded.pgnTags = map[string]string{}
	}
	loaded.pgnTags["Result"] = result

	*g = *loaded
	return nil
}

// LoadPGNGame is a helper that constructs a fresh board, loads the PGN, and
// returns the populated board.
func LoadPGNGame(pgn string) (*Game, error) {
	g := &Game{}
	if err := g.LoadPGN(pgn); err != nil {
		return nil, err
	}
	return g, nil
}

func cloneStringMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func pgnVariant(tag string) Variant {
	switch strings.ToLower(strings.ReplaceAll(strings.TrimSpace(tag), " ", "")) {
	case "chess960", "fischerrandom":
		return VariantChess960
	default:
		return VariantStandard
	}
}

// PGNTags returns a copy of the PGN tag pairs loaded with the game.
func (g *Game) PGNTags() map[string]string {
	tags := cloneStringMap(g.pgnTags)
	if tags == nil {
		tags = map[string]string{}
	}
	tags["Result"] = g.PGNResult()
	return tags
}

// PGNResult returns the loaded or inferred PGN result.
func (g *Game) PGNResult() string {
	if g.pgnResult != "" {
		return g.pgnResult
	}
	return inferPGNResult(g)
}

// PGN exports the game as a PGN with tag pairs and main-line movetext.
func (g *Game) PGN() string {
	tags := g.PGNTags()
	result := tags["Result"]
	var b strings.Builder
	for _, key := range orderedPGNTagKeys(tags) {
		fmt.Fprintf(&b, "[%s \"%s\"]\n", key, escapePGNTagValue(tags[key]))
	}
	b.WriteByte('\n')
	b.WriteString(wrapPGNMovetext(g.formatPGNMovetext(result), pgnMovetextLineWidth))
	return b.String()
}

func (g *Game) recordPGNMove(m Move) {
	if g.pgnStartFullMove == 0 {
		g.pgnStartTurn = g.turn
		g.pgnStartFullMove = g.fullMoves
		g.pgnStartFEN = g.FEN()
		if g.variant == VariantChess960 {
			if g.pgnTags == nil {
				g.pgnTags = map[string]string{}
			}
			g.pgnTags["Variant"] = "Chess960"
			g.pgnTags["SetUp"] = "1"
			g.pgnTags["FEN"] = g.pgnStartFEN
		}
	}
	g.pgnMoves = append(g.pgnMoves, g.GetMoveSan(m))
	g.pgnResult = ""
	if g.pgnTags != nil {
		delete(g.pgnTags, "Result")
	}
}

func (g *Game) formatPGNMovetext(result string) string {
	if result == "" {
		result = "*"
	}
	if len(g.pgnMoves) == 0 {
		return result
	}

	startTurn := g.pgnStartTurn
	if startTurn == NO_COLOR {
		startTurn = WHITE
	}
	fullMove := g.pgnStartFullMove
	if fullMove < 1 {
		fullMove = 1
	}

	var b strings.Builder
	turn := startTurn
	for i, move := range g.pgnMoves {
		if i > 0 {
			b.WriteByte(' ')
		}
		if turn == WHITE {
			fmt.Fprintf(&b, "%d. ", fullMove)
		} else if i == 0 {
			fmt.Fprintf(&b, "%d... ", fullMove)
		}
		b.WriteString(move)
		if turn == BLACK {
			fullMove++
			turn = WHITE
		} else {
			turn = BLACK
		}
	}
	if result != "" {
		b.WriteByte(' ')
		b.WriteString(result)
	}
	return b.String()
}

func wrapPGNMovetext(text string, width int) string {
	if width <= 0 {
		return text
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}
	var b strings.Builder
	lineLen := 0
	for i, word := range words {
		if i == 0 {
			b.WriteString(word)
			lineLen = len(word)
			continue
		}
		if lineLen > 0 && lineLen+1+len(word) > width {
			b.WriteByte('\n')
			b.WriteString(word)
			lineLen = len(word)
			continue
		}
		b.WriteByte(' ')
		b.WriteString(word)
		lineLen += 1 + len(word)
	}
	return b.String()
}

func parsePGNTags(pgn string) map[string]string {
	tags := map[string]string{}
	for i := 0; i < len(pgn); {
		lineEnd := strings.IndexByte(pgn[i:], '\n')
		if lineEnd == -1 {
			lineEnd = len(pgn)
		} else {
			lineEnd += i
		}
		line := strings.TrimSpace(pgn[i:lineEnd])
		if key, value, ok := parsePGNTagLine(line); ok {
			tags[key] = value
		}
		if lineEnd == len(pgn) {
			break
		}
		i = lineEnd + 1
	}
	if len(tags) == 0 {
		return nil
	}
	return tags
}

func parsePGNTagLine(line string) (string, string, bool) {
	if len(line) < 4 || line[0] != '[' || line[len(line)-1] != ']' {
		return "", "", false
	}
	body := strings.TrimSpace(line[1 : len(line)-1])
	space := strings.IndexFunc(body, unicode.IsSpace)
	if space <= 0 {
		return "", "", false
	}
	key := body[:space]
	rest := strings.TrimSpace(body[space:])
	if len(rest) < 2 || rest[0] != '"' {
		return "", "", false
	}
	var b strings.Builder
	escaped := false
	for i := 1; i < len(rest); i++ {
		c := rest[i]
		if escaped {
			b.WriteByte(c)
			escaped = false
			continue
		}
		if c == '\\' {
			escaped = true
			continue
		}
		if c == '"' {
			if strings.TrimSpace(rest[i+1:]) != "" {
				return "", "", false
			}
			return key, b.String(), true
		}
		b.WriteByte(c)
	}
	return "", "", false
}

func orderedPGNTagKeys(tags map[string]string) []string {
	seven := []string{"Event", "Site", "Date", "Round", "White", "Black", "Result"}
	defaults := map[string]string{
		"Event":  "?",
		"Site":   "?",
		"Date":   "????.??.??",
		"Round":  "?",
		"White":  "?",
		"Black":  "?",
		"Result": "*",
	}
	for _, key := range seven {
		if tags[key] == "" {
			tags[key] = defaults[key]
		}
	}
	keys := append([]string(nil), seven...)
	var extra []string
	for key := range tags {
		found := false
		for _, standard := range seven {
			if key == standard {
				found = true
				break
			}
		}
		if !found {
			extra = append(extra, key)
		}
	}
	sort.Strings(extra)
	return append(keys, extra...)
}

func escapePGNTagValue(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	return value
}

func resolvePGNResult(tagResult, movetextResult string, g *Game) (string, error) {
	if tagResult != "" && !isPGNResult(tagResult) {
		return "", fmt.Errorf("invalid pgn result tag: %s", tagResult)
	}
	if movetextResult != "" && !isPGNResult(movetextResult) {
		return "", fmt.Errorf("invalid pgn movetext result: %s", movetextResult)
	}
	if tagResult != "" && movetextResult != "" && tagResult != movetextResult {
		return "", fmt.Errorf("pgn result mismatch: tag %s movetext %s", tagResult, movetextResult)
	}

	result := movetextResult
	if result == "" {
		result = tagResult
	}
	if result == "" {
		result = "*"
	}

	if err := validatePGNResultAgainstStatus(result, g); err != nil {
		return "", err
	}
	return result, nil
}

func validatePGNResultAgainstStatus(result string, g *Game) error {
	expected := ""
	switch g.Status() {
	case GameStatusCheckmate:
		if g.SideToMove() == BLACK {
			expected = "1-0"
		} else {
			expected = "0-1"
		}
	case GameStatusStalemate, GameStatusDrawInsufficientMaterial, GameStatusDrawFivefoldRepetition, GameStatusDrawSeventyFiveMoveRule:
		expected = "1/2-1/2"
	}
	if expected == "" || result == "*" || result == expected {
		return nil
	}
	return fmt.Errorf("pgn result %s does not match final position, expected %s", result, expected)
}

func inferPGNResult(g *Game) string {
	switch g.Status() {
	case GameStatusCheckmate:
		if g.SideToMove() == BLACK {
			return "1-0"
		}
		return "0-1"
	case GameStatusStalemate, GameStatusDrawInsufficientMaterial, GameStatusDrawFivefoldRepetition, GameStatusDrawSeventyFiveMoveRule:
		return "1/2-1/2"
	default:
		return "*"
	}
}

func extractFENFromPGN(pgn string) string {
	isSpace := func(b byte) bool { return b == ' ' || b == '\t' || b == '\r' }
	for i := 0; i < len(pgn); {
		lineStart := i
		// find line end
		lineEnd := strings.IndexByte(pgn[i:], '\n')
		if lineEnd == -1 {
			lineEnd = len(pgn)
		} else {
			lineEnd = i + lineEnd
		}

		// trim leading spaces manually
		for lineStart < lineEnd && isSpace(pgn[lineStart]) {
			lineStart++
		}

		// quick prefix check for [FEN
		if lineEnd-lineStart >= 5 && pgn[lineStart] == '[' && pgn[lineStart+1] == 'F' && pgn[lineStart+2] == 'E' && pgn[lineStart+3] == 'N' && pgn[lineStart+4] == ' ' {
			firstQuote := -1
			for j := lineStart + 5; j < lineEnd; j++ {
				if pgn[j] == '"' {
					firstQuote = j
					break
				}
			}
			if firstQuote != -1 {
				lastQuote := -1
				for j := lineEnd - 1; j > firstQuote; j-- {
					if pgn[j] == '"' {
						lastQuote = j
						break
					}
				}
				if lastQuote != -1 {
					fenStart := firstQuote + 1
					fenEnd := lastQuote
					for fenStart < fenEnd && isSpace(pgn[fenStart]) {
						fenStart++
					}
					for fenEnd > fenStart && isSpace(pgn[fenEnd-1]) {
						fenEnd--
					}
					return pgn[fenStart:fenEnd]
				}
			}
		}

		if lineEnd == len(pgn) {
			break
		}
		i = lineEnd + 1
	}
	return ""
}

func tokenizePGNMoves(pgn string) []string {
	var tokens []string
	var bld strings.Builder
	braceDepth, parenDepth := 0, 0
	inTag := false
	inLineComment := false
	lineStart := true

	flush := func() {
		if bld.Len() > 0 {
			tokens = append(tokens, bld.String())
			bld.Reset()
		}
	}

	for i := 0; i < len(pgn); i++ {
		c := pgn[i]
		if inLineComment {
			if c == '\n' {
				inLineComment = false
				lineStart = true
			}
			continue
		}
		if c == '\n' {
			lineStart = true
			inTag = false
			flush()
			continue
		}
		if lineStart {
			if c == ' ' || c == '\t' || c == '\r' {
				continue
			}
			inTag = c == '['
			lineStart = false
		}

		if inTag {
			continue
		}
		switch c {
		case ';':
			inLineComment = true
			flush()
			continue
		case '{':
			braceDepth++
			flush()
			continue
		case '}':
			if braceDepth > 0 {
				braceDepth--
			}
			flush()
			continue
		case '(':
			parenDepth++
			flush()
			continue
		case ')':
			if parenDepth > 0 {
				parenDepth--
			}
			flush()
			continue
		}

		if braceDepth > 0 || parenDepth > 0 {
			continue
		}

		if c == ' ' || c == '\t' || c == '\r' {
			flush()
			continue
		}

		bld.WriteByte(c)
	}
	flush()
	return tokens
}

// fastTokenizeMoves is a lightweight splitter for simple PGNs that only contain
// moves (e.g. "1. e4 e5 2. Nf3 Nc6") without tags, comments, or variations.
func fastTokenizeMoves(pgn string) []string {
	var tokens []string
	start := -1
	for i := 0; i <= len(pgn); i++ {
		var c byte
		if i < len(pgn) {
			c = pgn[i]
		}
		if i == len(pgn) || c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			if start >= 0 {
				tokens = append(tokens, pgn[start:i])
				start = -1
			}
			continue
		}
		if start == -1 {
			start = i
		}
	}
	return tokens
}

func isPGNResult(tok string) bool {
	switch tok {
	case "1-0", "0-1", "1/2-1/2", "*":
		return true
	default:
		return false
	}
}

func trimSANAnnotations(san string) string {
	if i := strings.IndexByte(san, '$'); i >= 0 {
		san = san[:i]
	}
	// Drop trailing check/mate and annotation glyphs.
	san = strings.TrimRightFunc(san, func(r rune) bool {
		return strings.ContainsRune("+#!?", r)
	})
	// Remove leading move numbers if any slipped through (e.g., "12.Nf3").
	san = strings.TrimLeftFunc(san, func(r rune) bool {
		return unicode.IsDigit(r) || r == '.'
	})
	return san
}

func getTargetSquare(san string) int {
	for i := len(san) - 1; i >= 0; i-- {
		c := san[i]
		if c >= '1' && c <= '8' {
			if i > 0 {
				f := san[i-1]
				if f >= 'a' && f <= 'h' {
					col := int(f - 'a')
					row := int('8' - c)
					return row*8 + col
				}
			}
		}
	}
	return -1
}
