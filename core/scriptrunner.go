package core

import (
	"context"

	"github.com/deepnoodle-ai/risor/v2"
)

// ScriptRunner manages script execution for headless contexts.
type ScriptRunner struct {
	script  string
	globals []risor.Option
}

// NewScriptRunner creates a new script runner with the given globals.
func NewScriptRunner(globals []risor.Option) *ScriptRunner {
	return &ScriptRunner{
		globals: globals,
	}
}

// LoadScript sets the script to be executed.
func (sr *ScriptRunner) LoadScript(script string) {
	sr.script = script
}

// Compile parses and compiles the loaded script without executing it.
// Host apps use it to validate scripts before running them (CI gates,
// --check flags). Returns the same errors Eval would report up front.
func (sr *ScriptRunner) Compile() error {
	ctx := context.Background()
	_, err := risor.Compile(ctx, sr.script, sr.globals...)
	return err
}

// Eval executes the loaded script.
//
// Returns:
//   - any: The result of script execution
//   - error: Any execution error
func (sr *ScriptRunner) Eval() (any, error) {
	ctx := context.Background()

	// Execute
	result, err := risor.Eval(ctx, sr.script, sr.globals...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Clear clears the loaded script.
func (sr *ScriptRunner) Clear() {
	sr.script = ""
}
