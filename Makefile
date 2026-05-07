test:
	go test ./...
shorttest:
	go test -short ./...
race:
	go test -race ./...
vet:
	go vet ./...
staticcheck:
	staticcheck ./...
perft:
	CHESSONGO_DEEP_PERFT=1 go test -run TestPerftKnownPositionsDeep ./...
fuzz-smoke:
	go test -run '^$$' -fuzz=FuzzLoadFENDoesNotPanic -fuzztime=30s ./...
benchmark:
	go test -test.bench=".*" -test.cpu="8" ./...
bench:
	go test -run '^$$' -bench='BenchmarkGenerateLegalMoves|BenchmarkPerft|BenchmarkCompare' -benchmem -count=5 ./... | tee bench-current.txt
bench-snapshot:
	mkdir -p docs/benchmarks
	go test -run '^$$' -bench='BenchmarkGenerateLegalMoves|BenchmarkPerft|BenchmarkCompare' -benchmem -count=5 ./... | tee docs/benchmarks/$$(date +%F)-snapshot.txt
bench-short:
	go test -run '^$$' -bench='BenchmarkGenerateLegalMoves|BenchmarkPerft' -benchmem -benchtime=500ms -count=1 ./...
bench-smoke:
	go test -run '^$$' -bench='BenchmarkGenerateLegalMoves|BenchmarkPerft/InitialDepth4' -benchmem -benchtime=100ms -count=1 ./... | tee bench-smoke.txt
	awk 'BEGIN { found_gen=0; found_perft=0 } /BenchmarkGenerateLegalMoves\// { found_gen=1; if ($$3 > 5000) { printf("legal move generation benchmark too slow: %s ns/op\n", $$3); exit 1 } } /BenchmarkPerft\/InitialDepth4-/ { found_perft=1; if ($$3 > 50000000) { printf("initial depth 4 perft benchmark too slow: %s ns/op\n", $$3); exit 1 } } END { if (!found_gen || !found_perft) { print "missing benchmark smoke results"; exit 1 } }' bench-smoke.txt
benchstat:
	benchstat bench-before.txt bench-current.txt
install-benchstat:
	go install golang.org/x/perf/cmd/benchstat@latest
install-staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@latest
