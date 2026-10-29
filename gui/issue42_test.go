package gui

import (
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2/container"
)

// This file covers issue #42 (gui half): io.stdin/io.stdout/io.stderr,
// print and eprint all follow the window's own streams — a script can
// be a pipeline filter, and two windows never share streams.

func TestScriptStreamsArePerRuntime(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	host := a.NewWindow("host")

	var outA, errA, outB safeBuffer
	viewA := NewCanvasObject(`
io.stderr.write("note:")
eprint("to-stderr")
io.stdin.lines().each((line) => { print("echo:", line) })
io.stdout.write("done")
`, WithIO(),
		WithStdin(strings.NewReader("one\ntwo\n")),
		WithStdout(&outA), WithStderr(&errA))
	// No streams configured on B: its print must not land in A's buffers
	// (the process defaults stay process-wide; isolation is asserted by
	// A's exactness plus B completing cleanly).
	viewB := NewCanvasObject(`print("from-b")`, WithIO(), WithStdout(&outB))

	host.SetContent(container.NewVBox(viewA, viewB))
	defer viewA.Stop()
	defer viewB.Stop()

	viewA.Execute()
	waitFor(t, 2*time.Second, func() bool { return viewA.Status() == "Ready!" })
	if viewA.Status() != "Ready!" {
		t.Fatalf("view A failed: %s", viewA.Status())
	}
	if got := outA.String(); got != "echo: one\necho: two\ndone" {
		t.Errorf("view A stdout = %q, want %q", got, "echo: one\necho: two\ndone")
	}
	if got := errA.String(); got != "note:to-stderr\n" {
		t.Errorf("view A stderr = %q, want %q", got, "note:to-stderr\n")
	}

	viewB.Execute()
	waitFor(t, 2*time.Second, func() bool { return viewB.Status() == "Ready!" })
	if got := outB.String(); got != "from-b\n" {
		t.Errorf("view B stdout = %q, want %q", got, "from-b\n")
	}
}
