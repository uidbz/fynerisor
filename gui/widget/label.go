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

var _ object.Object = &Label{}

const LabelType object.Type = "widget.Label"

type Label struct {
	instance *widget.Label
}

func (obj *Label) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Label) Type() object.Type {
	return LabelType
}

func (obj *Label) Inspect() string {
	return "widget.Label"
}

func (obj *Label) Interface() interface{} {
	return obj.instance
}

func (obj *Label) IsTruthy() bool {
	return true
}

func (obj *Label) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'widget.Label'")
}

func (obj *Label) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(LabelType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", LabelType, opType)
	return errObj, err
}

func (obj *Label) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Label) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Label) SetAttr(name string, value object.Object) error {
	switch name {
	case "Text":
		text, err := object.AsString(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		fyne.Do(func() {
			obj.instance.SetText(text)
		})
		return nil
	case "Wrapping":
		i, err := object.AsInt(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		fyne.Do(func() {
			obj.instance.Wrapping = fyne.TextWrap(i)
			obj.instance.Refresh()
		})
		return nil
	case "Truncation":
		i, err := object.AsInt(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		fyne.Do(func() {
			obj.instance.Truncation = fyne.TextTruncation(i)
			obj.instance.Refresh()
		})
		return nil
	case "Alignment":
		i, err := object.AsInt(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		fyne.Do(func() {
			obj.instance.Alignment = fyne.TextAlign(i)
			obj.instance.Refresh()
		})
		return nil
	case "Importance":
		i, err := object.AsInt(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		fyne.Do(func() {
			obj.instance.Importance = widget.Importance(i)
			obj.instance.Refresh()
		})
		return nil
	case "TextStyle":
		return fmt.Errorf("attribute error: TextStyle is read-only, set Bold/Italic/Monospace properties instead")
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", LabelType, name)
}

func (obj *Label) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Text":
		return object.NewString(obj.instance.Text), true
	case "Wrapping":
		return object.NewInt(int64(obj.instance.Wrapping)), true
	case "Truncation":
		return object.NewInt(int64(obj.instance.Truncation)), true
	case "Alignment":
		return object.NewInt(int64(obj.instance.Alignment)), true
	case "Importance":
		return object.NewInt(int64(obj.instance.Importance)), true
	case "TextStyle":
		return NewTextStyle(&obj.instance.TextStyle, obj.instance), true
	case "SetText":
		return object.NewBuiltin("widget.Label.SetText", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			text, err := object.AsString(args[0])
			if err != nil {
				return object.Errorf("type error: %v", err), nil
			}
			fyne.Do(func() {
				obj.instance.SetText(text)
			})
			return object.Nil, nil
		}), true
	}
	return nil, false
}

func NewLabel(text string) *Label {
	return &Label{instance: widget.NewLabel(text)}
}

func NewLabelWithData(data interface{ Get() (string, error) }) *Label {
	// Cast to binding.String
	bindingStr, ok := data.(binding.String)
	if !ok {
		// Fallback: create a label with empty text
		return &Label{instance: widget.NewLabel("")}
	}

	return &Label{instance: widget.NewLabelWithData(bindingStr)}
}
