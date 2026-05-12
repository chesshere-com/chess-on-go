# Changelog

All notable user-visible changes to this package should be documented here.

Releases use semantic version tags in the form `v*.*.*`, for example
`v0.1.0`.

## Unreleased

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
