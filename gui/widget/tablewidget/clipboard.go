package tablewidget

import "fyne.io/fyne/v2"

// writeClipboard copies text to the system clipboard through Fyne's
// clipboard — the same one the `clipboard` global uses (issue #35).
// The previous golang.design/x/clipboard stack added a second cgo
// clipboard (X11 headers, separate init) and a mobile build-tag split.
func writeClipboard(text string) {
	app := fyne.CurrentApp()
	if app == nil {
		return
	}
	app.Clipboard().SetContent(text)
}
