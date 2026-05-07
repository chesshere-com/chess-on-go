package chessongo_test

import (
	"errors"
	"fmt"

	chessongo "github.com/chesshere-com/chess-on-go"
)

func ExampleGame_TryMoveUCI() {
	game := chessongo.NewGame()
	if err := game.TryMoveUCI("e2e4"); err != nil {
		panic(err)
	}

	fmt.Println(game.SideToMove())
	fmt.Println(game.FEN())

	// Output:
	// black
	// rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1
}

func ExampleGame_Board() {
	game := chessongo.NewGame()
	board := game.Board()
	e1, err := chessongo.ParseSquare("e1")
	if err != nil {
		panic(err)
	}

	fmt.Println(board[e1])
	fmt.Println(game.CastlingRights().String())

	// Output:
	// wK
	// KQkq
}

func ExampleGame_BoardView() {
	game := chessongo.NewGame()
	view := game.BoardView()
	e1, err := chessongo.ParseSquare("e1")
	if err != nil {
		panic(err)
	}

	piece, ok := view.PieceAt(e1)
	fmt.Println(ok, piece)
	fmt.Println(view.Pieces(chessongo.WHITE).NumberOfSetBits())

	// Output:
	// true wK
	// 16
}

func ExampleGame_Clone() {
	game := chessongo.NewGame()
	analysis := game.Clone()
	if err := analysis.TryMoveSAN("e4"); err != nil {
		panic(err)
	}

	fmt.Println(game.FEN())
	fmt.Println(analysis.FEN())

	// Output:
	// rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1
	// rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1
}

func ExampleGame_PGN() {
	game := chessongo.NewGame()
	for _, uci := range []string{"e2e4", "e7e5", "g1f3"} {
		if err := game.TryMoveUCI(uci); err != nil {
			panic(err)
		}
	}

	fmt.Println(game.PGN())

	// Output:
	// [Event "?"]
	// [Site "?"]
	// [Date "????.??.??"]
	// [Round "?"]
	// [White "?"]
	// [Black "?"]
	// [Result "*"]
	//
	// 1. e4 e5 2. Nf3 *
}

func ExampleParseEPD() {
	record, err := chessongo.ParseEPD(`rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - bm e4 d4; id "initial"; perft 1 20;`)
	if err != nil {
		panic(err)
	}

	id, _ := record.ID()
	perft, _ := record.PerftExpectations()
	fmt.Println(id)
	fmt.Println(record.BestMoves())
	fmt.Println(perft[1])

	// Output:
	// initial
	// [e4 d4]
	// 20
}

func ExampleFENError() {
	game := &chessongo.Game{}
	err := game.LoadFEN("8/8/8/8/8/8/8/8 w - - 0 1")

	var fenErr *chessongo.FENError
	fmt.Println(errors.Is(err, chessongo.ErrInvalidFEN))
	fmt.Println(errors.As(err, &fenErr), fenErr.Field)

	// Output:
	// true
	// true piece placement
}

func ExampleSearchBoard_Perft() {
	board, err := chessongo.NewSearchBoard(chessongo.STARTING_POSITION_FEN)
	if err != nil {
		panic(err)
	}

	fmt.Println(board.Perft(2))

	// Output:
	// 400
}

func ExampleGame_DrawStatus() {
	game := &chessongo.Game{}
	if err := game.LoadFEN("4k3/8/8/8/8/8/6R1/4K3 w - - 99 1"); err != nil {
		panic(err)
	}
	if err := game.TryMoveUCI("g2g3"); err != nil {
		panic(err)
	}

	draw := game.DrawStatus()
	fmt.Println(draw.CanClaimFiftyMoveRule)
	fmt.Println(draw.SeventyFiveMoveRule)

	// Output:
	// true
	// false
}
