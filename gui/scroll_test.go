package gui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

// TestScrollToBottom verifies the ScrollToBottom/ScrollToTop bindings exposed
// to Risor scripts on container.Scroll (wraps Fyne's Scroll.ScrollToBottom /
// ScrollToTop via guithread.Do).
func TestScrollToBottom(t *testing.T) {
	a := test.NewApp()
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
	time.Sleep(50 * time.Millisecond)

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
	time.Sleep(50 * time.Millisecond)
}
