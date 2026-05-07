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
