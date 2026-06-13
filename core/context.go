// Package core provides headless Risor execution without GUI dependencies.
// This enables static compilation and smaller binaries for CLI tools and servers.
package core

import (
	"fmt"
	"sync"

	"github.com/deepnoodle-ai/risor/v2"
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
	appName        string // Name of the embedding application
	runner         *ScriptRunner

	// Module import system
	moduleCache map[string]*ImportedModule // Cache of imported modules
	importStack []string                   // Track currently importing modules for circular detection
	moduleMutex sync.Mutex                 // Protect cache from concurrent access
}

// NewContext creates a new Risor context for non-GUI applications.
//
// Parameters:
//   - opts: Optional configuration using functional options (WithHTTP, WithSQL, etc.)
//
// Returns:
//   - *Context: Execution context for Risor scripts
//
// Example:
//
//	ctx := core.NewContext(
//	    core.WithHTTP(),
//	    core.WithSQL(),
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
		moduleCache:    make(map[string]*ImportedModule),
		importStack:    []string{},
	}

	// Set base globals first
	app := newAppObjectForContext(cb)
	globals := map[string]any{
		"app":     app,
		"print":   newPrintBuiltin(),
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

// LoadScript sets the main script to be executed.
// Call this after all ImportScript() calls.
func (cb *Context) LoadScript(script string) {
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
	// Load the script if provided
	if script != "" {
		cb.runner.LoadScript(script)
	}

	return cb.runner.Eval()
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
