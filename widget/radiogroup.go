package widget

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

const RadioGroupType object.Type = "widget.RadioGroup"

type RadioGroup struct {
	instance *widget.RadioGroup
}

func NewRadioGroup(options []string, onChange func(string)) *RadioGroup {
	rg := &RadioGroup{
		instance: widget.NewRadioGroup(options, onChange),
	}
	return rg
}

func (obj *RadioGroup) Type() object.Type {
	return RadioGroupType
}

func (obj *RadioGroup) Inspect() string {
	return fmt.Sprintf("RadioGroup(selected=%q, options=%d)", obj.instance.Selected, len(obj.instance.Options))
}

func (obj *RadioGroup) Interface() interface{} {
	return obj.instance
}

func (obj *RadioGroup) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *RadioGroup) IsTruthy() bool {
	return true
}

func (obj *RadioGroup) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal RadioGroup")
}

func (obj *RadioGroup) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for RadioGroup: %v", opType), nil
}

func (obj *RadioGroup) Equals(other object.Object) bool {
	if other, ok := other.(*RadioGroup); ok {
		return obj.instance == other.instance
	}
	return false
}

func (obj *RadioGroup) Attrs() []object.AttrSpec {
	return nil
}

func (obj *RadioGroup) SetAttr(name string, value object.Object) error {
	switch name {
	case "Selected":
		selected, err := object.AsString(value)
		if err != nil {
			return err
		}
		fyne.Do(func() {
			obj.instance.Selected = selected
			obj.instance.Refresh()
		})
		return nil
	case "Options":
		options, err := object.AsStringSlice(value)
		if err != nil {
			return err
		}
		fyne.Do(func() {
			obj.instance.Options = options
			obj.instance.Refresh()
		})
		return nil
	case "Horizontal":
		horizontal, err := object.AsBool(value)
		if err != nil {
			return err
		}
		fyne.Do(func() {
			obj.instance.Horizontal = horizontal
			obj.instance.Refresh()
		})
		return nil
	case "Required":
		required, err := object.AsBool(value)
		if err != nil {
			return err
		}
		fyne.Do(func() {
			obj.instance.Required = required
			obj.instance.Refresh()
		})
		return nil
	default:
		return fmt.Errorf("attribute error: RadioGroup has no attribute %q", name)
	}
}

func (obj *RadioGroup) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Selected":
		return object.NewString(obj.instance.Selected), true
	case "Options":
		return object.NewStringList(obj.instance.Options), true
	case "Horizontal":
		return object.NewBool(obj.instance.Horizontal), true
	case "Required":
		return object.NewBool(obj.instance.Required), true

	case "SetSelected":
		return object.NewBuiltin("RadioGroup.SetSelected", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			selected, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			fyne.Do(func() {
				obj.instance.SetSelected(selected)
			})
			return object.Nil, nil
		}), true

	case "Append":
		return object.NewBuiltin("RadioGroup.Append", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			option, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			fyne.Do(func() {
				obj.instance.Append(option)
			})
			return object.Nil, nil
		}), true

	case "Disable":
		return object.NewBuiltin("RadioGroup.Disable", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Disable()
			})
			return object.Nil, nil
		}), true

	case "Enable":
		return object.NewBuiltin("RadioGroup.Enable", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Enable()
			})
			return object.Nil, nil
		}), true

	case "Refresh":
		return object.NewBuiltin("RadioGroup.Refresh", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Refresh()
			})
			return object.Nil, nil
		}), true
	}
	return nil, false
}
