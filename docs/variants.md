# Variants

`chess-on-go` uses standard chess by default. Variant behavior is enabled only
through explicit variant APIs so existing `NewGame`, `LoadFEN`, and
`NewGameFromFEN` callers keep standard rules.

## Supported Variants

- `VariantStandard`: normal chess and the default for existing APIs.
- `VariantChess960`: Fischer Random / Chess960 start positions and castling.
- `VariantKingOfTheHill`: standard movement plus center-square king wins.
- `VariantThreeCheck`: standard movement plus checks-given counters and
  third-check wins.

## API

Use `Variant()` to inspect a game:

```go
fmt.Println(game.Variant())
```

Use explicit constructors/loading for variants:

```go
game, err := chessongo.NewChess960Game(518)

game, err = chessongo.NewGameFromFENWithVariant(
	fen,
	chessongo.VariantKingOfTheHill,
)

err = game.LoadFENWithVariant(fen, chessongo.VariantThreeCheck)
```

Variant wins are reported with `GameStatusVariantWin`. Use `Winner()` to get
the winning color:

```go
if game.Status() == chessongo.GameStatusVariantWin {
	fmt.Println(game.Winner())
}
```

`Game.Snapshot()` includes the active `Variant`.

## Chess960

Use `NewChess960Game(positionID)` for position IDs from `0` to `959`.
Position `518` is the classical `RNBQKBNR` back rank.

```go
game, err := chessongo.NewChess960Game(518)
fen, err := chessongo.Chess960StartingFEN(0)
```

Chess960 back ranks must satisfy the standard constraints:

- bishops on opposite colors
- exactly one king between two rooks
- normal piece counts

Castling always ends on the standard final squares:

- White king side: king `g1`, rook `f1`
- White queen side: king `c1`, rook `d1`
- Black king side: king `g8`, rook `f8`
- Black queen side: king `c8`, rook `d8`

`FEN()` emits Shredder-FEN file-letter castling rights in Chess960 mode. For
example, position `518` starts with `HAha` rather than `KQkq`.

SAN castling remains `O-O` and `O-O-O`. UCI input accepts the Chess960
king-to-rook-square form.

PGN import/export uses:

```text
[Variant "Chess960"]
[SetUp "1"]
[FEN "..."]
```

## King Of The Hill

Load King of the Hill positions with `LoadFENWithVariant` or
`NewGameFromFENWithVariant`:

```go
game, err := chessongo.NewGameFromFENWithVariant(
	fen,
	chessongo.VariantKingOfTheHill,
)
```

Standard legal movement, check, castling, en passant, promotion, repetition, and
draw rules still apply. The additional win condition is immediate: a player wins
when their king occupies `d4`, `e4`, `d5`, or `e5`.

Positions loaded from FEN are checked immediately. If either king is already on
a center square, `Status()` returns `GameStatusVariantWin`, `IsTerminal()` is
true, and `Winner()` returns the winning color.

PGN import accepts variant names such as:

```text
[Variant "King of the Hill"]
```

PGN export for variant games includes `Variant`, `SetUp`, and `FEN` tags.

## Three-Check

Load Three-check games with `LoadFENWithVariant` or
`NewGameFromFENWithVariant`:

```go
game, err := chessongo.NewGameFromFENWithVariant(
	"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1 +0+0",
	chessongo.VariantThreeCheck,
)
```

Standard legal movement, check, castling, en passant, promotion, repetition, and
draw rules still apply. The additional win condition is immediate: a player wins
after giving their third check.

Three-check FEN has a required seventh field:

```text
+W+B
```

`W` is the number of checks given by White and `B` is the number of checks given
by Black. Valid values are `0` through `3`.

Examples:

```text
rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1 +0+0
4k3/8/8/8/8/8/Q7/4K3 w - - 0 1 +2+0
```

Standard FEN loading rejects the seventh field. Three-check loading rejects a
missing or malformed counter field.

Check counters are included in:

- `FEN()`
- make/undo
- `PositionKey()`
- `MarshalBinary` / `UnmarshalBinary`
- PGN setup tags

PGN import accepts variant names such as:

```text
[Variant "Three-check"]
[Variant "ThreeCheck"]
[Variant "3-check"]
[Variant "3check"]
```

When a Three-check game is exported after play, PGN includes `Variant`, `SetUp`,
and `FEN` tags so non-zero counters are preserved.

## Binary Serialization

`MarshalBinary` writes the current variant and variant state. The decoder still
accepts legacy standard-chess payloads and treats them as `VariantStandard`.

Binary data is intended for same-version persistence. Use FEN/PGN for stable
interchange with other tools or future package versions.

## Architecture Notes

The engine shares board representation, attacks, move generation, make/undo,
notation, hashing, and binary serialization across variants. Variant behavior is
implemented through internal rules hooks and small variant-specific files in the
root package.

This avoids duplicating the standard chess engine while still keeping variant
rules physically separate.

## Adding Another Variant

Add a new `Variant` constant, then implement only the hooks that differ from
standard chess:

- starting position generation
- FEN and PGN metadata
- castling specs, if castling differs
- legal move filters, if movement legality differs
- make/undo state updates, if the variant has extra state
- game-end and draw rules, if terminal conditions differ
- hash inputs, if the same board can mean different legal positions
- binary state, if the variant has extra persistent state

Keep standard chess as the default path, and add variant-specific tests before
touching shared move-generation code.
