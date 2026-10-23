package main

import (
	"errors"
	"fmt"

	chessongo "github.com/chesshere-com/chess-on-go"
)

func main() {
	game := chessongo.NewGame()
	move, err := game.ParseMoveSAN("e4")
	if err != nil {
		panic(err)
	}
	if err := game.TryMove(move); err != nil {
		panic(err)
	}
	fmt.Println(game.FEN())
	// Keep the actual legal move and undo in reverse order.
	game.UndoMove(move)
	fmt.Println("Restored:", game.FEN() == chessongo.STARTING_POSITION_FEN)

	if err := game.TryMoveUCI("e2e5"); errors.Is(err, chessongo.ErrIllegalMove) {
		fmt.Println("That move is not legal.")
	}
}
