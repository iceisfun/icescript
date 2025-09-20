//go:build ignore
// +build ignore

// Sieve of Eratosthenes implemented in Icescript.
// Usage: LIMIT=30000 go run -tags ignore ./docs/benchmarks/03-sieve.go
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
    limit := getenvInt("LIMIT", 20000)

    // Build a Range array [0..limit] for index-driven loops in script
    rng := make([]icescript.Value, limit+1)
    for i := 0; i <= limit; i++ {
        rng[i] = icescript.VInt(int64(i))
    }

    // Boolean flags array, initially all true; script will mutate it.
    flags := make([]icescript.Value, limit+1)
    for i := 0; i <= limit; i++ {
        flags[i] = icescript.VBool(true)
    }

    const src = `
func sieve(limit) {
  // flags and range provided by host as globals
  // set 0 and 1 to non-prime
  Flags[0] = false
  Flags[1] = false

  // Simple sieve: for i in 2..limit, if prime, mark multiples
  for i in Range {
    if (i < 2 || i > limit) {
      continue
    }
    if (Flags[i]) {
      for j in Range {
        if (j < 2) { continue }
        var m = i * j
        if (m <= limit) {
          Flags[m] = false
        }
      }
    }
  }

  // Count primes
  var count = 0
  for i in Range {
    if (i <= limit && i >= 2 && Flags[i]) { count = count + 1 }
  }
  return count
}
`

    vm := icescript.NewVM()
    vm.SetGlobal("Range", icescript.VArray(rng))
    vm.SetGlobal("Flags", icescript.VArray(flags))

    if err := vm.Compile(src); err != nil {
        panic(err)
    }

    // Warm-up
    if _, err := vm.Invoke("sieve", icescript.VInt(int64(limit))); err != nil {
        panic(err)
    }

    // Reset flags to true for the measured run
    for i := 0; i <= limit; i++ { flags[i] = icescript.VBool(true) }
    vm.SetGlobal("Flags", icescript.VArray(flags))

    start := time.Now()
    out, err := vm.Invoke("sieve", icescript.VInt(int64(limit)))
    if err != nil {
        panic(err)
    }
    dur := time.Since(start)

    fmt.Printf("sieve: limit=%d, total=%v, primes=%d\n", limit, dur, out.AsInt())
}

