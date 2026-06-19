package widget

import (
	"context"
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"github.com/uidbz/fynerisor/gui/guithread"
	"fyne.io/fyne/v2/widget"

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
//	let popup = widget.NewPopUp(content, window.Canvas)
//	popup.Show()
//
//	// Modal popup (blocks interaction with background)
//	let modal = widget.NewModalPopUp(content, window.Canvas)
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
		return object.Errorf("wrong number of arguments. got=%d, want=2", len(args)), nil
	}

	contentObj, ok := args[0].(interface{ CanvasObject() fyne.CanvasObject })
	if !ok {
		return object.Errorf("argument error: expected widget/container, got %s", args[0].Type()), nil
	}

	canvasObj, ok := args[1].(interface{ Canvas() fyne.Canvas })
	if !ok {
		return object.Errorf("argument error: expected canvas (use window.Canvas), got %s", args[1].Type()), nil
	}

	instance := widget.NewPopUp(contentObj.CanvasObject(), canvasObj.Canvas())

	return &PopUp{
		instance: instance,
	}, nil
}

func NewModalPopUp(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 2 {
		return object.Errorf("wrong number of arguments. got=%d, want=2", len(args)), nil
	}

	contentObj, ok := args[0].(interface{ CanvasObject() fyne.CanvasObject })
	if !ok {
		return object.Errorf("argument error: expected widget/container, got %s", args[0].Type()), nil
	}

	canvasObj, ok := args[1].(interface{ Canvas() fyne.Canvas })
	if !ok {
		return object.Errorf("argument error: expected canvas (use window.Canvas), got %s", args[1].Type()), nil
	}

	instance := widget.NewModalPopUp(contentObj.CanvasObject(), canvasObj.Canvas())

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

func (obj *PopUp) Inspect() string {
	return "widget.PopUp"
}

func (obj *PopUp) Interface() interface{} {
	return obj.instance
}

func (obj *PopUp) IsTruthy() bool {
	return true
}

func (obj *PopUp) Cost() int {
	return 0
}

func (obj *PopUp) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'widget.PopUp'")
}

func (obj *PopUp) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(PopUpType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", PopUpType, opType)
	return errObj, err
}

func (obj *PopUp) Equals(other object.Object) bool {
	return obj == other
}

func (obj *PopUp) Attrs() []object.AttrSpec {
	return nil
}

func (obj *PopUp) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", PopUpType, name)
}

func (obj *PopUp) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Show":
		return object.NewBuiltin("widget.PopUp.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}

			guithread.Do(func() {
				obj.instance.Show()
			})

			return object.Nil, nil
		}), true

	case "Hide":
		return object.NewBuiltin("widget.PopUp.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}

			guithread.Do(func() {
				obj.instance.Hide()
			})

			return object.Nil, nil
		}), true

	case "ShowAtPosition":
		return object.NewBuiltin("widget.PopUp.ShowAtPosition", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return object.Errorf("wrong number of arguments. got=%d, want=2", len(args)), nil
			}

			x, err := object.AsFloat(args[0])
			if err != nil {
				return nil, err
			}

			y, err := object.AsFloat(args[1])
			if err != nil {
				return nil, err
			}

			guithread.Do(func() {
				obj.instance.ShowAtPosition(fyne.NewPos(float32(x), float32(y)))
			})

			return object.Nil, nil
		}), true

	case "Move":
		return object.NewBuiltin("widget.PopUp.Move", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return object.Errorf("wrong number of arguments. got=%d, want=2", len(args)), nil
			}

			x, err := object.AsFloat(args[0])
			if err != nil {
				return nil, err
			}

			y, err := object.AsFloat(args[1])
			if err != nil {
				return nil, err
			}

			guithread.Do(func() {
				obj.instance.Move(fyne.NewPos(float32(x), float32(y)))
			})

			return object.Nil, nil
		}), true

	case "Resize":
		return object.NewBuiltin("widget.PopUp.Resize", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return object.Errorf("wrong number of arguments. got=%d, want=2", len(args)), nil
			}

			w, err := object.AsFloat(args[0])
			if err != nil {
				return nil, err
			}

			h, err := object.AsFloat(args[1])
			if err != nil {
				return nil, err
			}

			guithread.Do(func() {
				obj.instance.Resize(fyne.NewSize(float32(w), float32(h)))
			})

			return object.Nil, nil
		}), true
	}

	return nil, false
}
