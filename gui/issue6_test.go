package gui

import (
	"strings"
	"testing"

	"github.com/uidbz/fynerisor/core"
)

// Issue #6: the host app must be able to compile-check a script without
// executing it. Window.CompileScript delegates to ScriptRunner.Compile.
func TestIssue6CompileScript(t *testing.T) {
	sr := core.NewScriptRunner(nil)

	sr.LoadScript("let x = 1\nx + 1")
	if err := sr.Compile(); err != nil {
		t.Fatalf("valid script failed compile: %v", err)
	}

	sr.LoadScript("let x = 1\nx = } else {")
	err := sr.Compile()
	if err == nil {
		t.Fatal("invalid script compiled cleanly")
	}
	if !strings.Contains(err.Error(), "syntax") && !strings.Contains(err.Error(), "else") {
		t.Logf("error shape: %v", err)
	}
}
