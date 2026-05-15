package widget

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

const TextStyleType object.Type = "widget.TextStyle"

// TextStyle wraps fyne.TextStyle and provides access to Bold, Italic, Monospace, etc.
type TextStyle struct {
	style    *fyne.TextStyle
	widget   *widget.Label // Reference to parent widget for refresh
}

func NewTextStyle(style *fyne.TextStyle, widget *widget.Label) *TextStyle {
	return &TextStyle{
		style:  style,
		widget: widget,
	}
}

func (obj *TextStyle) Type() object.Type {
	return TextStyleType
}

func (obj *TextStyle) Inspect() string {
	return fmt.Sprintf("TextStyle(bold=%v, italic=%v, monospace=%v)",
		obj.style.Bold, obj.style.Italic, obj.style.Monospace)
}

func (obj *TextStyle) Interface() interface{} {
	return obj.style
}

func (obj *TextStyle) IsTruthy() bool {
	return true
}

func (obj *TextStyle) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal TextStyle")
}

func (obj *TextStyle) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for TextStyle: %v", opType), nil
}

func (obj *TextStyle) Equals(other object.Object) bool {
	if other, ok := other.(*TextStyle); ok {
		return obj.style == other.style
	}
	return false
}

func (obj *TextStyle) Attrs() []object.AttrSpec {
	return nil
}

func (obj *TextStyle) SetAttr(name string, value object.Object) error {
	switch name {
	case "Bold":
		bold, err := object.AsBool(value)
		if err != nil {
			return err
		}
		obj.style.Bold = bold
		if obj.widget != nil {
			fyne.Do(func() {
				obj.widget.Refresh()
			})
		}
		return nil
	case "Italic":
		italic, err := object.AsBool(value)
		if err != nil {
			return err
		}
		obj.style.Italic = italic
		if obj.widget != nil {
			fyne.Do(func() {
				obj.widget.Refresh()
			})
		}
		return nil
	case "Monospace":
		monospace, err := object.AsBool(value)
		if err != nil {
			return err
		}
		obj.style.Monospace = monospace
		if obj.widget != nil {
			fyne.Do(func() {
				obj.widget.Refresh()
			})
		}
		return nil
	case "TabWidth":
		tabWidth, err := object.AsInt(value)
		if err != nil {
			return err
		}
		obj.style.TabWidth = int(tabWidth)
		if obj.widget != nil {
			fyne.Do(func() {
				obj.widget.Refresh()
			})
		}
		return nil
	default:
		return fmt.Errorf("attribute error: TextStyle has no attribute %q", name)
	}
}

func (obj *TextStyle) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Bold":
		return object.NewBool(obj.style.Bold), true
	case "Italic":
		return object.NewBool(obj.style.Italic), true
	case "Monospace":
		return object.NewBool(obj.style.Monospace), true
	case "TabWidth":
		return object.NewInt(int64(obj.style.TabWidth)), true
	}
	return nil, false
}
