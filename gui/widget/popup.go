package widget

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &PopUp{}

const PopUpType object.Type = "widget.PopUp"

// PopUp wraps fyne's PopUp widget - a floating overlay above the UI.
//
// PopUp displays content in a floating window above other widgets with a shadow.
// Can be modal (blocks interaction) or non-modal.
//
// Example usage in Risor:
//
//	let content = container.NewVBox(
//	    widget.NewLabel("This is a popup!"),
//	    widget.NewButton("Close", () => popup.Hide())
//	)
//	let popup = widget.NewPopUp(content, window)
//	popup.Show()
//
//	// Modal popup (blocks interaction with background)
//	let modal = widget.NewModalPopUp(content, window)
//	modal.Show()
//
// Methods:
//   - Show(): Display the popup
//   - Hide(): Hide the popup
//   - ShowAtPosition(x, y): Show at specific screen coordinates
type PopUp struct {
	instance *widget.PopUp
}

func NewPopUp(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 2 {
		return nil, object.Errorf("wrong number of arguments. got=%d, want=2", len(args))
	}

	contentObj, ok := args[0].(interface{ CanvasObject() fyne.CanvasObject })
	if !ok {
		return nil, object.Errorf("argument error: expected widget/container, got %s", args[0].Type())
	}

	canvasObj, ok := args[1].(interface{ Canvas() fyne.Canvas })
	if !ok {
		return nil, object.Errorf("argument error: expected canvas (use window), got %s", args[1].Type())
	}
	canvas := canvasObj.Canvas()
	if canvas == nil {
		return nil, object.Errorf("argument error: no canvas available (pass gui.WithHostWindow when embedding with gui.NewCanvasObject)")
	}

	instance := widget.NewPopUp(contentObj.CanvasObject(), canvas)

	return &PopUp{
		instance: instance,
	}, nil
}

func NewModalPopUp(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 2 {
		return nil, object.Errorf("wrong number of arguments. got=%d, want=2", len(args))
	}

	contentObj, ok := args[0].(interface{ CanvasObject() fyne.CanvasObject })
	if !ok {
		return nil, object.Errorf("argument error: expected widget/container, got %s", args[0].Type())
	}

	canvasObj, ok := args[1].(interface{ Canvas() fyne.Canvas })
	if !ok {
		return nil, object.Errorf("argument error: expected canvas (use window), got %s", args[1].Type())
	}
	canvas := canvasObj.Canvas()
	if canvas == nil {
		return nil, object.Errorf("argument error: no canvas available (pass gui.WithHostWindow when embedding with gui.NewCanvasObject)")
	}

	instance := widget.NewModalPopUp(contentObj.CanvasObject(), canvas)

	return &PopUp{
		instance: instance,
	}, nil
}

func (obj *PopUp) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *PopUp) Type() object.Type {
	return PopUpType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *PopUp) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *PopUp) inspectUI() string {
	return "widget.PopUp"
}

func (obj *PopUp) Interface() interface{} {
	return obj.instance
}

func (obj *PopUp) IsTruthy() bool {
	return true
}

func (obj *PopUp) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", PopUpType, opType)
}

func (obj *PopUp) Equals(other object.Object) bool {
	return obj == other
}

func (obj *PopUp) Attrs() []object.AttrSpec {
	return popUpAttrs.

		// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
		Specs()
}

func (obj *PopUp) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *PopUp) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", PopUpType, name)
}

var popUpAttrs = object.NewAttrRegistry[*PopUp](string(PopUpType))

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *PopUp) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *PopUp) getAttrUI(name string) (object.Object, bool) {
	return popUpAttrs.GetAttr(obj, name)

}
func init() {
	objattrs.Show(popUpAttrs)
	objattrs.Hide(popUpAttrs)
	objattrs.Refresh(popUpAttrs)
	popUpAttrs.Define("ShowAtPosition").
		Args("arg1",

			"arg2").Returns("nil").Impl(func(obj *PopUp, ctx context.Context, args ...object.Object) (object.Object, error) {
		x, err := object.AsFloat(args[0])
		if err != nil {
			return nil,
				err

		}
		y, err := object.AsFloat(args[1])
		if err != nil {
			return nil, err
		}
		guithread.
			Do(func() {
				obj.instance.ShowAtPosition(fyne.NewPos(float32(x),
					float32(y)))
			})
		return object.Nil, nil
	})
	popUpAttrs.Define("Move").
		Args("arg1",
			"arg2",
		).Returns("nil").Impl(func(obj *PopUp, ctx context.
		Context, args ...object.Object) (object.
		Object, error) {
		x, err := object.AsFloat(args[0])
		if err != nil {
			return nil, err
		}
		y,
			err := object.AsFloat(args[1])
		if err != nil {
			return nil, err
		}
		guithread.Do(func() { obj.instance.Move(fyne.NewPos(float32(x), float32(y))) })
		return object.Nil, nil
	})
	popUpAttrs.Define("Resize").Args("arg1",
		"arg2",
	).Returns("nil").
		Impl(func(obj *PopUp, ctx context.Context, args ...object.Object) (object.
			Object, error) {
			w, err := object.AsFloat(args[0])
			if err != nil {
				return nil, err
			}
			h,

				err := object.AsFloat(args[1])

			if err != nil {
				return nil, err
			}
			guithread.
				Do(func() {
					obj.instance.Resize(fyne.NewSize(float32(w), float32(h)))
				})
			return object.Nil, nil
		})
}
