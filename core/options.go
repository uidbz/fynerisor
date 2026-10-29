package core

import (
	"io"

	"github.com/deepnoodle-ai/risor/v2"
	csvmod "github.com/uidbz/fynerisor/modules/csv"
	"github.com/uidbz/fynerisor/modules/exec"
	filepathmod "github.com/uidbz/fynerisor/modules/filepath"
	"github.com/uidbz/fynerisor/modules/http"
	iomod "github.com/uidbz/fynerisor/modules/io"
	jsonmod "github.com/uidbz/fynerisor/modules/json"
	osmod "github.com/uidbz/fynerisor/modules/os"
	stringsmod "github.com/uidbz/fynerisor/modules/strings"
	tiemod "github.com/uidbz/fynerisor/modules/tie"
	timemod "github.com/uidbz/fynerisor/modules/time"
)

// Options assemble a shared Config (issue #40); the same values are
// accepted by core.NewContext and by gui.NewWindow/NewCanvasObject.
// Capability presets (Safe/Files/Exec/HTTP/Full) live in presets.go.

// WithRisorOptions adds advanced Risor VM configuration options.
// These are opaque risor.Option objects for advanced Risor VM configuration.
// Unlike WithGlobal(), these options are NOT forwarded to imported modules.
//
// Use WithGlobal() instead for named globals that modules should access.
// Do not define script-visible names through
// WithRisorOptions(risor.WithEnv(...)): imported modules are compiled and
// run against the shared environment (the WithGlobal/module values), so
// such names would be undefined in modules; keep this option for VM
// tuning (issue #27).
func WithRisorOptions(globals ...risor.Option) Option {
	return NewOption(func(c *Config) {
		c.UserVMOptions = append(c.UserVMOptions, globals...)
	})
}

// WithAppName sets the application name exposed to Risor scripts via app.name.
//
// Usage in script:
//
//	if (app.name == "myapp") {
//	    print("Running in myapp")
//	}
func WithAppName(name string) Option {
	return NewOption(func(c *Config) { c.AppName = name })
}

// WithHTTP enables the HTTP module (get, post, put, delete, fetch).
//
// Usage in script:
//
//	let response = http.get("https://example.com/data")
func WithHTTP() Option {
	return Module("http", http.Module())
}

// WithHTTPImport enables importing Risor modules from HTTP(S) URLs.
// This is a separate option from WithHTTP() for security reasons:
// imported remote code executes with full access to enabled modules,
// so only enable it for trusted scripts or controlled environments.
//
// Usage in script:
//
//	require(["@httpimport"])
//	let remote = import("https://example.com/utils.risor")
func WithHTTPImport() Option {
	return markEnabled("httpimport")
}

// WithOS enables the OS module: goos, current_user, home_dir,
// open_browser, read_file, write_file, create, copy_file, read_dir,
// mkdir, mkdir_all, remove_all, is_dir, getwd, args. Prefer Files() or
// Files(true) (issue #40) to grant read-only access, and the exec
// module over os.exec.
func WithOS() Option {
	return Module("os", osmod.Module())
}

// WithScriptArgs sets the script's command-line arguments for *this
// runtime (os.args); they travel through the eval context, so two
// runtimes in one process never share args (issue #44).
func WithScriptArgs(args ...string) Option {
	return NewOption(func(c *Config) { c.ScriptArgs = args })
}

// WithAppVersion sets the version scripts see as app.version and that
// require(["v1.2"]) validates against, for *this runtime (issue #44).
func WithAppVersion(version string) Option {
	return NewOption(func(c *Config) { c.AppVersion = version })
}

// WithStdin sets where this runtime's io.stdin reads from (issue #42).
func WithStdin(in io.Reader) Option {
	return NewOption(func(c *Config) { c.Stdin = in })
}

// WithStdout sets where the script's print output goes for this runtime
// (issue #24); no process-wide redirection, each runtime captures its
// own. A *widget.Log is an io.Writer; io.MultiWriter combines sinks.
func WithStdout(out io.Writer) Option {
	return NewOption(func(c *Config) { c.Stdout = out })
}

// WithStderr sets where this runtime reports script-execution errors
// and library diagnostics (issue #24).
func WithStderr(err io.Writer) Option {
	return NewOption(func(c *Config) { c.Stderr = err })
}

// WithStrings enables the strings module.
func WithStrings() Option {
	return Module("strings", stringsmod.Module())
}

// WithFilepath enables the filepath module.
func WithFilepath() Option {
	return Module("filepath", filepathmod.Module())
}

// WithTime enables the time module (now, date, parse).
func WithTime() Option {
	return Module("time", timemod.Module())
}

// WithTie enables the Tie triple-store module.
//
// Usage in script:
//
//	let db = tie.connect("http://localhost:1161")
func WithTie() Option {
	return Module("tie", tiemod.Module())
}

// WithIO enables the io stream module (read, lines, write, copy;
// stdin/stdout/stderr handles — issues #41/#42).
func WithIO() Option {
	return Module("io", iomod.NewIO())
}

// WithExec enables the exec module for running external commands.
func WithExec() Option {
	return Module("exec", exec.Module())
}

// WithJSON enables the json transform module (parse, format, marshal,
// marshal_indent, valid; read/write are deprecated sugar — issue #41).
func WithJSON() Option {
	return Module("json", jsonmod.Module())
}

// WithCSV enables the csv transform module (parse, format; read/write
// are deprecated sugar — issue #41).
func WithCSV() Option {
	return Module("csv", csvmod.Module())
}

// WithGlobal adds a custom global variable to the Risor environment and
// registers it as a requireable module validatable with require(["@name"]).
//
// This works for any custom object: application APIs, database
// connections, configuration maps, or any Go value you want to expose
// to scripts. If a script requires @name where it is not provided,
// require() fails with a clear message instead of an undefined-variable
// error later.
//
// Example:
//
//	ctx := core.NewContext(core.WithGlobal("myapi", &MyAPI{}))
//
// Usage in script:
//
//	require(["v1.0", "@myapi"])
//	myapi.DoSomething()
func WithGlobal(name string, value any) Option {
	return Module(name, value)
}

// WithSQL was removed from core: it forced the MySQL/PostgreSQL/SQLite/SQL
// Server drivers into every binary that imported core, even when no database
// was used. Use the opt-in subpackage instead:
//
//	import "github.com/uidbz/fynerisor/core/db"
//	ctx := core.NewContext(db.WithSQL())
//
// The gui package likewise moved its SQL option to gui/db (issue #40).
