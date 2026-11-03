package widget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// Wrap returns the fynerisor script wrapper for an existing Fyne widget, or nil
// when the widget has no wrapper. It lets scene.Node delegate unknown
// attributes to the familiar widget API (page.Node("title").SetText("x")).
func Wrap(o fyne.CanvasObject, w WindowInterface) object.Object {
	switch x := o.(type) {
	case *widget.Label:
		return &Label{instance: x}
	case *widget.Button:
		return &Button{instance: x}
	case *widget.Entry:
		return &Entry{instance: x, w: w}
	case *widget.Check:
		return &Check{instance: x}
	case *widget.Select:
		return &Select{instance: x}
	case *widget.RadioGroup:
		return &RadioGroup{instance: x}
	case *widget.Slider:
		return &Slider{instance: x, w: w}
	case *widget.ProgressBar:
		return &ProgressBar{instance: x}
	case *widget.Icon:
		return &Icon{instance: x}
	case *widget.Hyperlink:
		return &Hyperlink{instance: x, w: w}
	case *widget.Separator:
		return &Separator{instance: x}
	case *widget.RichText:
		return &RichText{instance: x}
	}
	return nil
}
