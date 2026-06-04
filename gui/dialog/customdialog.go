package dialog

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &CustomDialog{}

const CustomDialogType object.Type = "dialog.CustomDialog"

// CustomDialog wraps dialog.CustomDialog for Risor scripting
type CustomDialog struct {
	instance *dialog.CustomDialog
	window   fyne.Window
}

func (obj *CustomDialog) Type() object.Type {
	return CustomDialogType
}

func (obj *CustomDialog) Inspect() string {
	return "dialog.CustomDialog"
}

func (obj *CustomDialog) Interface() interface{} {
	return obj.instance
}

func (obj *CustomDialog) IsTruthy() bool {
	return true
}

func (obj *CustomDialog) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'dialog.CustomDialog'")
}

func (obj *CustomDialog) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for %s: %v", CustomDialogType, opType), nil
}

func (obj *CustomDialog) Equals(other object.Object) bool {
	return obj == other
}

func (obj *CustomDialog) Attrs() []object.AttrSpec {
	return nil
}

func (obj *CustomDialog) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", CustomDialogType, name)
}

func (obj *CustomDialog) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Show":
		return object.NewBuiltin("dialog.CustomDialog.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Show()
			})
			return object.Nil, nil
		}), true
	case "Hide":
		return object.NewBuiltin("dialog.CustomDialog.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Refresh":
		return object.NewBuiltin("dialog.CustomDialog.Refresh", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Refresh()
			})
			return object.Nil, nil
		}), true
	case "SetDismissText":
		return object.NewBuiltin("dialog.CustomDialog.SetDismissText", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			text, err := object.AsString(args[0])
			if err != nil {
				return object.Errorf("type error: %v", err), nil
			}
			fyne.Do(func() {
				obj.instance.SetDismissText(text)
			})
			return object.Nil, nil
		}), true
	case "SetButtons":
		return object.NewBuiltin("dialog.CustomDialog.SetButtons", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			buttonsList, ok := args[0].(*object.List)
			if !ok {
				return object.Errorf("type error: buttons must be a list, got %s", args[0].Type()), nil
			}

			var buttons []fyne.CanvasObject
			for i, btn := range buttonsList.Value() {
				canvasObj, ok := btn.(interface{ CanvasObject() fyne.CanvasObject })
				if !ok {
					return object.Errorf("type error: button %d is not a canvas object", i), nil
				}
				buttons = append(buttons, canvasObj.CanvasObject())
			}

			fyne.Do(func() {
				obj.instance.SetButtons(buttons)
			})
			return object.Nil, nil
		}), true
	}
	return nil, false
}

func NewCustom(title, dismiss string, content fyne.CanvasObject, parent fyne.Window) *CustomDialog {
	return &CustomDialog{
		instance: dialog.NewCustom(title, dismiss, content, parent),
		window:   parent,
	}
}

func NewCustomWithoutButtons(title string, content fyne.CanvasObject, parent fyne.Window) *CustomDialog {
	return &CustomDialog{
		instance: dialog.NewCustomWithoutButtons(title, content, parent),
		window:   parent,
	}
}
