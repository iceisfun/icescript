# Benchmarks

This folder contains small, runnable benchmark programs that measure core interpreter costs.

Run them with Go’s "ignore" tag (to opt into the example files):

- Empty loop overhead: `go run -tags ignore ./docs/benchmarks/01-empty-loop.go`
- Function call cost (`add`): `go run -tags ignore ./docs/benchmarks/02-call-add.go`
- Sieve of Eratosthenes: `go run -tags ignore ./docs/benchmarks/03-sieve.go`

Tips
- Use `BENCH_N` to set iteration counts when supported (defaults shown in each file).
- Numbers here are rough, single-run indicators; run multiple times and on an idle system for steadier results.
