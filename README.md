# icescript

A simple, embeddable, bytecode-compiled scripting language for Go.

## Features

- Go-like syntax (familiar to developers)
- Bytecode VM (stack-based, performant)
- First-class functions and closures
- Context-aware execution (cancellable/timeout support)
- Easy Go interop (inject globals, invoke script functions)
- Designed for game engines and embedded applications

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "github.com/iceisfun/icescript/compiler"
    "github.com/iceisfun/icescript/lexer"
    "github.com/iceisfun/icescript/parser"
    "github.com/iceisfun/icescript/vm"
)

func main() {
    source := `
        var x = 10
        var y = 20
        print("x + y =", x + y)
    `

    // Parse
    l := lexer.New(source)
    p := parser.New(l)
    program := p.ParseProgram()
    if len(p.Errors()) > 0 {
        panic(p.Errors())
    }

    // Compile
    c := compiler.New()
    if err := c.Compile(program); err != nil {
        panic(err)
    }

    // Execute
    machine := vm.New(c.Bytecode())
    if err := machine.Run(context.Background()); err != nil {
        panic(err)
    }
}
```

## Game Engine Pattern

icescript is designed for the "run once, invoke many" pattern common in game engines:

```go
// 1. Compile and initialize script state
machine := vm.New(bytecode)
machine.Run(ctx) // Run once to set up functions and state

// 2. Get function references
onTick, _ := machine.GetGlobal("onTick")

// 3. Invoke repeatedly (e.g., every frame)
for {
    result, err := machine.Invoke(ctx, onTick, &object.Integer{Value: deltaTime})
    // ... use result
}
```

## Go Interop

Inject values and functions from Go:

```go
// Pre-define globals before compilation
c := compiler.New()
configSym := c.SymbolTable().Define("Config")

// After creating VM, inject values
machine.SetGlobal(configSym.Index, &object.String{Value: "production"})

// Inject Go functions
machine.SetGlobal(callbackSym.Index, &object.Builtin{
    Fn: func(ctx object.BuiltinContext, args ...object.Object) object.Object {
        // Your Go code here
        return &object.Integer{Value: 42}
    },
})
```

## Documentation

- [SYNTAX.md](SYNTAX.md) - Language syntax guide
- [SPEC.md](SPEC.md) - Design specification and architecture

## Examples

See the `examples/` directory for complete working examples:

| Example | Description |
|---------|-------------|
| `01_simple_run` | Basic script execution |
| `02_interop` | Go ↔ Script state sharing |
| `03_scripting` | Fibonacci and script logic |
| `04_collections` | Arrays and maps |
| `05_integration` | Global state initialization |
| `06_invoke` | Function callbacks (game engine pattern) |
| `07_call_with_call_with_call` | Deep call stacks and closures |
| `08_error_handling` | Panic and stack traces |
| `09_std_lib` | Built-in functions demo |

Run an example:
```bash
go run ./examples/01_simple_run/ ./examples/01_simple_run/script.ice
```

## Built-in Functions

| Function | Description |
|----------|-------------|
| `len(obj)` | Length of string or array |
| `print(...args)` | Print to stdout |
| `push(arr, val)` | Append to array (mutating) |
| `keys(hash)` | Get all keys from a hash |
| `contains(obj, val)` | Check membership in array, string, or hash |
| `panic(msg)` | Trigger runtime error |
| `sqrt(x)` | Square root |
| `hypot(x1, y1, x2, y2)` | Euclidean distance between two points |
| `atan2(y, x)` | Arctangent of y/x |
| `equalFold(s1, s2)` | Case-insensitive string comparison |
| `seed(val)` | Set random number generator seed |
| `random()` | Random float in [0, 1) |
| `now()` | Current time in milliseconds |
| `since(start)` | Milliseconds elapsed since start |

## License

MIT
