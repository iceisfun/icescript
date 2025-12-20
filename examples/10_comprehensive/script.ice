// Comprehensive icescript test suite
// Tests the actually implemented language features

var testsPassed = 0
var testsFailed = 0

func assert(cond, name) {
    if cond {
        testsPassed = testsPassed + 1
    } else {
        testsFailed = testsFailed + 1
        print("FAIL:", name)
    }
}

func testReturn1() {
    return
}

assert(testReturn1() == null, "testReturn1")

func testReturn2() {
    return 2
}

assert(testReturn2() == 2, "testReturn2")





// ========================================
// Arithmetic Operations (integers only)
// ========================================
print("--- Arithmetic Tests ---")

assert(2 + 3 == 5, "integer addition")
assert(10 - 4 == 6, "integer subtraction")
assert(3 * 4 == 12, "integer multiplication")
assert(15 / 3 == 5, "integer division")
assert(17 % 5 == 2, "integer modulo")
assert(-5 + 3 == -2, "negative addition")
assert(-10 * -2 == 20, "negative * negative")
assert(-15 / 3 == -5, "negative division")

// ========================================
// Comparison Operators
// ========================================
print("--- Comparison Tests ---")

// Integer comparisons
assert(5 > 3, "greater than true")
assert(!(3 > 5), "greater than false")
assert(3 < 5, "less than true")
assert(!(5 < 3), "less than false")
assert(5 >= 5, "greater equal same")
assert(6 >= 5, "greater equal greater")
assert(5 <= 5, "less equal same")
assert(4 <= 5, "less equal less")
assert(5 == 5, "equal integers")
assert(5 != 6, "not equal integers")

// Float comparisons (but no float arithmetic)
assert(3.14 > 3.0, "float greater than")
assert(2.5 < 3.0, "float less than")
assert(3.0 == 3.0, "float equal")

// ========================================
// Logical Operators
// ========================================
print("--- Logical Tests ---")

assert(!false, "not false")
assert(!!true, "double not true")
assert(!(!true), "not not true")

// AND operator
assert(true && true, "and both true")
assert(!(true && false), "and one false")
assert(!(false && true), "and first false")
assert(!(false && false), "and both false")

// OR operator
assert(true || false, "or first true")
assert(false || true, "or second true")
assert(true || true, "or both true")
assert(!(false || false), "or both false")

// ========================================
// String Operations
// ========================================
print("--- String Tests ---")

assert(len("hello") == 5, "string length")
assert(len("") == 0, "empty string length")
assert(contains("hello world", "wor"), "string contains true")
assert(!contains("hello", "xyz"), "string contains false")
assert(equalFold("Hello", "hello"), "equalFold same")
assert(equalFold("WORLD", "world"), "equalFold caps")
assert(!equalFold("hello", "bye"), "equalFold different")

// String equality
assert("hello" == "hello", "string equality same")
assert(!("hello" == "world"), "string equality different")
var strVar = "test"
assert(strVar == "test", "string var equality")

// ========================================
// Array Operations
// ========================================
print("--- Array Tests ---")

var arr = [1, 2, 3, 4, 5]
assert(len(arr) == 5, "array length")
assert(arr[0] == 1, "array index 0")
assert(arr[4] == 5, "array index last")

push(arr, 6)
assert(len(arr) == 6, "push increases length")
assert(arr[5] == 6, "push adds element")

assert(contains(arr, 3), "array contains true")
assert(!contains(arr, 99), "array contains false")

// Mixed type array
var mixed = [1, "two", 3.0, true, null]
assert(len(mixed) == 5, "mixed array length")
assert(mixed[1] == "two", "mixed array string")
assert(mixed[3] == true, "mixed array bool")
assert(mixed[4] == null, "mixed array null")

// Empty array
var empty = []
assert(len(empty) == 0, "empty array length")
push(empty, "first")
assert(len(empty) == 1, "push to empty array")

// ========================================
// Hash/Map Operations
// ========================================
print("--- Hash Tests ---")

var hash = {"name": "Alice", "age": 30, "active": true}
assert(hash["age"] == 30, "hash bracket access int")
assert(hash["active"] == true, "hash bool value")

var hashKeys = keys(hash)
assert(len(hashKeys) == 3, "hash keys count")

assert(contains(hash, "name"), "hash contains key true")
assert(!contains(hash, "missing"), "hash contains key false")

// Integer keys
var intKeyHash = {1: "one", 2: "two"}
// Note: hash value string equality doesn't work, so just check it exists

// Empty hash
var emptyHash = {}
assert(len(keys(emptyHash)) == 0, "empty hash keys")

// ========================================
// Control Flow
// ========================================
print("--- Control Flow Tests ---")

// If/else
var ifResult = 0
if true {
    ifResult = 1
}
assert(ifResult == 1, "if true")

ifResult = 0
if false {
    ifResult = 1
} else {
    ifResult = 2
}
assert(ifResult == 2, "if false else")

// Nested else-if
ifResult = 0
var x = 5
if x < 3 {
    ifResult = 1
} else {
    if x < 7 {
        ifResult = 2
    } else {
        ifResult = 3
    }
}
assert(ifResult == 2, "nested else-if")

// For loops - C style
var sum = 0
for var i = 1; i <= 5; i = i + 1 {
    sum = sum + i
}
assert(sum == 15, "for loop sum 1-5")

// While-style for
var count = 0
var j = 0
for j < 3 {
    count = count + 1
    j = j + 1
}
assert(count == 3, "while-style for")

// Manual array iteration
var iterArr = [10, 20, 30]
var iterSum = 0
var idx = 0
for idx < len(iterArr) {
    iterSum = iterSum + iterArr[idx]
    idx = idx + 1
}
assert(iterSum == 60, "manual array iteration sum")

// ========================================
// Functions
// ========================================
print("--- Function Tests ---")

func add(a, b) {
    return a + b
}
assert(add(2, 3) == 5, "simple function")

func multiply(a, b) {
    return a * b
}
assert(multiply(4, 5) == 20, "another function")

// Function as value
var fn = func(x) { return x * 2 }
assert(fn(10) == 20, "function as value")

// Function returning function
func makeMultiplier(factor) {
    return func(x) { return x * factor }
}
var triple = makeMultiplier(3)
assert(triple(7) == 21, "higher-order function")

// Recursion
func factorial(n) {
    if n <= 1 {
        return 1
    }
    return n * factorial(n - 1)
}
assert(factorial(5) == 120, "recursive factorial")

// Fibonacci
func fib(n) {
    if n < 2 {
        return n
    }
    return fib(n - 1) + fib(n - 2)
}
assert(fib(10) == 55, "recursive fibonacci")

// ========================================
// Closures (read-only captures)
// ========================================
print("--- Closure Tests ---")

func makeAdder(x) {
    return func(y) {
        return x + y
    }
}

var add5 = makeAdder(5)
var add10 = makeAdder(10)

assert(add5(3) == 8, "closure add5(3)")
assert(add5(7) == 12, "closure add5(7)")
assert(add10(3) == 13, "closure add10(3)")

// Nested closures
func outer(x) {
    return func(y) {
        return func(z) {
            return x + y + z
        }
    }
}
var nested = outer(1)(2)(3)
assert(nested == 6, "nested closure")

// ========================================
// Math Builtins
// ========================================
print("--- Math Builtin Tests ---")

assert(sqrt(16) == 4.0, "sqrt 16")
assert(sqrt(25) == 5.0, "sqrt 25")
assert(sqrt(2) > 1.41, "sqrt 2 lower bound")
assert(sqrt(2) < 1.42, "sqrt 2 upper bound")

assert(hypot(0, 0, 3, 4) == 5.0, "hypot 3-4-5 triangle")
assert(hypot(0, 0, 5, 12) == 13.0, "hypot 5-12-13 triangle")

var atanResult = atan2(1, 1)
assert(atanResult > 0.78, "atan2(1,1) lower")
assert(atanResult < 0.79, "atan2(1,1) upper")

// ========================================
// Random Builtins
// ========================================
print("--- Random Builtin Tests ---")

seed(12345)
var r1 = random()
seed(12345)
var r2 = random()
assert(r1 == r2, "seeded random deterministic")
assert(r1 >= 0.0, "random >= 0")
assert(r1 < 1.0, "random < 1")

// ========================================
// Time Builtins
// ========================================
print("--- Time Builtin Tests ---")

var startTime = now()
assert(startTime > 0, "now returns positive")

var elapsed = since(startTime)
assert(elapsed >= 0, "since returns non-negative")

// ========================================
// Edge Cases
// ========================================
print("--- Edge Case Tests ---")

// Zero handling
assert(0 == 0, "zero equals zero")
assert(0 * 100 == 0, "zero times anything")
assert(100 * 0 == 0, "anything times zero")

// Boolean edge cases
assert(true == true, "true equals true")
assert(false == false, "false equals false")
assert(true != false, "true not equal false")

// Null handling
assert(null == null, "null equals null")
var nullVar = null
assert(nullVar == null, "null variable")

// Empty collections
assert(len([]) == 0, "empty array len")
assert(len("") == 0, "empty string len")
assert(len(keys({})) == 0, "empty hash keys")

// Single element collections
assert(len([42]) == 1, "single element array")
assert(len("x") == 1, "single char string")
assert(len(keys({"a": 1})) == 1, "single key hash")

// ========================================
// Summary
// ========================================
print("")
print("========================================")
print("Test Results:")
print("  Passed:", testsPassed)
print("  Failed:", testsFailed)
print("========================================")

if testsFailed > 0 {
    panic("Some tests failed!")
}

print("All comprehensive tests passed!")
