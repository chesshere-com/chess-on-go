package chessongo

import "testing"

func FuzzLoadFENDoesNotPanic(f *testing.F) {
	seeds := []string{
		"",
		" ",
		"8/8/8/8/8/8/8/8 w - - 0 1",
		STARTING_POSITION_FEN,
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNRR w KQkq - 0 1",
		"/////// w - - 0 1",
		"not a fen",
		"4k3/8/8/8/8/8/8/4K3 w - - 0 1",
		"4k3/8/8/8/8/8/8/4K3 x - - 0 1",
		"4k3/8/8/8/8/8/8/4K3 w K - 0 1",
		"4k3/8/8/8/8/8/8/4K3 w - z9 0 1",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, fen string) {
		defer func() {
			if recovered := recover(); recovered != nil {
				t.Fatalf("LoadFEN panicked for %q: %v", fen, recovered)
			}
		}()
		g := &Game{}
		_ = g.LoadFEN(fen)
	})
}
