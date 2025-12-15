
print("Initializing script state...")
var counter = 0

// Global function to be called from Go
func onTick(delta) {
    counter = counter + 1
    print("Tick:", counter, "Delta:", delta)
    return counter
}

func add(a, b) {
    return a + b
}

func slow() {
    var v = 0
    for var i = 1; i < 10000000; i = i + 1 {
        v = add(i, i%1000)
        print(".")
    }
}