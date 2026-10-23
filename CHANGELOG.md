# Changelog

All notable user-visible changes to this package should be documented here.

Releases use semantic version tags in the form `v*.*.*`, for example
`v0.1.0`.

## Unreleased

## v0.22.2 - 2026-10-24

- Launch the [documentation website](https://chesshere-com.github.io/chess-on-go/)
  with usage guides, API guidance, six runnable examples, and performance
  analysis. Add automated GitHub Pages builds and deployment.
- Add an isolated, reproducible benchmark comparison against
  `github.com/corentings/chess/v2` v2.6.1, including workload equivalence tests,
  raw measurements, and explanations of API and feature differences.
- Cache standard castling geometry, skip unavailable castling rights before
  building move data, and avoid copying variant rule tables in hot paths.
- Remove redundant legal-move generation and candidate simulations from SAN
  parsing. Match SAN using stack buffers, and reuse the resulting board's check
  state when recording PGN moves.
- Build FEN using one output buffer and parse its fields without a slice
  allocation. Add reproducible FEN/SAN benchmarks and before/after snapshots.
- Fix clones losing legal and pseudo-legal move lists; all mutable clone state
  is now copied independently.
- Fix Chess960 castling requests selecting an ordinary king step with the same
  destination, and avoid treating that step as SAN disambiguation.
- Hash en-passant rights only when a legal capture exists, fixing repetition
  detection with pinned pawns and discovered checks.
- Correct material draws for knight-versus-knight, King of the Hill, and
  Three-check positions.
- Fix PGN loading of attached move numbers, zero-based castling notation, and
  punctuation inside comments. Reject unknown variants and movetext after a
  result; preserve custom starting FEN on export and clear stale results on undo.
- Fix EPD parsing with variable whitespace, semicolons in quoted operands,
  escaped quotes, and unterminated quoted operands.
- Reject invalid binary positions, castling metadata, and repetition counts;
  check binary sizes without integer overflow and reject counter truncation.
  Validate the en-passant rank and vacated starting square in FEN and binary data.
- Prevent move-counter overflow, return zero for negative perft depths, and
  exclude defended king recaptures from static exchange evaluation.
- Add regression tests, binary fuzzing, and PGN export/reload fuzz checks; fix
  scheduled CI fuzzing to target one package. Correct outdated API documentation.
- Accept single-core benchmark output in the smoke check and propagate failed
  benchmark commands through output-capture pipelines.

### Upgrade notes

- Public API signatures are unchanged. Malformed FEN, PGN, EPD, and binary
  inputs that were previously accepted may now return errors.
- Corrected en-passant hashing changes position keys for some positions.
  Rebuild persisted repetition histories or hash tables when upgrading; position
  keys and binary snapshots are intended for same-version use. Use FEN or PGN
  for portable storage.

## v0.22.1 - 2026-05-31

- Added MIT License to resolve pkg.go.dev license restrictions.

## v0.22.0 - 2026-05-25

- Added explicit Chess960 support with position generation, Shredder-FEN
  castling rights, legal castling generation/execution/undo, notation/PGN
  integration, and variant-aware snapshots/hashes.
- Added King of the Hill and Three-check variants with variant-aware status,
  winners, FEN/PGN handling, hashes, undo, and binary serialization.

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
