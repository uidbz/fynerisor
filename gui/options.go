package gui

import (
	"github.com/deepnoodle-ai/risor/v2"
	csvmod "github.com/uidbz/fynerisor/modules/csv"
	"github.com/uidbz/fynerisor/modules/exec"
	filepathmod "github.com/uidbz/fynerisor/modules/filepath"
	"github.com/uidbz/fynerisor/modules/http"
	iomod "github.com/uidbz/fynerisor/modules/io"
	jsonmod "github.com/uidbz/fynerisor/modules/json"
	"github.com/uidbz/fynerisor/modules/os"
	"github.com/uidbz/fynerisor/modules/sql"
	"github.com/uidbz/fynerisor/modules/strings"
	"github.com/uidbz/fynerisor/modules/tie"
	"github.com/uidbz/fynerisor/modules/time"
)

// Module options (WithHTTP, WithSQL, etc.) work with both.
// Callback options (WithStatusCallback, WithResultCallback) only work with Window.
type Option interface {
	applyToWindow(*Window)
}

type moduleOption struct {
	// fn configures an imported environment. Built-in module options write their
	// global objects into env (so they can be passed to both the main script and
	// any imported module VMs). User-supplied opaque options (WithGlobals) are
	// appended to userGlobals instead, since their values can't be introspected.
	fn func(env map[string]any, userGlobals *[]risor.Option, modules map[string]bool)
}

func (o moduleOption) applyToWindow(w *Window) {
	o.fn(w.env, &w.userGlobals, w.enabledModules)
}

type windowOption struct {
	fn func(*Window)
}

func (o windowOption) applyToWindow(w *Window) {
	o.fn(w)
}

type appNameOption struct {
	name string
}

func (o appNameOption) applyToWindow(w *Window) {
	w.appName = o.name
}

// WithRisorOptions adds advanced Risor VM configuration options.
// These are opaque risor.Option objects for advanced Risor VM configuration.
// Unlike WithGlobal(), these options are NOT forwarded to imported modules.
//
// Use WithGlobal() instead for named globals that modules should access.
//
// Example:
//
//	customFuncs := map[string]any{
//	    "myFunc": func() string { return "hello" },
//	}
//	window := fynerisor.NewWindow(w,
//	    fynerisor.WithRisorOptions(risor.WithEnv(customFuncs)),
//	)
func WithRisorOptions(globals ...risor.Option) Option {
	return moduleOption{
		fn: func(env map[string]any, userGlobals *[]risor.Option, modules map[string]bool) {
			*userGlobals = append(*userGlobals, globals...)
		},
	}
}

// WithStatusCallback sets a callback that is called when script execution status changes.
// The callback receives status strings like "Ready!", "ERROR: ...", etc.
//
// Example:
//
//	window := fynerisor.NewWindow(w,
//	    fynerisor.WithStatusCallback(func(status string) {
//	        log.Println("Status:", status)
//	    }),
//	)
func WithStatusCallback(callback func(string)) Option {
	return windowOption{
		fn: func(w *Window) {
			w.statusCallback = callback
		},
	}
}

// WithResultCallback sets a callback that is called when script execution completes
// with a non-nil result value.
//
// Example:
//
//	window := fynerisor.NewWindow(w,
//	    fynerisor.WithResultCallback(func(result string) {
//	        log.Println("Result:", result)
//	    }),
//	)
func WithResultCallback(callback func(string)) Option {
	return windowOption{
		fn: func(w *Window) {
			w.resultCallback = callback
		},
	}
}

// WithAppName sets the application name exposed to Risor scripts via app.name.
// This allows scripts to detect which application is running them.
//
// Example:
//
//	window := fynerisor.NewWindow(w,
//	    fynerisor.WithAppName("lars"),
//	)
//
// Usage in script:
//
//	if (app.name == "lars") {
//	    print("Running in LARS")
//	}
func WithAppName(name string) Option {
	return appNameOption{name: name}
}

// WithHTTP enables the HTTP module for making HTTP requests from Risor scripts.
// The module provides functions: get, post, put, delete, and fetch.
//
// Example:
//
//	window := fynerisor.NewWindow(w,
//	    fynerisor.WithHTTP(),
//	)
//
// Usage in script:
//
//	let response = http.get("https://api.example.com/data")
//	print(response.status)
//	print(response.body)
func WithHTTP() Option {
	return moduleOption{
		fn: func(env map[string]any, userGlobals *[]risor.Option, modules map[string]bool) {
			env["http"] = http.Module()
			modules["http"] = true
		},
	}
}

// WithHTTPImport enables importing Risor modules from HTTP(S) URLs.
// This is a separate option from WithHTTP() for security reasons - you may want
// to allow HTTP requests from scripts but not allow loading arbitrary code from URLs.
//
// When enabled, scripts can use import() with HTTP(S) URLs:
//
//	let utils = import("https://example.com/utils.risor")
//	utils.someFunction()
//
// Security considerations:
//   - Only enable this for trusted scripts or controlled environments
//   - Imported modules execute with full access to enabled modules (http, sql, os, etc.)
//   - HTTPS is recommended for production to prevent code injection
//   - Consider implementing a whitelist of allowed domains in your application
//
// Example:
//
//	window := fynerisor.NewWindow(w,
//	    fynerisor.WithHTTP(),        // Enable http module for requests
//	    fynerisor.WithHTTPImport(),  // Enable importing from URLs
//	)
//
// Usage in script:
//
//	require(["@httpimport"])
//	let remote = import("https://cdn.example.com/mylib.risor")
//	remote.doSomething()
func WithHTTPImport() Option {
	return moduleOption{
		fn: func(env map[string]any, userGlobals *[]risor.Option, modules map[string]bool) {
			// This is a marker module - actual functionality is in import.go
			modules["httpimport"] = true
		},
	}
}

// WithOS enables the OS module for accessing OS functionality from Risor scripts.
// The module provides functions: goos, current_user, and open_browser.
//
// Example:
//
//	window := fynerisor.NewWindow(w,
//	    fynerisor.WithOS(),
//	)
//
// Usage in script:
//
//	let platform = os.goos()
//	let user = os.current_user()
//	os.open_browser("https://example.com")
func WithOS() Option {
	return moduleOption{
		fn: func(env map[string]any, userGlobals *[]risor.Option, modules map[string]bool) {
			env["os"] = os.Module()
			modules["os"] = true
		},
	}
}

// WithJSON enables the json module for JSON encoding/decoding from Risor scripts.
// The module provides functions: parse, marshal, marshal_indent, valid, read, write.
func WithJSON() Option {
	return moduleOption{
		fn: func(env map[string]any, userGlobals *[]risor.Option, modules map[string]bool) {
			env["json"] = jsonmod.Module()
			modules["json"] = true
		},
	}
}

// WithCSV enables the csv module for CSV encoding/decoding from Risor scripts.
// The module provides functions: parse, format, read, write. By default parse
// treats the first row as headers (list of maps); pass {header: false} for a
// list of lists.
func WithCSV() Option {
	return moduleOption{
		fn: func(env map[string]any, userGlobals *[]risor.Option, modules map[string]bool) {
			env["csv"] = csvmod.Module()
			modules["csv"] = true
		},
	}
}

// WithStrings enables the strings module for string manipulation from Risor scripts.
func WithStrings() Option {
	return moduleOption{
		fn: func(env map[string]any, userGlobals *[]risor.Option, modules map[string]bool) {
			env["strings"] = strings.Module()
			modules["strings"] = true
		},
	}
}

// WithFilepath enables the filepath module for file path manipulation from Risor scripts.
func WithFilepath() Option {
	return moduleOption{
		fn: func(env map[string]any, userGlobals *[]risor.Option, modules map[string]bool) {
			env["filepath"] = filepathmod.Module()
			modules["filepath"] = true
		},
	}
}

// WithTime enables the time module for date and time operations from Risor scripts.
// The module provides functions: now, date, and parse.
//
// Example:
//
//	window := fynerisor.NewWindow(w,
//	    fynerisor.WithTime(),
//	)
//
// Usage in script:
//
//	let now = time.now()
//	let date = time.date(2026, 5, 1)
//	let parsed = time.parse("2026-05-01")
func WithTime() Option {
	return moduleOption{
		fn: func(env map[string]any, userGlobals *[]risor.Option, modules map[string]bool) {
			env["time"] = time.Module()
			modules["time"] = true
		},
	}
}

// WithSQL enables the SQL module for database connectivity from Risor scripts.
// The module provides functions: connect, query, exec, and close.
//
// Supports: MySQL, PostgreSQL, SQLite, SQL Server
//
// Example:
//
//	window := fynerisor.NewWindow(w,
//	    fynerisor.WithSQL(),
//	)
//
// Usage in script:
//
//	let conn = sql.connect("sqlite3://./test.db")
//	let rows = conn.query("SELECT * FROM users")
//	conn.exec("INSERT INTO users (name) VALUES (?)", "Alice")
//	conn.close()
func WithSQL() Option {
	return moduleOption{
		fn: func(env map[string]any, userGlobals *[]risor.Option, modules map[string]bool) {
			env["sql"] = sql.Module()
			modules["sql"] = true
		},
	}
}

// WithTie enables the Tie module for triple store operations from Risor scripts.
// The module provides a client for the tie triple store database.
//
// Example:
//
//	window := fynerisor.NewWindow(w,
//	    fynerisor.WithTie(),
//	)
//
// Usage in script:
//
//	let db = tie.connect("http://localhost:1161")
//	db.add("pizza", "topping", "cheese")
//	db.sync()
//	print(db.get("pizza"))
func WithTie() Option {
	return moduleOption{
		fn: func(env map[string]any, userGlobals *[]risor.Option, modules map[string]bool) {
			env["tie"] = tie.Module()
			modules["tie"] = true
		},
	}
}

// WithIO enables the io module for file operations from Risor scripts.
// The module provides functions: cp (copy file).
//
// Example:
//
//	window := fynerisor.NewWindow(w,
//	    fynerisor.WithIO(),
//	)
//
// Usage in script:
//
//	io.cp("source.txt", "destination.txt")
func WithIO() Option {
	return moduleOption{
		fn: func(env map[string]any, userGlobals *[]risor.Option, modules map[string]bool) {
			env["io"] = iomod.NewIO()
			modules["io"] = true
		},
	}
}

// WithGlobal adds a custom global variable to the Risor environment and
// registers it as a requireable module that can be validated with require(["@name"]).
//
// This allows you to expose custom Go types and their methods to scripts, while
// enabling scripts to explicitly declare their dependencies for validation.
//
// The global will be:
//   - Added to the Risor environment (accessible as a global variable)
//   - Registered in enabledModules (can be validated with require())
//   - Provide clear error messages if required but not available
//
// Example:
//
//	type MyApp struct { ... }
//	func (app *MyApp) GetAttr(name string) (object.Object, bool) { ... }
//
//	myApp := &MyApp{...}
//	w := fynerisor.NewApp("My Application",
//	    fynerisor.WithGlobal("myapp", myApp),
//	)
//
// Usage in script:
//
//	require(["v1.0", "@gui", "@myapp"])  // Validates @myapp is available
//	myapp.DoSomething()                  // Use the custom API
//
// If a script tries to require @myapp in a different application that doesn't
// provide it, the require() call will fail with a clear error message instead
// of causing undefined variable errors later.
//
// This works for any custom object: application APIs, database connections,
// configuration objects, or any other Go type you want to expose to scripts.
func WithGlobal(name string, value any) Option {
	return moduleOption{
		fn: func(env map[string]any, userGlobals *[]risor.Option, modules map[string]bool) {
			env[name] = value
			modules[name] = true // Register for require() validation
		},
	}
}

// WithExec enables the exec module for running external commands from Risor scripts.
func WithExec() Option {
	return moduleOption{
		fn: func(env map[string]any, userGlobals *[]risor.Option, modules map[string]bool) {
			env["exec"] = exec.Module()
			modules["exec"] = true
		},
	}
}
