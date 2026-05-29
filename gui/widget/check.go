package widget

import (
	"context"
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"

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

func (obj *Check) Inspect() string {
	return "widget.Check"
}

func (obj *Check) Interface() interface{} {
	return obj.instance
}

func (obj *Check) IsTruthy() bool {
	return true
}

func (obj *Check) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'widget.Check'")
}

func (obj *Check) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(CheckType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", CheckType, opType)
	return errObj, err
}

func (obj *Check) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Check) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Check) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", CheckType, name)
}

func (obj *Check) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "SetChecked":
		return object.NewBuiltin("widget.Check.SetChecked", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=exactly 1", len(args)), nil
			}

			val, err := object.AsBool(args[0])
			if err != nil {
				return nil, err
			}

			obj.instance.SetChecked(val)

			return object.Nil, nil
		}), true
	case "Checked":
		return object.NewBool(obj.instance.Checked), true
	}
	return nil, false
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
