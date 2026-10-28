package gui

import (
	"testing"
	"time"
)

// TestClipboard verifies the clipboard global's Copy/Read builtins and the
// issue #18 round-trip: Read must return the copied content even though the
// main script body runs off the GUI thread.
func TestClipboard(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()

	w := a.NewWindow("Test")
	fw := NewWindow(w)
	fw.LoadScript(`
		if (clipboard.Copy == nil) { throw("clipboard.Copy missing") }
		if (clipboard.Read == nil) { throw("clipboard.Read missing") }
		clipboard.Copy("askbot answer text")
		let got = clipboard.Read()
		if (got != "askbot answer text") { throw("clipboard round-trip failed, got: " + got) }
	`)
	fw.Execute()
	time.Sleep(200 * time.Millisecond)

	if fw.Status() != "Ready!" {
		t.Fatalf("expected Ready!, got: %s", fw.Status())
	}
}
