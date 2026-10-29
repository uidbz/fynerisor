package core

import (
	"reflect"
	"strings"
	"testing"
)

// This file covers issue #23: AnalyzeRequirements is an AST walk, not
// a line-based Eval. Acceptance table from the issue: multi-line
// require, `let x = import(...)`, nested calls, comments containing
// `require(`, and a syntax error.

func TestAnalyzeRequirementsAcceptanceTable(t *testing.T) {
	tests := []struct {
		name        string
		script      string
		wantModules []string
		wantImports []string
		wantVersion string
		wantDynamic int // number of dynamic call sites reported
		wantErr     bool
	}{
		{
			name: "multi-line require list",
			script: `require([
    "v0.6",
    "@sql"
])`,
			wantModules: []string{"sql"},
			wantImports: []string{},
			wantVersion: "0.6",
		},
		{
			name:        "import assigned to a variable (the documented idiom)",
			script:      `let utils = import("utils.risor")`,
			wantModules: []string{},
			wantImports: []string{"utils.risor"},
		},
		{
			name: "nested calls in blocks and functions",
			script: `let setup = () => {
    if (true) {
        require("@http")
        let m = import("lib/nested.risor")
    }
}
require(["v1.2"])`,
			wantModules: []string{"http"},
			wantImports: []string{"lib/nested.risor"},
			wantVersion: "1.2",
		},
		{
			name: "require( inside a comment is not a requirement",
			script: `// require("@evil")
/* import("evil.risor") */
print("keep require( as text")`,
			wantModules: []string{},
			wantImports: []string{},
		},
		{
			name:        "syntax error is reported",
			script:      `require(["v0.2",`,
			wantErr:     true,
			wantModules: []string{},
			wantImports: []string{},
		},
		{
			name:        "dynamic arguments are reported, never evaluated",
			script:      `let p = "utils.risor"` + "\n" + `require(["v" + ".9"])` + "\n" + `let m = import(p)`,
			wantModules: []string{},
			wantImports: []string{},
			wantDynamic: 2,
		},
		{
			name: "nothing is executed: undefined names are fine",
			script: `require("@tie")
nonexistent_function(boom)`,
			wantModules: []string{"tie"},
			wantImports: []string{},
		},
		{
			name:        "template string arguments are dynamic",
			script:      `let v = "9"` + "\n" + "import(`lib_${v}.risor`)",
			wantModules: []string{},
			wantImports: []string{},
			wantDynamic: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqs, err := AnalyzeRequirements(tt.script)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected a parse error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(reqs.RequiredModules, tt.wantModules) {
				t.Errorf("RequiredModules = %v, want %v", reqs.RequiredModules, tt.wantModules)
			}
			if !reflect.DeepEqual(reqs.Imports, tt.wantImports) {
				t.Errorf("Imports = %v, want %v", reqs.Imports, tt.wantImports)
			}
			if tt.wantVersion != reqs.MinVersion {
				t.Errorf("MinVersion = %q, want %q", reqs.MinVersion, tt.wantVersion)
			}
			if len(reqs.Dynamic) != tt.wantDynamic {
				t.Errorf("Dynamic = %v, want %d entries", reqs.Dynamic, tt.wantDynamic)
			}
			for _, d := range reqs.Dynamic {
				if !strings.Contains(d, "line") {
					t.Errorf("dynamic site %q lacks a line number", d)
				}
			}
		})
	}
}

// require( inside a *string* is also just text:
func TestAnalyzeRequirementsStringLiteralNotCalled(t *testing.T) {
	reqs, err := AnalyzeRequirements(`print("require(@sql)")`)
	if err != nil {
		t.Fatal(err)
	}
	if len(reqs.RequiredModules) != 0 || len(reqs.Raw) != 0 {
		t.Errorf("string content parsed as a call: %+v", reqs)
	}
}
