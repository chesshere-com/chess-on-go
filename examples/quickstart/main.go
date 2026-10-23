package main

import (
	"fmt"

	chessongo "github.com/chesshere-com/chess-on-go"
)

func main() {
	game := chessongo.NewGame()
	for _, san := range []string{"e4", "e5", "Nf3"} {
		if err := game.TryMoveSAN(san); err != nil {
			panic(err)
		}
	}
	fmt.Println(game.FEN())
	fmt.Println("To move:", game.SideToMove())
	fmt.Println("Legal moves:", len(game.LegalMovesInto(nil)))
}
