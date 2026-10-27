package gui

import (
	"context"
	"sync"

	"github.com/deepnoodle-ai/risor/v2"
)

// ScriptRunner manages script execution.
// Shared by both Window and ContextBuilder.
type ScriptRunner struct {
	mu      sync.Mutex // guards script and base; the browser swaps scripts while a previous run may still read them (issue #15)
	script  string
	base    string // import base of the script (dir or URL); "" = process CWD (issue #8)
	globals []risor.Option
}

// NewScriptRunner creates a new script runner with the given globals.
func NewScriptRunner(globals []risor.Option) *ScriptRunner {
	return &ScriptRunner{
		globals: globals,
	}
}

// LoadScript sets the script to be executed. Relative imports resolve
// against the process working directory.
func (sr *ScriptRunner) LoadScript(script string) {
	sr.load(script, "")
}

// load sets the script and its import base together, so a run always sees
// a matching pair.
func (sr *ScriptRunner) load(script, base string) {
	sr.mu.Lock()
	sr.script, sr.base = script, base
	sr.mu.Unlock()
}

// snapshot returns the loaded script and its import base atomically.
func (sr *ScriptRunner) snapshot() (script, base string) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	return sr.script, sr.base
}

// currentScript returns the loaded script snapshot.
func (sr *ScriptRunner) currentScript() string {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	return sr.script
}

// Compile parses and compiles the loaded script without executing it.
// Host apps use it to validate scripts before running them (CI gates,
// --check flags). Returns the same errors Eval would report up front.
func (sr *ScriptRunner) Compile() error {
	return sr.CompileContext(context.Background())
}

// CompileContext compiles the loaded script with the given context.
func (sr *ScriptRunner) CompileContext(ctx context.Context) error {
	_, err := risor.Compile(ctx, sr.currentScript(), sr.globals...)
	return err
}

// Eval executes the loaded script.
//
// Returns:
//   - any: The result of script execution
//   - error: Any execution error
func (sr *ScriptRunner) Eval() (any, error) {
	return sr.EvalContext(context.Background())
}

// EvalContext executes the loaded script with the given context. Cancelling
// the context stops the Risor VM: the Risor VM checks ctx.Done() while it
// runs, so Stop()/re-Execute() can interrupt a running script (issue #9).
//
// Returns:
//   - any: The result of script execution
//   - error: Any execution error
func (sr *ScriptRunner) EvalContext(ctx context.Context) (any, error) {
	// Snapshot the script: a host (e.g. the browser on navigation) may call
	// Clear/LoadScript for the next run while this one is still executing.
	return evalScript(ctx, sr.currentScript(), sr.globals)
}

// evalScript evaluates script with the given risor options.
func evalScript(ctx context.Context, script string, opts []risor.Option) (any, error) {
	result, err := risor.Eval(ctx, script, opts...)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Clear clears the loaded script.
func (sr *ScriptRunner) Clear() {
	sr.load("", "")
}
