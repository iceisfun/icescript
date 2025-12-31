# icescript Specification

## 1. Design Goals

- **Embeddable**: Designed to be hosted in Go applications
- **Familiar**: C/Go-like syntax leveraging existing developer knowledge
- **Resilient**: Execution is interruptible via Go's `context.Context`
- **Interoperable**: Easy to bind host functions and share state
- **Performant**: Bytecode VM with proper closures (upvalues)

## 2. Architecture

icescript follows a traditional compilation pipeline:

```
Source Code → Lexer → Parser → Compiler → Bytecode → VM
```

### 2.1 Components

| Package | Purpose |
|---------|---------|
| `lexer` | Tokenizes source code |
| `token` | Token type definitions |
| `parser` | Builds Abstract Syntax Tree (Pratt parsing) |
| `ast` | AST node definitions |
| `compiler` | Compiles AST to bytecode |
| `opcode` | VM instruction definitions |
| `vm` | Stack-based bytecode executor |
| `object` | Runtime value types and builtins |

## 3. Type System

Dynamic typing with the following value types:

| Type | Description | Example |
|------|-------------|---------|
| `null` | Absence of value | `null` |
| `Integer` | 64-bit signed integer | `42`, `-7` |
| `Float` | 64-bit floating point | `3.14`, `-0.5` |
| `Boolean` | `true` or `false` | `true` |
| `String` | UTF-8 string | `"hello"` |
| `Array` | Dynamic array | `[1, 2, 3]` |
| `Hash` | Hash map | `{"key": "value"}` |
| `Closure` | Function with captured environment | `func(x) { return x + y }` |
| `Computed` | Host-defined behaviors | `UserObject` |

### 3.1 Equality Semantics

| Operands | Behavior |
|----------|----------|
| Primitives (`Int`, `Float`, `Bool`, `String`, `Null`) | **Value Equality**: `5 == 5` is `true`, `5 == 6` is `false`. |
| Cross-Primitive | **Type Mismatch**: `5 == "5"` is `false`. (Legacy behavior preserved) |
| Non-Primitive (Default) | **Runtime Error**: comparing arrays, hashes, or functions errors. |
| Host Objects (`UserObj`) | **Opt-in via Interface**: Errors by default. Host types implementing `ObjectEqual` interface invoke custom logic. |
| Mixed Primitive/Non-Primitive | **Runtime Error**: `UserObj == 5` errors. |

## 4. Embedding API

### 4.1 Basic Execution

```go
import (
    "context"
    "github.com/iceisfun/icescript/compiler"
    "github.com/iceisfun/icescript/lexer"
    "github.com/iceisfun/icescript/parser"
    "github.com/iceisfun/icescript/vm"
)

// Parse
l := lexer.New(source)
p := parser.New(l)
program := p.ParseProgram()
if len(p.Errors()) > 0 {
    // Handle parse errors
}

// Compile
c := compiler.New()
if err := c.Compile(program); err != nil {
    // Handle compile error
}

// Execute
machine := vm.New(c.Bytecode())
ctx := context.Background()
if err := machine.Run(ctx); err != nil {
    // Handle runtime error
}
```

### 4.2 Context Cancellation

The VM checks `ctx.Done()` every 1024 operations, enabling:
- Timeout handling with `context.WithTimeout()`
- Graceful shutdown with `context.WithCancel()`
- Per-request isolation

```go
ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
defer cancel()

err := machine.Run(ctx)
if err != nil {
    // Script timed out or was cancelled
}
```

### 4.3 State Injection (Pre-execution)

Inject Go values into script globals before execution:

```go
// 1. Create compiler and pre-define symbols
c := compiler.New()
configSym := c.SymbolTable().Define("Config")
callbackSym := c.SymbolTable().Define("MyCallback")

// 2. Compile (script can reference Config and MyCallback)
l := lexer.New(source)
p := parser.New(l)
c.Compile(p.ParseProgram())

// 3. Create VM and inject values
machine := vm.New(c.Bytecode())

machine.SetGlobal(configSym.Index, &object.String{Value: "production"})
machine.SetGlobal(callbackSym.Index, &object.Builtin{
    Fn: func(ctx object.BuiltinContext, args ...object.Object) object.Object {
        // Go code here
        return &object.Integer{Value: 42}
    },
})

// 4. Run
machine.Run(ctx)
```

### 4.4 Function Invocation (Game Engine Pattern)

Call script functions repeatedly from Go:

```go
// 1. Initialize script state (defines functions, globals)
machine := vm.New(bytecode)
machine.Run(ctx)

// 2. Get function references
onTick, err := machine.GetGlobal("onTick")
if err != nil {
    // Function not found
}

// 3. Invoke repeatedly
for {
    delta := &object.Integer{Value: int64(deltaTimeMs)}
    result, err := machine.Invoke(ctx, onTick, delta)
    if err != nil {
        // Handle error (including context cancellation)
        break
    }
    // Use result
}
```

### 4.5 Reading Globals

```go
value, err := machine.GetGlobal("counter")
if err != nil {
    // Global not found
}

// Type assert to access value
if intVal, ok := value.(*object.Integer); ok {
    fmt.Println("Counter:", intVal.Value)
}
```

## 5. Virtual Machine

### 5.1 Architecture

- **Stack-based**: Operations push/pop from a value stack
- **Frame-based**: Each function call creates a new frame
- **Limits**: 2048 stack slots, 65536 globals, 1024 call frames

### 5.2 Instruction Set (Selected)

| Category | Opcodes |
|----------|---------|
| Constants | `OpConstant`, `OpNull`, `OpTrue`, `OpFalse` |
| Arithmetic | `OpAdd`, `OpSub`, `OpMul`, `OpDiv`, `OpMod` |
| Comparison | `OpEqual`, `OpNotEqual`, `OpGreaterThan` |
| Logic | `OpBang`, `OpMinus` |
| Control | `OpJump`, `OpJumpNotTruthy` |
| Variables | `OpGetGlobal`, `OpSetGlobal`, `OpGetLocal`, `OpSetLocal` |
| Functions | `OpCall`, `OpReturn`, `OpReturnValue`, `OpClosure`, `OpGetFree` |
| Collections | `OpArray`, `OpHash`, `OpIndex`, `OpSlice` |
| Stack | `OpPop` |

### 5.3 Closures

Closures capture free variables (upvalues) at creation time:

```ice
func makeAdder(x) {
    return func(y) { return x + y }  // x is captured
}
var add5 = makeAdder(5)
add5(3)  // 8
```

The compiler identifies free variables and emits `OpClosure` with captured values. `OpGetFree` retrieves them at runtime.

## 6. Error Handling

### 6.1 Parse Errors

Collected and returned by `parser.Errors()`. Compilation should not proceed if errors exist.

### 6.2 Compile Errors

Returned by `compiler.Compile()`. Examples: undefined variables, invalid syntax.

### 6.3 Runtime Errors

The `panic(msg)` builtin triggers a runtime error with a full stack trace:

```
Runtime error: something went wrong
Stack trace:
  at fail (script.ice:3)
  at nested (script.ice:7)
  at main (script.ice:12)
```

Source maps generated during compilation map bytecode offsets to line numbers.

## 7. Standard Library

Minimal by design. Host applications should inject domain-specific functions.

### 7.1 Core Functions

| Function | Signature | Description |
|----------|-----------|-------------|
| `print` | `print(...args)` | Print to stdout |
| `len` | `len(obj) -> int` | Length of string or array |
| `push` | `push(arr, val) -> arr` | Append to array (mutating) |
| `keys` | `keys(hash) -> array` | Get all keys from hash |
| `contains` | `contains(obj, val) -> bool` | Check membership |
| `panic` | `panic(msg)` | Trigger runtime error |

### 7.2 Math Functions

| Function | Signature | Description |
|----------|-----------|-------------|
| `sqrt` | `sqrt(x) -> float` | Square root |
| `hypot` | `hypot(x1, y1, x2, y2) -> float` | Euclidean distance |
| `atan2` | `atan2(y, x) -> float` | Arctangent (radians) |

### 7.3 String Functions

| Function | Signature | Description |
|----------|-----------|-------------|
| `equalFold` | `equalFold(s1, s2) -> bool` | Case-insensitive comparison |

### 7.4 Random Functions

| Function | Signature | Description |
|----------|-----------|-------------|
| `seed` | `seed(val)` | Set RNG seed |
| `random` | `random() -> float` | Random float in [0, 1) |

### 7.5 Time Functions

| Function | Signature | Description |
|----------|-----------|-------------|
| `now` | `now() -> int` | Current time (ms since epoch) |
| `since` | `since(start) -> int` | Elapsed time (ms) |

## 8. Extending with Host Functions

Inject custom builtins for domain-specific functionality:

```go
mySym := c.SymbolTable().Define("myFunction")

machine.SetGlobal(mySym.Index, &object.Builtin{
    Fn: func(ctx object.BuiltinContext, args ...object.Object) object.Object {
        // Validate arguments
        if len(args) != 1 {
            return &object.Error{Message: "expected 1 argument"}
        }

        // Process
        str, ok := args[0].(*object.String)
        if !ok {
            return &object.Error{Message: "expected string"}
        }

        // Return result
        return &object.String{Value: strings.ToUpper(str.Value)}
    },
})
```

The `BuiltinContext` interface provides access to:
- `Rand() *rand.Rand` - Random number generator
- `Now() time.Time` - Current time
- `Get(key string) (any, bool)` - Retrieve value from context
- `Set(key string, value any)` - Store value in context
- `PrintPrefix() string` - Retrieve the configured print prefix (set via `VM.SetPrintPrefix`)

### 8.1 State Persistence

State can be persisted across builtin calls using `Get` and `Set`. This is useful for implementing iterators, long-running tasks, or complex state machines.

```go
// Example: A builtin that counts how many times it has been called
counterSym := c.SymbolTable().Define("countMe")

machine.SetGlobal(counterSym.Index, &object.Builtin{
    Fn: func(ctx object.BuiltinContext, args ...object.Object) object.Object {
        // Retrieve state
        val, ok := ctx.Get("my_counter")
        count := int64(0)
        if ok {
            count = val.(int64)
        }

        // Update state
        count++
        ctx.Set("my_counter", count)

        return &object.Integer{Value: count}
    },
})
```

## 9. Performance Considerations

- **Bytecode compilation**: Faster than tree-walking interpreters
- **Constant pool**: Deduplicates constants
- **Stack-based**: Efficient value passing without allocation
- **Preemptive cancellation**: Checks every 1024 ops (configurable overhead)
- **VM reuse**: Same VM instance can invoke multiple functions

## 10. Limitations

- Single-threaded execution (no goroutines in scripts)
- No module/import system
- No exception catching (panic terminates execution)
- No garbage collection (relies on Go's GC)
- Strings are double-quoted only (no backticks, no single quotes)
