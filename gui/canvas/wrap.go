package risorcanvas

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// Wrap returns the fynerisor script wrapper for an existing canvas object, or
// nil. Scene nodes use it to delegate unknown attributes to the canvas API.
func Wrap(o fyne.CanvasObject) object.Object {
	switch x := o.(type) {
	case *canvas.Text:
		return &Text{instance: x}
	case *canvas.Image:
		return &Image{instance: x, container: container.NewStack(x)}
	case *canvas.Rectangle:
		return &Rectangle{instance: x}
	case *canvas.Line:
		return &Line{instance: x}
	}
	return nil
}
