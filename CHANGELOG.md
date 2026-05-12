# Changelog

All notable user-visible changes to this package should be documented here.

Releases use semantic version tags in the form `v*.*.*`, for example
`v0.1.0`.

## Unreleased

- Add `Game.SEE(from, to Square) int` for static exchange evaluation.
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
