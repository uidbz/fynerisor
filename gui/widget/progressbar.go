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

var _ object.Object = &ProgressBar{}

const ProgressBarType object.Type = "widget.ProgressBar"

// ProgressBar wraps widget.ProgressBar for Risor scripting
type ProgressBar struct {
	instance *widget.ProgressBar
}

// NewProgressBar creates a new progress bar (default Min=0, Max=1)
func NewProgressBar() *ProgressBar {
	return &ProgressBar{instance: widget.NewProgressBar()}
}

// NewProgressBarWithData creates a new progress bar bound to a float data binding
func NewProgressBarWithData(data interface{ Get() (float64, error) }) *ProgressBar {
	// Cast to binding.Float
	bindingFloat, ok := data.(binding.Float)
	if !ok {
		// Fallback: create a regular progress bar
		return &ProgressBar{instance: widget.NewProgressBar()}
	}

	return &ProgressBar{instance: widget.NewProgressBarWithData(bindingFloat)}
}

func (obj *ProgressBar) Type() object.Type {
	return ProgressBarType
}

func (obj *ProgressBar) Inspect() string {
	return fmt.Sprintf("widget.ProgressBar(value=%f)", obj.instance.Value)
}

func (obj *ProgressBar) Interface() interface{} {
	return obj.instance
}

func (obj *ProgressBar) IsTruthy() bool {
	return true
}

func (obj *ProgressBar) Cost() int {
	return 0
}

func (obj *ProgressBar) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *ProgressBar) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'widget.ProgressBar'")
}

func (obj *ProgressBar) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(ProgressBarType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", ProgressBarType, opType)
	return errObj, err
}

func (obj *ProgressBar) Equals(other object.Object) bool {
	return obj == other
}

func (obj *ProgressBar) Attrs() []object.AttrSpec {
	return nil
}

func (obj *ProgressBar) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Value":
		return object.NewFloat(obj.instance.Value), true
	case "Min":
		return object.NewFloat(obj.instance.Min), true
	case "Max":
		return object.NewFloat(obj.instance.Max), true
	case "Hide":
		return object.NewBuiltin("widget.ProgressBar.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("widget.ProgressBar.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Show()
			})
			return object.Nil, nil
		}), true
	}
	return nil, false
}

func (obj *ProgressBar) SetAttr(name string, value object.Object) error {
	switch name {
	case "Value":
		val, err := object.AsFloat(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		fyne.Do(func() {
			obj.instance.SetValue(val)
		})
		return nil
	case "Min":
		val, err := object.AsFloat(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		obj.instance.Min = val
		fyne.Do(func() {
			obj.instance.Refresh()
		})
		return nil
	case "Max":
		val, err := object.AsFloat(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		obj.instance.Max = val
		fyne.Do(func() {
			obj.instance.Refresh()
		})
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", ProgressBarType, name)
}
