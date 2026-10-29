package container

import (
	"fyne.io/fyne/v2"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// canvasArg unwraps a script argument into a Fyne canvas object.
func canvasArg(a object.Object) (fyne.CanvasObject, bool) {
	co, ok := a.Interface().(fyne.CanvasObject)
	return co, ok
}
