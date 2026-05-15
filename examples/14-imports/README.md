# Imports Example

This example demonstrates how to split your Risor code across multiple files using `ImportScript()`.

## Features Demonstrated

- **Code Organization**: Split code into utility functions and main application
- **ImportScript()**: Load code from local files or HTTP(S) URLs
- **Script Composition**: Imported scripts are executed before the main script
- **Shared Scope**: All imported code shares the same global scope

## Running the Example

```bash
cd examples/12-imports
go run main.go
```

## How It Works

### 1. Create Utility Functions (simple-utils.risor)

```js
// Define reusable functions
let greet = (name) => {
    return `Hello, ${name}!`
}

let square = (x) => { return x * x }
```

### 2. Import in Go Code (main.go)

```go
fyneWindow := fynerisor.NewWindow(w)

// Import utilities before loading main script
err := fyneWindow.ImportScript("simple-utils.risor")
if err != nil {
    log.Fatal(err)
}

// Load main script
script, _ := os.ReadFile("simple-app.risor")
fyneWindow.LoadScript(string(script))
fyneWindow.Execute()
```

### 3. Use Imported Functions (simple-app.risor)

```js
// Use functions defined in simple-utils.risor
let greeting = greet("World")
let result = square(5)
```

## Key Concepts

### ImportScript()

```go
// Import from local file (relative or absolute path)
err := fyneWindow.ImportScript("utils.risor")

// Import from HTTP(S) URL
err := fyneWindow.ImportScript("https://example.com/lib.risor")

// Import multiple files (executed in order)
fyneWindow.ImportScript("config.risor")
fyneWindow.ImportScript("helpers.risor")
fyneWindow.ImportScript("widgets.risor")
```

### Execution Order

1. Imported scripts are executed **first**, in the order they were imported
2. Main script (from `LoadScript()`) is executed **last**
3. All code runs in the same Risor VM with shared global scope

### Script Composition

Internally, ImportScript() and LoadScript() concatenate all scripts:

```
[imported-script-1.risor]

[imported-script-2.risor]

[main-script.risor]
```

The combined script is executed as a single unit.

## Use Cases

### Utility Libraries

Create reusable functions:

```js
// math-utils.risor
let add = (a, b) => { return a + b }
let multiply = (a, b) => { return a * b }
```

### Widget Factories

Create common UI patterns:

```js
// widget-helpers.risor
let createLabeledEntry = (label) => {
    let entry = widget.NewEntry()
    return container.NewVBox(
        widget.NewLabel(label),
        entry
    )
}
```

### Configuration

Load configuration values:

```js
// config.risor
let API_URL = "https://api.example.com"
let MAX_ITEMS = 100
let THEME_COLOR = "blue"
```

### Remote Libraries

Load code from URLs:

```js
// In Go:
fyneWindow.ImportScript("https://mysite.com/shared-utils.risor")
```

## Important Notes

- **Order Matters**: Imports are executed in the order you call `ImportScript()`
- **Shared Scope**: All scripts share the same global variables
- **No Isolation**: Variables defined in imports are visible to main script and vice versa
- **Call Before LoadScript()**: Import scripts before calling `LoadScript()` for the main application
- **HTTP Support**: Can load scripts from HTTP(S) URLs
- **Error Handling**: ImportScript() returns an error if the file/URL can't be loaded

## Comparison with Traditional Imports

Unlike traditional module systems, this is **script concatenation**:

**NOT like this** (traditional modules):
```js
// NO: This doesn't exist in Risor
import { greet } from "./utils.risor"
```

**Instead, like this** (script composition):
```go
// In Go: Concatenate scripts before execution
fyneWindow.ImportScript("utils.risor")  // Defines greet()
fyneWindow.LoadScript(mainScript)        // Uses greet()
```

## File Structure

```
examples/12-imports/
├── main.go                 # Go application with ImportScript()
├── simple-utils.risor      # Utility functions
├── simple-app.risor        # Main application
└── README.md               # This file
```

## What This Example Shows

1. **Code Organization**: Splitting Risor code across files
2. **ImportScript()**: Loading code before main script
3. **Function Reuse**: Defining utilities once, using everywhere
4. **Widget Helpers**: Creating reusable UI components
5. **Shared Scope**: All code runs in same environment

## Limitations

- No module isolation (all variables are global)
- No selective imports (entire file is executed)
- Import order must be managed manually
- No circular import detection
- Relative paths are relative to current working directory

## Best Practices

1. **Import First**: Call `ImportScript()` before `LoadScript()`
2. **Avoid Name Conflicts**: Use descriptive variable names to prevent collisions
3. **Document Dependencies**: Comment which imports are required
4. **Keep Imports Simple**: Don't put complex logic in imported files
5. **Use for Libraries**: Best suited for utility functions and helpers
