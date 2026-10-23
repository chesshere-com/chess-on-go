# Benchmarking

This project keeps benchmark output in a `benchstat`-friendly format so performance changes can be reviewed with the same rigor as correctness changes.

## Quick Smoke

Use this before and after small local changes:

```sh
make bench-short
```

This runs the main move-generation and perft benchmarks once with a short benchmark duration.

## Reviewable Benchmark Run

For a reviewable comparison, capture a before/after pair:

```sh
go test -run '^$' -bench='BenchmarkGenerateLegalMoves|BenchmarkPerft|BenchmarkCompare' -benchmem -count=5 ./... | tee bench-before.txt

# make the change

make bench
make benchstat
```

`make bench` writes `bench-current.txt`. `make benchstat` compares `bench-before.txt` and `bench-current.txt`.

Use `make bench-snapshot` for a dated snapshot under `docs/benchmarks/`.

Install `benchstat` if needed:

```sh
make install-benchstat
```

## Snapshot Policy

Keep notable benchmark snapshots in this directory:

```text
docs/benchmarks/YYYY-MM-DD-description.txt
```

Snapshots should include:

- Git context if relevant
- Machine/CPU from benchmark output
- The exact command used
- Raw Go benchmark output
- Optional `benchstat` summary when comparing two runs

## Benchmarks To Watch

- `BenchmarkGenerateLegalMoves`: direct legal move generation, should stay zero-allocation.
- `BenchmarkPerft`: search-style traversal throughput and allocation pressure.
- `BenchmarkCompare`: context against `github.com/notnil/chess`.

Cached external move-list benchmarks are labeled separately. They are useful context, but they are not apples-to-apples regeneration benchmarks.

## October 2026 package audit

The audit covered move generation and traversal, move/undo state, cloning,
notation, FEN/PGN/EPD and binary serialization, hashing, draw and variant rules,
static exchange evaluation, the CLI, and CI commands. The main performance
changes remove repeated castling geometry construction, candidate simulations
during SAN parsing, duplicate move generation, and temporary formatting strings.

Baseline: commit `d4d0df625bb28a86bce66b9a7710898471531aab`.
Measurement: Apple M5 Pro, Go 1.27.1, darwin/arm64, `GOMAXPROCS=1`.
The original and updated test binaries were run in six alternating pairs with
150 ms per benchmark, without concurrent tests or fuzzing. The new FEN and SAN
benchmark definitions were included in both binaries. Values below are medians.

| Workload | Before | After | Speedup | Allocations before → after |
| --- | ---: | ---: | ---: | ---: |
| FEN export, initial position | 5.339 µs | 1.133 µs | 4.7× | 80 → 1 |
| Parse `Nf3` SAN, initial position | 64.964 µs | 116.8 ns | 556× | 94 → 0 |
| Format `Nf3` SAN, initial position | 2.153 µs | 112.7 ns | 19.1× | 4 → 1 |
| Generate moves, initial position | 432.4 ns | 277.8 ns | 1.6× | 0 → 0 |
| Generate moves, Kiwipete | 807.8 ns | 465.7 ns | 1.7× | 0 → 0 |
| Load PGN, 15 plies | 100.60 µs | 34.60 µs | 2.9× | 162 → 58 |
| Load PGN, 50 plies | 288.8 µs | 120.9 µs | 2.4× | 480 → 136 |

These comparisons were significant at `p <= 0.004` in `benchstat`. Host timing
variance was substantial, so the numbers describe these workloads on this
machine, not a universal speed multiplier. FEN loading, bare-king generation,
and SEE did not show statistically significant timing changes in this run.

Raw samples and exact commands are in
[the baseline](2026-10-23-audit-before.txt) and
[the updated results](2026-10-23-audit-after.txt). Recompute the comparison with:

```sh
benchstat docs/benchmarks/2026-10-23-audit-before.txt docs/benchmarks/2026-10-23-audit-after.txt
```

Perft was measured separately with eight alternating pairs and a longer
one-second benchmark duration to resolve the noise in the initial samples:

| Workload | Before | After | Time reduction |
| --- | ---: | ---: | ---: |
| Initial position, depth 4 (197,281 nodes) | 4.785 ms | 3.463 ms | 27.62% |
| Kiwipete, depth 4 (4,085,603 nodes) | 102.11 ms | 50.41 ms | 50.62% |

Both comparisons have `p < 0.001`. Allocation counts remain 3 per traversal
(512 bytes for undo history growth), while move generation remains allocation-free.
See [the perft baseline](2026-10-23-perft-before.txt) and
[the updated perft results](2026-10-23-perft-after.txt).

```sh
benchstat docs/benchmarks/2026-10-23-perft-before.txt docs/benchmarks/2026-10-23-perft-after.txt
```

Correctness checks included normal and race tests, `go vet`, staticcheck,
Go 1.25.8 compatibility, all deep perft fixtures, 100 randomized games with
undo/invariant checks, and SAN/move comparisons against `notnil/chess`.
FEN, Chess960 FEN, PGN, binary, and differential move-generation fuzzing
completed more than one million executions in total without failures.

Repetition handling was checked against
[FIDE article 9.2.3](https://handbook.fide.com/chapter/E012023), which distinguishes
positions by available moves, including legal en-passant captures. Variant
material-draw behavior was cross-checked against the
[python-chess variant implementations](https://python-chess.readthedocs.io/en/latest/_modules/chess/variant.html).
