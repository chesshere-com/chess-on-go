package chessongo

import (
	"testing"

	notnil "github.com/notnil/chess"
)

func BenchmarkCompareNotnilChessCachedValidMoves(b *testing.B) {
	b.ReportAllocs()
	fens := []string{
		STARTING_POSITION_FEN,
		"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
	}

	for _, fen := range fens {
		fen := fen
		b.Run(fen, func(b *testing.B) {
			option, err := notnil.FEN(fen)
			if err != nil {
				b.Fatalf("init fen: %v", err)
			}
			game := notnil.NewGame(option)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = game.ValidMoves()
			}
		})
	}
}

func BenchmarkCompareLoadFENAndGenerateLegalMoves(b *testing.B) {
	b.ReportAllocs()
	fens := []string{
		STARTING_POSITION_FEN,
		"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
	}

	for _, fen := range fens {
		fen := fen
		b.Run("chess-on-go/"+fen, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				game := &Game{}
				if err := game.LoadFEN(fen); err != nil {
					b.Fatalf("init fen: %v", err)
				}
				game.GenerateLegalMoves()
			}
		})
		b.Run("notnil-chess/"+fen, func(b *testing.B) {
			option, err := notnil.FEN(fen)
			if err != nil {
				b.Fatalf("init fen: %v", err)
			}
			for i := 0; i < b.N; i++ {
				game := notnil.NewGame(option)
				_ = game.ValidMoves()
			}
		})
	}
}
