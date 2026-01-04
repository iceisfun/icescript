// "Config" and "Callback" are NOT declared here with var.
// They are injected by the host.

print("Config is:", Config)

var result = Callback(1, 2)
print("Callback returned:", result)

if BoolFunc(2) {
    print("Correct EVEN")
} else {
    panic("failure")
}

if BoolFunc(3) {
    panic("failure")
} else {
    print("Correct ODD")
}

if BoolFunc(3)==false {
    print("Correct ODD")
} else {
    panic("failure")
}

if !BoolFunc(3) {
    print("Correct ODD")
} else {
    panic("failure")
}