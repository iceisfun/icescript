# icescript Syntax Guide

This document describes the syntax of the `icescript` language.

## Comments
```go
// Single line comment
/* Multi-line
   comment */
```

## Variables
Variables are dynamically typed but declared with `var`. Constants with `const`.
```go
var name = "icescript"
var age = 1
const VERSION = "2.0"
```

## Primitive Types
- **Integers**: `1`, `-50` (supports arithmetic: `+`, `-`, `*`, `/`, `%`)
- **Floats**: `3.14`, `-0.01` (comparisons only, no arithmetic)
- **Booleans**: `true`, `false`
- **Strings**: `"double quotes only"` (supports escape sequences: `\n`, `\t`, `\r`, `\"`, `\\`)
- **Null**: `null`

Note: Arithmetic operations (`+`, `-`, `*`, `/`, `%`) only work with integers. Float arithmetic and string concatenation via `+` are not currently implemented.

## Composites

### Arrays
```go
var list = [1, 2, "three"]
list[0]     // 1
len(list)   // 3
push(list, 4)  // mutates list, adds 4

// Slicing
list[0:2]   // [1, 2]
list[:2]    // [1, 2]
list[1:]    // [2, "three"]
```

### Maps
```go
var user = {
    "name": "Alice",
    "id": 123
}
user["name"]  // "Alice"
user["id"]    // 123
keys(user)    // ["name", "id"]
```

Note: Dot notation (`user.name`) is not currently supported. Use bracket notation.

## Functions
Functions are first-class values.
```go
func add(x, y) {
    return x + y
}

var sub = func(x, y) {
    return x - y
}
```

### Closures
Closures can capture variables from their enclosing scope (read-only):
```go
func makeAdder(x) {
    return func(y) {
        return x + y  // x is captured from outer scope
    }
}

var add5 = makeAdder(5)
add5(3)  // 8
add5(10) // 15
```

Note: Captured variables are read-only. You cannot modify them from within the closure.

## Control Structures

### If / Else
```go
if x > 0 {
    print("positive")
} else {
    print("not positive")
}
```

Note: `else if` is not directly supported. Use nested structures:
```go
if x > 0 {
    print("positive")
} else {
    if x < 0 {
        print("negative")
    } else {
        print("zero")
    }
}
```

### For Loops
```go
// C-style
for var i = 0; i < 5; i = i + 1 {
    print(i)
}

// While-style
var i = 0
for i < 5 {
    print(i)
    i = i + 1
}

// Infinite loop
for {
    // ...
}
```

Note: The following are NOT currently implemented:
- `break` and `continue` statements
- `for i, v := range arr` syntax (use manual iteration instead)

### Manual Array Iteration
```go
var arr = ["a", "b", "c"]
var i = 0
for i < len(arr) {
    print("Index:", i, "Value:", arr[i])
    i = i + 1
}
```

### Manual Map Iteration
```go
var m = {"a": 1, "b": 2}
var allKeys = keys(m)
var k = 0
for k < len(allKeys) {
    var key = allKeys[k]
    var val = m[key]
    print("Key:", key, "Value:", val)
    k = k + 1
}
```

## Operators

### Arithmetic
`+`, `-`, `*`, `/`, `%`

### Comparison
`==`, `!=`, `<`, `>`, `<=`, `>=`

### Unary
`!` (logical NOT), `-` (negation)

### Assignment
`=`

### Not Implemented
- `&&` and `||` (use nested `if` statements for complex logic)
- `++` and `--` (use `i = i + 1` instead)
- Compound assignment (`+=`, `-=`, etc.)

### Workaround for AND/OR Logic
```go
// Instead of: if a && b { ... }
if a {
    if b {
        // both true
    }
}

// Instead of: if a || b { ... }
var matched = false
if a {
    matched = true
}
if b {
    matched = true
}
if matched {
    // at least one true
}
```

## Built-in Functions

### Core
```go
print(args...)      // Print to stdout
len(obj)            // Length of string or array
push(arr, val)      // Append to array (mutating)
keys(hash)          // Get all keys from hash
contains(obj, val)  // Check if array/string/hash contains value
panic(msg)          // Trigger runtime error with stack trace
```

### Math
```go
sqrt(x)                  // Square root
hypot(x1, y1, x2, y2)    // Euclidean distance between two points
atan2(y, x)              // Arctangent of y/x (radians)
```

### Strings
```go
equalFold(s1, s2)   // Case-insensitive string comparison
```

### Random
```go
seed(val)    // Set RNG seed (for reproducibility)
random()     // Random float in [0, 1)
```

### Time
```go
now()          // Current time in milliseconds (Unix epoch)
since(start)   // Milliseconds elapsed since start timestamp
```

## String Operations
```go
var s = "hello world"
len(s)                       // 11
contains(s, "wor")           // true
equalFold(s, "HELLO WORLD")  // true
s == "hello world"           // true

```

Note: String indexing (`s[0]`) and slicing (`s[0:5]`) are NOT currently supported. String concatenation via `+` is also not supported.

## Error Handling
```go
func risky() {
    panic("something went wrong")
}

// When panic is called, execution stops and a stack trace is printed
risky()  // Runtime error with stack trace showing call chain
```

## Known Limitations

The following features are NOT currently implemented:

**Operators:**
- `&&` and `||` (logical AND/OR) - use nested `if` statements
- `++` and `--` (increment/decrement) - use `i = i + 1`
- `+=`, `-=` etc. (compound assignment)
- `else if` - use `else { if ... }`

**Arithmetic:**
- Float arithmetic (`3.14 + 2.0`) - only comparisons work for floats
- String concatenation (`"a" + "b"`)
- Mixed type arithmetic (`1 + 2.5`)

**Strings:**
- String indexing (`s[0]`)
- String slicing (`s[0:5]`)

**Other:**
- `break` and `continue` in loops
- `for k, v := range collection` syntax
- Dot notation for hash access (`hash.key`) - use `hash["key"]`

- Mutable closure captures (closures are read-only)
