# Strings

Icescript strings are UTF-8 byte sequences. The runtime does not enforce a particular character encoding beyond passing the bytes through to Go.

## Literal Forms

Strings can be delimited with double (`"`) or single (`'`) quotes. Escape sequences are supported inside either delimiter:

| Escape | Meaning |
|--------|---------|
| `\n`  | newline |
| `\r`  | carriage return |
| `\t`  | horizontal tab |
| `\\` | literal backslash |
| `\"` | double quote |
| `\'` | single quote |
| `\xHH` | byte value expressed as two hexadecimal digits |

Examples:

```icescript
var greeting = "hello
world"
var hex = "\x48\x69" // "Hi"
var quoted = 'He said 'hi''
```

## Concatenation

The `+` operator appends two strings. If either operand is a string and the other is a number or boolean, the runtime stringifies the non-string operand before concatenation:

```icescript
var msg = "count: " + 42
```

## Interop With Go

When a script value is converted back into Go via `Value.ToGo`, string values map onto Go `string`. Converting Go `[]byte` values to the script with `VFromGo` produces string literals as well.

## Indexing

Strings are not currently indexable or sliceable from within Icescript. Convert the string to an array in Go if per-byte processing is required.

## Null Handling

Variables default to `null`. Guard against `null` before concatenating by using `if` or providing defaults.

```icescript
if name == null {
  name = "stranger"
}
print("hello, " + name)
```
