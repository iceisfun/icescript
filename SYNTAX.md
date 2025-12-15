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
- **Integers**: `1`, `-50`
- **Floats**: `3.14`, `-0.01`
- **Booleans**: `true`, `false`
- **Strings**: `"double quotes"`, `` `backticks` ``
- **Null**: `null`

## Composites

### Arrays
```go
var list = [1, 2, "three"]
list[0] // 1
len(list) // 3
```

### Maps
```go
var user = {
    "name": "Alice",
    "id": 123
}
user["name"] // "Alice"
user.id      // "Alice" (dot access sugar)
```

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

## Control Structures

### If / Else
```go
if x > 0 {
    print("positive")
} else if x < 0 {
    print("negative")
} else {
    print("zero")
}
```

### For Loops
```go
// C-style
for var i = 0; i < 5; i++ {
    print(i)
}

// While-style
var i = 0
for i < 5 {
    print(i)
    i++
}

// Range (Array)
var arr = ["a", "b"]
for i, v := range arr {
    print(i, v)
}

// Range (Map)
var m = {"a": 1}
for k, v := range m {
    print(k, v)
}

// Infinite
for {
    break
}
```

## Operators
- Arithmetic: `+`, `-`, `*`, `/`, `%`
- Comparison: `==`, `!=`, `<`, `>`, `<=`, `>=`
- Logical: `&&`, `||`, `!`
- Assignment: `= `, `+=`, `-=`, `++`, `--`

## Built-ins
- `print(args...)`: Print to stdout
- `len(obj)`: Length of string, array, or map
- `push(arr, val)`: Add to array (mutating)
- `string(val)`: Convert to string
- `int(val)`: Convert to int
- `float(val)`: Convert to float
