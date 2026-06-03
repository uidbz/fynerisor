# WithGlobal API Change

## Summary

`WithGlobal()` now registers custom globals as requireable modules, enabling scripts to validate dependencies with `require(["@name"])`.

## What Changed

### Before (v0.5.0)

Two separate functions:
- `WithGlobal(name, value)` - Added global variable only
- `WithModule(name, value)` - Added global AND registered for require()

**Problem:** Confusing API with overlapping functionality. Why have two functions that do almost the same thing?

### After (v0.5.1+)

One unified function:
- `WithGlobal(name, value)` - Adds global variable AND registers for require()
- `WithModule()` - **REMOVED**

## Migration Guide

### If you were using `WithModule`:

**Before:**
```go
w := fynerisor.NewApp("My App",
    fynerisor.WithModule("myapp", myAppInstance),
)
```

**After:**
```go
w := fynerisor.NewApp("My App",
    fynerisor.WithGlobal("myapp", myAppInstance),  // Just rename WithModule → WithGlobal
)
```

Scripts work identically:
```risor
require(["v1.0", "@gui", "@myapp"])
myapp.DoSomething()
```

### If you were using `WithGlobal`:

**Before (v0.5.0):**
```go
w := fynerisor.NewApp("My App",
    fynerisor.WithGlobal("config", configObj),
)
```

Scripts could use it but NOT require it:
```risor
// This worked:
let val = config.get("key")

// This FAILED with error:
require(["@config"])  // ❌ "module @config is not enabled"
```

**After (v0.5.1+):**
```go
w := fynerisor.NewApp("My App",
    fynerisor.WithGlobal("config", configObj),  // Same code
)
```

Scripts can now require it:
```risor
require(["@config"])  // ✅ Now works!
let val = config.get("key")
```

## Why This Change?

1. **Simpler API** - One function instead of two
2. **Better name** - "WithGlobal" describes what it does (adds a global)
3. **Explicit dependencies** - Scripts can declare what they need
4. **Clear errors** - "require @config not enabled" instead of "undefined: config"

## Implementation Details

The change is minimal - just one line added to `WithGlobal`:

```go
func WithGlobal(name string, value any) Option {
	return moduleOption{
		fn: func(globalsList *[]risor.Option, modules map[string]bool) {
			customGlobals := map[string]any{
				name: value,
			}
			*globalsList = append(*globalsList, risor.WithEnv(customGlobals))
			modules[name] = true  // ← This line added
		},
	}
}
```

## Breaking Changes

**None** for most users:

- If you used `WithGlobal` - behavior improved (now requireable)
- If you used `WithModule` - just rename to `WithGlobal`

Only breaks scripts that intentionally avoided requiring a global. Solution: add `require(["@name"])` to those scripts.

## Examples

### Application with custom API:

```go
type MyApp struct { /* ... */ }

func (app *MyApp) GetAttr(name string) (object.Object, bool) {
    // Implement methods...
}

myApp := &MyApp{...}
w := fynerisor.NewApp("My Application",
    fynerisor.WithGlobal("myapp", myApp),
)
```

Script:
```risor
require(["v1.0", "@gui", "@myapp"])
myapp.DoSomething()
```

### Multiple custom globals:

```go
w := fynerisor.NewApp("My App",
    fynerisor.WithGlobal("app", appInstance),
    fynerisor.WithGlobal("config", configObj),
    fynerisor.WithGlobal("db", dbConnection),
)
```

Script:
```risor
require(["v1.0", "@gui", "@app", "@config", "@db"])
// All dependencies validated at startup
```

## Timeline

- **v0.5.0** - Introduced `WithModule()` alongside `WithGlobal()`
- **v0.5.1** - Modified `WithGlobal()` to register for require(), removed `WithModule()`
