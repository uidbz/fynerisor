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

var _ object.Object = &Slider{}

const SliderType object.Type = "widget.Slider"

// Slider wraps widget.Slider for Risor scripting
type Slider struct {
	instance *widget.Slider
	w WindowInterface
}

// NewSlider creates a new slider with min and max values
func NewSlider(min, max float64, w WindowInterface) *Slider {
	return &Slider{
		instance: widget.NewSlider(min, max),
		w:        w,
	}
}

func NewSliderWithData(min, max float64, data interface{ Get() (float64, error) }, w WindowInterface) *Slider {
	// Cast to binding.Float
	bindingFloat, ok := data.(binding.Float)
	if !ok {
		// Fallback: create a slider without binding
		return &Slider{
			instance: widget.NewSlider(min, max),
			w:        w,
		}
	}

	return &Slider{
		instance: widget.NewSliderWithData(min, max, bindingFloat),
		w:        w,
	}
}

func (obj *Slider) Type() object.Type {
	return SliderType
}

func (obj *Slider) Inspect() string {
	return fmt.Sprintf("widget.Slider(value=%f, min=%f, max=%f)", obj.instance.Value, obj.instance.Min, obj.instance.Max)
}

func (obj *Slider) Interface() interface{} {
	return obj.instance
}

func (obj *Slider) IsTruthy() bool {
	return true
}

func (obj *Slider) Cost() int {
	return 0
}

func (obj *Slider) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Slider) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'widget.Slider'")
}

func (obj *Slider) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(SliderType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", SliderType, opType)
	return errObj, err
}

func (obj *Slider) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Slider) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Slider) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Value":
		return object.NewFloat(obj.instance.Value), true
	case "Min":
		return object.NewFloat(obj.instance.Min), true
	case "Max":
		return object.NewFloat(obj.instance.Max), true
	case "Step":
		return object.NewFloat(obj.instance.Step), true
	case "OnChanged":
		return object.NewBuiltin("widget.Slider.OnChanged", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("slider: unable to get call function"), nil
			}

			obj.instance.OnChanged = func(value float64) {
				obj.w.Do(func() {
					callFunc(ctx, fn, []object.Object{object.NewFloat(value)})
				})
			}
			return object.Nil, nil
		}), true
	}
	return nil, false
}

func (obj *Slider) SetAttr(name string, value object.Object) error {
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
	case "Step":
		val, err := object.AsFloat(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		obj.instance.Step = val
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", SliderType, name)
}
