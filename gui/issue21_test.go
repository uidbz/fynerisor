package gui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	fyneiw "fyne.io/fyne/v2/widget"
)

// TestNewFormMixedListAndVariadic covers issue #21: list arguments must be
// taken from the current loop argument, not args[0].
func TestNewFormMixedListAndVariadic(t *testing.T) {
	a := newTestApp(t)
	defer a.Quit()
	w := a.NewWindow("Test")
	fw := NewWindow(w)

	script := `
let i1 = widget.NewFormItem("A:", widget.NewLabel("a"))
let i2 = widget.NewFormItem("B:", widget.NewLabel("b"))
let i3 = widget.NewFormItem("C:", widget.NewLabel("c"))
let i4 = widget.NewFormItem("D:", widget.NewLabel("d"))

let form = widget.NewForm(i1, [i2, i3], i4)
window.SetContent(form)
`

	fw.LoadScript(script)
	fw.Execute()
	time.Sleep(200 * time.Millisecond)

	if fw.Status() != "Ready!" {
		t.Fatalf("Expected Ready!, got %s", fw.Status())
	}

	forms := 0
	if cont, ok := w.Content().(*fyne.Container); ok {
		for _, obj := range cont.Objects {
			if form, ok := obj.(*fyneiw.Form); ok {
				forms++
				if len(form.Items) != 4 {
					t.Errorf("Expected 4 form items, got %d", len(form.Items))
				}
			}
		}
	}
	if forms != 1 {
		t.Errorf("Expected 1 form in content, got %d", forms)
	}
}
