package core

import (
	"strings"
	"testing"

	"github.com/deepnoodle-ai/risor/v2"
)

// This file covers issue #27: module compilation and the module VM must
// agree on the visible globals. Names added via WithGlobal are in the
// shared environment (visible to modules); names smuggled in through
// opaque WithRisorOptions are not, so they must fail at import time
// with a clear error instead of at runtime with a mystery nil.

func TestModuleSeesWithGlobalValues(t *testing.T) {
	dir := t.TempDir()
	mod := writeModule(t, dir, "uses.risor", `
let hi = () => cfg["name"] + "!"
`)

	cb := NewContext(WithGlobal("cfg", map[string]any{"name": "widget"}))
	res, err := cb.Eval(`let m = import("` + mod + `")
m.hi()
`)
	if err != nil {
		t.Fatalf("module could not use WithGlobal value: %v", err)
	}
	if got, _ := res.(string); got != "widget!" {
		t.Errorf("module returned %v, want widget!", res)
	}
}

func TestModuleFailsClearlyOnRisorOptionsGlobals(t *testing.T) {
	dir := t.TempDir()
	mod := writeModule(t, dir, "opaque.risor", `
let hi = () => smuggled + "!"
`)

	cb := NewContext(WithRisorOptions(risor.WithEnv(map[string]any{"smuggled": "x"})))

	// The main script can still see the option global...
	if _, err := cb.Eval(`smuggled`); err != nil {
		t.Fatalf("main script lost WithRisorOptions global: %v", err)
	}

	// ...but the module import now fails with a clear compile-time
	// error, not with a runtime surprise.
	_, err := cb.Eval(`let m = import("` + mod + `")`)
	if err == nil {
		t.Fatal("module import unexpectedly succeeded with an invisible global")
	}
	if !strings.Contains(err.Error(), "smuggled") {
		t.Errorf("error = %v, want it to name the undefined identifier", err)
	}
}
