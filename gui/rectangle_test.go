package gui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

// TestRectangle verifies the canvas.NewRectangle binding: constructor with
// fill color + corner radius, and the FillColor/StrokeColor/StrokeWidth/
// CornerRadius attributes (hex and named colors).
func TestRectangle(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	w := a.NewWindow("Test")
	fw := NewWindow(w)
	fw.LoadScript(`
		let bg = canvas.NewRectangle("#3b82f633", 8.0)
		if (bg == nil) { throw("NewRectangle returned nil") }
		if (bg.CornerRadius != 8.0) { throw("corner radius not set") }
		bg.FillColor = "#22c55e"
		bg.StrokeColor = "red"
		bg.StrokeWidth = 1.0
		bg.CornerRadius = 12.0
		if (bg.FillColor != "#22c55eff") { throw("FillColor roundtrip: " + bg.FillColor) }
		let failed = false
		try {
			canvas.NewRectangle("not-a-color", 4.0)
		} catch (err) {
			failed = true
		}
		if (!failed) { throw("invalid color did not error") }
		let stack = container.NewStack(bg, container.NewPadded(widget.NewLabel("hi")))
		window.SetContent(stack)
	`)
	fw.Execute()
	time.Sleep(50 * time.Millisecond)
}
