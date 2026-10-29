package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/deepnoodle-ai/risor/v2/pkg/ast"
	"github.com/deepnoodle-ai/risor/v2/pkg/parser"
)

// Requirements represents parsed script requirements
type Requirements struct {
	MinVersion      string   // Minimum version (e.g., "0.2.0")
	ExactVersion    string   // Exact version if ==v syntax used
	RequiredModules []string // List of required modules (e.g., ["sql", "http"])
	Imports         []string // List of import paths/URLs (e.g., ["utils.risor", "http://..."])
	Raw             []string // Raw requirement strings
	Dynamic         []string // require()/import() call sites whose arguments are
	// not string literals (they carry line numbers); the values
	// are unknowable statically and were NOT evaluated (issue #23)
}

// AnalyzeRequirements parses a Risor script and extracts require() and
// import() call sites *without executing anything* (issue #23). The
// script is parsed once with Risor's own parser and walked as an AST:
// `let m = import("utils.risor")`, multi-line `require([...])` and
// calls nested in blocks are all found; `require(` inside a comment is
// not. Arguments that are not string/list literals are recorded in
// Dynamic (with line numbers) rather than evaluated.
//
// A syntax error in the script is returned as an error (the old
// line-based version silently skipped lines it could not evaluate, and
// its error return was always nil).
//
// Example:
//
//	script := `
//	    require(["v0.2", "@sql", "@http"])
//	    let utils = import("utils.risor")
//	`
//	reqs, err := core.AnalyzeRequirements(script)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Requires version:", reqs.MinVersion)
//	fmt.Println("Requires modules:", reqs.RequiredModules)
//	fmt.Println("Imports:", reqs.Imports)
func AnalyzeRequirements(script string) (*Requirements, error) {
	reqs := &Requirements{
		RequiredModules: []string{},
		Imports:         []string{},
		Raw:             []string{},
		Dynamic:         []string{},
	}

	prog, err := parser.Parse(context.Background(), script, nil)
	if err != nil {
		return nil, fmt.Errorf("analyze: %w", err)
	}

	for node := range ast.Preorder(prog) {
		call, ok := node.(*ast.Call)
		if !ok {
			continue
		}
		fn, ok := call.Fun.(*ast.Ident)
		if !ok {
			continue // method calls like utils.require() are not ours
		}
		if fn.Name != "require" && fn.Name != "import" {
			continue
		}

		var literals []string
		static := true
		for _, arg := range call.Args {
			names, ok := literalRequirementStrings(arg)
			if !ok {
				static = false
				continue
			}
			literals = append(literals, names...)
		}
		if !static {
			reqs.Dynamic = append(reqs.Dynamic,
				fmt.Sprintf("%s() at line %d", fn.Name, call.Pos().Line))
		}

		for _, lit := range literals {
			if fn.Name == "import" {
				reqs.Imports = append(reqs.Imports, lit)
				continue
			}
			addRequirement(reqs, lit)
		}
	}

	return reqs, nil
}

// literalRequirementStrings extracts the requirement strings an
// argument denotes statically: a string literal, or a list literal of
// string literals (the require([...]) form). Anything dynamic
// (identifiers, concatenation, template strings, spreads) reports
// ok=false so the caller records it instead of evaluating it.
func literalRequirementStrings(arg ast.Node) ([]string, bool) {
	switch a := arg.(type) {
	case *ast.String:
		if len(a.Exprs) > 0 {
			return nil, false // template string: interpolated, dynamic
		}
		return []string{a.Value}, true
	case *ast.List:
		var out []string
		for _, item := range a.Items {
			names, ok := literalRequirementStrings(item)
			if !ok || len(names) != 1 {
				return nil, false
			}
			out = append(out, names[0])
		}
		return out, true
	default:
		return nil, false
	}
}

// UsesGlobals reports whether the script references any of the given
// names as a bare identifier (issue #43: the CLI uses it to decide
// whether a script can run through the core runtime; a parse error is
// returned so callers can surface it). String literals and comments do
// not count.
func UsesGlobals(script string, names ...string) (bool, error) {
	prog, err := parser.Parse(context.Background(), script, nil)
	if err != nil {
		return false, fmt.Errorf("analyze: %w", err)
	}
	want := make(map[string]bool, len(names))
	for _, n := range names {
		want[n] = true
	}
	for node := range ast.Preorder(prog) {
		if id, ok := node.(*ast.Ident); ok && want[id.Name] {
			return true, nil
		}
	}
	return false, nil
}

// addRequirement records one raw requirement string, splitting it into
// the version/module fields the struct exposes.
func addRequirement(reqs *Requirements, req string) {
	reqs.Raw = append(reqs.Raw, req)

	switch {
	case strings.HasPrefix(req, "==v"):
		reqs.ExactVersion = strings.TrimPrefix(req, "==v")
	case strings.HasPrefix(req, "v"):
		version := strings.TrimPrefix(req, "v")
		if reqs.MinVersion == "" || compareVersions(version, reqs.MinVersion) > 0 {
			reqs.MinVersion = version
		}
	case strings.HasPrefix(req, "@"):
		reqs.RequiredModules = append(reqs.RequiredModules, strings.TrimPrefix(req, "@"))
	}
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

	if len(r.Dynamic) > 0 {
		parts = append(parts, fmt.Sprintf("dynamic: %s", strings.Join(r.Dynamic, ", ")))
	}

	if len(parts) == 0 {
		return "no requirements"
	}

	return strings.Join(parts, "; ")
}
