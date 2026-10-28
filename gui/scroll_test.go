package gui

import (
	"testing"
	"time"
)

// TestScrollToBottom verifies the ScrollToBottom/ScrollToTop bindings exposed
// to Risor scripts on container.Scroll (wraps Fyne's Scroll.ScrollToBottom /
// ScrollToTop via guithread.Do).
func TestScrollToBottom(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()

	w := a.NewWindow("Test")
	fw := NewWindow(w)
	fw.LoadScript(`
		let big = widget.NewLabel("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10")
		let s = container.NewScroll(big)
		window.SetContent(s)
		s.ScrollToBottom()
		s.ScrollToTop()
	`)
	fw.Execute()
	defer fw.Stop()
	// Wait for the body to finish: Fyne's test driver runs UI work inline on
	// the VM goroutine, so the second window must not start while the first
	// body is still laying out.
	waitFor(t, 2*time.Second, func() bool { return fw.Status() == "Ready!" })
	if fw.Status() != "Ready!" {
		t.Fatalf("first script failed: %q", fw.Status())
	}

	// A missing attribute would resolve to nil and the call would fail the
	// script; assert both methods exist and are callable.
	w2 := a.NewWindow("Test2")
	fw2 := NewWindow(w2)
	fw2.LoadScript(`
		let s = container.NewScroll(widget.NewLabel("x"))
		if (s.ScrollToBottom == nil) { throw("ScrollToBottom missing") }
		if (s.ScrollToTop == nil) { throw("ScrollToTop missing") }
		s.ScrollToBottom()
		s.ScrollToTop()
	`)
	fw2.Execute()
	defer fw2.Stop()
	waitFor(t, 2*time.Second, func() bool { return fw2.Status() == "Ready!" })
	if fw2.Status() != "Ready!" {
		t.Fatalf("second script failed: %q", fw2.Status())
	}
}
