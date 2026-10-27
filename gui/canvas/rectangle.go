package risorcanvas

import (
	"errors"
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
	"github.com/uidbz/fynerisor/gui/guithread"
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

func (obj *Rectangle) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'canvas.rectangle'")
}

func (obj *Rectangle) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(RectangleType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", RectangleType, opType)
	return errObj, err
}

func (obj *Rectangle) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Rectangle) Attrs() []object.AttrSpec {
	return nil
}

var rectNamedColors = map[string]color.Color{
	"transparent": color.NRGBA{0, 0, 0, 0},
	"red":         color.NRGBA{255, 0, 0, 255},
	"green":       color.NRGBA{0, 128, 0, 255},
	"blue":        color.NRGBA{0, 0, 255, 255},
	"black":       color.NRGBA{0, 0, 0, 255},
	"white":       color.NRGBA{255, 255, 255, 255},
	"yellow":      color.NRGBA{255, 255, 0, 255},
	"cyan":        color.NRGBA{0, 255, 255, 255},
	"magenta":     color.NRGBA{255, 0, 255, 255},
	"gray":        color.NRGBA{128, 128, 128, 255},
	"grey":        color.NRGBA{128, 128, 128, 255},
	"orange":      color.NRGBA{255, 165, 0, 255},
	"purple":      color.NRGBA{128, 0, 128, 255},
}

// ParseColor parses a named color or a hex string ("#RGB", "#RRGGBB" or
// "#RRGGBBAA", with or without the leading '#').
func ParseColor(s string) (color.Color, error) {
	name := strings.ToLower(strings.TrimSpace(s))
	if c, ok := rectNamedColors[name]; ok {
		return c, nil
	}
	hex := strings.TrimPrefix(name, "#")
	if len(hex) == 3 {
		hex = strings.Repeat(string(hex[0]), 2) + strings.Repeat(string(hex[1]), 2) + strings.Repeat(string(hex[2]), 2) + "ff"
	} else if len(hex) == 6 {
		hex += "ff"
	}
	if len(hex) != 8 {
		return nil, fmt.Errorf("invalid color %q: use a name or #RGB/#RRGGBB/#RRGGBBAA", s)
	}
	parts := make([]uint8, 4)
	for i := 0; i < 4; i++ {
		v, err := strconv.ParseUint(hex[i*2:i*2+2], 16, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid color %q: %v", s, err)
		}
		parts[i] = uint8(v)
	}
	return color.NRGBA{R: parts[0], G: parts[1], B: parts[2], A: parts[3]}, nil
}

func colorToHex(c color.Color) string {
	if c == nil {
		return ""
	}
	r, g, b, a := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8), uint8(a>>8))
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

func (obj *Rectangle) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "FillColor":
		return object.NewString(colorToHex(obj.instance.FillColor)), true
	case "StrokeColor":
		return object.NewString(colorToHex(obj.instance.StrokeColor)), true
	case "StrokeWidth":
		return object.NewFloat(float64(obj.instance.StrokeWidth)), true
	case "CornerRadius":
		return object.NewFloat(float64(obj.instance.CornerRadius)), true
	}
	return nil, false
}

func NewRectangle(instance *canvas.Rectangle) *Rectangle {
	return &Rectangle{instance: instance}
}
