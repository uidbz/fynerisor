// Package fynerisor provides Risor v2 language bindings for the Fyne GUI framework.
//
// Fynerisor enables building cross-platform desktop applications using Risor v2
// scripts with full access to Fyne's widget and container system. Create interactive
// GUIs with buttons, forms, tables, charts, and more through simple factory functions.
//
// # Package Structure
//
// Fynerisor is split into two packages for flexibility:
//
//   - core: Headless Risor execution with no GUI dependencies (enables static compilation)
//   - gui: Full GUI capabilities with Fyne (requires dynamic linking for OpenGL/X11)
//
// The root package (github.com/uidbz/fynerisor) re-exports gui functionality for
// backward compatibility with existing code.
//
// # Version
//
// Current Version: 0.4.1
// Risor Compatibility: v2.1+ (arrow functions required)
// Fyne Compatibility: v2.7+
//
// # Features
//
//   - 34 widget bindings (60% coverage - all high/medium priority complete)
//   - Button importance levels and validation
//   - Entry validation with visual feedback
//   - SQL module for MySQL, PostgreSQL, SQLite, SQL Server
//   - HTTP module for REST API calls
//   - Layout containers (VBox, HBox, Border, Split, Scroll)
//   - Canvas objects and charts
//   - Thread-safe callback handling
//   - Property access and modification from scripts
//   - Constants global for Fyne enums
//
// # Quick Start
//
// Create a simple GUI application:
//
//	package main
//
//	import "github.com/uidbz/fynerisor"
//
//	func main() {
//	    fw := fynerisor.NewApp("Hello",
//	        fynerisor.WithHTTP(),
//	        fynerisor.WithSQL(),
//	    )
//
//	    script := `
//	        require(["v0.3", "@gui", "@http"])
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
// # Requirements
//
// Scripts can declare version and module requirements:
//
//	require(["v0.2", "@sql", "@http"])  // Multiple requirements
//	require("v0.2.0")                   // Minimum version
//	require("@sql")                     // Module must be enabled
//
// # Global Objects
//
// Risor scripts have access to the following global objects:
//
//   - window: Main window control (SetContent, SetStatus, Title property)
//   - widget: Factory for all widgets
//   - container: Factory for layouts
//   - canvas: Factory for canvas objects
//   - chart: Factory for charts
//   - constants: Fyne constants (button importance, etc.)
//   - app: Application metadata (app.name)
//
// # Options
//
// Enable additional modules using options:
//
//	fw := fynerisor.NewApp("My App",
//	    fynerisor.WithHTTP(),      // HTTP client
//	    fynerisor.WithSQL(),       // Database connectivity
//	    fynerisor.WithOS(),        // OS utilities
//	    fynerisor.WithIO(),        // File I/O (cp, etc.)
//	    fynerisor.WithStrings(),   // String manipulation
//	    fynerisor.WithFilepath(),  // Path utilities
//	    fynerisor.WithTime(),      // Time operations
//	    fynerisor.WithStatusCallback(func(status string) {
//	        log.Println("Status:", status)
//	    }),
//	    fynerisor.WithAppName("myapp"),
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
//   - 15-sql-test: Database connectivity
//   - 17-gridwrap: Grid layout
//   - 18-textgrid: Code display
//   - 19-richtext: Markdown formatting
//   - 20-button-importance: Button styling
//   - 21-form-validation: Entry validation
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
//	import "github.com/deepnoodle-ai/risor/v2"
//
//	customData := risor.WithEnv(map[string]any{
//	    "api": myAPIObject,
//	    "config": configData,
//	})
//
//	fw := fynerisor.NewApp("My App",
//	    fynerisor.WithGlobals(customData),
//	)
//
// Scripts can then access: api.someMethod() and config.someProperty
package fynerisor
