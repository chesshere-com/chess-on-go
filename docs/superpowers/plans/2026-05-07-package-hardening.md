# Package Hardening Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Harden the chess package for public use by improving rule correctness, validation, safe APIs, verification, and open-source polish.

**Architecture:** Keep the current bitboard representation and pseudo-move filtering approach. Add validation and safe wrappers around existing internals without a large restructure, then document the public surface.

**Tech Stack:** Go, `testing`, `testify`, GitHub Actions.

---

### Task 1: Core Rule Regression Tests

**Files:**
- Modify: `fen_test.go`
- Modify: `move-logic_test.go`
- Modify: `board_test.go`
- Modify: `binary_test.go`

- [ ] Add tests for invalid FEN shape, invalid kings/pawns/castling/en-passant/fullmove values.
- [ ] Add tests proving castling is not generated without the king and rook on their home squares.
- [ ] Add tests proving repetition ignores an en-passant field when no en-passant capture is legal.
- [ ] Add tests proving binary decoding rejects malformed versions, invalid pieces, and truncated history.
- [ ] Add tests for safe move APIs that reject illegal moves and preserve state.
- [ ] Run targeted tests and confirm the new tests fail before implementation.

### Task 2: Validation And Rule Fixes

**Files:**
- Modify: `fen.go`
- Modify: `board.go`
- Modify: `move-logic.go`
- Modify: `binary.go`

- [ ] Implement strict FEN validation while preserving successful parsing of legal positions.
- [ ] Refresh legal moves and status flags after `LoadFEN`.
- [ ] Require castling home-square king and rook occupancy before generating castling.
- [ ] Include en-passant in repetition hash only when an en-passant capture is legal.
- [ ] Version and validate binary encoding.
- [ ] Run targeted tests and then `go test ./...`.

### Task 3: Public API Cleanup

**Files:**
- Modify: `fen.go`
- Modify: `move-logic.go`
- Modify: `move.go`
- Modify: `board.go`

- [ ] Add `LoadFEN`, `ToFEN`, `IsStalemate`, `LegalMovesList`, `TryMove`, and coordinate move helpers.
- [ ] Keep older names as compatibility wrappers where possible.
- [ ] Add exported comments for new public methods.
- [ ] Run `go vet ./...`.

### Task 4: Open Source Polish

**Files:**
- Modify: `README.md`
- Modify: `Makefile`
- Modify: `benchmark_test.go`
- Modify: `perft_test.go`
- Create: `.github/workflows/ci.yml`

- [ ] Replace placeholder README with installation and usage examples.
- [ ] Move shallow perft into normal CI and keep deep perft behind `testing.Short`.
- [ ] Remove invalid empty-board benchmark FEN.
- [ ] Add CI for test, race, vet, and short perft.
- [ ] Run final verification: `gofmt`, `go test ./...`, `go test -race ./...`, `go vet ./...`, benchmarks where practical.
