# Application Versioning Example

Demonstrates how embedding applications can set their own version for script compatibility checks.

## Overview

When you embed fynerisor in your application, you can set your application's version so that scripts check compatibility against YOUR app version, not fynerisor's version.

## How It Works

### In Your Application (main.go)

```go
const AppVersion = "1.2.3"

func main() {
    // Set your application version
    fynerisor.SetAppVersion(AppVersion)
    
    w := fynerisor.NewApp("My App")
    // ...
}
```

### In Scripts (script.risor)

```risor
// Version check is now against YOUR app (1.2.3), not fynerisor (0.4.0)
require(["v1.2"])  // Checks: app version >= 1.2
```

## Version Check Rules

### Minimum Version (>=)
```risor
require(["v1.2"])      // Requires app v1.2.0 or higher
require(["v1.2.3"])    // Requires app v1.2.3 or higher
```

### Exact Version
```risor
require(["==v1.2.3"])  // Requires exactly app v1.2.3
```

## Use Cases

**Scenario 1: Application with Scripting**
```go
// Your app: MyApp v2.5.1
fynerisor.SetAppVersion("2.5.1")

// Script requires MyApp v2.5+
require(["v2.5"])  // ✅ Passes
```

**Scenario 2: API Changes**
```go
// Your app v3.0.0 has breaking changes from v2.x
fynerisor.SetAppVersion("3.0.0")

// Old scripts check for v2.x
require(["v2.0"])  // ❌ Fails - tells user to update script

// New scripts check for v3.x
require(["v3.0"])  // ✅ Passes
```

## Benefits

1. **Independent Versioning**: Your app version is separate from fynerisor's version
2. **Script Compatibility**: Scripts can require specific app versions
3. **API Evolution**: Manage breaking changes with version requirements
4. **Clear Errors**: Users get clear messages about version mismatches

## Version Info

Scripts can access version information via the `app` object:

```risor
print(app.version)  // Your app's version (if set)
```

## Without SetAppVersion()

If you don't call `SetAppVersion()`, version checks default to fynerisor's version (0.4.0).

## Run

```bash
go run main.go
```

This example sets app version to "1.2.3", and the script requires "v1.2" or higher.
