package core

import (
	"context"
	"sync"

	"github.com/deepnoodle-ai/risor/v2"
)

// ScriptRunner manages script execution. One implementation serves both
// the headless Context and the GUI Window (issue #32); the base field
// carries the script's import base so a run always sees a matching
// script/base pair (issue #8).
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

// LoadScriptFrom sets the script together with its import base (the
// containing directory of a file or a page URL), so relative import()
// calls resolve against the script's own location (issue #8).
func (sr *ScriptRunner) LoadScriptFrom(script, base string) {
	sr.load(script, base)
}

func (sr *ScriptRunner) load(script, base string) {
	sr.mu.Lock()
	sr.script, sr.base = script, base
	sr.mu.Unlock()
}

// Snapshot returns the loaded script and its import base atomically.
func (sr *ScriptRunner) Snapshot() (script, base string) {
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
	// Snapshot the script so a later LoadScript cannot race this run.
	return EvalScript(ctx, sr.currentScript(), sr.globals)
}

// EvalScript evaluates script with the given risor options. GUI runs use
// it with their per-run option snapshot (env taken at Execute time).
func EvalScript(ctx context.Context, script string, opts []risor.Option) (any, error) {
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
