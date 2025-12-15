
print("Initializing script state...")
var counter = 0

// Global function to be called from Go
var onTick = func(delta) {
    counter = counter + 1
    print("Tick:", counter, "Delta:", delta)
    return counter
}

var add = func(a, b) {
    return a + b
}

var slow = func() {
    var v = 0
    for var i = 1; i < 10000000; i = i + 1 {
        v = add(i, i%1000)
        print(".")
    }
}