package gui

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// recorder is a builtin that records its string arguments.
type recorder struct {
	mu   sync.Mutex
	seen []string
}

func (r *recorder) builtin() *object.Builtin {
	return object.NewBuiltin("rec", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		r.mu.Lock()
		defer r.mu.Unlock()
		for _, a := range args {
			r.seen = append(r.seen, object.PrintableValue(a).(string))
		}
		return object.Nil, nil
	})
}

func (r *recorder) has(s string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, v := range r.seen {
		if v == s {
			return true
		}
	}
	return false
}

// TestExecuteFromInsideRunningScript covers issue #16 (and a regression of
// the interim #15 fix): a script that triggers Execute itself — as
// browser.Open does — used to wait 2 s for its own goroutine, and the stale
// run then overwrote the new run's status.
func TestExecuteFromInsideRunningScript(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	w := a.NewWindow("Test")

	var mu sync.Mutex
	var statuses []string
	var fw *Window
	nav := object.NewBuiltin("nav", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		fw.LoadScript(`window.SetStatus("second")`)
		fw.Execute()
		return object.Nil, nil
	})
	fw = NewWindow(w, WithGlobal("nav", nav), WithStatusCallback(func(s string) {
		mu.Lock()
		statuses = append(statuses, s)
		mu.Unlock()
	}))
	defer fw.Stop()

	fw.LoadScript(`
nav()
window.SetStatus("stale")
`)
	start := time.Now()
	fw.Execute()
	waitFor(t, 3*time.Second, func() bool {
		mu.Lock()
		defer mu.Unlock()
		for _, s := range statuses {
			if s == "second" {
				return true
			}
		}
		return false
	})
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("re-Execute from inside the script took %v (self-wait)", elapsed)
	}
	time.Sleep(100 * time.Millisecond) // let the stale body unwind
	mu.Lock()
	defer mu.Unlock()
	for _, s := range statuses {
		if s == "stale" {
			t.Fatalf("stopped run wrote its status after being replaced: %v", statuses)
		}
	}
	if last := statuses[len(statuses)-1]; last != "Ready!" {
		t.Fatalf("final status %q, want the new run's \"Ready!\" (history %v)", last, statuses)
	}
}

// TestCallbacksOfReplacedRunAreDropped covers issue #16: widgets of a
// replaced run that are still reachable must not execute their handlers.
func TestCallbacksOfReplacedRunAreDropped(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	w := a.NewWindow("Test")
	rec := &recorder{}
	fw := NewWindow(w, WithGlobal("rec", rec.builtin()))
	defer fw.Stop()

	fw.LoadScript(`
let btn = widget.NewButton("old", () => { rec("old handler ran") })
window.SetContent(btn)
`)
	fw.Execute()
	waitFor(t, 2*time.Second, func() bool { return fw.Status() == "Ready!" })
	oldBtn := findButton(w.Content())
	if oldBtn == nil {
		t.Fatal("button not found")
	}

	// rec() after SetContent gives the test a happens-before edge with the
	// content swap (Fyne's test driver applies it inline on the VM goroutine).
	fw.LoadScript(`
window.SetContent(widget.NewLabel("new page"))
rec("new page shown")
`)
	fw.Execute()
	waitFor(t, 2*time.Second, func() bool { return rec.has("new page shown") })
	if findButton(w.Content()) != nil {
		t.Fatal("new run did not replace the content")
	}

	test.Tap(oldBtn)
	time.Sleep(100 * time.Millisecond)
	if rec.has("old handler ran") {
		t.Fatal("a replaced run's button handler executed")
	}
}

// TestReexecuteReimportsModules covers issue #16: imported modules are part
// of a run. A re-Execute (--watch reload, browser refresh) must import the
// module again instead of reusing the previous run's instance.
func TestReexecuteReimportsModules(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	w := a.NewWindow("Test")
	rec := &recorder{}
	fw := NewWindow(w, WithGlobal("rec", rec.builtin()))
	defer fw.Stop()

	dir := t.TempDir()
	mod := filepath.Join(dir, "m.risor")
	main := filepath.Join(dir, "main.risor")
	script := `let m = import("m.risor")
rec(m.v)
`
	if err := os.WriteFile(mod, []byte(`let v = "one"`), 0o644); err != nil {
		t.Fatal(err)
	}
	fw.LoadScriptFrom(script, main)
	fw.Execute()
	waitFor(t, 2*time.Second, func() bool { return rec.has("one") })
	if !rec.has("one") {
		t.Fatalf("first run did not import the module (status %q)", fw.Status())
	}

	if err := os.WriteFile(mod, []byte(`let v = "two"`), 0o644); err != nil {
		t.Fatal(err)
	}
	fw.LoadScriptFrom(script, main)
	fw.Execute()
	waitFor(t, 2*time.Second, func() bool { return rec.has("two") })
	if !rec.has("two") {
		t.Fatalf("re-Execute reused the previous run's module instance (status %q)", fw.Status())
	}
}
