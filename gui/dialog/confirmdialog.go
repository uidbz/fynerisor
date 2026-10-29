package dialog

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/fynerisor/gui/guithread"

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

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *ConfirmDialog) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *ConfirmDialog) inspectUI() string {
	return "dialog.ConfirmDialog"
}

func (obj *ConfirmDialog) Interface() interface{} {
	return obj.instance
}

func (obj *ConfirmDialog) IsTruthy() bool {
	return true
}

func (obj *ConfirmDialog) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", ConfirmDialogType, opType)
}

func (obj *ConfirmDialog) Equals(other object.Object) bool {
	return obj == other
}

func (obj *ConfirmDialog) Attrs() []object.AttrSpec {
	return confirmDialogAttrs.Specs()
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *ConfirmDialog) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *ConfirmDialog) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", ConfirmDialogType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *ConfirmDialog) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var confirmDialogAttrs = object.NewAttrRegistry[*ConfirmDialog](string(ConfirmDialogType))

func (obj *ConfirmDialog) getAttrUI(name string) (object.Object, bool) {
	return confirmDialogAttrs.GetAttr(obj, name)

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
func init() {
	confirmDialogAttrs.
		Define("Show").Args().Returns("nil").Impl(func(obj *ConfirmDialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		guithread.
			Do(func() { obj.instance.Show() })
		return object.Nil, nil
	})
	confirmDialogAttrs.
		Define("Hide").Args().Returns("nil").Impl(func(obj *ConfirmDialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		guithread.
			Do(func() { obj.instance.Hide() })
		return object.Nil, nil
	})
	confirmDialogAttrs.
		Define("Refresh").
		Args().
		Returns("nil").Impl(func(obj *ConfirmDialog,
		ctx context.Context, args ...object.Object) (object.Object, error) {
		guithread.
			Do(func() { obj.instance.Refresh() })
		return object.Nil, nil
	})
	confirmDialogAttrs.
		Define("Resize").
		Args("arg1",

			"arg2").Returns("nil").Impl(func(obj *ConfirmDialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		width, err := object.AsFloat(args[0])
		if err != nil {
			return nil, object.Errorf("argument error: width must be a number, got %s",

				args[0].Type())
		}
		height, err := object.AsFloat(args[1])
		if err != nil {
			return nil, object.Errorf("argument error: height must be a number, got %s",

				args[1].Type())
		}
		guithread.
			Do(func() {
				obj.instance.
					Resize(fyne.NewSize(float32(width), float32(height)))
			})
		return object.
			Nil, nil
	})
	confirmDialogAttrs.
		Define("SetDismissText").
		Args("arg1").Returns("nil").Impl(func(obj *ConfirmDialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		text, err := object.AsString(args[0])
		if err != nil {
			return nil, object.Errorf("type error: %v",

				err)
		}
		guithread.Do(func() {
			obj.instance.SetDismissText(text)
		})
		return object.Nil,

			nil
	})
	confirmDialogAttrs.
		Define("SetConfirmText").
		Args("arg1").Returns("nil").Impl(func(obj *ConfirmDialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		text, err := object.AsString(args[0])
		if err != nil {
			return nil, object.Errorf("type error: %v",

				err)
		}
		guithread.Do(func() {
			obj.instance.SetConfirmText(text)
		})
		return object.Nil,

			nil
	})
	confirmDialogAttrs.
		Define("SetConfirmImportance").Args("arg1").
		Returns("nil").Impl(func(obj *ConfirmDialog, ctx context.
		Context, args ...object.Object) (object.
		Object, error) {
		importance, err := object.AsInt(args[0])
		if err != nil {
			return nil, object.
				Errorf("type error: %v", err)
		}
		guithread.
			Do(func() {
				obj.instance.SetConfirmImportance(widget.Importance(importance))
			})
		return object.Nil, nil
	})
}
