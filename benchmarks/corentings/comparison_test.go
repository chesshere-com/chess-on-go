package comparison_test

import (
	"slices"
	"strings"
	"testing"

	ours "github.com/chesshere-com/chess-on-go"
	other "github.com/corentings/chess/v2"
)

const kiwipete = "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"
const shortPGN = "1. e4 e5 2. Nf3 Nc6 3. Bb5 a6 4. Ba4 Nf6 5. O-O Be7 6. Re1 b5 7. Bb3 d6 8. c3 *"

var positions = []struct {
	name, fen string
	depth     int
	nodes     uint64
}{
	{"Initial", ours.STARTING_POSITION_FEN, 4, 197281},
	{"Kiwipete", kiwipete, 3, 97862},
}

// Sinks keep all observed outputs live, including defensive copies.
var sinkString string
var sinkCount int
var sinkNodes uint64
var sinkOurs []ours.Move
var sinkOther []other.Move
var sinkOurMove ours.Move
var sinkOtherMove *other.Move
var sinkOurGame *ours.Game
var sinkOtherGame *other.Game

func loadOther(t testing.TB, fen string) *other.Game {
	t.Helper()
	option, err := other.FEN(fen)
	if err != nil {
		t.Fatal(err)
	}
	return other.NewGame(option)
}

// Use the non-copying legal move API for the competing traversal. Both
// implementations bulk-count depth one; neither tracks PGN or draw history.
func otherPerft(pos *other.Position, depth int) uint64 {
	if depth == 0 {
		return 1
	}
	moves := pos.ValidMovesUnsafe()
	if depth == 1 {
		return uint64(len(moves))
	}
	var nodes uint64
	for i := range moves {
		nodes += otherPerft(pos.Update(&moves[i]), depth-1)
	}
	return nodes
}

func TestEquivalentWorkloads(t *testing.T) {
	for _, p := range positions {
		t.Run(p.name, func(t *testing.T) {
			g, err := ours.NewGameFromFEN(p.fen)
			if err != nil {
				t.Fatal(err)
			}
			c := loadOther(t, p.fen)
			var a, b []string
			for _, m := range g.LegalMovesInto(nil) {
				a = append(a, m.UCI())
			}
			for _, m := range c.ValidMoves() {
				b = append(b, (other.UCINotation{}).Encode(c.Position(), &m))
			}
			slices.Sort(a)
			slices.Sort(b)
			if !slices.Equal(a, b) {
				t.Fatalf("different legal moves: %v / %v", a, b)
			}
			board, err := ours.NewSearchBoard(p.fen)
			if err != nil {
				t.Fatal(err)
			}
			if n := board.Perft(p.depth); n != p.nodes {
				t.Fatalf("ours: %d, want %d", n, p.nodes)
			}
			if n := otherPerft(c.Position(), p.depth); n != p.nodes {
				t.Fatalf("corentings: %d, want %d", n, p.nodes)
			}
			if board.FEN() != p.fen || c.FEN() != p.fen {
				t.Fatal("traversal changed root")
			}
		})
	}
	g := ours.NewGame()
	if err := g.LoadPGN(shortPGN); err != nil {
		t.Fatal(err)
	}
	option, err := other.PGN(strings.NewReader(shortPGN))
	if err != nil {
		t.Fatal(err)
	}
	c := other.NewGame(option)
	if g.FEN() != c.FEN() {
		t.Fatal("PGN final positions differ")
	}
	g = ours.NewGame()
	c = other.NewGame()
	a, err := g.ParseMoveSAN("Nf3")
	if err != nil {
		t.Fatal(err)
	}
	b, err := (other.AlgebraicNotation{}).Decode(c.Position(), "Nf3")
	if err != nil {
		t.Fatal(err)
	}
	if a.UCI() != (other.UCINotation{}).Encode(c.Position(), b) {
		t.Fatal("SAN moves differ")
	}
}

func BenchmarkCompare(b *testing.B) {
	for _, p := range positions {
		b.Run("LoadFENAndLegalMoves/"+p.name+"/chess-on-go", func(b *testing.B) {
			b.ReportAllocs()
			var buf [256]ours.Move
			for b.Loop() {
				g, err := ours.NewGameFromFEN(p.fen)
				if err != nil {
					b.Fatal(err)
				}
				sinkCount = len(g.LegalMovesInto(buf[:0]))
			}
		})
		b.Run("LoadFENAndLegalMoves/"+p.name+"/corentings", func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				// Parse inside the timed loop, just as in the chess-on-go case.
				option, err := other.FEN(p.fen)
				if err != nil {
					b.Fatal(err)
				}
				g := other.NewGame(option)
				sinkCount = len(g.Position().ValidMovesUnsafe())
			}
		})
		b.Run("Perft/"+p.name+"/chess-on-go", func(b *testing.B) {
			board, err := ours.NewSearchBoard(p.fen)
			if err != nil {
				b.Fatal(err)
			}
			board.Perft(p.depth) // Warm reusable undo storage outside measurement.
			b.ReportAllocs()
			for b.Loop() {
				sinkNodes = board.Perft(p.depth)
			}
		})
		b.Run("Perft/"+p.name+"/corentings", func(b *testing.B) {
			pos := loadOther(b, p.fen).Position()
			otherPerft(pos, p.depth) // Warm the root cache outside measurement.
			b.ReportAllocs()
			for b.Loop() {
				sinkNodes = otherPerft(pos, p.depth)
			}
		})
	}
	b.Run("FENExport/Initial/chess-on-go", func(b *testing.B) {
		g := ours.NewGame()
		b.ReportAllocs()
		for b.Loop() {
			sinkString = g.FEN()
		}
	})
	b.Run("FENExport/Initial/corentings", func(b *testing.B) {
		g := other.NewGame()
		b.ReportAllocs()
		for b.Loop() {
			sinkString = g.FEN()
		}
	})
	b.Run("CachedLegalMovesCopy/Initial/chess-on-go", func(b *testing.B) {
		g := ours.NewGame()
		g.LegalMovesInto(nil)
		b.ReportAllocs()
		for b.Loop() {
			sinkOurs = g.LegalMovesInto(nil)
		}
	})
	b.Run("CachedLegalMovesCopy/Initial/corentings", func(b *testing.B) {
		g := other.NewGame()
		g.ValidMoves()
		b.ReportAllocs()
		for b.Loop() {
			sinkOther = g.ValidMoves()
		}
	})
	b.Run("ParseSAN/Initial/chess-on-go", func(b *testing.B) {
		g := ours.NewGame()
		b.ReportAllocs()
		for b.Loop() {
			m, err := g.ParseMoveSAN("Nf3")
			if err != nil {
				b.Fatal(err)
			}
			sinkOurMove = m
		}
	})
	b.Run("ParseSAN/Initial/corentings", func(b *testing.B) {
		g := other.NewGame()
		g.ValidMoves()
		b.ReportAllocs()
		for b.Loop() {
			m, err := (other.AlgebraicNotation{}).Decode(g.Position(), "Nf3")
			if err != nil {
				b.Fatal(err)
			}
			sinkOtherMove = m
		}
	})
	b.Run("LoadPGN/15Plies/chess-on-go", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			g := &ours.Game{}
			if err := g.LoadPGN(shortPGN); err != nil {
				b.Fatal(err)
			}
			sinkOurGame = g
		}
	})
	b.Run("LoadPGN/15Plies/corentings", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			option, err := other.PGN(strings.NewReader(shortPGN))
			if err != nil {
				b.Fatal(err)
			}
			sinkOtherGame = other.NewGame(option)
		}
	})
}
