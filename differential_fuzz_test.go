package chessongo

import (
	"reflect"
	"testing"

	notnil "github.com/notnil/chess"
)

func FuzzDifferentialMoveGen(f *testing.F) {
	// Add seed corpus
	f.Add([]byte{0, 1, 2, 3, 4})
	f.Add([]byte{12, 5, 8, 2, 14, 0, 7})
	f.Add([]byte{0})

	f.Fuzz(func(t *testing.T, moveIndices []byte) {
		g := NewGame()
		option, err := notnil.FEN(STARTING_POSITION_FEN)
		if err != nil {
			t.Fatal(err)
		}
		other := notnil.NewGame(option)

		for _, b := range moveIndices {
			if g.isFinished || other.Outcome() != notnil.NoOutcome {
				break
			}

			// Generate and compare legal moves
			ours := sortedMoveUCIs(g.legalMoves)
			theirs := notnilUCIMoves(other)

			if !reflect.DeepEqual(ours, theirs) {
				t.Fatalf("Move mismatch!\nFEN: %s\nOurs: %v\nTheirs: %v", g.ToFEN(), ours, theirs)
			}

			if len(g.legalMoves) == 0 {
				break
			}

			// Select move based on byte value modulo move count
			idx := int(b) % len(g.legalMoves)
			chosenMove := g.legalMoves[idx]

			// Apply move to our game
			g.MakeMove(chosenMove)

			// Decode and apply move to notnil/chess game
			notation := notnil.UCINotation{}
			nnMove, err := notation.Decode(other.Position(), chosenMove.UCI())
			if err != nil {
				t.Fatalf("Failed to decode move %s: %v", chosenMove.UCI(), err)
			}
			if err := other.Move(nnMove); err != nil {
				t.Fatalf("Failed to apply move %s: %v", chosenMove.UCI(), err)
			}
		}
	})
}
