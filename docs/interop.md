# Go ↔︎ Icescript Interop

This guide walks through moving data and behavior between Go and Icescript.

## Binding Globals

Use `SetGlobal` to expose Go data in the script environment. Combine it with `VFromGo` or `MustVFromGo` for automatic conversion.

```go
player := Player{Name: "Hero", Life: 100}
vm.SetGlobal("Player", icescript.MustVFromGo(player))
```

Scripts can then read and write the object:

```icescript
Player.Life = Player.Life - 10
```

After running the script, pull updated data back into Go with `GetGlobal` and `ToGo`:

```go
var updated Player
if err := vm.GetGlobal("Player").ToGo(&updated); err != nil {
    log.Fatal(err)
}
```

`ToGo` works with scalars, structs (respecting `script:"alias"` tags), maps, slices, pointers, and interface{} destinations.

## Registering Host Functions

Host functions let scripts call into Go. Functions receive the current VM pointer plus evaluated arguments. Return an `icescript.Value` (or `icescript.VNull()`) and an `error` if something goes wrong.

```go
vm.RegisterHostFunc("distance", func(_ *icescript.VM, args []icescript.Value) (icescript.Value, error) {
    if len(args) != 4 {
        return icescript.VNull(), fmt.Errorf("distance expects 4 args")
    }
    ax, ay := args[0].AsFloat(), args[1].AsFloat()
    bx, by := args[2].AsFloat(), args[3].AsFloat()
    return icescript.VFloat(math.Hypot(bx-ax, by-ay)), nil
})
```

In a script:

```icescript
var d = distance(0, 0, 3, 4)
print("dist:", d)
```

Host errors propagate to the script caller with a stack trace, so you can provide rich diagnostics.

## Sharing State Between Calls

Globals live in the VM. Subsequent calls to `Invoke` continue using the same environment, making it easy to accumulate state:

```go
for i := 0; i < 10; i++ {
    if _, err := vm.Invoke("tick", icescript.VInt(int64(i))); err != nil {
        log.Fatal(err)
    }
}
```

## Manual Parsing and AST Access

If you need to inspect the parsed program (for tooling, static analysis, etc.) you can work with the parser directly:

```go
lx := icescript.New(scriptSource)
parser := icescript.NewParser(lx)
program, errs := parser.ParseProgram()
if len(errs) > 0 {
    // handle parse errors
}
fmt.Printf("Found %d functions
", len(program.Funcs))
```

The AST nodes are defined in `icescript/lex.go`. You can traverse and pretty-print them to build tooling or diagnostics.

## Stack Traces

When a runtime error occurs (`Invoke` returns an error), the message includes the failing source location and the active call stack:

```
runtime error at 12:5: division by zero
  at tick (8:3)
  at main (3:1)
```

Log or surface this text to users for easier debugging.
