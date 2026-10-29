// Package core provides headless Risor execution without GUI dependencies.
// This enables static compilation and smaller binaries for CLI tools and servers.
package core

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/deepnoodle-ai/risor/v2"

	iomod "github.com/uidbz/fynerisor/modules/io"
	osmod "github.com/uidbz/fynerisor/modules/os"
)

// Context builds Risor execution contexts for non-GUI applications.
// It provides access to modules (HTTP, SQL, OS, etc.) and import functionality
// without requiring Fyne or any GUI dependencies.
//
// This enables static compilation for headless scripts, CLI tools, or server
// applications that need Risor scripting with modules.
type Context struct {
	globals        []risor.Option
	env            map[string]any // Merged built-in globals (shared with imported module VMs)
	userGlobals    []risor.Option // Opaque user-supplied globals (via WithGlobals)
	enabledModules map[string]bool
	appName        string   // Name of the embedding application
	appVersion     string   // Embedding app version for require() checks; "" = deprecated global (issue #44)
	scriptArgs     []string // This context's os.args(); carried in the eval context (issue #44)
	stdin          io.Reader
	stdout         io.Writer // print's sink; nil = process stdout (issue #42)
	stderr         io.Writer // eprint's sink; nil = process stderr (issue #42)
	runner         *ScriptRunner

	// Module import system
	moduleCache map[string]*ImportedModule // Cache of imported modules, keyed by canonical path/URL
	importStack []ImportFrame              // In-progress imports: cycle detection + base resolution
	scriptBase  string                     // Base of the entry script (dir or URL); "" = process CWD
	moduleMutex sync.Mutex                 // Protect cache from concurrent access
}

// NewContext creates a new Risor context for non-GUI applications.
//
// Parameters:
//   - opts: Optional configuration using functional options (WithHTTP, WithOS, etc.;
//     for SQL use the opt-in core/db subpackage)
//
// Returns:
//   - *Context: Execution context for Risor scripts
//
// Example:
//
//	ctx := core.NewContext(
//	    core.WithHTTP(),
//	    db.WithSQL(), // import "github.com/uidbz/fynerisor/core/db"
//	    core.WithOS(),
//	)
//
//	script := `
//	    require(["@http", "@sql"])
//	    let data = http.get("https://api.example.com/data").json()
//	    print(data)
//	`
//
//	result, err := ctx.Eval(script)
func NewContext(opts ...Option) *Context {
	cb := &Context{
		enabledModules: make(map[string]bool),
		appName:        "fynerisor", // default
		appVersion:     GetAppVersion(),
		moduleCache:    make(map[string]*ImportedModule),
		importStack:    []ImportFrame{},
	}

	// Set base globals first
	app := newAppObjectForContext(cb)
	globals := map[string]any{
		"app":     app,
		"print":   newPrintBuiltin(),
		"eprint":  newEprintBuiltin(),
		"require": newRequireBuiltinForContext(cb),
		"import":  cb.newImportBuiltin(),
	}

	// Build a single merged environment containing the standard library and all
	// built-in globals. This same map is passed to imported module VMs so that
	// functions defined in modules can access globals like http, sql, etc.
	cb.env = risor.Builtins()
	for k, v := range globals {
		cb.env[k] = v
	}

	// Apply functional options after (so they can populate cb.env / cb.userGlobals)
	for _, opt := range opts {
		opt.applyToContext(cb)
	}

	// Compose the final risor options: the merged env first, then any opaque
	// user-supplied globals (from WithGlobals).
	cb.globals = append([]risor.Option{risor.WithEnv(cb.env)}, cb.userGlobals...)

	// Create script runner
	cb.runner = NewScriptRunner(cb.globals)

	return cb
}

// ImportScript is deprecated and has been removed in favor of runtime import().
// Use the import() function directly in scripts for module-scoped imports:
//
//	let utils = import("utils.risor")
//	utils.myFunction()
//
// This provides proper namespacing and prevents global scope pollution.
// The old concatenation-based import system is no longer supported.
func (cb *Context) ImportScript(source string) error {
	return fmt.Errorf("ImportScript() is deprecated: use import() function in scripts instead")
}

// LoadScript sets the main script to be executed. Relative import() paths
// in the script resolve against the process working directory; use
// LoadScriptFrom when the script's own location is known (issue #8).
func (cb *Context) LoadScript(script string) {
	cb.scriptBase = ""
	cb.runner.LoadScript(script)
}

// LoadScriptFrom sets the main script together with its source (file path
// or URL). Relative import() calls in the script resolve against that
// source: the containing directory for local files, the source URL (same
// origin only) for http(s) sources. See issue #8.
func (cb *Context) LoadScriptFrom(script, source string) {
	cb.scriptBase = ImportBaseFor(source)
	cb.runner.LoadScript(script)
}

// Eval evaluates a Risor script and returns the result.
//
// Parameters:
//   - script: The Risor script source code
//
// Returns:
//   - any: The script result (converted to Go types)
//   - error: Any evaluation error
//
// Example:
//
//	// Direct eval
//	ctx := core.NewContext(core.WithHTTP())
//	result, err := ctx.Eval(`http.get("https://example.com").status`)
//
//	// With imports (runtime module scoping)
//	result, err := ctx.Eval(`
//	    let utils = import("utils.risor")
//	    utils.myUtil(42)
//	`)
func (cb *Context) Eval(script string) (any, error) {
	return cb.EvalContext(context.Background(), script)
}

// EvalContext evaluates a Risor script with the given context. Cancelling
// the context stops the Risor VM, which lets callers bound or interrupt
// headless runs (issue #9).
func (cb *Context) EvalContext(ctx context.Context, script string) (any, error) {
	// Load the script if provided
	if script != "" {
		cb.runner.LoadScript(script)
	}

	return cb.runner.EvalContext(cb.runContext(ctx))
}

// runContext stamps this Context's per-runtime state onto the eval
// context so builtins read it instead of process globals (issue #44):
// the module VMs and every builtin share this context.
func (cb *Context) runContext(ctx context.Context) context.Context {
	if cb.scriptArgs != nil {
		ctx = osmod.ContextWithArgs(ctx, cb.scriptArgs)
	}
	ctx = iomod.ContextWithStreams(ctx, cb)
	return ctx
}

// CompileScript compiles the given script against this Context's
// environment (globals, modules) without executing it (issue #43:
// `fynerisor check`). The returned error carries the parse/compile
// diagnostics.
func (cb *Context) CompileScript(script string) error {
	_, err := risor.Compile(context.Background(), script, cb.globals...)
	return err
}

// Stdin/Stdout/Stderr implement modules/io.StreamSource (issue #42):
// io.stdin, io.stdout/io.stderr and print/eprint follow this Context's
// configured streams, with the process streams as defaults.
func (cb *Context) Stdin() io.Reader {
	if cb.stdin != nil {
		return cb.stdin
	}
	return os.Stdin
}

func (cb *Context) Stdout() io.Writer {
	if cb.stdout != nil {
		return cb.stdout
	}
	return os.Stdout
}

func (cb *Context) Stderr() io.Writer {
	if cb.stderr != nil {
		return cb.stderr
	}
	return os.Stderr
}

// EvalWithImports is deprecated. Use runtime import() function instead:
//
//	script := `
//	    let utils = import("utils.risor")
//	    require("@http")
//	    let result = utils.myUtil(42)
//	`
//	result, err := ctx.Eval(script)
//
// The new import() provides proper module scoping with namespace isolation.
func (cb *Context) EvalWithImports(script string, fetchFunc func(path string) (string, error)) (any, error) {
	return nil, fmt.Errorf("EvalWithImports() is deprecated: use import() function in scripts for module-scoped imports")
}

// EnabledModules returns a map of enabled module names.
func (cb *Context) EnabledModules() map[string]bool {
	modules := make(map[string]bool, len(cb.enabledModules))
	for k, v := range cb.enabledModules {
		modules[k] = v
	}
	return modules
}
