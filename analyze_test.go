package fynerisor

import (
	"testing"
)

func TestAnalyzeRequirements(t *testing.T) {
	tests := []struct {
		name            string
		script          string
		wantMinVersion  string
		wantModules     []string
		wantImports     []string
		wantRawCount    int
	}{
		{
			name: "single version requirement",
			script: `require("v0.2")
let btn = widget.NewButton("Test", () => {})`,
			wantMinVersion: "0.2",
			wantModules:    []string{},
			wantImports:    []string{},
			wantRawCount:   1,
		},
		{
			name: "single module requirement",
			script: `require("@sql")
let conn = sql.connect("...")`,
			wantMinVersion: "",
			wantModules:    []string{"sql"},
			wantImports:    []string{},
			wantRawCount:   1,
		},
		{
			name: "list of requirements",
			script: `require(["v0.2", "@sql", "@http"])
let btn = widget.NewButton("Test", () => {})`,
			wantMinVersion: "0.2",
			wantModules:    []string{"sql", "http"},
			wantImports:    []string{},
			wantRawCount:   3,
		},
		{
			name: "multiple require calls",
			script: `require("v0.2")
require("@sql")
require("@http")
let btn = widget.NewButton("Test", () => {})`,
			wantMinVersion: "0.2",
			wantModules:    []string{"sql", "http"},
			wantImports:    []string{},
			wantRawCount:   3,
		},
		{
			name: "single import",
			script: `import("utils.risor")
require("v0.2")
let x = utilFunc()`,
			wantMinVersion: "0.2",
			wantModules:    []string{},
			wantImports:    []string{"utils.risor"},
			wantRawCount:   1,
		},
		{
			name: "list of imports",
			script: `import(["utils.risor", "helpers.risor"])
let x = utilFunc()`,
			wantMinVersion: "",
			wantModules:    []string{},
			wantImports:    []string{"utils.risor", "helpers.risor"},
			wantRawCount:   0,
		},
		{
			name: "imports and requirements",
			script: `import("utils.risor")
require(["v0.2", "@http"])
let x = utilFunc()`,
			wantMinVersion: "0.2",
			wantModules:    []string{"http"},
			wantImports:    []string{"utils.risor"},
			wantRawCount:   2,
		},
		{
			name: "no requirements",
			script: `let btn = widget.NewButton("Test", () => {})
window.SetContent(btn)`,
			wantMinVersion: "",
			wantModules:    []string{},
			wantImports:    []string{},
			wantRawCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqs, err := AnalyzeRequirements(tt.script)
			if err != nil {
				t.Fatalf("AnalyzeRequirements() error = %v", err)
			}

			if reqs.MinVersion != tt.wantMinVersion {
				t.Errorf("MinVersion = %v, want %v", reqs.MinVersion, tt.wantMinVersion)
			}

			if len(reqs.RequiredModules) != len(tt.wantModules) {
				t.Errorf("RequiredModules count = %v, want %v", len(reqs.RequiredModules), len(tt.wantModules))
			} else {
				for i, mod := range tt.wantModules {
					if reqs.RequiredModules[i] != mod {
						t.Errorf("RequiredModules[%d] = %v, want %v", i, reqs.RequiredModules[i], mod)
					}
				}
			}

			if len(reqs.Imports) != len(tt.wantImports) {
				t.Errorf("Imports count = %v, want %v", len(reqs.Imports), len(tt.wantImports))
			} else {
				for i, imp := range tt.wantImports {
					if reqs.Imports[i] != imp {
						t.Errorf("Imports[%d] = %v, want %v", i, reqs.Imports[i], imp)
					}
				}
			}

			if len(reqs.Raw) != tt.wantRawCount {
				t.Errorf("Raw count = %v, want %v", len(reqs.Raw), tt.wantRawCount)
			}

			// Test String() method
			str := reqs.String()
			if str == "" && (tt.wantRawCount > 0 || len(tt.wantImports) > 0) {
				t.Error("String() returned empty for non-empty requirements")
			}
		})
	}
}
