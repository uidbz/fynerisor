package gui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	fynewidget "fyne.io/fyne/v2/widget"
)

func findEntries(obj fyne.CanvasObject) []*fynewidget.Entry {
	var entries []*fynewidget.Entry
	switch o := obj.(type) {
	case *fynewidget.Entry:
		entries = append(entries, o)
	case *fyne.Container:
		for _, c := range o.Objects {
			entries = append(entries, findEntries(c)...)
		}
	}
	return entries
}

// TestMultiLineEntrySetMinRowsVisible covers issue #53: a script can size a
// multi-line entry explicitly instead of accepting Fyne's default height.
func TestMultiLineEntrySetMinRowsVisible(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	w := a.NewWindow("Test")
	fw := NewWindow(w)
	defer fw.Stop()

	fw.LoadScript(`
let small = widget.NewMultiLineEntry()
small.SetMinRowsVisible(1)
small.SetText("one row")

let big = widget.NewMultiLineEntry()
big.SetMinRowsVisible(8)
big.SetText("eight rows")

window.SetContent(container.NewVBox(small, big))
window.SetStatus("ready")
`)
	fw.Execute()
	waitFor(t, 2*time.Second, func() bool { return fw.Status() == "ready" })

	entries := findEntries(w.Content())
	if len(entries) != 2 {
		t.Fatalf("found %d entries in content, want 2", len(entries))
	}
	small, big := entries[0], entries[1]
	if !big.MultiLine {
		t.Fatal("second entry is not multi-line")
	}
	if big.MinSize().Height <= small.MinSize().Height {
		t.Fatalf("SetMinRowsVisible did not grow the entry: one row = %v, eight rows = %v",
			small.MinSize(), big.MinSize())
	}
}
