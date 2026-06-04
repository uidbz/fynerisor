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
	"github.com/uidbz/fynerisor/modules/time"
)

// Option configures a Context during creation.
type Option interface {
	applyToContext(*Context)
}

type moduleOption struct {
	fn func(globals *[]risor.Option, modules map[string]bool)
}

func (o moduleOption) applyToContext(cb *Context) {
	o.fn(&cb.globals, cb.enabledModules)
}

type appNameOption struct {
	name string
}

func (o appNameOption) applyToContext(cb *Context) {
	cb.appName = o.name
}

// WithGlobals adds custom global objects to the Risor script environment.
// The globals parameter should be created using risor.WithEnv().
//
// Example:
//
//	customGlobals := map[string]any{
//	    "myAPI": myAPIObject,
//	}
//	ctx := core.NewContext(
//	    core.WithGlobals(risor.WithEnv(customGlobals)),
//	)
func WithGlobals(globals ...risor.Option) Option {
	return moduleOption{
		fn: func(globalsList *[]risor.Option, modules map[string]bool) {
			*globalsList = append(*globalsList, globals...)
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
//	ctx := core.NewContext(core.WithTime())
//
// Usage in script:
//
//	let now = time.now()
//	let date = time.date(2026, 5, 1)
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
