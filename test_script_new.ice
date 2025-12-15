// Test optional semicolons
var a = 1
var b = 2
var c = a + b
print(c) // Expect 3

// Test short declaration
d := 4
print(d) // Expect 4

e := a + b + d
print(e) // Expect 7

// Test in loop
sum := 0
for i := 1; i <= 5; i = i + 1 {
    sum = sum + i
}
print(sum) // Expect 15

// Test function with short decl
func test() {
    x := 10
    print(x)
}
test() // Expect 10
