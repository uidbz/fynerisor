package core

import (
	"github.com/deepnoodle-ai/risor/v2"
	"github.com/uidbz/fynerisor/modules/exec"
	filepathmod "github.com/uidbz/fynerisor/modules/filepath"
	"github.com/uidbz/fynerisor/modules/http"
	iomod "github.com/uidbz/fynerisor/modules/io"
	"github.com/uidbz/fynerisor/modules/os"
	"github.com/uidbz/fynerisor/modules/sql"
	"github.com/uidbz/fynerisor/modules/strings"
	"github.com/uidbz/fynerisor/modules/tie"
	"github.com/uidbz/fynerisor/modules/time"
)

// Option configures a Context during creation.
type Option interface {
	applyToContext(*Context)
}

type moduleOption struct {
	// fn configures an imported environment. Built-in module options write their
	// global objects into env (so they can be passed to both the main script and
	// any imported module VMs). User-supplied opaque options (WithGlobals) are
	// appended to userGlobals instead, since their values can't be introspected.
	fn func(env map[string]any, userGlobals *[]risor.Option, modules map[string]bool)
}

func (o moduleOption) applyToContext(cb *Context) {
	o.fn(cb.env, &cb.userGlobals, cb.enabledModules)
}

type appNameOption struct {
	name string
}

func (o appNameOption) applyToContext(cb *Context) {
	cb.appName = o.name
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
//	ctx := core.NewContext(
//	    core.WithRisorOptions(risor.WithEnv(customFuncs)),
//	)
func WithRisorOptions(globals ...risor.Option) Option {
	return moduleOption{
		fn: func(env map[string]any, userGlobals *[]risor.Option, modules map[string]bool) {
			*userGlobals = append(*userGlobals, globals...)
		},
	}
}

// WithAppName sets the application name exposed to Risor scripts via app.name.
// This allows scripts to detect which application is running them.
//
// Example:
//
//	ctx := core.NewContext(
//	    core.WithAppName("myapp"),
//	)
//
// Usage in script:
//
//	if (app.name == "myapp") {
//	    print("Running in myapp")
//	}
func WithAppName(name string) Option {
	return appNameOption{name: name}
}

// WithHTTP enables the HTTP module for making HTTP requests from Risor scripts.
// The module provides functions: get, post, put, delete, and fetch.
//
// Example:
//
//	ctx := core.NewContext(core.WithHTTP())
//
// Usage in script:
//
//	let response = http.get("https://api.example.com/data")
//	print(response.status, response.body)
func WithHTTP() Option {
	return moduleOption{
		fn: func(env map[string]any, userGlobals *[]risor.Option, modules map[string]bool) {
			env["http"] = http.Module()
			modules["http"] = true
		},
	}
}

// WithHTTPImport enables importing Risor modules from HTTP(S) URLs.
// This is a separate option from WithHTTP() for security reasons.
//
// When enabled, scripts can use import() with HTTP(S) URLs:
//
//	let utils = import("https://example.com/utils.risor")
//	utils.someFunction()
//
// Security considerations:
//   - Only enable this for trusted scripts or controlled environments
//   - Imported modules execute with full access to enabled modules
//   - HTTPS is recommended for production
//
// Example:
//
//	ctx := core.NewContext(
//	    core.WithHTTP(),
//	    core.WithHTTPImport(),
//	)
//
// Usage in script:
//
//	require(["@httpimport"])
//	let remote = import("https://cdn.example.com/mylib.risor")
func WithHTTPImport() Option {
	return moduleOption{
		fn: func(env map[string]any, userGlobals *[]risor.Option, modules map[string]bool) {
			modules["httpimport"] = true
		},
	}
}

// WithOS enables the OS module for accessing OS functionality from Risor scripts.
// The module provides functions: goos, current_user, open_browser, read_file,
// write_file, and read_dir.
//
// Example:
//
//	ctx := core.NewContext(core.WithOS())
//
// Usage in script:
//
//	let platform = os.goos()
//	let user = os.current_user()
//	os.write_file("test.txt", "Hello!")
func WithOS() Option {
	return moduleOption{
		fn: func(env map[string]any, userGlobals *[]risor.Option, modules map[string]bool) {
			env["os"] = os.Module()
			modules["os"] = true
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
//	ctx := core.NewContext(core.WithTime())
//
// Usage in script:
//
//	let now = time.now()
//	let date = time.date(2026, 5, 1)
func WithTime() Option {
	return moduleOption{
		fn: func(env map[string]any, userGlobals *[]risor.Option, modules map[string]bool) {
			env["time"] = time.Module()
			modules["time"] = true
		},
	}
}

// WithSQL enables the SQL module for database connectivity from Risor scripts.
// Supports: MySQL, PostgreSQL, SQLite, SQL Server
//
// Example:
//
//	ctx := core.NewContext(core.WithSQL())
//
// Usage in script:
//
//	let conn = sql.connect("sqlite3::memory:")
//	conn.exec("CREATE TABLE users (id INT, name TEXT)")
//	let rows = conn.query("SELECT * FROM users").collect()
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
//	ctx := core.NewContext(core.WithTie())
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

// WithIO enables the IO module for file I/O operations from Risor scripts.
// The module provides functions: cp and read_all.
//
// Example:
//
//	ctx := core.NewContext(core.WithIO())
//
// Usage in script:
//
//	io.cp("source.txt", "dest.txt")
//	let content = io.read_all("file.txt")
func WithIO() Option {
	return moduleOption{
		fn: func(env map[string]any, userGlobals *[]risor.Option, modules map[string]bool) {
			env["io"] = iomod.NewIO()
			modules["io"] = true
		},
	}
}

// WithExec enables the exec module for running external commands from Risor scripts.
// The module provides functions: command, look_path, and exec.
//
// Example:
//
//	ctx := core.NewContext(core.WithExec())
//
// Usage in script:
//
//	let result = exec(["ls", "-la"])
//	print(result.stdout)
func WithExec() Option {
	return moduleOption{
		fn: func(env map[string]any, userGlobals *[]risor.Option, modules map[string]bool) {
			env["exec"] = exec.Module()
			modules["exec"] = true
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
//	type MyAPI struct { ... }
//	func (api *MyAPI) GetAttr(name string) (object.Object, bool) { ... }
//
//	myAPI := &MyAPI{...}
//	ctx := core.NewContext(
//	    core.WithGlobal("myapi", myAPI),
//	)
//
// Usage in script:
//
//	require(["v1.0", "@myapi"])  // Validates @myapi is available
//	myapi.DoSomething()          // Use the custom API
//
// If a script tries to require @myapi in a different application that doesn't
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
