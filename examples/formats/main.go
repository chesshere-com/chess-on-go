package main

import (
	"fmt"

	chessongo "github.com/chesshere-com/chess-on-go"
)

func main() {
	game := &chessongo.Game{}
	if err := game.LoadPGN("1.e4 e5 2.Nf3 Nc6 3.Bb5 a6 *"); err != nil {
		panic(err)
	}
	fmt.Println(game.FEN())
	fmt.Println(game.PGN())

	// Binary saves the current position and repetition counts, not PGN or undo.
	data, err := game.MarshalBinary()
	if err != nil {
		panic(err)
	}
	var restored chessongo.Game
	if err := restored.UnmarshalBinary(data); err != nil {
		panic(err)
	}
	fmt.Println("Same position:", restored.FEN() == game.FEN())

	record, err := chessongo.ParseEPD("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq -" + " bm e4 d4; id \"initial\";")
	if err != nil {
		panic(err)
	}
	fmt.Println("Best moves:", record.BestMoves())
}
