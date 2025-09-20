//go:build ignore
// +build ignore

// Function call overhead: calls a simple internal add(a,b) inside a loop.
// Usage: BENCH_N=500000 go run -tags ignore ./docs/benchmarks/02-call-add.go
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
    N := getenvInt("BENCH_N", 500000)

    // Iteration array for the script to range over
    it := make([]icescript.Value, N)
    for i := 0; i < N; i++ {
        it[i] = icescript.VInt(0)
    }

    const src = `
func add(a, b) { return a + b }

func bench() {
  var sum = 0
  for _ in Iter {
    sum = add(sum, 1)
  }
  return sum
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
    out, err := vm.Invoke("bench")
    if err != nil {
        panic(err)
    }
    dur := time.Since(start)

    nsPerCall := float64(dur.Nanoseconds()) / float64(N)
    fmt.Printf("call add: N=%d, total=%v, ns/call=%.1f, result=%d\n", N, dur, nsPerCall, out.AsInt())
}

