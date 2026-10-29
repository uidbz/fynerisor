package gui

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// This file covers issue #31: library diagnostics (callback errors,
// dropped callbacks, panics) must reach a per-runtime sink — the
// window's stderr writer by default, or WithErrorHandler — never a
// bare fmt.Print to the process stdout, where they would corrupt CLI
// pipelines and masquerade as script output in capture widgets.

// raisingBuiltin is a script-callable function that fails with err.
func raisingBuiltin(msg string) *object.Builtin {
	return object.NewBuiltin("boom", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		return nil, errors.New(msg)
	})
}

const failingHandlerScript = `
let btn = widget.NewButton("trigger", () => { boom() })
window.SetContent(btn)
`

func tapTrigger(t *testing.T, view *CanvasObject) {
	t.Helper()
	btn := findButton(view.Window().GetContentContainer())
	if btn == nil {
		t.Fatal("button not found")
	}
	test.Tap(btn)
}

// TestCallbackErrorGoesToWindowStderr proves the default sink: a
// failing handler surfaces on the window's Stderr(), not on Stdout()
// and not on the process stderr.
func TestCallbackErrorGoesToWindowStderr(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	host := a.NewWindow("host")

	var out, errs safeBuffer
	view := NewCanvasObject(failingHandlerScript,
		WithStdout(&out), WithStderr(&errs), WithGlobal("boom", raisingBuiltin("handler failed")))
	host.SetContent(container.NewVBox(view))
	defer view.Stop()

	view.Execute()
	waitFor(t, 2*time.Second, func() bool { return view.Status() == "Ready!" })
	tapTrigger(t, view)

	waitFor(t, 2*time.Second, func() bool { return strings.Contains(errs.String(), "handler failed") })
	if !strings.Contains(errs.String(), "handler failed") {
		t.Errorf("stderr got %q, want it to contain the handler error", errs.String())
	}
	if out.String() != "" {
		t.Errorf("diagnostic leaked to the script stdout writer: %q", out.String())
	}
}

// TestWithErrorHandlerReceivesRuntimeErrors proves the handler form:
// every diagnostic (callback errors and pre-Execute drops) arrives at
// the host function, and nothing is written to stderr writers.
func TestWithErrorHandlerReceivesRuntimeErrors(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	host := a.NewWindow("host")

	var mu sync.Mutex
	var reported []string
	handler := func(err error) {
		mu.Lock()
		defer mu.Unlock()
		reported = append(reported, err.Error())
	}

	var errs safeBuffer
	view := NewCanvasObject(failingHandlerScript,
		WithStderr(&errs), WithErrorHandler(handler), WithGlobal("boom", raisingBuiltin("sink me")))
	host.SetContent(container.NewVBox(view))
	defer view.Stop()

	// queueCallback before Execute drops the callback and reports it.
	view.window.queueCallback("Probe", func() {})

	view.Execute()
	waitFor(t, 2*time.Second, func() bool { return view.Status() == "Ready!" })
	tapTrigger(t, view)

	waitFor(t, 2*time.Second, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(reported) >= 2
	})
	mu.Lock()
	defer mu.Unlock()
	joined := strings.Join(reported, "|")
	if !strings.Contains(joined, "Probe: called before Execute") {
		t.Errorf("handler got %q, want the pre-Execute drop", joined)
	}
	if !strings.Contains(joined, "sink me") {
		t.Errorf("handler got %q, want the callback error", joined)
	}
	if errs.String() != "" {
		t.Errorf("handler is set yet stderr writer got %q", errs.String())
	}
}
