package widget

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"
	"github.com/uidbz/fynerisor/gui/vmexec"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &Button{}

const ButtonType object.Type = "widget.Button"

type Button struct {
	instance *widget.Button
}

func (obj *Button) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Button) Type() object.Type {
	return ButtonType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Button) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Button) inspectUI() string {
	return "widget.Button"
}

func (obj *Button) Interface() interface{} {
	return obj.instance
}

func (obj *Button) IsTruthy() bool {
	return true
}

func (obj *Button) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", ButtonType, opType)
}

func (obj *Button) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Button) Attrs() []object.AttrSpec {
	return buttonAttrs.Specs()
}

// buttonAttrs declares the widget.Button surface (issue #38).
var buttonAttrs = objattrs.Register(string(ButtonType), object.NewAttrRegistry[*Button](string(ButtonType)))

func init() {
	objattrs.Add(buttonAttrs)

	buttonAttrs.Define("Text").
		Doc("The button's label.").
		Getter(func(obj *Button) object.Object { return object.NewString(obj.instance.Text) })
	buttonAttrs.Define("Importance").
		Doc("Visual importance level (a constants.Importance* value).").
		Getter(func(obj *Button) object.Object { return object.NewInt(int64(obj.instance.Importance)) })
	buttonAttrs.Define("Disabled").
		Doc("Whether the button is disabled.").
		Getter(func(obj *Button) object.Object { return object.NewBool(obj.instance.Disabled()) })

	buttonAttrs.Define("SetText").
		Doc("Set the button label.").
		Arg("text").
		Returns("nil").
		Impl(func(obj *Button, ctx context.Context, args ...object.Object) (object.Object, error) {
			text, err := object.AsString(args[0])
			if err != nil {
				return nil, object.Errorf("type error: %v", err)
			}
			guithread.Do(func() {
				obj.instance.SetText(text)
			})
			return object.Nil, nil
		})
	buttonAttrs.Define("Disable").
		Doc("Disable (grey out) the button.").
		Returns("nil").
		Impl(func(obj *Button, ctx context.Context, args ...object.Object) (object.Object, error) {
			guithread.Do(func() {
				obj.instance.Disable()
			})
			return object.Nil, nil
		})
	buttonAttrs.Define("Enable").
		Doc("Re-enable a disabled button.").
		Returns("nil").
		Impl(func(obj *Button, ctx context.Context, args ...object.Object) (object.Object, error) {
			guithread.Do(func() {
				obj.instance.Enable()
			})
			return object.Nil, nil
		})
	buttonAttrs.Define("OnTapped").
		Doc("Set the tap handler; fires on the script VM goroutine.").
		Arg("callback").
		Returns("nil").
		Impl(func(obj *Button, ctx context.Context, args ...object.Object) (object.Object, error) {
			cb, err := vmexec.NewCallback(ctx, "widget.Button.OnTapped", args[0])
			if err != nil {
				return nil, err
			}

			// Event: queued on the script's VM goroutine (same path as the
			// handler passed to widget.NewButton), never run on the UI thread.
			guithread.Do(func() {
				obj.instance.OnTapped = func() { cb.Fire() }
			})
			return object.Nil, nil
		})
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Button) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Button) setAttrUI(name string, value object.Object) error {
	switch name {
	case "Text":
		text, err := object.AsString(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
			obj.instance.SetText(text)
		})
		return nil
	case "Importance":
		// widget.Importance values: High=1, Medium=2, Low=3, DangerHigh=4, WarningHigh=5, SuccessHigh=6
		i, err := object.AsInt(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
			obj.instance.Importance = widget.Importance(i)
			obj.instance.Refresh()
		})
		return nil
	case "Disabled":
		b, err := object.AsBool(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
			if b {
				obj.instance.Disable()
			} else {
				obj.instance.Enable()
			}
		})
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", ButtonType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Button) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *Button) getAttrUI(name string) (object.Object, bool) {
	return buttonAttrs.GetAttr(obj, name)
}

func NewButton(label string, tapped func()) *Button {
	return &Button{instance: widget.NewButton(label, tapped)}
}
