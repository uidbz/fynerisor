package widget

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &Check{}

const CheckType object.Type = "widget.Check"

type Check struct {
	instance *widget.Check
}

func (obj *Check) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Check) Type() object.Type {
	return CheckType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Check) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Check) inspectUI() string {
	return "widget.Check"
}

func (obj *Check) Interface() interface{} {
	return obj.instance
}

func (obj *Check) IsTruthy() bool {
	return true
}

func (obj *Check) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", CheckType, opType)
}

func (obj *Check) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Check) Attrs() []object.AttrSpec {
	return checkAttrs.

		// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
		Specs()
}

func (obj *Check) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Check) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", CheckType, name)
}

var checkAttrs = object.NewAttrRegistry[*Check](string(CheckType))

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Check) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *Check) getAttrUI(name string) (object.Object, bool) {
	return checkAttrs.GetAttr(obj, name)

}

func NewCheck(label string, changed func(bool)) *Check {
	return &Check{instance: widget.NewCheck(label, changed)}
}

func NewCheckWithData(label string, data interface{ Get() (bool, error) }) *Check {
	// Cast to binding.Bool
	bindingBool, ok := data.(binding.Bool)
	if !ok {
		// Fallback: create a check without binding
		return &Check{instance: widget.NewCheck(label, nil)}
	}

	return &Check{instance: widget.NewCheckWithData(label, bindingBool)}
}
func init() {
	objattrs.Add(checkAttrs)
	checkAttrs.Define("SetChecked").Args("arg1").
		Returns("nil").Impl(func(obj *Check, ctx context.Context, args ...object.Object) (object.Object, error) {
		val, err := object.
			AsBool(args[0])
		if err != nil {
			return nil,
				err
		}
		obj.instance.SetChecked(val)
		return object.
			Nil, nil
	})
	checkAttrs.Define("Checked").Getter(
		func(obj *Check) object.Object {
			return object.NewBool(obj.
				instance.Checked)
		})
}
