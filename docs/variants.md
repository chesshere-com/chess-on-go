# Variants

## Current Variants

- `VariantStandard`: default behavior for `NewGame`, `LoadFEN`, and `NewGameFromFEN`.
- `VariantChess960`: Fischer Random / Chess960 support, including generated starting positions, Shredder-FEN castling rights, SAN castling notation, UCI castling parsing, make/undo, hashing, and PGN `Variant` tags.
- `VariantKingOfTheHill`: standard chess movement plus a win when a king reaches `d4`, `e4`, `d5`, or `e5`.
- `VariantThreeCheck`: standard chess movement plus checks-given counters and a win after the third check.

## Chess960 Notes

Use `NewChess960Game(positionID)` for position IDs from `0` to `959`.
Position `518` is the classical `RNBQKBNR` back rank.

Chess960 castling always ends with the king and rook on their standard final
squares:

- White king side: king `g1`, rook `f1`
- White queen side: king `c1`, rook `d1`
- Black king side: king `g8`, rook `f8`
- Black queen side: king `c8`, rook `d8`

`FEN()` emits Shredder-FEN file-letter castling rights in Chess960 mode. For
example, position 518 starts with `HAha` rather than `KQkq`.

## King Of The Hill Notes

Load King of the Hill positions with `LoadFENWithVariant` or `NewGameFromFENWithVariant`:

```go
game, err := chessongo.NewGameFromFENWithVariant(fen, chessongo.VariantKingOfTheHill)
```

The game ends immediately when either king reaches one of the four center
squares. `Winner()` returns the color whose king reached the center and
`Status()` returns `GameStatusVariantWin`.

## Three-Check Notes

Three-check FEN uses a seventh field for checks given by White and Black:

```text
rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1 +0+0
```

The game ends immediately after a player gives their third check. Check counters
are included in make/undo, FEN, Zobrist hashes, PGN setup tags, and binary
serialization.

## Adding The Next Variant

Add a new `Variant` constant, then implement only the rule hooks that differ
from standard chess:

- starting position generation
- FEN castling/variant metadata
- castling specs, if castling differs
- move generation filters, if legal piece movement differs
- game-end/draw rules, if terminal conditions differ
- notation/PGN tags, if notation differs
- hash inputs, if the same board state can mean different legal positions

Keep standard chess as the default path, and add variant-specific tests before
touching shared move-generation code.
