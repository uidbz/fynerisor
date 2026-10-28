package gui

import (
	"testing"
	"time"
)

// TestCanvasText verifies canvas.NewText (colored text) and the Label
// Selectable attribute.
func TestCanvasText(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()

	w := a.NewWindow("Test")
	fw := NewWindow(w)
	fw.LoadScript(`
		let t = canvas.NewText("Askbot", "#22c55e")
		if (t == nil) { throw("NewText returned nil") }
		if (t.Color != "#22c55eff") { throw("Color roundtrip: " + t.Color) }
		t.Color = "#3b82f6"
		t.Size = 18.0
		if (t.Size != 18.0) { throw("Size roundtrip") }
		let lbl = widget.NewLabel("select me")
		lbl.Selectable = true
		if (lbl.Selectable != true) { throw("Selectable roundtrip") }
		window.SetContent(container.NewVBox(t, lbl))
	`)
	fw.Execute()
	time.Sleep(50 * time.Millisecond)
}
