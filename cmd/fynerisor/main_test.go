package main

import (
	"errors"
	"testing"

	"github.com/uidbz/fynerisor/core"
)

// Unit tests for the CLI's pure helpers (issue #43).

func TestFormatScriptError(t *testing.T) {
	// Parse error: position comes from the parser error type.
	_, err := core.AnalyzeRequirements("let x = ;")
	if err == nil {
		t.Fatal("expected parse error")
	}
	got := formatScriptError("app.risor", err)
	if got != `app.risor:1:9: parse error: invalid syntax (unexpected ";")` {
		t.Errorf("formatScriptError = %q", got)
	}

	// Plain error: no positions anywhere.
	plain := errors.New("boom")
	if got := formatScriptError("app.risor", plain); got != "app.risor: boom" {
		t.Errorf("formatScriptError plain = %q", got)
	}

	// Compiler error keeps its embedded location line.
	cb := core.NewContext()
	if err := cb.CompileScript(`print(undefined_name)`); err != nil {
		got := formatScriptError("app.risor", err)
		if got != `app.risor:1:7: compile error: undefined variable "undefined_name"` {
			t.Errorf("formatScriptError compile = %q", got)
		}
	} else {
		t.Fatal("expected compile error")
	}
}

func TestScriptUsesGUI(t *testing.T) {
	cases := []struct {
		script string
		want   bool
	}{
		{`print("hello")`, false},
		{`require(["@gui"])`, true},
		{`let b = widget.NewButton("x", () => {})`, true},
		{`print(window.Title)`, true},
		{`let m = import("lib.risor")`, false},
		{`print("widget.NewButton is only text")`, false},
		{`go(() => { print(1) })`, true},
	}
	for _, c := range cases {
		got, err := scriptUsesGUI(c.script)
		if err != nil {
			t.Errorf("%q: unexpected error %v", c.script, err)
			continue
		}
		if got != c.want {
			t.Errorf("scriptUsesGUI(%q) = %v, want %v", c.script, got, c.want)
		}
	}
}
