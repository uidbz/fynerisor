package gui

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/uidbz/fynerisor/gui/vmexec"
)

// TestHostDispatchedHandlerCanCreateWidgets covers issue #54: a host stores a
// script handler and later dispatches it through window.Do + vmexec.Call.
// The call installs the VM call function, so the handler can create widgets
// that take callbacks instead of failing with "unable to get call function".
func TestHostDispatchedHandlerCanCreateWidgets(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	w := a.NewWindow("Test")

	type registration struct {
		fn   *object.Closure
		call object.CallFunc
		run  *vmexec.Run
	}
	var mu sync.Mutex
	var reg registration

	register := object.NewBuiltin("register", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		fn, ok := args[0].(*object.Closure)
		if !ok {
			return nil, object.Errorf("register: expected a function")
		}
		call, ok := object.GetCallFunc(ctx)
		if !ok {
			return nil, object.Errorf("register: unable to get call function")
		}
		mu.Lock()
		reg = registration{fn: fn, call: call, run: vmexec.FromContext(ctx)}
		mu.Unlock()
		return object.Nil, nil
	})

	fw := NewWindow(w, WithGlobal("register", register))
	defer fw.Stop()
	fw.LoadScript(`
register((msg) => {
	let button = widget.NewButton("Rebuilt", () => {})
	window.SetContent(container.NewVBox(button, widget.NewLabel(msg)))
	window.SetStatus(msg)
})
window.SetStatus("registered")
`)
	fw.Execute()
	waitFor(t, 2*time.Second, func() bool { return fw.Status() == "registered" })

	mu.Lock()
	got := reg
	mu.Unlock()
	if got.fn == nil || got.call == nil || got.run == nil {
		t.Fatal("script did not register a handler")
	}

	done := make(chan struct{})
	fw.Do(func() {
		defer close(done)
		if _, err := vmexec.Call(got.call, got.run.Context(), got.fn, object.NewString("rebuilt")); err != nil {
			t.Errorf("host dispatch failed: %v", err)
		}
	})
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("host dispatch did not run")
	}

	waitFor(t, 2*time.Second, func() bool { return fw.Status() == "rebuilt" })
	if btn := findButton(w.Content()); btn == nil || btn.Text != "Rebuilt" {
		t.Fatalf("handler did not rebuild the content with a button: %#v", btn)
	}

	// The run-bound form rejects handlers of a stopped run instead of
	// entering a different VM.
	fw.Stop()
	if _, err := got.run.Call(got.call, got.fn, object.NewString("late")); !errors.Is(err, vmexec.ErrStopped) {
		t.Fatalf("stopped run dispatch error = %v, want ErrStopped", err)
	}
}
