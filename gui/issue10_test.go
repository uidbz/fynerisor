package gui

import (
	"context"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// TestTapAfterScriptErrorNoFreeze covers issue #10: a script that raises
// after creating widgets used to leave the callback channel undrained, so
// the 11th interaction anywhere froze the UI thread forever.
func TestTapAfterScriptErrorNoFreeze(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Test")
	fw := NewWindow(w)

	fw.LoadScript(`
let btn = widget.NewButton("tap", () => { print("tapped") })
window.SetContent(btn)
throw("boom after building UI")
`)
	fw.Execute()
	time.Sleep(200 * time.Millisecond)

	if fw.Status() != "ERROR: boom after building UI" {
		t.Fatalf("expected an ERROR status, got: %q", fw.Status())
	}

	btn := findButton(w.Content())
	if btn == nil {
		t.Fatal("button not found in content")
	}

	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			test.Tap(btn)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("tapping after a script error froze the UI thread")
	}
}

// TestTapAfterStopNoFreeze covers issue #10: widgets still on screen after
// Stop() used to fill the un-drained channel and freeze after 10
// interactions.
func TestTapAfterStopNoFreeze(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Test")
	fw := NewWindow(w)

	fw.LoadScript(`
let btn = widget.NewButton("tap", () => { print("tapped") })
window.SetContent(btn)
`)
	fw.Execute()
	time.Sleep(200 * time.Millisecond)

	if fw.Status() != "Ready!" {
		t.Fatalf("expected Ready!, got: %q", fw.Status())
	}

	fw.Stop()

	btn := findButton(w.Content())
	if btn == nil {
		t.Fatal("button not found in content")
	}

	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			test.Tap(btn)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("tapping after Stop froze the UI thread")
	}
}

// TestTapsDuringSlowBodyNoFreeze covers issue #10: sends must not block the
// UI thread while the main body is still running; the queued callbacks run
// once the body finishes.
func TestTapsDuringSlowBodyNoFreeze(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Test")

	release := make(chan struct{})
	waitgo := object.NewBuiltin("waitgo", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		select {
		case <-release:
		case <-ctx.Done():
		}
		return object.Nil, nil
	})

	fw := NewWindow(w, WithGlobal("waitgo", waitgo))
	fw.LoadScript(`
let btn = widget.NewButton("tap", () => { window.SetStatus("tapped") })
window.SetContent(btn)
waitgo()
`)
	fw.Execute()
	time.Sleep(100 * time.Millisecond)

	btn := findButton(w.Content())
	if btn == nil {
		t.Fatal("button not found in content")
	}

	// Tap more times than the old channel buffer (10) while the body is
	// still inside waitgo(): the old code blocked the 11th send.
	done := make(chan struct{})
	go func() {
		for i := 0; i < 15; i++ {
			test.Tap(btn)
			time.Sleep(5 * time.Millisecond)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("taps blocked while the main body was running")
	}

	close(release)

	// The queued callbacks are processed after the body finishes.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if fw.Status() == "tapped" {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("queued callbacks were not processed after the body finished, status=%q", fw.Status())
}
