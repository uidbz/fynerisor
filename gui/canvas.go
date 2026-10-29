package gui

import (
	"context"
	"fmt"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"
	"image/color"
	"strings"

	"fyne.io/fyne/v2/canvas"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
	risorcanvas "github.com/uidbz/fynerisor/gui/canvas"
)

var _ object.Object = &Canvas{}

const CanvasType object.Type = "canvas"

// Canvas is a factory object that provides canvas drawing functions to Risor scripts.
// It implements object.Object and exposes methods via GetAttr() for creating
// canvas objects like lines and images.
//
// Available in Risor scripts as the global 'canvas' object.
//
// Example usage in Risor:
//
//	let line = canvas.NewLine("red")
//	let img = canvas.NewImageFromURI("file:///path/to/image.png")
//
// Thread Safety: All canvas object creation is thread-safe.
type Canvas struct {
}

func (obj *Canvas) Type() object.Type {
	return CanvasType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Canvas) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Canvas) inspectUI() string {
	return "canvas"
}

func (obj *Canvas) Interface() interface{} {
	return nil
}

func (obj *Canvas) IsTruthy() bool {
	return true
}

func (obj *Canvas) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", CanvasType, opType)
}

func (obj *Canvas) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Canvas) Attrs() []object.AttrSpec {
	return append(canvasAttrs.Specs(),
		object.AttrSpec{Name: "NewImageFromURI", Doc: "Create an Image loading from a URI string; the fetch runs off the UI thread", Args: []string{"uri"}, Returns: "canvas.Image"},
	)
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Canvas) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Canvas) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", CanvasType, name)
}

var namedColors = map[string]color.Color{
	"red":     color.RGBA{255, 0, 0, 255},
	"green":   color.RGBA{0, 255, 0, 255},
	"blue":    color.RGBA{0, 0, 255, 255},
	"black":   color.RGBA{0, 0, 0, 255},
	"white":   color.RGBA{255, 255, 255, 255},
	"yellow":  color.RGBA{255, 255, 0, 255},
	"cyan":    color.RGBA{0, 255, 255, 255},
	"magenta": color.RGBA{255, 0, 255, 255},
}

func getNamedColor(name string) color.Color {
	c, ok := namedColors[strings.ToLower(name)]
	if !ok {
		return namedColors["black"]
	}
	return c
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Canvas) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var canvasAttrs = objattrs.Register(string(CanvasType), object.NewAttrRegistry[*Canvas](string(CanvasType)))

func (obj *Canvas) getAttrUI(name string) (object.Object, bool) {
	if o, ok := canvasAttrs.GetAttr(obj, name); ok {
		return o, true
	}
	switch name {

	case "NewImageFromURI":
		// OffUI: loading may hit the network or disk; the image is not on
		// screen yet, so it is built on the script's goroutine.
		return guithread.OffUI(object.NewBuiltin("canvas.NewImageFromURI", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}
			path, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}

			img, err := risorcanvas.NewImageFromURIContext(ctx, path)
			if err != nil {
				return nil, object.NewError(err)
			}

			return img, nil
		})), true
	}
	return nil, false

}

func NewFyneCanvas() *Canvas {
	return &Canvas{}
}
func init() {
	canvasAttrs.
		Define("NewLine").Variadic("args").Returns("nil").Impl(func(obj *Canvas,
		ctx context.Context, args ...object.Object) (object.
		Object, error) {
		if len(args) != 1 {
			return nil,
				object.Errorf("wrong number of arguments. got=%d, want=1",
					len(args))
		}
		color, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}
		line := canvas.NewLine(
			getNamedColor(color))
		return risorcanvas.NewLine(line), nil
	})
	canvasAttrs.
		Define("NewText").Variadic("args").Returns("nil").Impl(func(obj *Canvas,
		ctx context.Context, args ...object.Object) (object.
		Object, error) {
		if len(args) != 2 {
			return nil,
				object.Errorf("wrong number of arguments. got=%d, want=2",
					len(args))
		}
		text, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}
		colorName, err := object.
			AsString(args[1])
		if err != nil {
			return nil, err
		}
		c, err := risorcanvas.ParseColor(colorName)
		if err != nil {
			return nil, err
		}
		txt := canvas.NewText(text, c)
		return risorcanvas.NewText(txt), nil
	},
	)
	canvasAttrs.
		Define("NewRectangle").Variadic("args").Returns("nil").Impl(func(obj *Canvas, ctx context.
		Context,

		args ...object.Object) (object.
		Object, error) {
		if len(args) != 2 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=2",
				len(args))
		}
		colorName, err :=
			object.
				AsString(args[0])
		if err !=

			nil {
			return nil, err
		}
		cornerRadius,
			err := object.AsFloat(args[1])
		if err != nil {
			return nil, err
		}
		c, err := risorcanvas.
			ParseColor(colorName)
		if err != nil {
			return nil, err
		}
		rect := canvas.NewRectangle(c)
		rect.CornerRadius = float32(cornerRadius)
		return risorcanvas.NewRectangle(rect), nil
	})
}
