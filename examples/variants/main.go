package main

import (
	"fmt"

	chessongo "github.com/chesshere-com/chess-on-go"
)

func main() {
	chess960, err := chessongo.NewChess960Game(518)
	if err != nil {
		panic(err)
	}
	fmt.Println(chess960.Variant(), chess960.FEN())

	hill, err := chessongo.NewGameFromFENWithVariant(
		chessongo.STARTING_POSITION_FEN,
		chessongo.VariantKingOfTheHill,
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(hill.Variant(), hill.Status())

	threeCheck, err := chessongo.NewGameFromFENWithVariant(
		chessongo.STARTING_POSITION_FEN+" +0+0",
		chessongo.VariantThreeCheck,
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(threeCheck.Variant(), threeCheck.FEN())
}
