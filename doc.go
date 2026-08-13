// Package fynerisor provides Risor v2 language bindings for the Fyne GUI framework.
//
// Fynerisor enables building cross-platform desktop applications using Risor v2
// scripts with full access to Fyne's widget and container system. Create interactive
// GUIs with buttons, forms, tables, charts, and more through simple factory functions.
//
// # Package Structure
//
// Fynerisor is split into two packages:
//
//   - core: Headless Risor execution with no GUI dependencies (enables static compilation)
//   - gui: Full GUI capabilities with Fyne (requires dynamic linking for OpenGL/X11)
//
// Always import the specific package you need: core for headless, gui for desktop applications.
//
// # Version
//
// Current Version: 0.7.0
// Risor Compatibility: v2.1+ (arrow functions required)
// Fyne Compatibility: v2.7+
//
// # Features
//
//   - 48+ widget bindings (100% coverage - all priority levels complete)
//   - Keyboard shortcuts (window.AddShortcut, MenuItem.Shortcut display)
//   - Data binding system (String, Bool, Int, Float)
//   - Dialog support (file, confirm, color, form, custom)
//   - Module-scoped imports with namespace isolation
//   - SQL module for MySQL, PostgreSQL, SQLite, SQL Server
//   - HTTP module for REST API calls
//   - Layout containers (VBox, HBox, Border, Split, Scroll, Grid, Stack, etc.)
//   - Canvas objects and charts (bar, line, scatter, histogram, box plot)
//   - Reusable browser package for loading Risor scripts from URLs (see browser/)
//   - Thread-safe callback handling
//   - Widget visibility control (Hide/Show methods)
//   - Constants global for Fyne enums
//
// # Quick Start
//
// Create a simple GUI application:
//
//	package main
//
//	import "github.com/uidbz/fynerisor/gui"
//
//	func main() {
//	    fw := gui.NewApp("Hello",
//	        gui.WithHTTP(),
//	        gui.WithSQL(),
//	    )
//
//	    script := `
//	        require(["v0.5", "@gui", "@http"])
//
//	        let btn = widget.NewButton("Click Me", () => {
//	            window.SetStatus("Button clicked!")
//	        })
//	        btn.Importance = constants.ImportanceHigh
//
//	        window.SetContent(btn)
//	    `
//
//	    fw.LoadScript(script)
//	    fw.Execute()
//	    fw.ShowAndRun()
//	}
//
// # Headless Mode
//
// Use the core package for headless execution without GUI dependencies.
// This enables static compilation and is useful for CLI tools, batch processing,
// or server-side scripting:
//
//	package main
//
//	import "github.com/uidbz/fynerisor/core"
//
//	func main() {
//	    ctx := core.NewContext(
//	        core.WithHTTP(),
//	        core.WithSQL(),
//	    )
//
//	    script := `
//	        require(["v0.4", "@http", "@sql"])
//	        let response = http.get("https://api.example.com/data")
//	        print(response.json())
//	    `
//
//	    ctx.LoadScript(script)
//	    ctx.Eval()
//	}
//
// The core package excludes all GUI code, allowing your binary to be statically linked
// without OpenGL, X11, or other GUI dependencies.
//
// # Browser Package
//
// The browser package provides a reusable, browser-style UI for applications that
// load Risor scripts from HTTP(S) or file:// URLs. It handles navigation (address
// bar, back/forward/refresh/home), history, source view, and pluggable menu and
// authentication providers. Scripts loaded in the browser can navigate
// programmatically and read URL query parameters via a "browser" global:
//
//	browser.Open("https://example.com/page")  // Navigate to a URL
//	browser.GetURL()                           // Current URL
//	browser.params["myarg"]                    // Query parameter value
//
// See cmd/fynerisor-browser for a complete reference implementation, which can
// also be packaged as an Android app with "fyne package -os android".
//
// # Requirements
//
// Scripts can declare version and module requirements:
//
//	require(["v0.7", "@sql", "@http"])  // Multiple requirements
//	require("v0.7.0")                   // Minimum version
//	require("@sql")                     // Module must be enabled
//
// # Global Objects
//
// Risor scripts have access to the following global objects:
//
//   - window: Main window control (SetContent, SetStatus, AddShortcut, etc.)
//   - widget: Factory for all widgets
//   - container: Factory for layouts
//   - canvas: Factory for canvas objects
//   - chart: Factory for charts
//   - dialog: Dialog functions (file, confirm, color, etc.)
//   - binding: Data binding for reactive UIs
//   - constants: Fyne constants (button importance, etc.)
//   - fyne: Fyne utilities (NewMenu, NewMainMenu, etc.)
//   - app: Application metadata (app.name, app.version)
//   - print: Output function
//   - require: Version and module validation
//   - import: Module imports with namespace isolation
//
// # Options
//
// Enable additional modules using options:
//
//	fw := gui.NewApp("My App",
//	    gui.WithHTTP(),           // HTTP client
//	    gui.WithSQL(),            // Database connectivity
//	    gui.WithTie(),            // Tie triple store database
//	    gui.WithOS(),             // OS utilities
//	    gui.WithIO(),             // File I/O (cp, etc.)
//	    gui.WithExec(),           // Execute commands
//	    gui.WithStrings(),        // String manipulation
//	    gui.WithFilepath(),       // Path utilities
//	    gui.WithTime(),           // Time operations
//	    gui.WithHTTPImport(),     // Import modules from URLs
//	    gui.WithGlobal("myapi", apiObject),  // Custom named global
//	    gui.WithStatusCallback(func(status string) {
//	        log.Println("Status:", status)
//	    }),
//	    gui.WithAppName("myapp"),
//	)
//
// # Risor v2 Syntax
//
// Fynerisor requires Risor v2 with arrow function syntax:
//
//	// Correct
//	let onClick = () => { print("clicked") }
//	let process = (x) => x * 2
//
//	// Wrong - func keyword not supported
//	func onClick() { print("clicked") }
//
// No for/while loops - use functional methods:
//
//	// Correct
//	let doubled = [1, 2, 3].map(x => x * 2)
//	list.each(item => print(item))
//
//	// Wrong - no loop keywords
//	for i in list { }
//	while (condition) { }
//
// # Thread Safety
//
// The Risor VM is single-threaded and NOT thread-safe. All widget callbacks
// execute synchronously in the UI thread. Never spawn goroutines or use
// async patterns inside callbacks.
//
// See CONCURRENCY.md for detailed information about the race condition bug
// that was fixed in v0.2.0.
//
// # SQL Module
//
// The SQL module supports MySQL, PostgreSQL, SQLite, and SQL Server:
//
//	let conn = sql.connect("sqlite3::memory:")
//	conn.exec("CREATE TABLE users (id INT, name TEXT)")
//	conn.exec("INSERT INTO users VALUES (?, ?)", 1, "Alice")
//
//	// Query returns streaming iterator with .map() and .each()
//	let names = conn.query("SELECT * FROM users").map(row => row["name"])
//
//	// Or use .each() for side effects
//	conn.query("SELECT * FROM users").each(row => print(row["name"]))
//
//	// Or .collect() to load all into memory first
//	let rows = conn.query("SELECT * FROM users").collect()
//
//	conn.close()
//
// # HTTP Module
//
// Make HTTP requests from Risor scripts:
//
//	let response = http.get("https://api.example.com/data")
//	let data = response.json()  // parse JSON response
//	print(data["items"])
//
// # Examples
//
// See the examples/ directory for 21 complete working examples:
//
//   - 01-hello-world: Simple label display
//   - 02-button-callback: Interactive button with state
//   - 03-form-input: Form with validation
//   - 04-table-display: Paginated table
//   - 05-progress-widgets: Progress bars and sliders
//   - 11-list: Virtualized scrolling list
//   - 12-tree: Hierarchical tree widget
//   - 13-http-fetch: HTTP requests and JSON
//   - 14-module-imports: Module-scoped imports with namespaces
//   - 15-sql-test: Database connectivity
//   - 17-gridwrap: Grid layout
//   - 18-textgrid: Code display
//   - 19-richtext: Markdown formatting
//   - 20-button-importance: Button styling
//   - 21-form-validation: Entry validation
//   - 25-data-binding: Reactive UI with data binding
//   - 27-app-versioning: Application version checking
//   - 28-custom-struct: Expose custom Go types
//   - 30-dialogs: File, confirm, color picker dialogs
//   - 32-table-widgets: Widget mode for table cells
//   - 33-image-gallery: Images in table cells
//   - 34-keyboard-shortcuts: Global shortcuts and menu integration
//
// # CLI Tool
//
// The fynerisor command-line tool runs Risor GUI scripts:
//
//	fynerisor app.risor
//	fynerisor --watch app.risor              # auto-reload on changes
//	fynerisor --width 1024 --height 768 app.risor
//	fynerisor --headless app.risor           # for testing
//	fynerisor --globals data.json app.risor  # custom globals
//
// # Custom Globals
//
// Inject custom data or functions into scripts:
//
//	import (
//	    "github.com/deepnoodle-ai/risor/v2"
//	    "github.com/uidbz/fynerisor/gui"
//	)
//
//	customData := risor.WithEnv(map[string]any{
//	    "api": myAPIObject,
//	    "config": configData,
//	})
//
//	fw := gui.NewApp("My App",
//	    gui.WithRisorOptions(customData),
//	)
//
// Scripts can then access: api.someMethod() and config.someProperty
package fynerisor
