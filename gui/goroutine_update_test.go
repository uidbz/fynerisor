package gui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	fynewidget "fyne.io/fyne/v2/widget"
)

// TestGoroutineWidgetUpdate reproduces the real-world pattern: a button
// handler spawns go() which, after work, disables/enables widgets directly
// (no window.Do). Verifies the updates actually land.
func TestGoroutineWidgetUpdate(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	w := a.NewWindow("Test")
	fw := NewWindow(w)

	script := `
let btn = widget.NewButton("Run", () => {})
btn.Disable()
window.SetContent(btn)

go(() => {
	btn.Enable()
	window.SetStatus("enabled")
})
`
	fw.LoadScript(script)
	fw.Execute()

	// The go() closure runs on the callback loop; the status write gives
	// this test its happens-before edge before reading widget state.
	waitFor(t, 2*time.Second, func() bool { return fw.Status() == "enabled" })

	btn := findButton(w.Content())
	if btn == nil {
		t.Fatal("button not found in content")
	}
	if btn.Disabled() {
		t.Fatal("button still disabled - goroutine Enable() did not take effect")
	}
}

func findButton(obj fyne.CanvasObject) *fynewidget.Button {
	switch o := obj.(type) {
	case *fynewidget.Button:
		return o
	case *fyne.Container:
		for _, c := range o.Objects {
			if b := findButton(c); b != nil {
				return b
			}
		}
	}
	return nil
}
