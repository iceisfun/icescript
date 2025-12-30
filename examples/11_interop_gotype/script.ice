// "Config" and "NewGotype" are injected by the host.

print("Config is:", Config)

// Create a new opaque user object
var g = NewGotype()
print("Created object:", g)

// Read initial state
var state = GetState(g)
print("Initial state:", state)

// Modify state via host function
print("Incrementing state...")
IncState(g)

// Verify new state
var newState = GetState(g)
print("New state:", newState)
print("Object after mutation:", g)

if (newState != state + 1) {
    print("ERROR: State did not increment correctly!")
} else {
    print("SUCCESS: State incremented correctly.")
}
