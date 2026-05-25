# Variants

## Current Variants

- `VariantStandard`: default behavior for `NewGame`, `LoadFEN`, and `NewGameFromFEN`.
- `VariantChess960`: Fischer Random / Chess960 support, including generated starting positions, Shredder-FEN castling rights, SAN castling notation, UCI castling parsing, make/undo, hashing, and PGN `Variant` tags.

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
