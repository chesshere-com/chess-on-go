# Contributing

Thanks for improving `chess-on-go`. This package is used by
[www.chesshere.com](https://www.chesshere.com), so correctness and repeatable
verification matter.

## Development Commands

Run the normal verification set before sending changes:

```sh
make test
make race
make vet
make staticcheck
```

Run deeper correctness checks before releases or larger move-generation changes:

```sh
make perft
make fuzz-smoke
```

Run benchmark smoke checks before and after performance-sensitive changes:

```sh
make bench-smoke
```

For reviewable benchmark comparisons:

```sh
go test -run '^$' -bench='BenchmarkGenerateLegalMoves|BenchmarkPerft|BenchmarkCompare' -benchmem -count=5 ./... | tee bench-before.txt

# make the change

make bench
make benchstat
```

Use `make bench-snapshot` when a benchmark result should be kept under
`docs/benchmarks/`.

## Public API Guidelines

- Prefer stable accessors over direct `Game` field access.
- Add tests for behavior changes before changing implementation.
- Keep public APIs conservative and documented.
- Keep generated or scratch files out of the repository.
- Do not change standard chess behavior to support variants without a dedicated
  opt-in variant API.

## Correctness Expectations

Changes touching move generation, FEN, PGN, EPD, castling, en passant,
promotion, check detection, repetition, or draw rules should include focused
tests. Add known-position perft or fixture coverage when practical.
