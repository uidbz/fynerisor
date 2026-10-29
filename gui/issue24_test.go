package gui

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"

	"fyne.io/fyne/v2/container"

	risorwidget "github.com/uidbz/fynerisor/gui/widget"
)

// This file covers issue #24: stdout must be a per-runtime io.Writer,
// not a process-global os.Stdout pipe hijack. The old CaptureStdout
// replaced os.Stdout with a pipe, so capture leaked into host output,
// leaked between CanvasObjects (whoever captured last won), and only
// stopped leaking by owning goroutines. print now writes to the
// window's own Stdout(); WithStdout/WithStderr configure it from Go and
// window.SetStdout from script.

// safeBuffer is an io.Writer safe for concurrent writes.
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *safeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *safeBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// TestPrintIsPerRuntime proves the capture acceptance criterion: two
// embedded views each capture their own print output; nothing leaks
// into the process stdout and nothing leaks between the views.
func TestPrintIsPerRuntime(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	host := a.NewWindow("host")

	var bufA, bufB safeBuffer
	viewA := NewCanvasObject(`print("from-a")`, WithStdout(&bufA))
	viewB := NewCanvasObject(`print("from-b")`, WithStdout(&bufB))
	host.SetContent(container.NewVBox(viewA, viewB))
	defer viewA.Stop()
	defer viewB.Stop()

	viewA.Execute()
	waitFor(t, 2*time.Second, func() bool { return viewA.Status() == "Ready!" })
	viewB.Execute()
	waitFor(t, 2*time.Second, func() bool { return viewB.Status() == "Ready!" })

	if got := bufA.String(); got != "from-a\n" {
		t.Errorf("view A captured %q, want %q", got, "from-a\n")
	}
	if got := bufB.String(); got != "from-b\n" {
		t.Errorf("view B captured %q, want %q", got, "from-b\n")
	}
}

// TestSetStdoutLogWidget covers the preferred script form: print lines
// land in the Log widget via its io.Writer implementation, without the
// VM being re-entered.
func TestSetStdoutLogWidget(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	host := a.NewWindow("host")

	var log *risorwidget.Log
	capture := object.NewBuiltin("capture", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		if l, ok := args[0].(*risorwidget.Log); ok {
			log = l
		}
		return object.Nil, nil
	})

	var def safeBuffer
	view := NewCanvasObject(`
let log = widget.NewLog(50)
capture(log)
window.SetStdout(log)
print("line one")
print("line", "two")
`, WithStdout(&def), WithGlobal("capture", capture))
	host.SetContent(container.NewVBox(view))
	defer view.Stop()

	view.Execute()
	waitFor(t, 2*time.Second, func() bool { return view.Status() == "Ready!" })
	if view.Status() != "Ready!" {
		t.Fatalf("script failed: %s", view.Status())
	}

	if log == nil {
		t.Fatal("Log widget was not captured")
	}
	lines := strings.Join(log.Lines(), "\n")
	if !strings.Contains(lines, "line one") || !strings.Contains(lines, "line two") {
		t.Errorf("Log widget got %q, want the printed lines", lines)
	}
	if def.String() != "" {
		t.Errorf("default stdout got %q while redirected to the Log", def.String())
	}
}

// TestSetStdoutNilRestoresDefault checks the nil form and that a
// redirect installed by a superseded run is dropped for the next run.
func TestSetStdoutNilRestoresDefault(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	host := a.NewWindow("host")

	var def safeBuffer
	view := NewCanvasObject(`
let log = widget.NewLog(50)
window.SetStdout(log)
print("to-log")
`, WithStdout(&def))
	host.SetContent(container.NewVBox(view))
	defer view.Stop()

	view.Execute()
	waitFor(t, 2*time.Second, func() bool { return view.Status() == "Ready!" })
	if view.Status() != "Ready!" {
		t.Fatalf("script failed: %s", view.Status())
	}
	if def.String() != "" {
		t.Errorf("default got %q while redirected to the Log", def.String())
	}

	// A redirect left installed by run 1 must not survive into run 2:
	// print lands on the default again. (Poll the buffer, not Status:
	// run 1 already left it at "Ready!".)
	view.LoadScript(`print("run2")`)
	view.Execute()
	waitFor(t, 2*time.Second, func() bool { return def.String() != "" })
	if def.String() != "run2\n" {
		t.Errorf("after reload default got %q, want %q", def.String(), "run2\n")
	}
}

// TestSetStdoutClosureForm covers the legacy closure form: each complete
// line is queued onto the run (never dropped, issue #51) and the closure
// re-enters the VM only through the normal event queue.
func TestSetStdoutClosureForm(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	host := a.NewWindow("host")

	var mu sync.Mutex
	var seen []string
	seenB := object.NewBuiltin("got_line", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		mu.Lock()
		defer mu.Unlock()
		s, _ := object.AsString(args[0])
		seen = append(seen, s)
		return object.Nil, nil
	})

	view := NewCanvasObject(`
window.SetStdout((line) => { got_line(line) })
print("relay-1")
print("relay-2")
`, WithGlobal("got_line", seenB))
	host.SetContent(container.NewVBox(view))
	defer view.Stop()

	view.Execute()
	waitFor(t, 2*time.Second, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(seen) >= 2
	})
	mu.Lock()
	defer mu.Unlock()
	if len(seen) < 2 || seen[0] != "relay-1" || seen[1] != "relay-2" {
		t.Errorf("closure got %v, want [relay-1 relay-2]", seen)
	}
}

// TestLogWidgetWrite covers the Log io.Writer directly: partial lines
// stay buffered until the newline completes them. (Concurrent safety of
// the line store is exercised under -race by TestSetStdoutLogWidget:
// the VM goroutine appends while this test's goroutine reads Lines.)
// Parallel writers are deliberately not hammered here: the test driver
// runs fyne.Do inline, so raw concurrent Refresh calls race inside
// Fyne itself, not in this code.
func TestLogWidgetWrite(t *testing.T) {
	log := risorwidget.NewLog(100)
	log.Write([]byte("partial"))       // buffered, not a line yet
	log.Write([]byte("-done\nnext\n")) // completes + adds one
	if _, err := log.Write([]byte("tail")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	all := log.Lines()
	want := []string{"partial-done", "next"}
	if strings.Join(all, "|") != strings.Join(want, "|") {
		t.Fatalf("Log holds %q, want %q", strings.Join(all, "|"), strings.Join(want, "|"))
	}
}
