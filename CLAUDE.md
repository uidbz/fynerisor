# Claude Code Guide for Fynerisor

This file provides context and guidelines for AI assistants working on fynerisor.

## Project Overview

Fynerisor provides Risor v2 language bindings for the Fyne GUI framework, enabling cross-platform desktop applications written in Risor scripts.

**Current Version:** 0.8.0  
**Language:** Go 1.25+  
**Dependencies:** Fyne v2.8+, Risor v2.x

## Architecture

### Package Structure

```
fynerisor/
├── core/           # Headless Risor execution (no GUI deps)
├── gui/            # GUI functionality with Fyne
│   ├── widget/     # Widget wrappers
│   ├── container/  # Container wrappers
│   ├── canvas/     # Canvas object wrappers
│   ├── chart/      # Chart wrappers
│   ├── binding/    # Data binding wrappers
│   └── dialog/     # Dialog wrappers
├── modules/        # Optional Risor modules (http, sql, os, etc.)
├── cmd/fynerisor/  # CLI tool
├── examples/       # Example applications
└── docs/           # Documentation
```

### Critical Architectural Rules

1. **core package MUST have zero GUI dependencies**
   - Enables static compilation
   - Used for headless scripts, CLI tools, servers
   - Only imports: risor, standard library

2. **Shared functionality goes in core, not gui**
   - Version management → core/version.go
   - Script analysis → core/analyze.go
   - Module types → core/module.go
   - gui package imports from core when needed

3. **No import cycles between core and gui**
   - core never imports gui
   - gui can import core for shared types/functions

## Conventions

### Naming

- **Options:** `WithFeature()` for enabling features
  - `WithGlobal(name, value)` - Named globals (forwarded to modules)
  - `WithRisorOptions(opts...)` - Opaque Risor VM config (not forwarded)
- **Factory functions:** `New*()` (e.g., `NewWindow`, `NewApp`)
- **Wrapper types:** Match Fyne names (e.g., `Button`, `Label`, `MenuItem`)

### File Organization

- **options.go** - Option types and With* functions
- **widget/*.go** - One file per widget wrapper
- **module.go** - ImportedModule type for import system
- **import.go** - import() builtin implementation
- **require.go** - require() builtin implementation

## Common Tasks

### Adding a New Widget

1. Create `gui/widget/widgetname.go`
2. Implement wrapper struct with `object.Object` interface:
   ```go
   type WidgetName struct {
       instance *widget.WidgetName
       w        *Window  // Only if callbacks needed
   }
   ```
3. Implement required methods:
   - `Type()`, `Inspect()`, `Interface()`, `IsTruthy()`, `Cost()`
   - `GetAttr()`, `SetAttr()`, `Attrs()`
   - `Equals()`, `RunOperation()`, `MarshalJSON()`
4. Add factory function in `gui/widget.go`:
   ```go
   case "NewWidgetName":
       return object.NewBuiltin("widget.NewWidgetName", NewWidgetName(obj.w)), true
   ```
5. Add to `Attrs()` list
6. Create example in `examples/XX-widgetname/`
7. Test manually

### Adding a Module Option

1. Add to both `core/options.go` and `gui/options.go`:
   ```go
   func WithModuleName() Option {
       return moduleOption{
           fn: func(env map[string]any, userGlobals *[]risor.Option, modules map[string]bool) {
               env["modulename"] = modulename.Module()
               modules["modulename"] = true
           },
       }
   }
   ```
2. Import the module package at the top
3. Update documentation in comments
4. Add to examples where relevant

### Updating Version

The single source of truth for code is `core.Version` in `core/version.go`.
Both CLIs (`cmd/fynerisor`, `cmd/fynerisor-browser`) read it via `core.Version`,
so they never need a manual bump. Only these files carry a hardcoded copy:

1. Update `core/version.go` - `const Version = "X.Y.Z"` (authoritative)
2. Update the version strings in the docs (they cannot reference Go code):
   - `doc.go` - `Current Version:` comment
   - `llms.txt` - `Version:` line
   - `docs/README.md` - `**Current Version**:` line
   - `CLAUDE.md` - `**Current Version:**` line (top of this file)
3. Update the `Version` in both `FyneApp.toml` packaging files (read by `fyne package`;
   e.g. sets the APK versionName, so they cannot use Go code):
   - `cmd/fynerisor/FyneApp.toml`
   - `cmd/fynerisor-browser/FyneApp.toml`
4. Update `CHANGELOG.md` - Add release notes under new version heading
5. Update examples that use `require()` if major/minor version changed

### Creating an Example

1. Create directory: `examples/NN-example-name/`
2. Create files:
   - `main.go` - Go code to load and run script
   - `main.risor` or `script.risor` - Risor script
   - `README.md` - What it demonstrates, how to run
3. Keep examples simple and focused on one concept
4. Use `require(["v0.6"])` at the top of scripts
5. Add to `examples/README.md` list

## Threading and Concurrency

**CRITICAL:** Read `docs/CONCURRENCY.md` before touching widget callbacks.

### Rules

1. **All Fyne GUI operations must run on UI thread**
   - Use `fyne.Do(func() { ... })` for UI updates from goroutines
   - Widget callbacks are already on UI thread

2. **Risor VM is single-threaded per instance**
   - Never call Risor callbacks from multiple goroutines
   - Use `w.functionCalls` channel to queue callbacks

3. **Callback pattern for widgets:**
   ```go
   callFunc, ok := object.GetCallFunc(ctx)
   if !ok {
       return nil, fmt.Errorf("unable to get call function")
   }
   
   widget.OnEvent = func() {
       w.functionCalls <- func() {
           fyne.Do(func() {
               _, err := callFunc(ctx, callback, []object.Object{})
               if err != nil {
                   w.SetStatus("ERROR: " + err.Error())
               }
           })
       }
   }
   ```

## Module Import System

### Architecture (v0.6+)

- **Module-scoped imports** with namespace isolation
- Each import creates an `ImportedModule` object
- Exported functions run in their own module VM
- Module cache prevents re-execution

### Key Files

- `core/module.go` - ImportedModule type (shared)
- `gui/import.go` - import() builtin for Window
- `core/import.go` - import() builtin for Context

### Important

- Modules receive the same `env` map as main script
- User globals from `WithRisorOptions()` are NOT forwarded (opaque)
- Named globals from `WithGlobal()` ARE forwarded (in env map)

## Testing

```bash
# Build all packages
go build ./...

# Run tests
go test ./...

# Run specific package tests
go test ./core
go test ./gui

# Run examples manually
cd examples/01-hello-world
go run main.go
```

## Risor v2 Patterns

### In Go Code

```go
// Get call function for invoking Risor closures
callFunc, ok := object.GetCallFunc(ctx)
if !ok {
    return nil, fmt.Errorf("unable to get call function")
}

// Call Risor function
result, err := callFunc(ctx, closure, []object.Object{arg1, arg2})

// Convert Go types to Risor objects
object.NewString("text")
object.NewInt(42)
object.NewFloat(3.14)
object.NewBool(true)
object.NewList([]object.Object{...})

// Convert Risor objects to Go types
str, err := object.AsString(obj)
num, err := object.AsInt(obj)
flt, err := object.AsFloat(obj)
bol, err := object.AsBool(obj)
```

### In Risor Scripts (Examples)

```javascript
// Use 'function' keyword at top/global scope
function processData(data) {
    return data.map(x => x * 2)
}

function handleClick() {
    print("Button clicked!")
}

// Use arrow functions for callbacks and inline functions
let button = widget.NewButton("Click", () => {
    print("Inline callback")
})

// Arrow functions in functional iteration
items.map(x => x * 2)
items.filter(x => x > 10)
items.each(x => print(x))

// Arrow functions assigned to variables
let onClick = () => { print("clicked") }
let transform = (x) => x * 2

// Module imports
let utils = import("utils.risor")
utils.myFunction()

// Requirements
require(["v0.6", "@http", "@sql"])
```

## Common Pitfalls

1. **Import cycles** - gui imports core (OK), core MUST NOT import gui
2. **MenuItem doesn't have Refresh()** - Just set properties, don't call Refresh
3. **Deprecated Fyne APIs** - Check Fyne deprecation warnings
4. **Module cache** - Remember imports are cached, same path = same instance
5. **Thread safety** - Always use functionCalls channel for Risor callbacks
6. **WithGlobals vs WithGlobal** - Now called WithRisorOptions vs WithGlobal

## Git Workflow

```bash
# Build and test before committing
go build ./...
go test ./...

# Commit format
git commit -m "type: brief description

Detailed explanation if needed.

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"

# Types: feat, fix, docs, refactor, test, chore
```

## Breaking Changes

When making breaking changes:
1. Document in commit message: `BREAKING CHANGE: ...`
2. Update CHANGELOG.md under Unreleased
3. Consider if major version bump is needed
4. Update all affected examples
5. Update documentation (doc.go, llms.txt, README.md)

## Key Files to Know

- `CHANGELOG.md` - All version history (single source of truth)
- `doc.go` - Package documentation (godoc)
- `llms.txt` - AI assistant reference
- `docs/CONCURRENCY.md` - Threading guidelines
- `docs/EXAMPLES.md` - Example descriptions
- `cmd/fynerisor/main.go` - CLI tool

## Resources

- Fyne docs: https://docs.fyne.io/
- Risor docs: https://risor.io/docs/
- This project: https://github.com/uidbz/fynerisor

## Questions?

If something is unclear:
1. Check existing code for patterns
2. Look at similar widgets/modules
3. Read docs/CONCURRENCY.md for threading
4. Check CHANGELOG.md for recent changes
5. Ask the user for clarification
