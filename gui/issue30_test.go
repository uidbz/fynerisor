package gui

import (
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2/container"
)

// This file covers issue #30: argument/type checks in script-visible
// builtins must *raise* (catchable via try/catch, halting on failure)
// instead of handing the script an error *value* that surfaces later
// far from its cause.

// TestArityErrorRaisesAndIsCatchable: the widget.NewButton("Save")
// mistake from the issue — missing callback — is now a raised error a
// script can catch.
func TestArityErrorRaisesAndIsCatchable(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	host := a.NewWindow("host")

	var out safeBuffer
	view := NewCanvasObject(`
let outcome = "not-raised"
try {
    widget.NewButton("no-callback")
} catch e {
    outcome = "caught: " + e.message()
}
print(outcome)
`, WithStdout(&out))
	host.SetContent(container.NewVBox(view))
	defer view.Stop()

	view.Execute()
	waitFor(t, 2*time.Second, func() bool { return view.Status() == "Ready!" })
	if view.Status() != "Ready!" {
		t.Fatalf("script failed: %s", view.Status())
	}
	got := out.String()
	if !strings.Contains(got, "caught: wrong number of arguments") {
		t.Errorf("output = %q, want the caught 'wrong number of arguments' error", got)
	}
}

// TestUncaughtArityErrorFailsNoisily: the same mistake without a try
// block stops the run with an error (it used to keep going with an
// error object in a variable, failing later at SetContent).
func TestUncaughtArityErrorFailsNoisily(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	host := a.NewWindow("host")

	view := NewCanvasObject(`
let b = widget.NewButton("no-callback")
window.SetContent(42)
print("should never get here")
`)
	host.SetContent(container.NewVBox(view))
	defer view.Stop()

	view.Execute()
	waitFor(t, 2*time.Second, func() bool { return strings.HasPrefix(view.Status(), "ERROR") })
	if !strings.Contains(view.Status(), "wrong number of arguments") {
		t.Errorf("status = %q, want the arity error (fail noisily at the cause)", view.Status())
	}
}
