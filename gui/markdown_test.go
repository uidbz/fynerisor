package gui

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// TestMarkdownWrapping verifies the Wrapping/Truncation attributes exposed on
// the Markdown widget (backed by RichText). With Wrapping=Word the MinSize
// width collapses (longest word) instead of the full unwrapped paragraph, so a
// containing Scroll sizes to the viewport and the text autowraps instead of
// producing a horizontal scrollbar.
func TestMarkdownWrapping(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	// Go-level check of the underlying fyne behavior the binding relies on.
	long := "This is a fairly long markdown paragraph that should autowrap when Wrapping is set to word instead of forcing a horizontal scrollbar."
	rt := widget.NewRichTextFromMarkdown(long)
	offW := rt.MinSize().Width
	rt.Wrapping = fyne.TextWrapWord
	rt.Refresh()
	wordW := rt.MinSize().Width
	if !(wordW < offW) {
		t.Fatalf("Wrapping=Word did not shrink MinSize width: off=%v word=%v", offW, wordW)
	}

	w := a.NewWindow("Test")
	w.Resize(fyne.NewSize(300, 400))
	fw := NewWindow(w)
	fw.LoadScript(`
		let md = widget.NewMarkdown("This is a fairly long markdown paragraph that should autowrap when Wrapping is set to word instead of forcing a horizontal scrollbar.")
		if (md.Wrapping != 0) { throw("default Wrapping should be off") }
		md.Wrapping = constants.TextWrapWord
		if (md.Wrapping != constants.TextWrapWord) { throw("Wrapping did not round-trip") }
		md.Truncation = constants.TextTruncateEllipsis
		window.SetContent(container.NewScroll(md))
	`)
	fw.Execute()
	time.Sleep(50 * time.Millisecond)
}
