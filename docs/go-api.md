# Go Integration API

The `github.com/iceisfun/icescript/icescript` package exposes a small API surface for embedding the interpreter inside Go programs.

## Values

```go
import "github.com/iceisfun/icescript/icescript"
```

`icescript.Value` is a tagged union representing all runtime values. Inspect the kind with the `Kind` field (one of `NullKind`, `IntKind`, `FloatKind`, `BoolKind`, `StringKind`, `ArrayKind`, `ObjectKind`). Convenience constructors build specific values:

```go
icescript.VNull()
icescript.VInt(123)
icescript.VFloat(3.14)
icescript.VBool(true)
icescript.VString("hello")
icescript.VArray([]icescript.Value{...})
icescript.VObject(map[string]icescript.Value{...})
```

Helper methods convert between native Go types and script values:

* `Value.String()` renders a human-readable version of the value.
* `Value.AsInt()`, `AsFloat()`, `AsBool()` coerce values when possible.
* `Value.ToGo(out any) error` populates a Go variable from a script value. The destination must be a pointer. Struct fields can opt into custom names using the `script:"alias"` tag or be skipped entirely with `script:"-"`.

Use `VFromGo(any)` / `MustVFromGo(any)` to wrap Go data into script values. Supported inputs include numbers, booleans, strings, byte slices (converted to strings), slices/arrays, maps with string keys, structs, and nested combinations of those types.

## The Virtual Machine

`vm := icescript.NewVM()` creates an interpreter instance. Key methods:

* `RegisterHostFunc(name string, fn HostFunc)` registers a Go function that scripts can call. A `HostFunc` has the signature `func(*icescript.VM, []icescript.Value) (icescript.Value, error)`.
* `SetGlobal(name string, value Value)` injects a global value. Reassigning the same name overwrites the previous value.
* `GetGlobal(name string) Value` fetches a global (returns `null` if missing). Combine with `ToGo` to hydrate Go structs.
* `Compile(src string) error` parses source code and prepares it for execution.
* `Invoke(name string, args ...Value) (Value, error)` calls a script function by name.

`LoadProgram(*Program)` is also available if you parse source yourself via `NewParser` / `ParseProgram` (see [`examples/05-ast-print.go`](examples/05-ast-print.go)).

## Error Reporting

Both compile-time (`Compile`) and runtime (`Invoke`, host functions) errors are returned as Go `error` values. Runtime errors include a stack trace that records script file positions and function names.

## Example

```go
vm := icescript.NewVM()
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

vm.SetGlobal("Player", icescript.MustVFromGo(playerStruct))

if err := vm.Compile(scriptSource); err != nil {
    log.Fatal(err)
}

if _, err := vm.Invoke("main"); err != nil {
    log.Fatal(err)
}
```
