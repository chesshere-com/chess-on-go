package main

import (
	"fmt"

	chessongo "github.com/chesshere-com/chess-on-go"
)

func main() {
	board, err := chessongo.NewSearchBoard(chessongo.STARTING_POSITION_FEN)
	if err != nil {
		panic(err)
	}
	fmt.Println("Depth 4:", board.Perft(4)) // 197281

	// Keep a separate move buffer at each level of a recursive search.
	moves := board.LegalMoves(make([]chessongo.Move, 0, 256))
	for _, move := range moves {
		board.MakeMove(move)
		nodes := board.Perft(2)
		board.UndoMove(move)
		fmt.Printf("%s: %d\n", move.UCI(), nodes)
	}
}
