# Icescript Language Basics

## Functions

Scripts are organized as top-level `func` declarations. A function lists its name and zero or more comma-separated parameter names. There are no type annotations—values are dynamically typed at runtime.

```icescript
func demo(x, y) {
  return x + y
}
```

Use `return` to leave a function early. When no explicit value is provided the runtime returns `null`.

## Variables

Declare locals with `var name = expression`. If the initializer is omitted the variable defaults to `null`. Variables live in the function scope.

```icescript
var total = 0
var mystery // -> null
```

Assignments mutate locals, globals, object fields, and array elements. Nested writes such as `actor.position.x = 42` are supported.

## Conditionals

Use `if` to branch on a truthy value. Non-zero numbers, non-empty strings, non-empty arrays/objects, and `true` are considered truthy.

```icescript
if (enemy.health <= 0) {
  print("victory!")
} else {
  print("keep fighting")
}
```

## Loops

The `for` statement iterates over array values:

```icescript
var nums = [1, 2, 3]
for n in nums {
  print(n)
}
```

Loop variables are reassigned on each iteration.

## Literals

* Numbers: decimal integers or floats (`1`, `3.14`, `2.0`).
* Booleans: `true`, `false`.
* Null: `null`.
* Arrays: `[expr, ...]`.
* Objects: `{ key: value, ... }`.
* Strings: see [`strings.md`](strings.md).

## Operators

| Category      | Operators | Notes |
|---------------|-----------|-------|
| Arithmetic    | `+`, `-`, `*`, `/`, `%` | `/` always returns a float and errors on divide-by-zero. `%` works on integers only. |
| Comparison    | `==`, `!=`, `<`, `<=`, `>`, `>=` | `==` / `!=` compare stringified values. Relational operators compare numerically. |
| Logical       | `&&`, `||` | Operands are coerced with `AsBool()`. |
| Member Access | `object.field` | Errors if the left side is not an object or the field is missing. |
| Indexing      | `array[index]` | Bounds-checked; indices are truncated to integers. |

The `+` operator concatenates strings when either operand is a string.

## Built-in Functions

The runtime provides a single built-in: `print`, which is typically supplied by the embedding Go program. Additional host functions can be registered from Go. See [`interop.md`](interop.md) for details.

## Error Handling

Runtime errors (unknown identifiers, bounds checks, division by zero, etc.) abort evaluation and surface a stack trace that includes file position information.

## Comments

* Line comments: `// until end-of-line`
* Block comments: `/* nested comments are not supported */`

## Example

```icescript
func main() {
  var total = 0
  var data = [1, 2, 3, 4]

  for n in data {
    if (n % 2 == 0) {
      total = total + n
    }
  }

  return total
}
```
