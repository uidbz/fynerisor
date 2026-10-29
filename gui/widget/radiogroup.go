package widget

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"

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

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *RadioGroup) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *RadioGroup) inspectUI() string {
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

func (obj *RadioGroup) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for RadioGroup: %v", opType)
}

func (obj *RadioGroup) Equals(other object.Object) bool {
	if other, ok := other.(*RadioGroup); ok {
		return obj.instance == other.instance
	}
	return false
}

func (obj *RadioGroup) Attrs() []object.AttrSpec {
	return radioGroupAttrs.

		// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
		Specs()
}

func (obj *RadioGroup) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *RadioGroup) setAttrUI(name string, value object.Object) error {
	switch name {
	case "Selected":
		selected, err := object.AsString(value)
		if err != nil {
			return err
		}
		guithread.Do(func() {
			obj.instance.Selected = selected
			obj.instance.Refresh()
		})
		return nil
	case "Options":
		options, err := object.AsStringSlice(value)
		if err != nil {
			return err
		}
		guithread.Do(func() {
			obj.instance.Options = options
			obj.instance.Refresh()
		})
		return nil
	case "Horizontal":
		horizontal, err := object.AsBool(value)
		if err != nil {
			return err
		}
		guithread.Do(func() {
			obj.instance.Horizontal = horizontal
			obj.instance.Refresh()
		})
		return nil
	case "Required":
		required, err := object.AsBool(value)
		if err != nil {
			return err
		}
		guithread.Do(func() {
			obj.instance.Required = required
			obj.instance.Refresh()
		})
		return nil
	default:
		return fmt.Errorf("attribute error: RadioGroup has no attribute %q", name)
	}
}

var radioGroupAttrs = object.NewAttrRegistry[*RadioGroup](string(RadioGroupType))

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *RadioGroup) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *RadioGroup) getAttrUI(name string) (object.Object, bool) {
	return radioGroupAttrs.GetAttr(obj, name)

}
func init() {
	objattrs.Add(radioGroupAttrs)
	radioGroupAttrs.
		Define("Selected").Doc("The selected index.").Getter(func(obj *RadioGroup) object.Object {
		return object.NewString(obj.
			instance.Selected)
	})
	radioGroupAttrs.
		Define("Options").Getter(func(obj *RadioGroup) object.Object {
		return object.
			NewStringList(obj.instance.Options)
	})
	radioGroupAttrs.
		Define("Horizontal").
		Getter(
			func(obj *RadioGroup) object.Object {
				return object.NewBool(obj.instance.Horizontal)
			})
	radioGroupAttrs.
		Define("Required").Getter(func(obj *RadioGroup) object.Object {
		return object.
			NewBool(obj.instance.Required)
	})
	radioGroupAttrs.
		Define("SetSelected").Args(
		"arg1",
	).Returns("nil").Impl(func(obj *RadioGroup, ctx context.Context, args ...object.Object) (object.Object, error) {
		selected, err := object.
			AsString(args[0])
		if err != nil {
			return nil, err
		}
		guithread.Do(func() {
			obj.instance.SetSelected(selected)
		})
		return object.Nil, nil
	})
	radioGroupAttrs.
		Define("Append").Args("arg1").Returns("nil").Impl(func(obj *RadioGroup, ctx context.Context, args ...object.Object) (
		object.
			Object, error) {
		option, err := object.AsString(args[0])
		if err != nil {
			return nil,
				err

		}
		guithread.Do(func() {
			obj.instance.Append(option)

		})
		return object.Nil, nil
	})
	radioGroupAttrs.
		Define("Disable").Doc("Disable the object.").Args().Returns("nil").Impl(func(obj *RadioGroup, ctx context.Context, args ...object.Object) (object.Object, error) {
		guithread.
			Do(func() { obj.instance.Disable() })

		return object.Nil, nil
	})
	radioGroupAttrs.
		Define("Enable").Doc("Re-enable the object.").Args().Returns("nil").Impl(
		func(obj *RadioGroup, ctx context.Context, args ...object.Object) (object.Object, error) {
			guithread.
				Do(func() { obj.instance.Enable() })

			return object.Nil, nil
		})
}
