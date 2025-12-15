var fib = func(x) {
	if x == 0 {
		return 0
	}
	if x == 1 {
		return 1
	}
	return fib(x - 1) + fib(x - 2)
}

var start = 10
print("Calculating fib(", start, ")...")
var result = fib(start)
print("Result:", result)
