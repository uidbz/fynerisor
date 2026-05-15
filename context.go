package fynerisor

import (
	"context"

	"github.com/deepnoodle-ai/risor/v2"
)

// ContextBuilder builds Risor contexts for non-GUI applications.
// It provides access to the same module system (HTTP, SQL, OS, etc.) and import
// functionality as fynerisor Window, but without GUI dependencies.
//
// This is useful for headless scripts, CLI tools, or server applications that
// need Risor scripting with modules but don't require a GUI.
type ContextBuilder struct {
	globals        []risor.Option
	enabledModules map[string]bool
	appName        string // Name of the embedding application
	runner         *ScriptRunner
}

// NewContext creates a new Risor context builder for non-GUI applications.
//
// Parameters:
//   - opts: Optional configuration using functional options (WithHTTP, WithSQL, etc.)
//
// Returns:
//   - *ContextBuilder: Builder for creating Risor contexts
//
// Example:
//
//	ctx := fynerisor.NewContext(
//	    fynerisor.WithHTTP(),
//	    fynerisor.WithSQL(),
//	    fynerisor.WithOS(),
//	)
//
//	script := `
//	    require(["@http", "@sql"])
//	    let data = http.get("https://api.example.com/data").json()
//	    print(data)
//	`
//
//	result, err := ctx.Eval(script)
func NewContext(opts ...Option) *ContextBuilder {
	cb := &ContextBuilder{
		enabledModules: make(map[string]bool),
		appName:        "fynerisor", // default
	}

	// Set base globals first
	app := newAppObjectForContext(cb)
	globals := map[string]any{
		"app":     app,
		"print":   newPrintBuiltin(),
		"require": newRequireBuiltinForContext(cb),
		"import":  newImportBuiltin(),
	}

	cb.globals = []risor.Option{risor.WithEnv(risor.Builtins()), risor.WithEnv(globals)}

	// Apply functional options after (so they can append to cb.globals)
	for _, opt := range opts {
		opt.applyToContext(cb)
	}

	// Create script runner
	cb.runner = NewScriptRunner(cb.globals)

	return cb
}

// ImportScript loads a script from a path or URL and adds it to the import list.
// The script will be executed before the main script when Eval() is called.
//
// Parameters:
//   - source: Path to a local file or HTTP(S) URL
//
// Returns:
//   - error: Any error encountered while fetching the script
//
// Example:
//
//	ctx := fynerisor.NewContext(fynerisor.WithHTTP())
//	ctx.ImportScript("utils.risor")
//	ctx.ImportScript("https://example.com/helpers.risor")
//	result, err := ctx.Eval(mainScript)
func (cb *ContextBuilder) ImportScript(source string) error {
	return cb.runner.ImportScript(source)
}

// LoadScript sets the main script to be executed.
// Call this after all ImportScript() calls.
func (cb *ContextBuilder) LoadScript(script string) {
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
//	ctx := fynerisor.NewContext(fynerisor.WithHTTP())
//	result, err := ctx.Eval(`http.get("https://example.com").status`)
//
//	// With imports
//	ctx.ImportScript("utils.risor")
//	ctx.LoadScript(`let result = myUtil(42)`)
//	result, err := ctx.Eval("")
func (cb *ContextBuilder) Eval(script string) (any, error) {
	// If no script was loaded via LoadScript, use the provided script
	if len(cb.runner.scriptParts) == 0 && script != "" {
		cb.runner.LoadScript(script)
	}

	return cb.runner.Eval()
}

// EvalWithImports analyzes the script for imports, loads them in order,
// then evaluates the main script. This method creates a shared Risor context
// so imported scripts can define functions/variables used by the main script.
//
// Parameters:
//   - script: The Risor script source code
//   - fetchFunc: Function to fetch import sources by path/URL
//
// Returns:
//   - any: The script result (converted to Go types)
//   - error: Any evaluation error
//
// Example:
//
//	ctx := fynerisor.NewContext(fynerisor.WithContextHTTP())
//
//	fetchFunc := func(path string) (string, error) {
//	    data, err := os.ReadFile(path)
//	    return string(data), err
//	}
//
//	script := `
//	    import("utils.risor")
//	    require("@http")
//	    let result = myUtil(42)
//	`
//
//	result, err := ctx.EvalWithImports(script, fetchFunc)
func (cb *ContextBuilder) EvalWithImports(script string, fetchFunc func(path string) (string, error)) (any, error) {
	// Analyze script for imports
	reqs, err := AnalyzeRequirements(script)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	// Create a shared Risor context by evaluating imports first
	// Each eval in the same context shares the global scope
	combinedScript := ""

	// Load imports in order
	for _, importPath := range reqs.Imports {
		importScript, err := fetchFunc(importPath)
		if err != nil {
			return nil, err
		}
		combinedScript += importScript + "\n"
	}

	// Add main script
	combinedScript += script

	// Execute all scripts in one context so imports can define functions
	result, err := risor.Eval(ctx, combinedScript, cb.globals...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// EnabledModules returns a map of enabled module names.
func (cb *ContextBuilder) EnabledModules() map[string]bool {
	modules := make(map[string]bool, len(cb.enabledModules))
	for k, v := range cb.enabledModules {
		modules[k] = v
	}
	return modules
}
