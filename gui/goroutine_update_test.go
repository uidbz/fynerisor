package gui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	fynewidget "fyne.io/fyne/v2/widget"
)

// TestGoroutineWidgetUpdate reproduces the real-world pattern: a button
// handler spawns go() which, after work, disables/enables widgets directly
// (no window.Do). Verifies the updates actually land.
func TestGoroutineWidgetUpdate(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Test")
	fw := NewWindow(w)

	script := `
let btn = widget.NewButton("Run", () => {})
btn.Disable()
window.SetContent(btn)

go(() => {
	btn.Enable()
})
`
	fw.LoadScript(script)
	fw.Execute()
	time.Sleep(300 * time.Millisecond)

	// Pump any queued GUI functions the way the real main loop would.
	drainFuncCalls(fw, 200*time.Millisecond)

	btn := findButton(w.Content())
	if btn == nil {
		t.Fatal("button not found in content")
	}
	if btn.Disabled() {
		t.Fatal("button still disabled - goroutine Enable() did not take effect")
	}
}

func drainFuncCalls(fw *Window, d time.Duration) {
	deadline := time.After(d)
	for {
		select {
		case f := <-fw.functionCalls:
			f()
		case <-deadline:
			return
		default:
			return
		}
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
