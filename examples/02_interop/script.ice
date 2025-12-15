// "Config" and "Callback" are NOT declared here with var.
// They are injected by the host.

print("Config is:", Config)

var result = Callback(1, 2)
print("Callback returned:", result)
