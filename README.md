# icescript

icescript is a small embedded scripting language implemented in Go. It ships with a handwritten lexer, Pratt parser, and a compact interpreter so you can add lightweight scripting to Go applications without external dependencies. Arrays, objects, conditionals, host function bindings, and helpful runtime stack traces are all supported out of the box.

## Features

- Pratt-style expression parser supporting arithmetic, comparisons, logical operators, object and array literals.
- Control flow primitives: `if`/`else`, `for` loops over arrays, and early `return`.
- Loop control with `break` and `continue`.
- Seamless Go interop with `VFromGo` / `Value.ToGo` for structs, maps, slices, strings, and scalars.
- Host function bridge (`RegisterHostFunc`) so scripts can call back into Go.
- Descriptive runtime errors with function stack traces and source positions.
- Immutable host-defined constants with annotated string output (e.g. `65::AncientTunnels`).
- Built-in helpers: `sqrt`, `distance`, `sin`, `cos`, `atan`, `abs`, `len`, `contains`, `lower`, `upper`, `trim`, `sleep`

## Install

```bash
go get github.com/iceisfun/icescript
```

## Quick Start

```go
package main

import (
    "fmt"

    "github.com/iceisfun/icescript/icescript"
)

func main() {
    const src = `
func main() {
  print("hello from icescript")
  return null
}
`

    vm := icescript.NewVM()
    if err := vm.SetConstants(map[string]any{
        "AncientTunnels": 65,
    }); err != nil {
        panic(err)
    }
    vm.RegisterHostFunc("print", func(_ *icescript.VM, args []icescript.Value) (icescript.Value, error) {
        for i, v := range args {
            if i > 0 {
                fmt.Print(" ")
            }
            fmt.Print(v.String())
        }
        fmt.Println()
        return icescript.VNull(), nil
    })

    if err := vm.Compile(src); err != nil {
        panic(err)
    }

    if _, err := vm.Invoke("main"); err != nil {
        panic(err)
    }
}
```

Constants stringify as `<value>::<NAME>` so printing `AncientTunnels` yields `65::AncientTunnels`.

### Binding Go Values

```go
type Player struct {
    Name string
    Life int
    X, Y float64
}

func main() {
    script := `
func move(dx, dy) {
  Player.X = Player.X + dx
  Player.Y = Player.Y + dy
  return null
}
`

    vm := icescript.NewVM()
    if err := vm.SetConstants(map[string]any{
        "AncientTunnels": 65,
    }); err != nil {
        panic(err)
    }
    vm.SetGlobal("Player", icescript.MustVFromGo(Player{Name: "Hero", Life: 100}))

    if err := vm.Compile(script); err != nil {
        panic(err)
    }

    if _, err := vm.Invoke("move", icescript.VFloat(3), icescript.VFloat(-2)); err != nil {
        panic(err)
    }

    var out Player
    if err := vm.GetGlobal("Player").ToGo(&out); err != nil {
        panic(err)
    }

    fmt.Printf("updated player: %+v\n", out)
}
```

## Docs

- [Language guide](docs/language.md) – syntax, statements, operators.
- [Strings](docs/strings.md) – literal forms and escapes.
- [Go API](docs/go-api.md) – public Go surface area.
- [Interop](docs/interop.md) – binding globals/functions, reading stack traces.
- [Examples](docs/examples) – runnable embedding scenarios.

## Development

```bash
go test ./...
```

## License

MIT License. See the full text below.

```
MIT License

Copyright (c) 2025 iceisfun

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
```
