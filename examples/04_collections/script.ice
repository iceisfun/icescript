// Arrays
print("--- Arrays ---")
var arr = [1, 2, 3, "four", true]
print("Array:", arr)
print("Index 0:", arr[0])
print("Length:", len(arr))

print("Iterating Array:")
var i = 0
for i < len(arr) {
    print("Index", i, "Value:", arr[i])
    i = i + 1
}

// Maps
print("\n--- Maps ---")
var m = {"name": "Alice", "age": 30, "admin": true}
print("Map:", m)
print("Name:", m["name"])
print("Age:", m["age"])

print("Iterating Map Keys:")
var allKeys = keys(m)
var k = 0
for k < len(allKeys) {
    var key = allKeys[k]
    var val = m[key]
    print("Key:", key, "Value:", val)
    k = k + 1
}

// Slices
print("\n--- Slices ---")
var slice1 = arr[1:3]
print("Slice [1:3]:", slice1)
var slice2 = arr[:2]
print("Slice [:2]:", slice2)
var slice3 = arr[3:]
print("Slice [3:]:", slice3) // (arr has 5 elements, so index 3, 4)
