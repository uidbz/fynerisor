package widget

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"
	"github.com/uidbz/fynerisor/gui/vmexec"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &Slider{}

const SliderType object.Type = "widget.Slider"

// Slider wraps widget.Slider for Risor scripting
type Slider struct {
	instance *widget.Slider
	w        WindowInterface
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

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Slider) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Slider) inspectUI() string {
	return fmt.Sprintf("widget.Slider(value=%f, min=%f, max=%f)", obj.instance.Value, obj.instance.Min, obj.instance.Max)
}

func (obj *Slider) Interface() interface{} {
	return obj.instance
}

func (obj *Slider) IsTruthy() bool {
	return true
}

func (obj *Slider) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Slider) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", SliderType, opType)
}

func (obj *Slider) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Slider) Attrs() []object.AttrSpec {
	return sliderAttrs.Specs()
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
var sliderAttrs = objattrs.Register(string(SliderType), object.NewAttrRegistry[*Slider](string(SliderType)))

func (obj *Slider) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *Slider) getAttrUI(name string) (object.Object, bool) {
	// Change event: coalesced, latest value wins — a fast drag
	// produces far more updates than the VM can run (issue #51).
	return sliderAttrs.GetAttr(obj, name)
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Slider) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Slider) setAttrUI(name string, value object.Object) error {
	switch name {
	case "Value":
		val, err := object.AsFloat(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
			obj.instance.SetValue(val)
		})
		return nil
	case "Min":
		val, err := object.AsFloat(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		obj.instance.Min = val
		guithread.Do(func() {
			obj.instance.Refresh()
		})
		return nil
	case "Max":
		val, err := object.AsFloat(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		obj.instance.Max = val
		guithread.Do(func() {
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
func init() {
	objattrs.Add(sliderAttrs)
	sliderAttrs.Define("Value").Getter(func(obj *Slider) object.Object {
		return object.NewFloat(obj.
			instance.Value)
	})
	sliderAttrs.Define("Min").
		Getter(func(obj *Slider) object.Object {
			return object.NewFloat(obj.instance.
				Min)
		})
	sliderAttrs.Define("Max").
		Getter(func(obj *Slider) object.Object {
			return object.NewFloat(obj.instance.
				Max)
		})
	sliderAttrs.Define("Step").Getter(func(obj *Slider) object.Object {
		return object.NewFloat(obj.
			instance.Step)
	})
	sliderAttrs.Define("OnChanged").Doc(
		"Set the change handler; fires on the script VM goroutine.",
	).Args("arg1").Returns("nil").Impl(func(obj *Slider, ctx context.Context, args ...object.Object) (object.Object,

		error) {
		cb, err := vmexec.NewCallback(ctx, "widget.Slider.OnChanged",
			args[0])
		if err !=
			nil {
			return nil, err
		}
		obj.instance.
			OnChanged = func(value float64) {
			cb.FireCoalesced(object.NewFloat(value))
		}
		return object.Nil,
			nil
	})
}
