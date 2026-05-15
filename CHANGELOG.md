# Changelog

All notable user-visible changes to this package should be documented here.

Releases use semantic version tags in the form `v*.*.*`, for example
`v0.1.0`.

## v0.21.1 - 2026-05-15

- Fix `parseFENNumber` overflow: halfmove/fullmove tokens with enough digits
  to exceed `math.MaxInt` previously wrapped silently to a negative value,
  which `ToFEN` then serialized and `LoadFEN` subsequently rejected — breaking
  the `LoadFEN`/`ToFEN` round-trip. Such tokens now produce a `FENError` with
  `Field` set to `FENFieldHalfMoveClock` or `FENFieldFullMoveNumber`.

## v0.2.0 - 2026-05-12

- **Breaking:** remove deprecated public `Game` fields (`Fen`, `WhitePieces`, `BlackPieces`,
  `Whites`, `Blacks`, `Occupied`, `Squares`, `EnPassant`, `Castling`, `Ply`,
  `HalfMoves`, `FullMoves`, `Turn`, `PseudoMoves`, `LegalMoves`,
  `PositionHistory`, `ZobristHash`, `IsCheck`, `IsCheckmate`, `IsStalement`,
  `IsMaterialDraw`, `IsThreefoldRepetition`, `IsFiftyMoveRule`,
  `IsSeventyFiveMoveRule`, `IsFinished`, `History`) and the deprecated
  `E_INVALID_FEN` constant. Use the stable method API (`Game.FEN`,
  `Game.Snapshot`, `Game.BoardView`, `Game.Pieces`, `Game.PieceAt`,
  `Game.SideToMove`, `Game.CastlingRights`, `Game.EnPassantSquare`,
  `Game.HalfMoveClock`, `Game.FullMoveNumber`, `Game.PositionKey`,
  `Game.LegalMovesList`, `Game.LegalMovesInto`, `Game.Status`,
  `Game.IsStalemate`, `Game.DrawStatus`, `Game.IsTerminal`) and
  `errors.Is(err, ErrInvalidFEN)` instead.

## v0.1.1 - 2026-05-12

- Add `Game.SEE(from, to Square) int` for static exchange evaluation.

## v0.1.0 - 2026-05-07

- Added stable public accessors for board inspection, game snapshots, draw
  state, and Zobrist position keys.
- Added UCI and SAN move parsing helpers.
- Added structured FEN and illegal-move errors.
- Added PGN import/export coverage for tags, comments, NAGs, variations,
  FEN/SetUp tags, results, and line wrapping.
- Added EPD parsing and typed opcode helpers.
- Added a small CLI for FEN validation, legal moves, and perft.
- Added magic-bitboard sliding attacks and search/perft-oriented APIs.
- Added formal benchmark instructions and benchmark snapshots under
  `docs/benchmarks/`.
- Added CI coverage for tests, race, vet, staticcheck, benchmark smoke, deep
  perft, randomized invariants, and scheduled fuzzing.
