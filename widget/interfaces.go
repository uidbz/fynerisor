package widget

import "fyne.io/fyne/v2"

// WindowInterface defines the interface that widgets need from the Window type
// to dispatch callbacks to the main goroutine
type WindowInterface interface {
	Do(fn func())
}

// IsCanvasObject interface for objects that can be converted to fyne.CanvasObject
type IsCanvasObject interface {
	CanvasObject() fyne.CanvasObject
}
