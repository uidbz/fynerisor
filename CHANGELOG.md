# Changelog

## [Unreleased]

### Added

**Browser Package:**
- **browser/** - Generic, reusable browser UI for fynerisor applications
  - Navigation: address bar, back/forward/refresh/home buttons
  - History management with browser-style back/forward stacks
  - URL handling: automatic normalization and relative path resolution
  - Source view toggle: split layout showing content and source code side-by-side
  - **Programmatic navigation from scripts:**
    - `browser.Open(url)` - Navigate to a URL from Risor scripts
    - `browser.GetURL()` - Get current URL
    - `browser.SetStatus(text)` - Set status bar text
  - Hybrid plugin system:
    - MenuProvider interface for custom menu items
    - AuthProvider interface for custom authentication UI
    - SourceViewProvider interface for source code display
    - Callbacks for navigation lifecycle (OnNavigate, OnError, etc.)
  - Built on fynerisor's gui.Window for seamless integration

**Reference Browser:**
- **cmd/fynerisor-browser/** - Reference implementation demonstrating browser package usage
  - Loads Risor scripts from HTTP(S) or file:// URLs
  - Shows plugin architecture and integration patterns
  - Programmatic navigation enabled via `browser` global object
  - Serves as starting point for custom browser applications
  - ~350 lines with clear comments

**Window Enhancements:**
- **window.RegisterGlobal(name, value)** - Add global objects after window creation
  - Useful for objects that need to reference window or components created after initialization
  - Example: registering browser object for programmatic navigation

**Error Handling:**
- **Script execution error display** - Errors now show in clean UI with "Copy Error to Clipboard" button
- **browser.CopyToClipboard(text)** - Copy text to system clipboard from scripts
- **Terminal error logging** - All errors logged to terminal for debugging

### Fixed

**Canvas Image Loading:**
- **canvas.NewImageFromURI()** - Better error messages for invalid URIs
  - Relative paths now return clear error: "relative paths are not supported, use absolute file:// URI or HTTP(S) URL"
  - Panic recovery for malformed URIs prevents crashes
  - Empty URI check with helpful error message

## [0.6.1] - 2026-06-23

### Fixed

**Goroutine Safety:**
- **Fixed concurrent VM access in goroutines** - go() spawned goroutines that call GUI
  functions (e.g., `label.SetText()`) no longer cause concurrent VM access panics.
  GUI updates from goroutines are now properly routed through fyne.Do() to ensure
  thread-safe execution on the main UI thread.
- **Fixed stdout capture deadlock** - Large output from spawned goroutines no longer
  causes deadlocks when stdout capture is enabled. Output lines longer than buffer
  size are now handled gracefully.
- **Panic recovery in callbacks** - Script callbacks (button clicks, form submits, etc.)
  that panic are now caught and logged instead of crashing the application.

### Changed
- Goroutine stdout capture now routes through Go channel to avoid concurrent VM access
- Long output lines are truncated with warning instead of blocking

## [0.6.0] - 2026-06-15

### Added

**Keyboard Shortcuts:**
- **window.AddShortcut(shortcut, callback)** - Register global keyboard shortcuts
  - Simple string-based API: `"Ctrl+S"`, `"Alt+Shift+N"`, etc.
  - Works without menus - shortcuts are always active
  - Supports all modifier keys: Ctrl, Alt, Shift, Super/Cmd
  - Cross-platform modifier aliases: Ctrl/Control, Alt/Option, Super/Cmd/Command
  - Supports letters (A-Z), numbers (0-9), function keys (F1-F12)
  - Special keys: Return, Escape, Tab, Space, Backspace, Delete, arrows, etc.
  - Multi-modifier shortcuts: `"Ctrl+Shift+S"`, `"Alt+Shift+F1"`
- **window.RemoveShortcut(shortcut)** - Remove registered shortcuts dynamically
- **MenuItem.Shortcut** - Display shortcut hints in menu items
  - Shows shortcut text like "Ctrl+S" next to menu item labels
  - Display-only property (actual shortcuts registered via window.AddShortcut)
  - Example: 34-keyboard-shortcuts with menu integration

### Fixed

**Module Imports (`import()`):**
- **Module-level references now work** - Functions in an imported module can
  reference other module-level variables and functions (e.g.
  `let circleArea = (r) => PI * square(r)`). Previously such references resolved
  against the wrong globals array, returning incorrect values or crashing with
  `index out of range`. Exported functions are now executed inside their own
  module VM where global references resolve correctly.
- **Modules receive host globals** - Imported module VMs are now seeded with the
  same merged environment as the main script, so module functions can use
  `widget`, `http`, `os`, `sql`, etc. Use `WithGlobal("name", value)` (singular)
  for custom globals that imported modules should also access; opaque
  `WithRisorOptions(risor.WithEnv(...))` values are not forwarded into module VMs.

## [0.5.1] - 2026-06-12

### Added

**Widget Table Cells:**
- **Widget mode for tables** - Use any Fyne widget type in table cells
  - CreateCell callback for widget creation (called once per cell position)
  - UpdateCell callback for data binding (maps filtered rows automatically)
  - Supports all widget types: Button, Icon, Label, Entry, Check, Select, etc.
  - Canvas image support for displaying images in table cells
  - Example: 32-table-widgets with interactive buttons and icons
  - Example: 33-image-gallery with thumbnail images in tables

**Table Export Enhancements:**
- **Multi-format export** - Export table data to CSV, XLSX, or JSON
  - CSV: Standard comma-separated format
  - XLSX: Excel-compatible spreadsheet format
  - JSON: Structured data with column names as keys
  - Configurable export path and filename
  - Column selection for export
  - Export current page or all data
  - Default export directory: ./exports (current working directory)

**OS Module Enhancements:**
- **os.read_dir(path)** - List directory contents
  - Returns list of maps with entry.name and entry.is_dir fields
  - Available via gui.WithOS() option
- **os.getwd()** - Get current working directory
- **os.exec(command, args)** - Execute external commands

### Fixed

**Table Filtering with Widgets:**
- **Automatic row mapping** - Filtering now works correctly with widget mode
  - Filtered rows automatically map to original data indices
  - Widget state preserved during filtering
  - No script changes required - mapping is transparent
  - Widget placeholder columns (e.g., "[Icon]", "[Button]") excluded from filter search

### Changed
- Export functionality integrated into table widget footer
- Widget cells cached and reused for better performance

### Added

**Dialog Support:**
- **Dialog package** - Full Fyne dialog support for user interactions
  - Show* convenience functions: ShowInformation, ShowError, ShowConfirm, ShowFileOpen, ShowFileSave, ShowFolderOpen, ShowColorPicker, ShowForm, ShowCustom, ShowCustomConfirm
  - New* constructors for advanced control: NewFileOpen, NewFileSave, NewConfirm, NewCustom, NewColorPicker, NewForm
  - Dialog wrapper types with full API: FileDialog (SetFileName, SetFilter, SetLocation), ConfirmDialog (SetConfirmImportance), CustomDialog (SetButtons), ColorPickerDialog (Advanced mode, SetColor), FormDialog (Submit)
  - All callbacks properly marshalled to GUI thread
  - File dialogs return paths as strings, color picker returns RGB map
  - Example: 30-dialogs with both basic and advanced usage

**Widget Visibility Control:**
- **Hide() and Show() methods** - Added to all 48 widgets and containers
  - Dynamic visibility control for all UI elements
  - Thread-safe via fyne.Do()
  - Consistent API across buttons, labels, entries, forms, tables, containers, etc.

**Command Execution:**
- **Exec module** - Run external commands from Risor scripts
  - exec(["command", "args"]) - Execute and return result with stdout/stderr/pid
  - exec.command(["ls", "-la"]) - Create command object with control methods
  - exec.look_path("cmd") - Find command path in PATH
  - Result and Command types with methods: run(), output(), json()
  - Configuration via options map: dir, env, stdin, stdout, stderr
  - Available in both core and gui packages via WithExec()

## [0.5.0] - 2026-05-29

### Breaking Changes

**Package Restructure for Static Compilation:**
- Split fynerisor into `core` (headless) and `gui` (GUI) packages
- **All imports must be updated** - root package no longer provides functionality
- Use explicit package imports: `github.com/uidbz/fynerisor/gui` or `github.com/uidbz/fynerisor/core`

### Added

**Headless Mode with Static Compilation:**
- **`core` package** - Risor execution with zero Fyne dependencies
  - `core.NewContext()` for headless script execution
  - All modules available: HTTP, SQL, OS, Strings, Filepath, Time, IO
  - Enables static compilation without OpenGL/X11 dependencies
  - Perfect for CLI tools, batch processing, server-side scripts
- **`gui` package** - Full GUI capabilities with Fyne framework
  - All original functionality preserved
  - `gui.NewWindow()` and `gui.NewApp()` for GUI applications
  - Subpackages: widget, binding, container, canvas, chart
- **Backward compatibility** - Root package re-exports gui functionality
  - Existing imports continue to work: `import "github.com/uidbz/fynerisor"`
  - No code changes required for existing applications

### Changed
- Moved NAMING_CONVENTIONS.md to docs/ directory
- Updated documentation for new package structure
- Updated example 16-context-builder to use `core` package

### Migration Guide

**Update all imports - REQUIRED:**

```go
// OLD (v0.4.x)
import "github.com/uidbz/fynerisor"
fw := fynerisor.NewApp("My App", fynerisor.WithHTTP())
core.SetAppVersion("1.0.0")

// NEW (v0.5.0) - GUI applications
import (
    "github.com/uidbz/fynerisor/core"
    "github.com/uidbz/fynerisor/gui"
)
fw := gui.NewApp("My App", gui.WithHTTP())
core.SetAppVersion("1.0.0")
```

**For headless/static compilation:**
```go
// Use core package for CLI tools and server-side scripts
import "github.com/uidbz/fynerisor/core"
ctx := core.NewContext(core.WithHTTP(), core.WithSQL())
result, err := ctx.Eval(`print("Hello!")`)
```

## [0.4.2] - 2026-05-19

### Fixed
- **CLI installation** - Fixed `go install github.com/uidbz/fynerisor/cmd/fynerisor@latest` 
  - Previous v0.4.1 tag was cached incorrectly by Go proxy
  - CLI tool now properly installable via go install
- **Missing example** - Added 01-hello-world example that was blocked by .gitignore
- **Documentation** - Added syntax highlighting to all Risor code blocks (using js highlighting)
- **Git author** - Fixed commit attribution for GitHub

### Added
- **CLI theme flag** - Added `--theme` flag to CLI for dark/light theme selection
- **Logo and screenshot** - Added visual assets to README
- **Prerequisites documentation** - Added link to Fyne prerequisites in installation section

### Summary
**Patch Release - CLI Installation Fix**
- ✅ CLI tool now works with `go install`
- ✅ All 29 examples included
- ✅ Complete and production-ready

## [0.4.1] - 2026-05-07 [YANKED]

**Note:** This version had issues with Go module proxy caching. Use v0.4.2 instead.

### Added

**Custom Struct Support:**
- **WithGlobal(name, object)** - Expose custom Go types with methods as global variables in Risor scripts
  - Implement `object.Object` interface on your Go types
  - Scripts can call methods on your custom objects (e.g., `db.query("SELECT * FROM users")`)
  - Example: 28-custom-struct - Complete pattern with UserDatabase

**Application Versioning:**
- **SetAppVersion(version)** - Allow embedding applications to control their own version checking
  - Separates application version from fynerisor library version
  - Scripts can use `require(["v2.5"])` to check compatibility with the host application
  - Example: 27-app-versioning

### Fixed
- Fixed custom-struct example to use Risor v2 syntax (`.each()` instead of for loops)
- Fixed example numbering - moved 28-imports to 14-imports to fill gap
- Added status callback to custom-struct example for better error visibility

### Documentation
- Added comprehensive custom struct integration guide
- Documented Risor v2 functional iteration patterns
- Added debugging section with status callback examples
- Updated WithGlobal() documentation in options.go

### Examples
- All 29 examples tested and working
- 28 widget/feature examples
- 1 custom struct integration example

### Summary
**Backward Compatible Release**
- ✅ Custom Go type integration via WithGlobal()
- ✅ Application versioning support via SetAppVersion()
- ✅ All examples updated and verified working
- ✅ Fully backward compatible with v0.4.0

### Added

**Custom Struct Support:**
- **WithGlobal(name, object)** - Expose custom Go types with methods as global variables in Risor scripts
  - Implement `object.Object` interface on your Go types
  - Scripts can call methods on your custom objects (e.g., `db.query("SELECT * FROM users")`)
  - Example: 28-custom-struct - Complete pattern with UserDatabase

**Application Versioning:**
- **SetAppVersion(version)** - Allow embedding applications to control their own version checking
  - Separates application version from fynerisor library version
  - Scripts can use `require(["v2.5"])` to check compatibility with the host application
  - Example: 27-app-versioning

### Fixed
- Fixed custom-struct example to use Risor v2 syntax (`.each()` instead of for loops)
- Fixed example numbering - moved 28-imports to 14-imports to fill gap
- Added status callback to custom-struct example for better error visibility

### Documentation
- Added comprehensive custom struct integration guide
- Documented Risor v2 functional iteration patterns
- Added debugging section with status callback examples
- Updated WithGlobal() documentation in options.go

### Examples
- All 29 examples tested and working
- 28 widget/feature examples
- 1 custom struct integration example

### Summary
**Backward Compatible Release**
- ✅ Custom Go type integration via WithGlobal()
- ✅ Application versioning support via SetAppVersion()
- ✅ All examples updated and verified working
- ✅ Fully backward compatible with v0.4.0

## [0.4.0] - 2026-05-06

### Added

**🎉 Data Binding System (Complete):**
- **binding.NewString()** - String data binding with optional initial value
- **binding.NewBool()** - Boolean data binding with optional initial value
- **binding.NewInt()** - Integer data binding with optional initial value
- **binding.NewFloat()** - Float data binding with optional initial value
- All bindings support: `Get()`, `Set(value)`, `AddListener(callback)`

**Widgets with Data Binding:**
- **widget.NewLabelWithData(binding.String)** - Auto-updating labels
- **widget.NewCheckWithData(label, binding.Bool)** - Bound checkboxes
- **widget.NewSliderWithData(min, max, binding.Float)** - Bound sliders
- Enables automatic two-way synchronization between data and UI

**Container Types (Complete 100% Coverage):**
- **container.NewCenter** - Center-aligned content layout
- **container.NewMax** - Maximum size layout (fills available space)
- **container.NewStack** - Layered stack of widgets
- **container.NewPadded** - Padded container with standard spacing
- **container.NewGridWithColumns** - Fixed column grid layout
- **container.NewGridWithRows** - Fixed row grid layout
- All 10 standard Fyne container types now available

**Widget Enhancements:**
- **Label.SetText()** - Method variant for setting label text
- **Button.SetText()** - Method variant for setting button text
- **Entry.OnChanged()** - Callback fires on every text change
- **Label Properties** - Wrapping, Truncation, Alignment, Importance now fully supported

**Constants & Enums:**
- **TextAlign** - Leading, Center, Trailing constants
- Complete constants example (example 24-constants)

**Examples:**
- 25-data-binding: Basic data binding with Label + Entry
- 26-data-binding-types: Comprehensive demo of all binding types

### Fixed
- Menu example (23-menu): Simplified to use button trigger instead of right-click
- Popup example (22-popup): Fixed variable scope issues in closures
- Constants example (24-constants): Added missing Label properties
- GridWrap example (17-gridwrap): Fixed overlapping text with proper Button sizing
- All 26 examples now working correctly

### Summary
**100% Feature Complete!**
- ✅ All high, medium, and low priority widgets (37 widgets)
- ✅ All container types (10/10)
- ✅ Complete data binding system (4 types)
- ✅ 26 working examples
- ✅ Production ready!

## [0.3.0] - 2026-05-06

### Added

**Advanced Widgets (95% Coverage Complete):**
- **GridWrap Widget** - Grid layout with virtualized rendering
  - `Length()`, `CreateItem()`, `UpdateItem()` callbacks
  - `Select()`, `Unselect()`, `UnselectAll()` methods
  - `OnSelected()`, `OnUnselected()` event handlers
  - Example: 17-gridwrap
  
- **TextGrid Widget** - Monospace text grid for code/tabular data
  - Line numbers and whitespace display options
  - Row operations: `SetRow()`, `Row()`, `RowCount()`
  - Configurable tab width
  - Example: 18-textgrid
  
- **RichText Widget** - Formatted text with markdown support
  - `ParseMarkdown()` for dynamic content
  - Text wrapping and truncation control
  - Scrollable content
  - Note: Direct segment manipulation not exposed (use markdown)
  - Example: 19-richtext

**Widget Enhancements:**
- **Button.Importance** - Visual hierarchy and semantic meaning
  - Standard levels: `ImportanceHigh`, `ImportanceMedium`, `ImportanceLow`
  - Semantic levels: `SuccessImportance`, `WarningImportance`, `DangerImportance`
  - Example: 20-button-importance
  
- **Button.Disabled** - Enable/disable button state programmatically
  
- **Entry.SetValidator()** - Custom validation with visual feedback
  - Return error string for invalid input
  - Return `nil` for valid input
  - Fyne shows red border automatically
  - Example: 21-form-validation

**Constants Global:**
- New global object `constants` for Fyne enums
- Currently provides button importance values
- Extensible for future constants

### Changed
- CLI module renamed from `fynerisor-cli` to `fynerisor`
- Installation instructions updated with `go install` command
- Documentation examples fixed for current API

### Fixed
- Module options (WithOS, WithHTTP, etc.) now work correctly in NewContext
- SQL query example no longer uses `.collect()` incorrectly
- Non-GUI usage examples corrected

## [0.2.1] - 2026-05-05

### Added
- **IO Module**: File operations support
  - `io.cp(src, dst)` - Copy files
  - Enabled via `WithIO()` option
  - Require with `@io` in scripts
  
- **App Metadata Object**: `app.name` global
  - Exposes embedding application name to scripts
  - Set via `WithAppName("myapp")` option
  - Default value: `"fynerisor"`
  - Enables conditional logic based on runtime environment
  
- **Go Function**: Spawn goroutines from scripts
  - `go(() => { ... })` - Run closure in background
  - Replaces removed `go` keyword from Risor v1
  - Safe for long-running operations
  
- **Window.Do()**: Queue GUI updates from background threads
  - `window.Do(() => { ... })` - Execute on GUI thread
  - Matches Fyne's threading model
  - Safe alternative to direct widget updates from goroutines
  
- **Window.Clear()**: Clear accumulated scripts
  - Clears all loaded scripts and imports
  - Critical for navigation-based apps (like goto)
  - Prevents script accumulation when loading new pages

### Fixed
- **window.OnDropped()**: Now properly registers with Fyne window
  - Fixed missing `SetOnDropped()` call
  - Drag-and-drop now works correctly
  
- **Options application order**: Fixed initialization sequence
  - AppName and callbacks applied before globals
  - Module options applied after base globals
  - Fixes issue where custom globals weren't available

### Changed
- Documentation restructured to `docs/` directory
- Updated README with current API and examples

## [0.2.0] - 2026-05-04

### Breaking Changes
- **Risor v2 Migration**: All scripts must use arrow function syntax
  - `func()` keyword removed, use `() => {}` or `(x) => x`
  - No `for`/`while` loops, use `.map()`, `.each()`, `.filter()`, `.reduce()`

### Added
- **ContextBuilder for non-GUI usage**
  - `NewContext()` creates Risor contexts without GUI dependencies
  - Access to all modules (HTTP, SQL, OS, etc.) and import system
  - `.Eval()` and `.EvalWithImports()` methods
  - Context-specific options: `WithContextHTTP()`, `WithContextSQL()`, etc.
  - Enables headless scripts, CLI tools, and server applications
  - Example: 16-context-builder
- **SQL Module**: Database connectivity support
  - MySQL, PostgreSQL, SQLite, SQL Server drivers
  - Connection via `sql.connect(dsn)`
  - Query execution: `conn.query()`, `conn.exec()`
  - Row iterator with `.collect()` method for list conversion
  - Example: 15-sql-test
  - Integrated into CLI with `fynerisor.WithSQL()` option

- **Tree widget** with hierarchical data support
  - ChildUIDs, IsBranch, CreateNode, UpdateNode callbacks
  - OpenBranch, CloseBranch, Select, Unselect methods
  - Example: 12-tree

- **List widget** for scrolling lists
  - Length, CreateItem, UpdateItem, OnSelected callbacks
  - Example: 13-list

- **SQL row iterator streaming methods**
  - `.map(fn)` - Transform rows in single pass (efficient)
  - `.each(fn)` - Iterate with side effects
  - `.collect()` - Load all rows into list (when needed)

- **Enhanced require() function**
  - Accepts string or list of requirements
  - Version requirements: `require("v0.2.0")` or `require("==v0.2.0")`
  - Module requirements: `require("@sql")`, `require("@http")`, etc.
  - Combined: `require(["v0.2", "@sql", "@http"])`
  - Validates enabled modules at script startup

- **AnalyzeRequirements() function**
  - Parse script source to extract requirements
  - Returns Requirements struct with version and module info
  - Useful for external tools and script validation
  - Example: `reqs, _ := fynerisor.AnalyzeRequirements(script)`

- **NewApp() convenience function**
  - One-line creation of Fyne app and fynerisor window
  - Simplifies common use case of creating a new application
  - Example: `fw := fynerisor.NewApp("My App", fynerisor.WithHTTP())`
  - Use NewWindow() for advanced scenarios requiring app customization

- **Window.Resize() method**
  - Set custom window size after creation
  - Example: `fw.Resize(800, 600)`
  - Works with both NewApp and NewWindow

- **import() global function**
  - Scripts can declare imports: `import("utils.risor")` or `import(["a.risor", "b.risor"])`
  - Supports local files and HTTP URLs
  - No-op at runtime (marker for static analysis)
  - CLI automatically handles imports via AnalyzeRequirements
  - Imports loaded in declaration order before script execution

- **AnalyzeRequirements() returns imports**
  - Requirements.Imports field lists all import paths/URLs
  - Enables automatic dependency resolution
  - Example: `reqs.Imports = ["utils.risor", "http://..."]`

- CLI error reporting without --verbose flag
  - All ERROR messages now display by default
  - --verbose still shows full execution status

- Comprehensive llms.txt for LLM code generation
  - Complete widget reference
  - Risor v2 syntax examples
  - Example prompts for GUI creation
  - Common patterns and constraints

### Fixed
- **CRITICAL**: Fixed race condition in Tree and List widgets causing VM stack corruption
  - Changed UpdateNode, OnSelected, OnUnselected in Tree to synchronous calls
  - Changed UpdateItem, OnSelected in List to synchronous calls
  - Issue: Mixed sync/async callbacks caused concurrent VM access
  - Symptom: `panic: index out of range [-1]` at `vm.pop()` after 30-100 interactions
  - See CONCURRENCY.md for details and guidelines
  
- Fixed Tree widget map access for Risor v2
  - Use `.get()` method instead of direct indexing (which throws errors)
  - Risor v2 throws errors on missing keys, unlike Risor v1
  
- Fixed accordion example (08-accordion)
  - Removed `while` loops (not supported in Risor)
  - Risor only supports functional iteration (for...in, map, filter)
  - Simplified example to focus on core accordion features

### Changed
- All Tree and List widget callbacks now use synchronous pattern
- Updated version requirement system
  - `require("v0.2")` for minimum version
  - `require("==v0.2.0")` for exact match
  - Added clear error messages for version mismatches

### Documentation
- Added CONCURRENCY.md with threading guidelines for widget development
- Added comprehensive widget audit (all 28 widgets checked)
- Updated examples with Risor v2 syntax
- Added version checking to relevant examples

## Previous Releases

See git history for earlier changes.
