package chessongo

import (
	"fmt"
	"strconv"
	"strings"
)

// EPDRecord is one Extended Position Description record.
type EPDRecord struct {
	FEN        string
	Operations map[string][]string
}

// Operation returns a copy of the operands for opcode.
func (r EPDRecord) Operation(opcode string) []string {
	values := r.Operations[opcode]
	return append([]string(nil), values...)
}

// BestMoves returns the `bm` operands.
func (r EPDRecord) BestMoves() []string {
	return r.Operation("bm")
}

// AvoidMoves returns the `am` operands.
func (r EPDRecord) AvoidMoves() []string {
	return r.Operation("am")
}

// ID returns the first `id` operand.
func (r EPDRecord) ID() (string, bool) {
	values := r.Operations["id"]
	if len(values) == 0 {
		return "", false
	}
	return values[0], true
}

// PerftExpectations returns perft depth/node expectations from `perft` or `D<n>` opcodes.
func (r EPDRecord) PerftExpectations() (map[int]uint64, error) {
	out := map[int]uint64{}
	if values := r.Operation("perft"); len(values) > 0 {
		if len(values)%2 != 0 {
			return nil, fmt.Errorf("invalid epd perft operands")
		}
		for i := 0; i < len(values); i += 2 {
			depth, err := strconv.Atoi(values[i])
			if err != nil || depth < 0 {
				return nil, fmt.Errorf("invalid epd perft depth %q", values[i])
			}
			nodes, err := strconv.ParseUint(values[i+1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid epd perft nodes %q", values[i+1])
			}
			out[depth] = nodes
		}
	}
	for opcode, values := range r.Operations {
		if len(opcode) < 2 || (opcode[0] != 'D' && opcode[0] != 'd') || len(values) == 0 {
			continue
		}
		depth, err := strconv.Atoi(opcode[1:])
		if err != nil || depth < 0 {
			continue
		}
		nodes, err := strconv.ParseUint(values[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid epd %s nodes %q", opcode, values[0])
		}
		out[depth] = nodes
	}
	return out, nil
}

// Game returns a game loaded from the EPD position.
func (r EPDRecord) Game() (*Game, error) {
	return NewGameFromFEN(r.FEN)
}

// ParseEPD parses one Extended Position Description record.
func ParseEPD(line string) (EPDRecord, error) {
	line = strings.TrimSpace(line)
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return EPDRecord{}, invalidFEN("epd requires four position fields")
	}
	fen := strings.Join(fields[:4], " ") + " 0 1"
	g := &Game{}
	if err := g.LoadFEN(fen); err != nil {
		return EPDRecord{}, err
	}

	rest := line
	for _, field := range fields[:4] {
		rest = strings.TrimSpace(rest)
		rest = rest[len(field):]
	}
	ops, err := parseEPDOperations(strings.TrimSpace(rest))
	if err != nil {
		return EPDRecord{}, err
	}
	return EPDRecord{FEN: fen, Operations: ops}, nil
}

// LoadEPDRecords parses multiple EPD records, skipping blank and comment lines.
func LoadEPDRecords(input string) ([]EPDRecord, error) {
	var records []EPDRecord
	for lineNo, line := range strings.Split(input, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		record, err := ParseEPD(line)
		if err != nil {
			return nil, fmt.Errorf("epd line %d: %w", lineNo+1, err)
		}
		records = append(records, record)
	}
	return records, nil
}

func parseEPDOperations(text string) (map[string][]string, error) {
	ops := map[string][]string{}
	start := 0
	quoted, escaped := false, false
	for i := 0; i <= len(text); i++ {
		if i < len(text) {
			c := text[i]
			if escaped {
				escaped = false
				continue
			}
			if quoted && c == '\\' {
				escaped = true
				continue
			}
			if c == '"' {
				quoted = !quoted
				continue
			}
			if c != ';' || quoted {
				continue
			}
		}
		if quoted || escaped {
			return nil, fmt.Errorf("invalid epd quoted operand")
		}
		raw := text[start:i]
		start = i + 1
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		tokens, err := splitEPDOperation(raw)
		if err != nil {
			return nil, err
		}
		if len(tokens) == 0 {
			continue
		}
		ops[tokens[0]] = append([]string(nil), tokens[1:]...)
	}
	return ops, nil
}

func splitEPDOperation(text string) ([]string, error) {
	var tokens []string
	for i := 0; i < len(text); {
		for i < len(text) && (text[i] == ' ' || text[i] == '\t' || text[i] == '\r') {
			i++
		}
		if i >= len(text) {
			break
		}
		if text[i] != '"' {
			start := i
			for i < len(text) && text[i] != ' ' && text[i] != '\t' && text[i] != '\r' {
				i++
			}
			tokens = append(tokens, text[start:i])
			continue
		}

		i++
		var b strings.Builder
		escaped := false
		closed := false
		for ; i < len(text); i++ {
			c := text[i]
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
				i++
				tokens = append(tokens, b.String())
				closed = true
				break
			}
			b.WriteByte(c)
		}
		if !closed || (i < len(text) && text[i] != ' ' && text[i] != '\t' && text[i] != '\r') {
			return nil, fmt.Errorf("invalid epd quoted operand")
		}
	}
	return tokens, nil
}
