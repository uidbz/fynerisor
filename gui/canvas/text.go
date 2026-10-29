package risorcanvas

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"
)

var _ object.Object = &Text{}

const TextType object.Type = "canvas.text"

// Text wraps a Fyne canvas.Text — a colored, non-wrapping text object. Unlike
// widget.Label it exposes an arbitrary Color, so scripts can tint a heading to
// match a custom background (e.g. a chat bubble hue).
type Text struct {
	instance *canvas.Text
}

func (obj *Text) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Text) Type() object.Type {
	return TextType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Text) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Text) inspectUI() string {
	return "canvas.text"
}

func (obj *Text) Interface() interface{} {
	return obj.instance
}

func (obj *Text) IsTruthy() bool {
	return true
}

func (obj *Text) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", TextType, opType)
}

func (obj *Text) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Text) Attrs() []object.AttrSpec {
	return textAttrs.Specs()
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Text) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Text) setAttrUI(name string, value object.Object) error {
	switch name {
	case "Text":
		s, err := object.AsString(value)
		if err != nil {
			return err
		}
		guithread.Do(func() {
			obj.instance.Text = s
			obj.instance.Refresh()
		})
		return nil
	case "Color":
		s, err := object.AsString(value)
		if err != nil {
			return err
		}
		c, err := ParseColor(s)
		if err != nil {
			return err
		}
		guithread.Do(func() {
			obj.instance.Color = c
			obj.instance.Refresh()
		})
		return nil
	case "Size":
		f, err := object.AsFloat(value)
		if err != nil {
			return err
		}
		guithread.Do(func() {
			obj.instance.TextSize = float32(f)
			obj.instance.Refresh()
		})
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", TextType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Text) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var textAttrs = object.NewAttrRegistry[*Text](string(TextType))

func (obj *Text) getAttrUI(name string) (object.Object, bool) {
	return textAttrs.GetAttr(obj, name)

}

func NewText(instance *canvas.Text) *Text {
	return &Text{instance: instance}
}
func init() {
	objattrs.Add(textAttrs)
	textAttrs.Define("Text").Doc("The text.").Getter(func(obj *Text) object.Object {
		return object.NewString(obj.instance.Text)
	})
	textAttrs.Define("Color").Getter(func(
		obj *Text) object.Object {
		return object.NewString(colorToHex(obj.instance.Color))
	})
	textAttrs.Define("Size").Getter(func(
		obj *Text) object.Object {
		return object.NewFloat(float64(obj.instance.TextSize))
	})
}
