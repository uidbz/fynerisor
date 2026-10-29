package dialog

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"

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

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *FormDialog) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *FormDialog) inspectUI() string {
	return "dialog.FormDialog"
}

func (obj *FormDialog) Interface() interface{} {
	return obj.instance
}

func (obj *FormDialog) IsTruthy() bool {
	return true
}

func (obj *FormDialog) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", FormDialogType, opType)
}

func (obj *FormDialog) Equals(other object.Object) bool {
	return obj == other
}

func (obj *FormDialog) Attrs() []object.AttrSpec {
	return formDialogAttrs.Specs()
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *FormDialog) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *FormDialog) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", FormDialogType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *FormDialog) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var formDialogAttrs = objattrs.Register(string(FormDialogType), object.NewAttrRegistry[*FormDialog](string(FormDialogType)))

func (obj *FormDialog) getAttrUI(name string) (object.Object, bool) {
	return formDialogAttrs.GetAttr(obj, name)

}

func NewForm(title, confirm, dismiss string, items []*widget.FormItem, callback func(bool), parent fyne.Window) *FormDialog {
	return &FormDialog{
		instance: dialog.NewForm(title, confirm, dismiss, items, callback, parent),
		window:   parent,
	}
}
func init() {
	formDialogAttrs.
		Define("Show").Args().Returns("nil").Impl(func(
		obj *FormDialog, ctx context.
			Context, args ...object.
			Object) (object.Object, error) {
		guithread.
			Do(func() { obj.instance.Show() })
		return object.
			Nil, nil
	})
	formDialogAttrs.
		Define("Hide").Args().Returns("nil").Impl(func(
		obj *FormDialog, ctx context.
			Context, args ...object.
			Object) (object.Object, error) {
		guithread.
			Do(func() { obj.instance.Hide() })
		return object.
			Nil, nil
	})
	formDialogAttrs.
		Define("Refresh").Args().Returns("nil").Impl(func(obj *FormDialog, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {
		guithread.
			Do(func() { obj.instance.Refresh() })
		return object.Nil, nil
	})
	formDialogAttrs.
		Define("Resize").Args("arg1",

		"arg2").Returns("nil").Impl(func(obj *FormDialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		width, err := object.AsFloat(args[0])
		if err !=
			nil {
			return nil, object.Errorf("argument error: width must be a number, got %s",

				args[0].Type())
		}
		height, err := object.AsFloat(args[1])
		if err != nil {
			return nil, object.Errorf("argument error: height must be a number, got %s",

				args[1].Type())
		}
		guithread.Do(
			func() {
				obj.instance.Resize(fyne.NewSize(float32(width), float32(height)))
			})
		return object.Nil, nil
	})
	formDialogAttrs.
		Define("SetDismissText").Args("arg1").Returns("nil").Impl(func(obj *FormDialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		text, err := object.AsString(args[0])
		if err !=
			nil {
			return nil, object.Errorf("type error: %v",

				err)
		}
		guithread.Do(func() {
			obj.instance.SetDismissText(text)
		})
		return object.
			Nil, nil
	})
	formDialogAttrs.
		Define("Submit").Args().Returns("nil").Impl(func(obj *FormDialog, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {
		guithread.
			Do(func() { obj.instance.Submit() })
		return object.
			Nil, nil
	})
}
