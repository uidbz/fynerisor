package gui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

// TestClipboard verifies the clipboard global's Copy/Read builtins run without
// error under the test app.
func TestClipboard(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	w := a.NewWindow("Test")
	fw := NewWindow(w)
	fw.LoadScript(`
		if (clipboard.Copy == nil) { throw("clipboard.Copy missing") }
		if (clipboard.Read == nil) { throw("clipboard.Read missing") }
		clipboard.Copy("askbot answer text")
		clipboard.Read()
	`)
	fw.Execute()
	time.Sleep(50 * time.Millisecond)
}
