package chessongo

import "testing"

func FuzzUnmarshalBinary(f *testing.F) {
	f.Add([]byte{})
	for _, variant := range []Variant{VariantStandard, VariantChess960, VariantKingOfTheHill, VariantThreeCheck} {
		fen := STARTING_POSITION_FEN
		if variant == VariantThreeCheck {
			fen += " +0+0"
		}
		g, err := NewGameFromFENWithVariant(fen, variant)
		if err != nil {
			f.Fatal(err)
		}
		data, err := g.MarshalBinary()
		if err != nil {
			f.Fatal(err)
		}
		f.Add(data)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		var g Game
		if err := g.UnmarshalBinary(data); err != nil {
			return
		}
		reloaded, err := NewGameFromFENWithVariant(g.FEN(), g.Variant())
		if err != nil {
			t.Fatalf("decoded invalid position %q: %v", g.FEN(), err)
		}
		if g.PositionKey() != reloaded.PositionKey() {
			t.Fatal("binary position key differs from FEN")
		}
	})
}
