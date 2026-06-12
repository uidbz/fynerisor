package gui

import (
	"github.com/deepnoodle-ai/risor/v2"
	filepathmod "github.com/uidbz/fynerisor/modules/filepath"
	"github.com/uidbz/fynerisor/modules/http"
	iomod "github.com/uidbz/fynerisor/modules/io"
	"github.com/uidbz/fynerisor/modules/os"
	"github.com/uidbz/fynerisor/modules/sql"
	"github.com/uidbz/fynerisor/modules/strings"
	"github.com/uidbz/fynerisor/modules/time"
	"github.com/uidbz/fynerisor/modules/exec"
)

// Module options (WithHTTP, WithSQL, etc.) work with both.
// Callback options (WithStatusCallback, WithResultCallback) only work with Window.
type Option interface {
	applyToWindow(*Window)
}

type moduleOption struct {
	fn func(globals *[]risor.Option, modules map[string]bool)
}

func (o moduleOption) applyToWindow(w *Window) {
	o.fn(&w.globals, w.enabledModules)
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

// WithGlobals adds custom global objects to the Risor script environment.
// The globals parameter should be created using risor.WithEnv().
//
// Example:
//
//	customGlobals := map[string]any{
//	    "myAPI": myAPIObject,
//	}
//	window := fynerisor.NewWindow(w,
//	    fynerisor.WithGlobals(risor.WithEnv(customGlobals)),
//	)
func WithGlobals(globals ...risor.Option) Option {
	return moduleOption{
		fn: func(globalsList *[]risor.Option, modules map[string]bool) {
			*globalsList = append(*globalsList, globals...)
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
		fn: func(globalsList *[]risor.Option, modules map[string]bool) {
			httpModule := http.Module()
			httpGlobals := map[string]any{
				"http": httpModule,
			}
			*globalsList = append(*globalsList, risor.WithEnv(httpGlobals))
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
		fn: func(globalsList *[]risor.Option, modules map[string]bool) {
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
		fn: func(globalsList *[]risor.Option, modules map[string]bool) {
			osModule := os.Module()
			osGlobals := map[string]any{
				"os": osModule,
			}
			*globalsList = append(*globalsList, risor.WithEnv(osGlobals))
			modules["os"] = true
		},
	}
}

// WithStrings enables the strings module for string manipulation from Risor scripts.
func WithStrings() Option {
	return moduleOption{
		fn: func(globalsList *[]risor.Option, modules map[string]bool) {
			stringsModule := strings.Module()
			stringsGlobals := map[string]any{
				"strings": stringsModule,
			}
			*globalsList = append(*globalsList, risor.WithEnv(stringsGlobals))
			modules["strings"] = true
		},
	}
}

// WithFilepath enables the filepath module for file path manipulation from Risor scripts.
func WithFilepath() Option {
	return moduleOption{
		fn: func(globalsList *[]risor.Option, modules map[string]bool) {
			filepathModule := filepathmod.Module()
			filepathGlobals := map[string]any{
				"filepath": filepathModule,
			}
			*globalsList = append(*globalsList, risor.WithEnv(filepathGlobals))
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
		fn: func(globalsList *[]risor.Option, modules map[string]bool) {
			timeModule := time.Module()
			timeGlobals := map[string]any{
				"time": timeModule,
			}
			*globalsList = append(*globalsList, risor.WithEnv(timeGlobals))
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
		fn: func(globalsList *[]risor.Option, modules map[string]bool) {
			sqlModule := sql.Module()
			sqlGlobals := map[string]any{
				"sql": sqlModule,
			}
			*globalsList = append(*globalsList, risor.WithEnv(sqlGlobals))
			modules["sql"] = true
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
		fn: func(globalsList *[]risor.Option, modules map[string]bool) {
			ioModule := iomod.NewIO()
			ioGlobals := map[string]any{
				"io": ioModule,
			}
			*globalsList = append(*globalsList, risor.WithEnv(ioGlobals))
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
		fn: func(globalsList *[]risor.Option, modules map[string]bool) {
			customGlobals := map[string]any{
				name: value,
			}
			*globalsList = append(*globalsList, risor.WithEnv(customGlobals))
			modules[name] = true // Register for require() validation
		},
	}
}


// WithExec enables the exec module for running external commands from Risor scripts.
func WithExec() Option {
	return moduleOption{
		fn: func(globalsList *[]risor.Option, modules map[string]bool) {
			execModule := exec.Module()
			execGlobals := map[string]any{
				"exec": execModule,
			}
			*globalsList = append(*globalsList, risor.WithEnv(execGlobals))
			modules["exec"] = true
		},
	}
}
