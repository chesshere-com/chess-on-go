# Compatibility Policy

This document describes the intended public surface for `chess-on-go`.

## Current Compatibility Surface

The preferred public API is method-based:

- `Game.FEN`, `Game.SideToMove`, `Game.Status`, `Game.IsTerminal`
- `Game.Variant`, `Game.Winner`
- `Game.BoardView`, `Game.Snapshot`, `Game.PieceAt`, `Game.Pieces`
- `Game.LegalMovesInto`, `Game.TryMoveUCI`, `Game.TryMoveSAN`
- `Game.DrawStatus`, `Game.CanClaimFiftyMoveRule`,
  `Game.CanClaimThreefoldRepetition`
- `Game.PositionKey`
- `NewGameFromFENWithVariant`, `LoadFENWithVariant`
- `NewChess960Game`, `Chess960StartingFEN`, `Chess960BackRank`
- `VariantStandard`, `VariantChess960`, `VariantKingOfTheHill`,
  `VariantThreeCheck`
- `GameStatusVariantWin`

The `Game` struct still has exported fields for older callers. Those fields are
compatibility surface, but they should be treated as read-only implementation
details. Mutating them directly can create states that no legal chess position
can produce.

## Deprecated Exported Fields

Deprecated `Game` fields are expected to remain available through a future v1
line so existing callers have a migration window. New code should not use them.

If a v2 is ever introduced, these fields are candidates for unexporting or
replacement with read-only view types.

## Low-Level Constants And Lookup Tables

Some exported constants and lookup tables are historical compatibility surface.
Examples include coordinate maps, piece-rune maps, masks, and attack tables.
Prefer the documented helper APIs when writing new code:

- `ParseSquare`, `Square.String`, and `Square.UCI`
- `PieceFromRune`, `Piece.String`, and `Piece.ToRune`
- `BoardView`, `Snapshot`, and `PositionKey`

These low-level values may remain exported through v1, but callers should avoid
depending on their mutability or exact representation.

## Variant Compatibility

Standard chess remains the default for `NewGame`, `LoadFEN`, and
`NewGameFromFEN`. Variants are opt-in through `Variant` APIs.

Variant-aware FEN and PGN behavior is part of the public compatibility surface:

- Chess960 FEN uses Shredder-FEN castling file letters on export.
- Three-check FEN requires the seventh `+W+B` check-counter field.
- King of the Hill and Three-check use `GameStatusVariantWin` and `Winner()`
  for variant wins.
- `NewGame`, `LoadFEN`, and `NewGameFromFEN` remain standard-only defaults.
- Reserved future variant constants may exist before implementation, but they
  are not considered supported until `LoadFENWithVariant` accepts them.

`MarshalBinary` and `UnmarshalBinary` include the active variant and variant
state. Binary payloads are same-version persistence data, not a long-term stable
wire format. Use FEN or PGN when data must survive package upgrades or
interoperate with other tools.

## Versioning

Before the first semver tag, public API changes should still be reviewed as if
the package were consumed externally. After a v1 tag, incompatible API changes
should wait for a v2.
