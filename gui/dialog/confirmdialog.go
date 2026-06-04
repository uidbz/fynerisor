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

var _ object.Object = &ConfirmDialog{}

const ConfirmDialogType object.Type = "dialog.ConfirmDialog"

// ConfirmDialog wraps dialog.ConfirmDialog for Risor scripting
type ConfirmDialog struct {
	instance *dialog.ConfirmDialog
	window   fyne.Window
}

func (obj *ConfirmDialog) Type() object.Type {
	return ConfirmDialogType
}

func (obj *ConfirmDialog) Inspect() string {
	return "dialog.ConfirmDialog"
}

func (obj *ConfirmDialog) Interface() interface{} {
	return obj.instance
}

func (obj *ConfirmDialog) IsTruthy() bool {
	return true
}

func (obj *ConfirmDialog) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'dialog.ConfirmDialog'")
}

func (obj *ConfirmDialog) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for %s: %v", ConfirmDialogType, opType), nil
}

func (obj *ConfirmDialog) Equals(other object.Object) bool {
	return obj == other
}

func (obj *ConfirmDialog) Attrs() []object.AttrSpec {
	return nil
}

func (obj *ConfirmDialog) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", ConfirmDialogType, name)
}

func (obj *ConfirmDialog) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Show":
		return object.NewBuiltin("dialog.ConfirmDialog.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Show()
			})
			return object.Nil, nil
		}), true
	case "Hide":
		return object.NewBuiltin("dialog.ConfirmDialog.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Refresh":
		return object.NewBuiltin("dialog.ConfirmDialog.Refresh", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Refresh()
			})
			return object.Nil, nil
		}), true
	case "SetDismissText":
		return object.NewBuiltin("dialog.ConfirmDialog.SetDismissText", func(ctx context.Context, args ...object.Object) (object.Object, error) {
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
	case "SetConfirmText":
		return object.NewBuiltin("dialog.ConfirmDialog.SetConfirmText", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			text, err := object.AsString(args[0])
			if err != nil {
				return object.Errorf("type error: %v", err), nil
			}
			fyne.Do(func() {
				obj.instance.SetConfirmText(text)
			})
			return object.Nil, nil
		}), true
	case "SetConfirmImportance":
		return object.NewBuiltin("dialog.ConfirmDialog.SetConfirmImportance", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			importance, err := object.AsInt(args[0])
			if err != nil {
				return object.Errorf("type error: %v", err), nil
			}
			fyne.Do(func() {
				obj.instance.SetConfirmImportance(widget.Importance(importance))
			})
			return object.Nil, nil
		}), true
	}
	return nil, false
}

func NewConfirm(title, message string, callback func(bool), parent fyne.Window) *ConfirmDialog {
	return &ConfirmDialog{
		instance: dialog.NewConfirm(title, message, callback, parent),
		window:   parent,
	}
}

func NewCustomConfirm(title, confirm, dismiss string, content fyne.CanvasObject, callback func(bool), parent fyne.Window) *ConfirmDialog {
	return &ConfirmDialog{
		instance: dialog.NewCustomConfirm(title, confirm, dismiss, content, callback, parent),
		window:   parent,
	}
}
