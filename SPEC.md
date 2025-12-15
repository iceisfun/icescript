# icescript v2 Specification

## 1. Design Goals
- **Embeddable**: Designed to be hosted in Go applications.
- **Familiar**: C/Go-like syntax to leverage existing LLM training data.
- **Resilient**: Execution must be interruptible (Context awareness).
- **Interoperable**: Easy to bind host functions and share state.
- **Performant**: Bytecode VM with proper closures (upvalues).

## 2. Syntax & Semantics

The syntax is a subset of Go, with some simplifications (no explicit types for variables, dynamic typing).

### 2.1 Variables & Types
Dynamic typing. Values can be:
- `null` (or `nil`)
- `int` (64-bit)
- `float` (64-bit)
- `bool`
- `string`
- `array` (dynamic slice)
- `map` (hash map)
- `function` (first-class, supports closures)

```go
// Declarations (var optional if assigning? maybe stick to var for clarity)
var x = 10
const PI = 3.14

// Assignment
x = 20
```

### 2.2 Control Flow
Standard Go-like flow.

```go
// If
if x > 10 {
    print("big")
} else {
    print("small")
}

// Loops
// C-style
for var i = 0; i < 10; i++ { ... }

// Condition only (while)
for x < 100 { ... }

// Range (arrays/maps) - Vital for easy iteration
for i, v := range items { ... }
```

### 2.3 Functions
First-class functions with lexical scoping.

```go
func add(a, b) {
    return a + b
}

// Closures
func makeCounter() {
    var count = 0
    return func() {
        count++
        return count
    }
}
```

## 3. Embedding API (Host Side)

The Go API is the primary interface for the user.

### 3.1 Engine & VM

```go
// Engine holds compiled code/modules (thread-safe, reusable)
engine := icescript.NewEngine()
program, err := engine.Compile(src)

// VM is a lightweight execution context (one per thread/request)
vm := engine.NewVM()

// Context support for cancellation
ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
defer cancel()

// Execute
result, err := vm.Run(ctx, program)
```

### 3.2 Interop & State

```go
// Setting Global State (Pre-exec)
vm.Set("Config", map[string]any{"Enabled": true})

// Invoking specific functions
// Uses the same VM state (preserving globals modified by previous runs if reusing VM)
val, err := vm.Call(ctx, "onEvent", arg1, arg2)
```

## 4. Virtual Machine Architecture

### 4.1 Instruction Set (Bytecode)
Stack-based VM.

- **OP_CONST**: Load constant
- **OP_GET_GLOBAL / OP_SET_GLOBAL**: Global state access
- **OP_GET_LOCAL / OP_SET_LOCAL**: Stack-local access
- **OP_GET_UPVALUE**: Closure capture access
- **OP_CALL**: Function call
- **OP_CLOSURE**: Create function instance with captured upvalues

### 4.2 Cancellation
The VM loop will check `ctx.Done()` every $N$ instructions (e.g., every loop backedge or function call, or every 1024 ops) to ensure tight loops can be killed without overhead on every instruction.

## 5. Standard Library
Minimal set, expandable via host.
- `print(...)`
- `len(seq)`
- `append(arr, val)`
- `format(...)` (string formatting)
