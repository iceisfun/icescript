//go:build ignore
// +build ignore

// Optimized Sieve of Eratosthenes in Icescript using only `for in` + break/continue.
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

	// Build Range = [0..limit] for numeric-style loops.
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
  // Globals: Range, Flags
  Flags[0] = false
  Flags[1] = false

  // Outer loop up to sqrt(limit)
  for i in Range {
    if (i < 2) { continue }
    if (i * i > limit) { break }
    if (Flags[i]) {
      // Mark multiples starting at i*i, stepping by i
      var m = i * i
      for _ in Range {
        if (m > limit) { break }
        Flags[m] = false
        m = m + i
      }
    }
  }

  // Count primes
  var count = 0
  for k in Range {
    if (k >= 2 && k <= limit && Flags[k]) {
      count = count + 1
    }
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

	// Reset flags for measured run
	for i := 0; i <= limit; i++ {
		flags[i] = icescript.VBool(true)
	}
	vm.SetGlobal("Flags", icescript.VArray(flags))

	start := time.Now()
	out, err := vm.Invoke("sieve", icescript.VInt(int64(limit)))
	if err != nil {
		panic(err)
	}
	dur := time.Since(start)

	fmt.Printf("sieve: limit=%d, total=%v, primes=%d\n", limit, dur, out.AsInt())
}
