//go:build ignore
// +build ignore

// Naive recursive Fibonacci in Icescript.
// Usage: N=25 go run -tags ignore ./docs/benchmarks/04-fib.go
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
	n := getenvInt("N", 25)

	const src = `
func fib(n) {
  if (n <= 1) {
  	return n
  }
  return fib(n - 1) + fib(n - 2)
}
`

	vm := icescript.NewVM()
	if err := vm.Compile(src); err != nil {
		panic(err)
	}

	start := time.Now()
	out, err := vm.Invoke("fib", icescript.VInt(int64(n)))
	if err != nil {
		panic(err)
	}
	dur := time.Since(start)

	fmt.Printf("fib(%d) = %d, total=%v\n", n, out.AsInt(), dur)
}
