package chessongo

import "testing"

func FuzzLoadPGNDoesNotPanic(f *testing.F) {
	seeds := []string{
		"",
		" ",
		"1. e4 e5 2. Nf3 Nc6 3. Bb5 a6",
		"[Event \"Casual Game\"]\n[Result \"*\"]\n1. e4 e5 2. Nf3 Nc6 *",
		"1. e4 {comment} e5",
		"1. e4 (1... e5) 2. Nf3",
		"invalid moves or annotations",
		"1. e4? e5! 2. Nf3+- Nc6~",
		"1.e4 1...e5 2.Nf3 2...Nc6",
		"1. e4 {punctuation ( ; inside comment} e5",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, pgn string) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("LoadPGN panicked on %q: %v", pgn, r)
			}
		}()
		g := &Game{}
		if err := g.LoadPGN(pgn); err == nil {
			exported := g.PGN()
			replayed, err := LoadPGNGame(exported)
			if err != nil {
				t.Fatalf("cannot reload exported PGN %q: %v", exported, err)
			}
			if replayed.FEN() != g.FEN() || replayed.PGNResult() != g.PGNResult() {
				t.Fatalf("PGN round-trip changed position or result for %q", pgn)
			}
		}
	})
}
