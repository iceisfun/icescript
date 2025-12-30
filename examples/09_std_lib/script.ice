func assert(cond, msg) {
    if (!cond) {
        print("Assertion failed:", msg)
        panic(msg)
    }
}

func test() {
    return
    print("this should NOT run and should call panic next")
    panic("this is not running because the result of print is returning")
}

test()

// Math
assert(sqrt(16) == 4.0, "sqrt(16)")
assert(hypot(0, 0, 3, 4) == 5.0, "hypot(0,0,3,4)")
var atanVal = atan2(1, 1)
assert(atanVal > 0.7, "atan2(1,1) > 0.7")
assert(atanVal < 0.8, "atan2(1,1) < 0.8")

// Strings
assert(equalFold("Hello", "hello"), "equalFold case insensitive")
assert(!equalFold("Hello", "world"), "equalFold different")
assert(contains("Teamwork", "work"), "contains string")

// Collections
var arr = [1, 2, 3]
assert(contains(arr, 2), "contains array found")
assert(!contains(arr, 99), "contains array not found")

var hashVal = {"a": 1, "b": 2}
assert(contains(hashVal, "a"), "contains hash key found")
assert(!contains(hashVal, "z"), "contains hash key not found")

// RNG
seed(42)
var r1 = random()
seed(42)
var r2 = random()
assert(r1 == r2, "RNG deterministic")
assert(r1 >= 0.0, "random >= 0")
assert(r1 < 1.0, "random < 1")

// Time
var start = now()
// Busy wait simulation not easily possible without loop, but we can check existence
assert(start > 0, "now() returns > 0")
var diff = since(start)
assert(diff >= 0, "since() >= 0")
assert(
    diff >= 0,
    "since()
    >= 0"
)

print("All standard library tests passed!")
