# chess-on-go

`chess-on-go` is a Go chess rules package built around bitboards. It provides
legal move generation, FEN, PGN, EPD, SAN/UCI move handling, undo support,
draw-rule tracking, binary serialization, and search-oriented traversal helpers.

This package powers [www.chesshere.com](https://www.chesshere.com).

## Features

- Standard chess legal move generation using bitboards.
- Fast sliding attacks with magic bitboards.
- FEN load/export with structured validation errors.
- SAN and UCI move parsing/formatting.
- PGN import/export with headers, comments, NAGs, variations, FEN tags, and
  result validation.
- EPD parsing with typed helpers for common opcodes such as `bm`, `am`, `id`,
  and `perft`.
- Check, checkmate, stalemate, en passant, castling, promotion, insufficient
  material, repetition, and 50/75-move-rule handling.
- Opt-in Chess960, King of the Hill, and Three-check variant support.
- Read-only public accessors, snapshots, and draw-state helpers.
- Search/perft helpers that avoid game-status bookkeeping overhead.
- Static exchange evaluation via `Game.SEE(from, to)` for material-only swap analysis on a target square.
- Optional CLI for validating FEN, listing legal moves, and running perft.

## Install

```sh
go get github.com/chesshere-com/chess-on-go
```

```go
import chessongo "github.com/chesshere-com/chess-on-go"
```

## Quick Start

```go
package main

import (
	"fmt"

	chessongo "github.com/chesshere-com/chess-on-go"
)

func main() {
	game := chessongo.NewGame()

	for _, move := range game.LegalMovesInto(nil) {
		fmt.Println(move.UCI())
	}

	if err := game.TryMoveUCI("e2e4"); err != nil {
		panic(err)
	}

	fmt.Println(game.FEN())
	fmt.Println(game.SideToMove())
}
```

## Public API

New code should use accessor methods instead of reading or mutating exported
`Game` fields directly. The fields remain exported for compatibility with older
callers, but they are deprecated and can become internally inconsistent if
callers mutate them.

Use these stable APIs for normal integration work:

- `FEN`, `SideToMove`, `HalfMoveClock`, `FullMoveNumber`, `Status`, and
  `IsTerminal` for game metadata.
- `Variant` and `Winner` for variant-aware integrations.
- `Snapshot` for a defensive value copy of the current public game state.
- `BoardView`, `PieceAt`, `Board`, `Pieces`, `PiecesOfKind`,
  `OccupiedSquares`, `EnPassantSquare`, and `CastlingRights` for board
  inspection.
- `LegalMovesInto`, `LegalMovesList`, `TryMove`, `TryMoveUCI`, `TryMoveSAN`,
  and `TryMoveFromCoords` for legal move handling.
- `DrawStatus`, `CanClaimThreefoldRepetition`, `CanClaimFiftyMoveRule`,
  `IsFivefoldRepetitionDraw`, and `IsSeventyFiveMoveRuleDraw` for draw-rule
  state.
- `PositionKey` for same-version Zobrist position comparisons and hash-table
  keys.
- `Clone` when analysis code needs to branch without changing the original
  game.

## Load And Export FEN

```go
game, err := chessongo.NewGameFromFEN("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1")
if err != nil {
	return err
}

fmt.Println(game.FEN())
fmt.Println(game.Status())
```

`LoadFEN` validates the position before replacing the current game state.
Invalid board shapes, missing kings, illegal castling rights, malformed
en-passant fields, pawns on promotion ranks, invalid counters, and illegal
positions return errors. FEN errors support `errors.Is(err, ErrInvalidFEN)` and
structured field inspection with `errors.As`.

## Apply Moves

```go
game := chessongo.NewGame()

if err := game.TryMoveUCI("e2e4"); err != nil {
	return err
}

if err := game.TryMoveSAN("e5"); err != nil {
	return err
}
```

`TryMoveUCI` accepts coordinate notation such as `e2e4`, `e1g1`, and `a7a8q`.
`TryMoveSAN` accepts standard algebraic notation such as `Nf3`, `O-O`, `exd5`,
and `Qh5+`. Illegal move errors support `errors.Is(err, ErrIllegalMove)`.

## Inspect A Position

```go
game := chessongo.NewGame()
view := game.BoardView()

e1, _ := chessongo.ParseSquare("e1")
piece, ok := view.PieceAt(e1)
if ok {
	fmt.Println(piece)
}

snapshot := game.Snapshot()
fmt.Println(snapshot.PositionKey)
fmt.Println(len(snapshot.LegalMoves))
```

`BoardView` and `Snapshot` return value copies. Mutating their returned board or
move slices does not mutate the game.

## Draw And Game Status

```go
status := game.Status()
draw := game.DrawStatus()

fmt.Println(status)
fmt.Println(draw.CanClaimThreefoldRepetition)
fmt.Println(draw.CanClaimFiftyMoveRule)
fmt.Println(draw.FivefoldRepetition)
fmt.Println(draw.SeventyFiveMoveRule)
```

`Status` reports terminal and check states. Threefold repetition and the
50-move rule are claimable draws; fivefold repetition and the 75-move rule are
automatic draws.

## PGN

```go
game := &chessongo.Game{}
if err := game.LoadPGN("1. e4 e5 2. Nf3 Nc6 3. Bb5 a6"); err != nil {
	return err
}

fmt.Println(game.FEN())
fmt.Println(game.PGN())
```

PGN loading applies the main line, preserves tag pairs, accepts headers and FEN
tags, ignores comments, NAGs, and variations, and validates result tags against
movetext and forced terminal positions. Use `PGN`, `PGNTags`, and `PGNResult`
to export the played main line.

## EPD

```go
record, err := chessongo.ParseEPD("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - bm e4; id \"initial\";")
if err != nil {
	return err
}

fmt.Println(record.FEN)
fmt.Println(record.BestMoves())
```

Use `ParseEPD` or `LoadEPDRecords` for Extended Position Description test
suites. EPD records expose normalized FEN and opcode operands such as `bm`,
`am`, `id`, and `perft`.

## Search And Perft

For search-style traversal, use `SearchBoard` or `MakeMoveFast` /
`UndoMoveFast`. These APIs maintain board state while skipping repetition,
PGN, and game-over bookkeeping that engine-style traversal usually does not
need.

```go
board, err := chessongo.NewSearchBoard(chessongo.STARTING_POSITION_FEN)
if err != nil {
	return err
}

nodes := board.Perft(4)
fmt.Println(nodes)
```

For normal game play, use `TryMove`, `TryMoveUCI`, `TryMoveSAN`, `MakeMove`,
and `UndoMove` so draw, checkmate, repetition, and PGN state stay current.

## CLI

The repository includes a small command-line tool:

```sh
go run ./cmd/chessongo fen "4k3/8/8/8/8/8/8/4K3 w - - 0 1"
go run ./cmd/chessongo legal "4k3/8/8/8/8/8/8/4K3 w - - 0 1"
go run ./cmd/chessongo perft -depth 2 "4k3/8/8/8/8/8/8/4K3 w - - 0 1"
```

## Supported Rules

The package supports standard chess by default and variants through explicit
constructors or variant loading APIs:

```go
game, err := chessongo.NewChess960Game(518)
fen, err := chessongo.Chess960StartingFEN(0)
game, err = chessongo.NewGameFromFENWithVariant(fen, chessongo.VariantChess960)

standardFEN := chessongo.STARTING_POSITION_FEN
hill, err := chessongo.NewGameFromFENWithVariant(standardFEN, chessongo.VariantKingOfTheHill)

threeCheckFEN := chessongo.STARTING_POSITION_FEN + " +0+0"
threeCheck, err := chessongo.NewGameFromFENWithVariant(threeCheckFEN, chessongo.VariantThreeCheck)
```

Chess960 FEN export uses Shredder-FEN castling file letters so rook origins are
not lost. Three-check FEN uses a seventh `+W+B` field for checks given, such as
`+0+0`.

See [docs/variants.md](docs/variants.md) for variant-specific FEN, PGN,
status, hashing, and binary serialization notes.

Covered rule areas include:

- Legal move generation.
- Check, checkmate, and stalemate.
- King of the Hill center-square wins.
- Three-check counters and third-check wins.
- Castling legality, including attacked transit/destination squares.
- En passant legality, including discovered-check edge cases.
- Promotions.
- Insufficient material.
- Claimable threefold repetition and 50-move rule.
- Automatic fivefold repetition and 75-move rule.
- Variant-aware PGN, FEN, Zobrist hashes, snapshots, and binary serialization.

## Testing

```sh
make test
make race
make vet
make staticcheck
```

Shallow known-position perft checks run with the normal test suite. Deeper perft
checks are available with:

```sh
make perft
```

Short fuzzing and benchmark smoke checks are available with:

```sh
make fuzz-smoke
make bench-smoke
```

## Benchmarking

Benchmark output is kept in a `benchstat`-friendly format. For a reviewable
benchmark comparison:

```sh
go test -run '^$' -bench='BenchmarkGenerateLegalMoves|BenchmarkPerft|BenchmarkCompare' -benchmem -count=5 ./... | tee bench-before.txt

# make the change

make bench
make benchstat
```

Use `make bench-snapshot` to keep a dated benchmark snapshot under
`docs/benchmarks/`.

## Contributing And Releases

- See [CONTRIBUTING.md](CONTRIBUTING.md) for development commands and
  correctness expectations.
- See [docs/compatibility.md](docs/compatibility.md) for the compatibility
  policy around deprecated exported fields and historical lookup tables.
- See [docs/release.md](docs/release.md) before tagging a release.
- Track user-visible changes in [CHANGELOG.md](CHANGELOG.md).

## Development Notes

- Prefer stable accessors over mutable `Game` fields.
- Prefer `LegalMovesInto` when reusing caller-owned buffers.
- Prefer `SearchBoard` for engine-style perft/search traversal.
- Use `PositionKey` for in-memory position keys, and FEN for persisted
  positions.
