//go:build ignore
// +build ignore

// Empty loop overhead: iterates over a prebuilt array and does nothing.
// Usage: BENCH_N=1000000 go run -tags ignore ./docs/benchmarks/01-empty-loop.go
package main

import (
    "fmt"
    "os"
    "strconv"
    "time"

    "github.com/iceisfun/icescript/icescript"
)

func getenvInt(name string, def int) int {
    if s := os.Getenv(name); s != "" {
        if n, err := strconv.Atoi(s); err == nil && n > 0 {
            return n
        }
    }
    return def
}

func main() {
    // Number of loop iterations inside the script function
    N := getenvInt("BENCH_N", 1000000)

    // Build iteration array once so the script can range over it.
    it := make([]icescript.Value, N)
    for i := 0; i < N; i++ {
        it[i] = icescript.VInt(0)
    }

    const src = `
func bench() {
  // Loop over Iter and do nothing
  for _ in Iter {
    // no-op
  }
  return 0
}
`

    vm := icescript.NewVM()
    vm.SetGlobal("Iter", icescript.VArray(it))

    if err := vm.Compile(src); err != nil {
        panic(err)
    }

    // Warm-up
    if _, err := vm.Invoke("bench"); err != nil {
        panic(err)
    }

    start := time.Now()
    if _, err := vm.Invoke("bench"); err != nil {
        panic(err)
    }
    dur := time.Since(start)

    nsPerIter := float64(dur.Nanoseconds()) / float64(N)
    fmt.Printf("empty loop: N=%d, total=%v, ns/op=%.1f\n", N, dur, nsPerIter)
}

