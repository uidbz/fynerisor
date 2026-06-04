package dialog

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &FormDialog{}

const FormDialogType object.Type = "dialog.FormDialog"

// FormDialog wraps dialog.FormDialog for Risor scripting
type FormDialog struct {
	instance *dialog.FormDialog
	window   fyne.Window
}

func (obj *FormDialog) Type() object.Type {
	return FormDialogType
}

func (obj *FormDialog) Inspect() string {
	return "dialog.FormDialog"
}

func (obj *FormDialog) Interface() interface{} {
	return obj.instance
}

func (obj *FormDialog) IsTruthy() bool {
	return true
}

func (obj *FormDialog) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'dialog.FormDialog'")
}

func (obj *FormDialog) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for %s: %v", FormDialogType, opType), nil
}

func (obj *FormDialog) Equals(other object.Object) bool {
	return obj == other
}

func (obj *FormDialog) Attrs() []object.AttrSpec {
	return nil
}

func (obj *FormDialog) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", FormDialogType, name)
}

func (obj *FormDialog) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Show":
		return object.NewBuiltin("dialog.FormDialog.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Show()
			})
			return object.Nil, nil
		}), true
	case "Hide":
		return object.NewBuiltin("dialog.FormDialog.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Refresh":
		return object.NewBuiltin("dialog.FormDialog.Refresh", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Refresh()
			})
			return object.Nil, nil
		}), true
	case "SetDismissText":
		return object.NewBuiltin("dialog.FormDialog.SetDismissText", func(ctx context.Context, args ...object.Object) (object.Object, error) {
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
	case "Submit":
		return object.NewBuiltin("dialog.FormDialog.Submit", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Submit()
			})
			return object.Nil, nil
		}), true
	}
	return nil, false
}

func NewForm(title, confirm, dismiss string, items []*widget.FormItem, callback func(bool), parent fyne.Window) *FormDialog {
	return &FormDialog{
		instance: dialog.NewForm(title, confirm, dismiss, items, callback, parent),
		window:   parent,
	}
}
