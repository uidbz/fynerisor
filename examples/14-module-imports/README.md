# Module Imports Example

This example demonstrates the new module-scoped import system introduced in v0.6.0.

## Features Demonstrated

- **Module-scoped imports**: Use `let module = import("path")` to import modules with namespace isolation
- **Module caching**: Same import path returns the same module instance
- **Module functions**: Call functions from imported modules using dot notation
- **Module exports**: All top-level variables and functions in a module are automatically exported
- **UI integration**: Use module functions in callbacks and event handlers

## How It Works

### Module Definition (string_utils.risor)

```javascript
// Define utility functions
let upper = (s) => s.to_upper()
let lower = (s) => s.to_lower()
let reverse = (s) => s.split("").reduce("", (acc, char) => char + acc)
```

All top-level `let` declarations are automatically exported and accessible via the module object.

### Using Modules (main.risor)

```javascript
// Import the module
let strUtils = import("string_utils.risor")

// Call module functions
print(strUtils.upper("hello"))  // "HELLO"
print(strUtils.lower("WORLD"))  // "world"

// Use in UI callbacks
let btn = widget.NewButton("To Upper", () => {
    output.SetText(strUtils.upper(input.Text))
})
```

## Key Differences from v0.5.x

**Old way (v0.5.x - concatenation):**
```javascript
// utils.risor
let helper = () => "help"

// main.risor  
import("utils.risor")  // Concatenates into global scope
helper()               // Direct access (no namespace)
```

**New way (v0.6.0+ - module scoping):**
```javascript
// utils.risor
let helper = () => "help"

// main.risor
let utils = import("utils.risor")  // Returns module object
utils.helper()                     // Namespaced access
```

## Benefits

1. **Namespace isolation**: No global scope pollution
2. **Clear dependencies**: Explicitly see what modules are used
3. **Name collision prevention**: Multiple modules can export same names
4. **Better IDE support**: Dot notation enables autocomplete
5. **Caching**: Modules loaded once and reused

## Limitations

Module functions should be self-contained or only reference parameters. They cannot reliably reference module-level variables due to Risor v2's closure scope limitations:

**Works:**
```javascript
let add = (a, b) => a + b  // ✓ Self-contained
```

**Doesn't work reliably:**
```javascript
let PI = 3.14159
let circleArea = (r) => PI * r * r  // ✗ References module variable
```

**Workaround:**
```javascript
let circleArea = (r, pi) => pi * r * r  // ✓ Pass as parameter
```

## HTTP(S) Imports

You can also import modules from URLs when `WithHTTPImport()` is enabled:

```go
// Go code
fw := gui.NewApp("My App",
    gui.WithHTTP(),        // For http module
    gui.WithHTTPImport(),  // For importing from URLs
)
```

```javascript
// Risor script
require(["v0.6", "@httpimport"])
let utils = import("https://cdn.example.com/utils.risor")
utils.someFunction()
```

**Security Note:** Only enable `WithHTTPImport()` for trusted scripts. Imported modules execute with full access to all enabled modules (http, sql, os, etc.). Use HTTPS in production to prevent code injection attacks.

## Running the Example

```bash
cd examples/14-module-imports
go run main.go
```

The application will show a text input field with buttons to transform text using the imported string utility functions.

## Files

- `main.go` - Go application entry point
- `main.risor` - Main script that imports and uses the module
- `string_utils.risor` - Reusable string utility module

## See Also

- Example 16-context-builder: Using modules in headless (non-GUI) contexts
- [Module System Documentation](../../docs/MODULES.md) (if available)
