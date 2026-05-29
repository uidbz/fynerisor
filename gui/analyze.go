package gui

import (
	"context"
	"fmt"
	"strings"

	"github.com/deepnoodle-ai/risor/v2"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// Requirements represents parsed script requirements
type Requirements struct {
	MinVersion      string   // Minimum version (e.g., "0.2.0")
	ExactVersion    string   // Exact version if ==v syntax used
	RequiredModules []string // List of required modules (e.g., ["sql", "http"])
	Imports         []string // List of import paths/URLs (e.g., ["utils.risor", "http://..."])
	Raw             []string // Raw requirement strings
}

// AnalyzeRequirements parses a Risor script and extracts all require() calls.
// This function is useful for external applications that need to understand
// script dependencies before execution.
//
// It compiles the script and inspects require() calls without executing the
// full script logic.
//
// Example:
//
//	script := `
//	    require(["v0.2", "@sql", "@http"])
//	    let btn = widget.NewButton("Hello", () => {})
//	`
//	reqs, err := fynerisor.AnalyzeRequirements(script)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Requires version:", reqs.MinVersion)
//	fmt.Println("Requires modules:", reqs.RequiredModules)
func AnalyzeRequirements(script string) (*Requirements, error) {
	reqs := &Requirements{
		RequiredModules: []string{},
		Imports:         []string{},
		Raw:             []string{},
	}

	// Create a mock require function that captures requirements
	capturedReqs := []string{}
	mockRequire := object.NewBuiltin("require", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) != 1 {
			return object.Nil, nil
		}

		// Handle string argument
		if str, ok := args[0].(*object.String); ok {
			capturedReqs = append(capturedReqs, str.Value())
			return object.Nil, nil
		}

		// Handle list argument
		if list, ok := args[0].(*object.List); ok {
			for _, item := range list.Value() {
				if str, ok := item.(*object.String); ok {
					capturedReqs = append(capturedReqs, str.Value())
				}
			}
			return object.Nil, nil
		}

		return object.Nil, nil
	})

	// Create a mock import function that captures imports
	capturedImports := []string{}
	mockImport := object.NewBuiltin("import", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) != 1 {
			return object.Nil, nil
		}

		// Handle string argument
		if str, ok := args[0].(*object.String); ok {
			capturedImports = append(capturedImports, str.Value())
			return object.Nil, nil
		}

		// Handle list argument
		if list, ok := args[0].(*object.List); ok {
			for _, item := range list.Value() {
				if str, ok := item.(*object.String); ok {
					capturedImports = append(capturedImports, str.Value())
				}
			}
			return object.Nil, nil
		}

		return object.Nil, nil
	})

	// Create minimal environment with mock functions
	env := map[string]any{
		"require": mockRequire,
		"import":  mockImport,
	}

	// Try to evaluate just the require/import calls
	// We'll parse line by line looking for require() and import() calls
	lines := strings.Split(script, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "require(") || strings.HasPrefix(line, "import(") {
			// Extract the requirement/import string
			// Try to evaluate just this line
			_, err := risor.Eval(context.Background(), line, risor.WithEnv(env))
			if err != nil {
				// Ignore errors - script might reference undefined variables
				continue
			}
		}
	}

	// Parse captured requirements
	for _, req := range capturedReqs {
		reqs.Raw = append(reqs.Raw, req)

		// Parse version requirement
		if strings.HasPrefix(req, "==v") {
			version := strings.TrimPrefix(req, "==v")
			reqs.ExactVersion = version
		} else if strings.HasPrefix(req, "v") {
			version := strings.TrimPrefix(req, "v")
			if reqs.MinVersion == "" || compareVersions(version, reqs.MinVersion) > 0 {
				reqs.MinVersion = version
			}
		} else if strings.HasPrefix(req, "@") {
			moduleName := strings.TrimPrefix(req, "@")
			reqs.RequiredModules = append(reqs.RequiredModules, moduleName)
		}
	}

	// Store captured imports (in order)
	reqs.Imports = capturedImports

	return reqs, nil
}

// compareVersions compares two version strings (without 'v' prefix)
// Returns: -1 if v1 < v2, 0 if equal, 1 if v1 > v2
func compareVersions(v1, v2 string) int {
	p1, _ := parseVersion(v1)
	p2, _ := parseVersion(v2)

	for i := 0; i < 3; i++ {
		if p1[i] < p2[i] {
			return -1
		}
		if p1[i] > p2[i] {
			return 1
		}
	}
	return 0
}

// RequiresGUI returns true if the script requires a GUI window (@gui)
func (r *Requirements) RequiresGUI() bool {
	for _, mod := range r.RequiredModules {
		if mod == "gui" {
			return true
		}
	}
	return false
}

// String returns a human-readable representation of requirements
func (r *Requirements) String() string {
	var parts []string

	if r.ExactVersion != "" {
		parts = append(parts, fmt.Sprintf("version ==v%s", r.ExactVersion))
	} else if r.MinVersion != "" {
		parts = append(parts, fmt.Sprintf("version >=v%s", r.MinVersion))
	}

	if len(r.RequiredModules) > 0 {
		parts = append(parts, fmt.Sprintf("modules: %s", strings.Join(r.RequiredModules, ", ")))
	}

	if len(r.Imports) > 0 {
		parts = append(parts, fmt.Sprintf("imports: %s", strings.Join(r.Imports, ", ")))
	}

	if len(parts) == 0 {
		return "no requirements"
	}

	return strings.Join(parts, "; ")
}
