print("--- Script Start ---")
print("Received Version:", version)
print("Received Config:", config)

// Logic: Use config, update version
var newVersion = version + 1
print("Updating version to:", newVersion)
version = newVersion

// Update config
// Since we don't have map mutation (m[k]=v) yet, we replace the map
var newConfig = {"env": config["env"], "status": "processed"}
config = newConfig

print("--- Script End ---")
