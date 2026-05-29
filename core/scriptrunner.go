package core

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/deepnoodle-ai/risor/v2"
)

// ScriptRunner manages script execution with import support.
// Shared by both Window and ContextBuilder.
type ScriptRunner struct {
	scriptParts []string
	globals     []risor.Option
}

// NewScriptRunner creates a new script runner with the given globals.
func NewScriptRunner(globals []risor.Option) *ScriptRunner {
	return &ScriptRunner{
		globals:     globals,
		scriptParts: []string{},
	}
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
//	runner.ImportScript("utils.risor")
//	runner.ImportScript("https://example.com/helpers.risor")
func (sr *ScriptRunner) ImportScript(source string) error {
	var script string
	var err error

	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		script, err = fetchHTTPScript(source)
	} else {
		data, readErr := os.ReadFile(source)
		if readErr != nil {
			return readErr
		}
		script = string(data)
		err = nil
	}

	if err != nil {
		return err
	}

	sr.scriptParts = append(sr.scriptParts, script)
	return nil
}

// LoadScript adds the main script to be executed.
// This should be called after all ImportScript() calls.
func (sr *ScriptRunner) LoadScript(script string) {
	sr.scriptParts = append(sr.scriptParts, script)
}

// Eval executes all imported scripts followed by the main script.
// Returns the result of the final script.
//
// Returns:
//   - any: The result of script execution
//   - error: Any execution error
func (sr *ScriptRunner) Eval() (any, error) {
	ctx := context.Background()

	// Combine all script parts
	combinedScript := strings.Join(sr.scriptParts, "\n")

	// Execute
	result, err := risor.Eval(ctx, combinedScript, sr.globals...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Clear clears all loaded scripts (both imports and main script).
func (sr *ScriptRunner) Clear() {
	sr.scriptParts = []string{}
}

// fetchHTTPScript fetches a script from an HTTP(S) URL.
func fetchHTTPScript(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
