//go:build !android && !ios

package tablewidget

import "golang.design/x/clipboard"

// writeClipboard copies text to the system clipboard on desktop platforms.
func writeClipboard(text string) {
	clipboard.Write(clipboard.FmtText, []byte(text))
}
