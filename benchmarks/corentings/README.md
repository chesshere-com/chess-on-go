# Comparison with corentings/chess

This isolated Go module pins `github.com/corentings/chess/v2` at **v2.6.1**
(tag commit `6771e791c92afd4383708b601cc88c333cde6954`) and replaces chess-on-go
with the local repository. It does not add a package runtime dependency.

```sh
cd benchmarks/corentings
go test ./...
GOMAXPROCS=1 go test -run '^TestEquivalentWorkloads$' \
  -bench '^BenchmarkCompare$' -benchmem -benchtime=250ms -count=8 .
```

The committed run is `docs/benchmarks/2026-10-24-corentings.txt`.
It measured chess-on-go at `fc6ae9c6837ff8d2748bd6e66a1bdba87df9890e`,
Go 1.27.1, darwin/arm64, Apple M5 Pro, with one logical processor.
Eight sequential samples per case; setup excluded except for load benchmarks.
The run did not randomize order or establish statistical significance.

Workloads:

- **FEN load + legal moves:** both parse and construct within the timed loop.
  chess-on-go copies cached moves into reusable storage; corentings uses its
  non-copying accessor. Constructor validation/bookkeeping differ.
- **Perft:** same positions, depths, expected leaf counts, and depth-one
  bulk-counting. chess-on-go makes/unmakes; corentings creates child positions
  with `Update` and uses `ValidMovesUnsafe`. Both roots are warmed. No PGN or
  draw history is maintained. Warm-up means zero allocations for chess-on-go
  does not include constructor/initial undo-storage allocations.
- **FEN export:** serialize the initial position to a retained string.
- **Cached legal move copy:** a new defensive slice from each library. This is
  copying, not fresh move generation. The move representations differ in size.
- **SAN:** parse Nf3 with warmed legal moves, without applying it.
- **PGN:** same unannotated 15-ply main line and final FEN. corentings retains a
  richer move tree; chess-on-go retains a main line.

Tests check the same legal move sets, expected perft counts, unchanged roots,
SAN resolution, and final PGN position. Sinks retain benchmark outputs.

To reproduce the historical engine with a later checkout of the harness, create
a separate checkout of the measured chess-on-go commit and point this module's
local `replace` at that directory. To compare changes, keep the hardware,
compiler, inputs, and timing boundaries fixed and collect repeated runs.

The website's full table and chart are generated directly from the raw samples.
The FEN-export timing ranges overlap; the report describes them as similar
instead of claiming a significant speed advantage.
