package core

import (
	"fmt"
	"testing"

	osmod "github.com/uidbz/fynerisor/modules/os"
)

// This file covers issue #44 (core half): script args and the app
// version must be per-Context, not process globals. Two Contexts in
// one process (parallel tests, a GUI plus a headless batch) must not
// see each other's configuration.

func TestScriptArgsAndVersionArePerContext(t *testing.T) {
	// Poison the deprecated global: contexts carrying their own values
	// must ignore it, and a context with *no* args must not inherit it
	// when other runtimes set one (the old behavior leaked "the last
	// host that configured wins" into every script).
	osmod.SetScriptArgs([]string{"poison"})
	defer osmod.SetScriptArgs(nil)

	a := NewContext(WithOS(), WithScriptArgs("alpha", "--force"), WithAppVersion("1.2.3"))

	res, err := a.Eval(`os.args()`)
	if err != nil {
		t.Fatalf("os.args(): %v", err)
	}
	if got := fmt.Sprintf("%v", res); got != "[alpha --force]" {
		t.Errorf("context A os.args() = %s, want [alpha --force]", got)
	}

	res, err = a.Eval(`app.version`)
	if err != nil {
		t.Fatalf("app.version: %v", err)
	}
	if got := fmt.Sprintf("%v", res); got != "1.2.3" {
		t.Errorf("context A app.version = %s, want 1.2.3", got)
	}

	// require() version checks consult *this* context's app version.
	if _, err := a.Eval(`require(["v1.2"])`); err != nil {
		t.Errorf("require(v1.2) against app version 1.2.3 failed: %v", err)
	}

	b := NewContext(WithOS())

	res, err = b.Eval(`os.args()`)
	if err != nil {
		t.Fatalf("os.args() (context B): %v", err)
	}
	// Context B carries no args of its own: it falls back to the
	// deprecated global (kept as the documented default for one
	// release), and crucially never sees context A's args.
	if got := fmt.Sprintf("%v", res); got != "[poison]" {
		t.Errorf("context B os.args() = %s, want the global fallback [poison]", got)
	}

	// A configured context never sees the poisoned global:
	res, err = a.Eval(`os.args()[0]`)
	if err != nil {
		t.Fatalf("os.args()[0] (context A): %v", err)
	}
	if got := fmt.Sprintf("%v", res); got != "alpha" {
		t.Errorf("context A os.args()[0] = %s, want alpha (global leaked)", got)
	}

	// Context B has no app version: require checks fynerisor's own
	// version (v0.9.0), so v1.2 must fail even though context A passed.
	if _, err := b.Eval(`require(["v1.2"])`); err == nil {
		t.Error("require(v1.2) unexpectedly passed on context B")
	}
}
