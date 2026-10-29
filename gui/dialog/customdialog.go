package dialog

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"

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

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *CustomDialog) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *CustomDialog) inspectUI() string {
	return "dialog.CustomDialog"
}

func (obj *CustomDialog) Interface() interface{} {
	return obj.instance
}

func (obj *CustomDialog) IsTruthy() bool {
	return true
}

func (obj *CustomDialog) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", CustomDialogType, opType)
}

func (obj *CustomDialog) Equals(other object.Object) bool {
	return obj == other
}

func (obj *CustomDialog) Attrs() []object.AttrSpec {
	return customDialogAttrs.Specs()
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *CustomDialog) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *CustomDialog) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", CustomDialogType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *CustomDialog) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var customDialogAttrs = objattrs.Register(string(CustomDialogType), object.NewAttrRegistry[*CustomDialog](string(CustomDialogType)))

func (obj *CustomDialog) getAttrUI(name string) (object.Object, bool) {
	return customDialogAttrs.GetAttr(obj, name)

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
func init() {
	customDialogAttrs.
		Define(
			"Show").Args().Returns("nil").Impl(func(obj *CustomDialog, ctx context.
		Context, args ...object.Object) (object.
		Object, error) {
		guithread.Do(func() {
			obj.instance.
				Show()
		})
		return object.Nil, nil
	})
	customDialogAttrs.
		Define(
			"Hide").Args().Returns("nil").Impl(func(obj *CustomDialog, ctx context.
		Context, args ...object.Object) (object.
		Object, error) {
		guithread.Do(func() {
			obj.instance.
				Hide()
		})
		return object.Nil, nil
	})
	customDialogAttrs.
		Define(
			"Refresh").
		Args().
		Returns("nil").Impl(
		func(obj *CustomDialog, ctx context.Context, args ...object.Object) (object.
			Object, error) {
			guithread.Do(func() {
				obj.
					instance.Refresh()
			})
			return object.Nil, nil

		})
	customDialogAttrs.
		Define(
			"Resize").Args("arg1",

		"arg2").Returns("nil").Impl(func(obj *CustomDialog, ctx context.Context, args ...object.
		Object) (object.Object, error) {
		width, err :=
			object.AsFloat(args[0])
		if err != nil {
			return nil, object.Errorf("argument error: width must be a number, got %s",

				args[0].Type())
		}
		height,

			err := object.AsFloat(args[1])
		if err != nil {
			return nil, object.
				Errorf("argument error: height must be a number, got %s",

					args[1].Type())
		}

		guithread.Do(func() {
			obj.instance.Resize(fyne.
				NewSize(float32(width),
					float32(height)))
		})
		return object.Nil, nil
	})
	customDialogAttrs.
		Define(
			"SetDismissText",
		).
		Args("arg1").Returns("nil").Impl(func(obj *CustomDialog, ctx context.Context, args ...object.
		Object) (object.Object, error) {
		text, err :=
			object.AsString(args[0])
		if err != nil {
			return nil, object.Errorf("type error: %v",

				err)
		}
		guithread.Do(func() {
			obj.instance.SetDismissText(
				text)
		})
		return object.Nil, nil
	})
	customDialogAttrs.
		Define(
			"SetButtons",
		).Args(
		"arg1").Returns("nil").Impl(func(obj *CustomDialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		buttonsList, ok :=
			args[0].(*object.List)
		if !ok {
			return nil,

				object.Errorf("type error: buttons must be a list, got %s",

					args[0].Type())
		}
		var buttons []fyne.
			CanvasObject
		for i, btn := range buttonsList.Value() {
			canvasObj,
				ok := btn.(interface {
				CanvasObject() fyne.CanvasObject
			})
			if !ok {
				return nil, object.Errorf("type error: button %d is not a canvas object",
					i)
			}

			buttons = append(buttons, canvasObj.CanvasObject())
		}
		guithread.Do(func() {
			obj.instance.SetButtons(buttons)
		})
		return object.Nil, nil
	})
}
