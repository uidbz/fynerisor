package risorcanvas

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"
	"github.com/uidbz/fynerisor/internal/colors"
)

var _ object.Object = &Rectangle{}

const RectangleType object.Type = "canvas.rectangle"

// Rectangle wraps a Fyne canvas.Rectangle. Scripts set FillColor, StrokeColor,
// StrokeWidth and CornerRadius as attributes (hex "#RGB"/"#RRGGBB"/"#RRGGBBAA"
// or named colors), e.g. a rounded background behind a chat message:
//
//	let bg = canvas.NewRectangle("#3b82f633", 8.0)
//	container.NewStack(bg, container.NewPadded(content))
type Rectangle struct {
	instance *canvas.Rectangle
}

func (obj *Rectangle) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Rectangle) Type() object.Type {
	return RectangleType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Rectangle) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Rectangle) inspectUI() string {
	return "canvas.rectangle"
}

func (obj *Rectangle) Interface() interface{} {
	return obj.instance
}

func (obj *Rectangle) IsTruthy() bool {
	return true
}

func (obj *Rectangle) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", RectangleType, opType)
}

func (obj *Rectangle) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Rectangle) Attrs() []object.AttrSpec {
	return rectangleAttrs.Specs()
}

// ParseColor parses a named color or a hex string ("#RGB", "#RRGGBB" or
// "#RRGGBBAA", with or without the leading '#').
func ParseColor(s string) (color.Color, error) { return colors.Parse(s) }

func colorToHex(c color.Color) string {
	if c == nil {
		return ""
	}
	return colors.Hex(c)
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Rectangle) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Rectangle) setAttrUI(name string, value object.Object) error {
	switch name {
	case "FillColor":
		s, err := object.AsString(value)
		if err != nil {
			return err
		}
		c, err := ParseColor(s)
		if err != nil {
			return err
		}
		guithread.Do(func() {
			obj.instance.FillColor = c
			obj.instance.Refresh()
		})
		return nil
	case "StrokeColor":
		s, err := object.AsString(value)
		if err != nil {
			return err
		}
		c, err := ParseColor(s)
		if err != nil {
			return err
		}
		guithread.Do(func() {
			obj.instance.StrokeColor = c
			obj.instance.Refresh()
		})
		return nil
	case "StrokeWidth":
		f, err := object.AsFloat(value)
		if err != nil {
			return err
		}
		guithread.Do(func() {
			obj.instance.StrokeWidth = float32(f)
			obj.instance.Refresh()
		})
		return nil
	case "CornerRadius":
		f, err := object.AsFloat(value)
		if err != nil {
			return err
		}
		guithread.Do(func() {
			obj.instance.CornerRadius = float32(f)
			obj.instance.Refresh()
		})
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", RectangleType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Rectangle) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var rectangleAttrs = objattrs.Register(string(RectangleType), object.NewAttrRegistry[*Rectangle](string(RectangleType)))

func (obj *Rectangle) getAttrUI(name string) (object.Object, bool) {
	return rectangleAttrs.GetAttr(obj, name)

}

func NewRectangle(instance *canvas.Rectangle) *Rectangle {
	return &Rectangle{instance: instance}
}
func init() {
	objattrs.Add(rectangleAttrs)
	rectangleAttrs.Define("FillColor").Getter(func(obj *Rectangle) object.Object {
		return object.NewString(colorToHex(obj.instance.
			FillColor))
	})
	rectangleAttrs.Define("StrokeColor").
		Getter(func(obj *Rectangle) object.Object {
			return object.
				NewString(colorToHex(obj.
					instance.
					StrokeColor))
		})
	rectangleAttrs.Define("StrokeWidth").
		Getter(func(obj *Rectangle) object.Object {
			return object.
				NewFloat(float64(obj.instance.
					StrokeWidth))
		})
	rectangleAttrs.Define("CornerRadius").
		Getter(func(obj *Rectangle) object.Object {
			return object.
				NewFloat(float64(obj.instance.
					CornerRadius))
		})
}
