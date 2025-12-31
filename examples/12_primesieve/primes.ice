print("--- Prime Sieve ---")
var start = now()
var n = 5000000
print("Finding primes up to:", n)

// Initialize array with true
// We need indices 0 to n, so size n+1
var isPrime = []
push(isPrime, true, n + 1)

isPrime[0] = false
isPrime[1] = false

var p = 2
for p * p < n + 1 {
    if isPrime[p] {
        // Mark multiples
        var multiple = p * p
        for multiple < n + 1 {
            isPrime[multiple] = false
            multiple = multiple + p
        }
    }
    p = p + 1
}

print("Primes:")
var count = 0
var i = 0
for i < n + 1 {
    if isPrime[i] {
        count = count + 1
    }
    i = i + 1
}
print("Total count:", count, "in", since(start), "ms")
