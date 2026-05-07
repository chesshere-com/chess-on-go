package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	chessongo "github.com/chesshere-com/chess-on-go"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stderr)
		return 2
	}

	switch args[0] {
	case "fen":
		return runFEN(args[1:], stdout, stderr)
	case "legal":
		return runLegal(args[1:], stdout, stderr)
	case "perft":
		return runPerft(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", args[0])
		usage(stderr)
		return 2
	}
}

func runFEN(args []string, stdout, stderr io.Writer) int {
	fen := strings.Join(args, " ")
	if fen == "" {
		fmt.Fprintln(stderr, "missing FEN")
		return 2
	}
	game, err := chessongo.NewGameFromFEN(fen)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "valid %s\n", game.FEN())
	return 0
}

func runLegal(args []string, stdout, stderr io.Writer) int {
	fen := strings.Join(args, " ")
	if fen == "" {
		fmt.Fprintln(stderr, "missing FEN")
		return 2
	}
	game, err := chessongo.NewGameFromFEN(fen)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	moves := game.LegalMovesInto(nil)
	labels := make([]string, len(moves))
	for i, move := range moves {
		labels[i] = move.UCI()
	}
	sort.Strings(labels)
	for _, label := range labels {
		fmt.Fprintln(stdout, label)
	}
	return 0
}

func runPerft(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("perft", flag.ContinueOnError)
	fs.SetOutput(stderr)
	depth := fs.Int("depth", 1, "perft depth")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *depth < 0 {
		fmt.Fprintln(stderr, "depth must be non-negative")
		return 2
	}
	fen := strings.Join(fs.Args(), " ")
	if fen == "" {
		fmt.Fprintln(stderr, "missing FEN")
		return 2
	}
	board, err := chessongo.NewSearchBoard(fen)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "nodes %d\n", board.Perft(*depth))
	return 0
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "usage:")
	fmt.Fprintln(w, "  chessongo fen <fen>")
	fmt.Fprintln(w, "  chessongo legal <fen>")
	fmt.Fprintln(w, "  chessongo perft -depth N <fen>")
}
