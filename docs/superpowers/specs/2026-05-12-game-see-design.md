# Game.SEE — Static Exchange Evaluation Design

> Design spec for adding `Game.SEE(from, to Square) int` to package
> `chessongo`. After upstream ships this, ChessHere's
> `internal/gameanalysis/see.go` (~360 lines) deletes and
> `SeeOnDestination` calls `g.SEE(from, to)` directly.

## Goal

Provide a public, documented `Game.SEE(from, to)` method on `*Game` that
returns the static exchange evaluation of capturing on `to` with the piece
currently on `from`, in centipawns from the moving side's perspective.

## Public API

```go
// In package chessongo, new file see.go.

// SEE returns the static exchange evaluation of capturing on `to` with the
// piece currently on `from`, in centipawns from the moving side's perspective.
// Positive values mean the moving side gains material; negative values mean it
// loses material after optimal recaptures.
//
// SEE is square-based, not move-based: it derives the moving side's color from
// the piece on `from`, not from g.SideToMove. Callers passing promotion or
// en-passant moves supply the move's from/to squares; SEE detects both from
// the source piece and engine state (g.EnPassant, the rank of `to`).
//
// SEE assumes both sides recapture with their cheapest legal attacker until
// continuing would lose material. It honors absolute pins (precomputed once
// at entry) and reveals x-ray attackers as front pieces are removed from a
// working occupancy. Mid-sequence pawn captures that reach the promotion rank
// receive the Q-P bonus; promotions are always to queen.
//
// SEE does not validate move legality: it does not check that the initial
// capture is a real legal move, nor does it filter en-passant discovered
// checks. Callers should only pass moves the engine considers legal in the
// current position.
//
// Returns 0 (no error, no panic) when either square is out of range or the
// piece on `from` is empty.
func (g *Game) SEE(from, to Square) int
```

### Behavior contract

| # | Requirement |
|---|---|
| 1 | If `from` or `to` is out of range (`>= 64`), or `g.Squares[from] == EMPTY`, return 0 without panicking. |
| 2 | Mover's value is taken from `seePieceValue`. The captured piece is the piece currently at `to`, except for en passant: when the mover is a pawn and `to == g.EnPassant && g.EnPassant != 0`, the victim is a pawn one rank behind `to` from the mover's perspective. |
| 3 | Pawn captures landing on rank 1 or 8 promote. The mover's *effective value on `to`* becomes the queen value, and the initial gain receives the `Q-P` promotion bonus. Mid-sequence pawn recapturers that reach the promotion rank receive the same bonus. The choice is always queen — no under-promotion search. |
| 4 | X-ray attacker reveals are produced by mutating a working occupancy bitboard and recomputing sliding attackers against it on every step. Sliding attackers use the package's existing magic-bitboard helpers (`rookAttacks`, `bishopAttacks`). |
| 5 | Absolute pins are honored. A piece pinned against its own king along ray R may participate in the swap only if `to` lies on R. Pins are computed once at SEE entry for both sides (a one-shot snapshot) and not updated as the swap progresses. |
| 6 | When a king captures, the swap terminates immediately after recording that capture's gain. |
| 7 | The swap also terminates when the side to move has no legal attacker on `to` (after pin filtering and after the most recent occupancy update). |
| 8 | Negamax collapse: from deepest to shallowest, `if -gain[d] < gain[d-1] { gain[d-1] = -gain[d] }`. Return `gain[0]`. |
| 9 | Hard-cap depth at 32 (`gain [32]int`); break if `d >= 31` before incrementing. |

### Centipawn values

```go
var seePieceValue = [7]int{
    0,     // EMPTY
    100,   // PAWN
    320,   // KNIGHT
    330,   // BISHOP
    500,   // ROOK
    900,   // QUEEN
    20000, // KING — large enough that capturing the king never appears
           //        profitable to the opposing side in negamax.
}
```

The table is unexported. Callers depend on the *sign* of `SEE`, not the exact
numbers. Documented as fixed across releases.

## Internal architecture

**File layout:** single new file `see.go` in package `chessongo` (~250 lines
including doc). No changes to existing files. No new exported symbols
besides `Game.SEE`.

**Helpers (all unexported, all in `see.go`):**

```go
var seePieceValue = [7]int{0, 100, 320, 330, 500, 900, 20000}

// Top-level method.
func (g *Game) SEE(from, to Square) int

// Find the lowest-valued attacker of `to` from `side`, given the current
// working occupancy and pin snapshot. Returns (attackerSquare, kind, found).
// Scans in order: PAWN, KNIGHT, BISHOP, ROOK, QUEEN, KING. Uses the
// package's magic-bitboard sliding-attack helpers.
func (g *Game) seeLeastValuableAttacker(
    to Square, side Color, occ Bitboard,
    pinned Bitboard, pinRays *[64]Bitboard,
) (Square, Piece, bool)

// Per-color pin snapshot: bitboard of pinned squares plus, for each pinned
// square, the ray it may legally move along (including the pinning slider).
// Uses RAY_MASKS and nearestRayBlocker from legal_fast.go.
func (g *Game) seeComputePins(side Color) (Bitboard, [64]Bitboard)
```

**Why this shape (not direct reuse of existing helpers):**

- `attackersToWithOccupied(square, color, occ)` returns all attackers in one
  bitboard. SEE needs *least-valuable* with pin filtering applied per
  candidate. The cleanest expression is a kind-ordered scanner that ANDs
  each kind's bitboard with the attacker mask and pin-filters in place.
- `fillAbsolutePinInfo` is keyed to the side-to-move's `legalMoveInfo`. SEE
  needs pin snapshots for *both* colors, with per-square ray lookup. A
  dedicated `seeComputePins` is simpler than generalizing the legal-move
  helper, and keeps SEE independent of legal-move-generation state.
- `seeComputePins` and `seeLeastValuableAttacker` reuse the package's
  existing primitives (`RAY_MASKS`, `DIRECTION_LSB_MSP`, `nearestRayBlocker`,
  `rookAttacks`, `bishopAttacks`, `KNIGHT_ATTACKS_FROM`, `KING_ATTACKS_FROM`,
  `pawnAttackersTo`). No new lookup tables.

### Algorithm

Classic swap algorithm with one-shot pin snapshot. Inside `Game.SEE`:

1. Bounds check + empty-from guard → return 0.
2. Identify mover: `moverPiece := g.Squares[from]`, `moverColor :=
   moverPiece.Color()`, `moverKind := moverPiece.Kind()`.
3. Detect en passant: `isEP := moverKind == PAWN && to == g.EnPassant &&
   g.EnPassant != 0`. If EP, the captured-square is `to+8` for white or
   `to-8` for black; captured value is a pawn.
4. Detect promotion: `isPromo := moverKind == PAWN && (toBB & (RANK1_MASK |
   RANK8_MASK)) != 0`. If promo, add `Q-P` to the initial captured value and
   set `moverValueOnTo = Q`.
5. Initialize `occ := g.Occupied`, then clear `from` from `occ`. If EP,
   clear the EP victim square from `occ`.
6. Compute pin snapshots: `whitePinned, whitePinRays :=
   g.seeComputePins(WHITE)` and same for black.
7. `gain[0] = capturedValue`. `pieceOnTo = moverValueOnTo`. `d = 0`. `side =
   opponentOf(moverColor)`.
8. Swap loop:
   - Pick LVA via `seeLeastValuableAttacker(to, side, occ, pinned-for-side,
     pinRays-for-side)`. If none, break.
   - `d++`. `gain[d] = pieceOnTo - gain[d-1]`.
   - If `attackerKind == PAWN && (toBB & (RANK1_MASK | RANK8_MASK)) != 0`,
     add `Q-P` to `gain[d]` and set `nextPieceOnTo = Q`. Else
     `nextPieceOnTo = seePieceValue[attackerKind]`.
   - Clear the attacker's square from `occ`.
   - If `attackerKind == KING`, break (no further recaptures possible).
   - `side = opponentOf(side)`. `pieceOnTo = nextPieceOnTo`. Break if
     `d >= 31`.
9. Negamax collapse: `for d > 0 { if -gain[d] < gain[d-1] { gain[d-1] =
   -gain[d] }; d-- }`. Return `gain[0]`.

### `seeLeastValuableAttacker` details

For each kind in order PAWN, KNIGHT, BISHOP, ROOK, QUEEN, KING:

- Compute the set of squares of that kind/color that *could* attack `to`
  given the current `occ` (using magic bitboards for sliders, lookup tables
  for jumpers, `pawnAttackersTo` for pawns).
- AND with `occ` to filter out pieces already removed from the swap.
- For each set bit, apply pin filter: if `pinned & bb != 0` and
  `pinRays[sq] & toBB == 0`, skip. Return the first surviving square.

Note: the queen's attacker set is `(bishopAttacks | rookAttacks) & queens`.

### `seeComputePins` details

For one side, find each piece of that color absolutely pinned by an enemy
slider to its own king. For each pinned piece:

- `pinned |= (1 << pinnedSq)`
- `pinRays[pinnedSq] = ray-from-king-through-pinner` (the full set of
  squares on the ray from the king up to and including the pinning slider,
  excluding the king square)

Returns `(Bitboard, [64]Bitboard)`. Implementation walks all 8 directions
from the king. Reuses `nearestRayBlocker` (legal_fast.go:754) and direction
constants.

## Acceptance tests

New test file `see_test.go` using `testify/require`. All tests run under
`go test ./...`.

### Canonical table

| # | FEN | UCI | Expected | What it verifies |
|---|-----|-----|----------|------------------|
| 1 | `1k1r4/1pp4p/p7/4p3/8/P5P1/1PP4P/2K1R3 w - - 0 1` | `e1e5` | `+100` | Rook×pawn, no recapture. |
| 2 | `1k1r3q/1ppn3p/p4b2/4p3/8/P2N2P1/1PP1R1BP/2K1Q3 w - - 0 1` | `d3e5` | `-220` | Multi-step swap, lowest-valued attackers chosen in correct order. |
| 3 | `4k3/8/8/4p3/8/8/4R3/4K3 w - - 0 1` | `e2e5` | `+100` | Plain rook×pawn. |
| 4 | `4k3/8/4r3/4p3/8/8/4R3/4K3 w - - 0 1` | `e2e5` | `-400` | Rook×pawn, opposing rook recaptures; net loss = R−P. |
| 5 | `4k3/8/8/3pP3/8/8/8/4K3 w - d6 0 1` | `e5d6` | `+100` | **En passant.** FEN corrected from the original upstream-request doc, which placed the captured pawn one rank too high and was rejected by `LoadFEN`. |
| 6 | `4k3/8/8/8/8/8/3p4/4K3 b - - 0 1` | `d2d1` | `-100` | **Promotion + king recapture.** Pawn pushes to d1 (promotes to Q for SEE purposes). King on e1 recaptures the new queen. Walkthrough: `gain[0] = 0 + (Q-P) = 800`; `gain[1] = Q - 800 = 100`; negamax `gain[0] := -gain[1] = -100`. Locked at `-100`; the upstream-request doc's table value `-800` contradicts its own arithmetic and is wrong. |
| 7 | `4k3/8/4q3/8/4N3/8/4R3/4K3 w - - 0 1` | `e4e6` | `+900` | Square-based capture: SEE evaluates "knight on e4 takes queen on e6" by piece values, not by knight-move geometry. The white rook on e2 is friendly, the black king is far; no recapture. |
| 8 | `4k3/8/4q3/8/4N3/8/4r3/4K3 w - - 0 1` | `e4e6` | `+580` | Same square-based framing. Black rook on e2 becomes the recapturer once the knight is removed from `occ` (the file is clear after the e6 queen is also removed). Net = Q (900) − N (320) = 580. **Not** a true x-ray reveal — the black rook was always an attacker of e6 once the knight leaves; see test #11 for an honest x-ray. |
| 9 | any starting FEN, `from == EMPTY` square | n/a | `0` | Defensive: empty source → 0, no panic. |
| 10 | any FEN, `from = Square(64)` or `to = Square(64)` | n/a | `0` | Defensive: out-of-range → 0, no panic. |

### Edge cases beyond the canonical table

| # | FEN | UCI | Expected | What it verifies |
|---|-----|-----|----------|------------------|
| 11 | `4k3/4r3/8/4p3/8/8/4R3/4R2K w - - 0 1` | `e2e5` | `+100` | **Honest x-ray reveal.** White rook e2 × black pawn e5. Black rook e7 (always an attacker) recaptures; that costs black a rook. Then white rook on e1 **becomes** an attacker only after e2 is removed from `occ` (initially blocked by the e2 rook) — captures the black rook. No further black attackers. Walkthrough: `gain[0]=100, gain[1]=R-100=400, gain[2]=R-400=100`. Negamax: `gain[1] := -gain[2] = -100`; `gain[0] := -gain[1] = 100`. Final `+100`. |
| 12a | `1k6/8/1n6/3p4/2P5/8/8/1R2K3 w - - 0 1` | `c4d5` | `+100` | **Pinned attacker filtered out.** White pawn c4 captures black pawn d5. The would-be recapturer is the black knight on b6 (`KNIGHT_ATTACKS_FROM[d5]` includes b6). The knight is absolutely pinned along the b-file by the white rook b1 against the black king b8 — its pin ray is `{b1..b7}`, and d5 is **not** on that ray, so the pin filter excludes the knight from the swap. No other black recapturer exists, so SEE returns the initial pawn-gain `+100`. |
| 12b | `7k/8/1n6/3p4/2P5/8/8/1R2K3 w - - 0 1` | `c4d5` | `0` | **Pin-filter control.** Same position as 12a but the black king is moved to h8, so the knight on b6 is no longer pinned. It recaptures on d5, costing it the knight. `gain[0]=100, gain[1]=N-100=220`; wait — `pieceOnTo` at step 1 is the *pawn that just moved to d5*, value `P=100`. So `gain[1] = 100 - 100 = 0`. No further white recapturer; negamax `gain[0] := -gain[1] = 0`. Final `0`. Together with 12a, this pair proves the pin filter changes behavior. |
| 13 | `4k3/8/8/8/8/8/4p3/4K3 w - - 0 1` | `e1e2` | `+100` | **King-recapture terminates.** White king captures black pawn; no defenders. Single-ply swap, king terminates. |
| 14 | `1k1r4/1pp4p/p7/4p3/8/P5P1/1PP4P/2K1R3 b - - 0 1` | `e1e5` | `+100` | **Side-to-move agnostic.** Test #1 with side-to-move flipped to black. SEE derives the mover's color from `g.Squares[e1].Color()` (white) and returns the same `+100` regardless of `g.Turn`. |
| 15 | `4k2r/6P1/8/8/8/8/8/4K3 w - - 0 1` | `g7h8` | `+1300` | **Promotion capture.** Pawn on g7 captures rook on h8 and promotes. `gain[0] = R + (Q-P) = 500 + 800 = 1300`. Black king e8 is too far to recapture. Final `+1300`. |

**Implementer note on tests #11 and #12:** if any hand-walked expected
value above differs from what `SEE` produces, the bug is in the
implementation, not the fixture. Stop and debug rather than adjusting
the expected value; the canonical/edge tests have been verified by
manual walkthrough against the algorithm specified in §"Algorithm".

## Non-goals

SEE deliberately does *not*:

- Search the full legal-move list at each ply. It is a cheap heuristic, not
  a search.
- Track pin relationships that change mid-sequence (e.g., a pinner is
  captured earlier in the swap, "unpinning" a friend). The one-shot
  snapshot is intentional and matches what every other production engine
  does.
- Search under-promotions. Always queen.
- Treat check delivered by the recapturer as a reason to stop. Material-only.
- Account for back-rank tactics, defended squares, or anything that is not
  strict material exchange on the target square.
- Validate that the initial move is legal in the position. Filtering
  illegal moves (e.g., EP discovered checks) is the caller's responsibility.

## Acceptance criteria

- `(*Game).SEE(from, to Square) int` is exported, with the inline doc above.
- All test cases #1–#15 pass with the locked expected values, with #12
  finalized by the implementer per the implementer note.
- `go test ./...`, `go vet ./...`, `staticcheck ./...` all pass.
- No new exported symbols other than `Game.SEE`.
- `seePieceValue` and all helpers (`seeLeastValuableAttacker`,
  `seeComputePins`) remain unexported.
- `see.go` is a single new file at the package root. No changes to existing
  source files. Adding a one-line bullet to `README.md` and `CHANGELOG.md`
  is acceptable.
- Benchmark regression on `BenchmarkGenerateLegalMoves` and
  `BenchmarkPerft/InitialDepth4` must be ≤ 1%. SEE is not on either hot
  path; this is just a guard against accidentally importing `see.go`
  helpers into the legal-move path.

## Notes for the upstream maintainer

- The reference implementation lives at
  `/Users/jouda/projects/chesshere/internal/gameanalysis/see.go` in the
  ChessHere repo. It is a faithful port of the original
  `pkg/chessongo/see.go` from before the upstream module was split out,
  with adaptations to access only public (deprecated) fields. The upstream
  version can be simpler: it has access to magic-bitboard slider helpers
  (`rookAttacks`, `bishopAttacks`) and unexported `Bitboard` methods, so
  the slider re-walks and helper-function duplications in the port can be
  removed.
- After this ships, the follow-up ChessHere PR is mechanical: delete
  `internal/gameanalysis/see.go`, change `see(g, from, to)` to
  `g.SEE(from, to)` in `internal/gameanalysis/material.go`, and re-run
  `go test ./...`.
- A stronger correctness story than the test table is possible by adding a
  property-style test: SEE result is consistent with a brute-force
  recapture-search to depth N for randomly generated positions. The
  `random_invariant_test.go` style in this repo is a natural fit. Not
  required for this PR.
