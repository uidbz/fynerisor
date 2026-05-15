package fynerisor

import (
	"context"
	"errors"
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2/canvas"

	risorcanvas "github.com/uidbz/fynerisor/canvas"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
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

func (obj *Canvas) Inspect() string {
	return "canvas"
}

func (obj *Canvas) Interface() interface{} {
	return nil
}

func (obj *Canvas) IsTruthy() bool {
	return true
}

func (obj *Canvas) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'canvas'")
}

func (obj *Canvas) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(CanvasType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", CanvasType, opType)
	return errObj, err
}

func (obj *Canvas) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Canvas) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Canvas) SetAttr(name string, value object.Object) error {
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

func (obj *Canvas) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "NewLine":
		return object.NewBuiltin("canvas.NewLine", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			color, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}

			line := canvas.NewLine(getNamedColor(color))

			return risorcanvas.NewLine(line), nil
		}), true

	case "NewImageFromURI":
		return object.NewBuiltin("canvas.NewImageFromURI", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			path, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}

			img, err := risorcanvas.NewImageFromURI(path)
			if err != nil {
				return object.NewError(err), nil
			}

			return img, nil
		}), true
	}
	return nil, false
}

func NewFyneCanvas() *Canvas {
	return &Canvas{}
}
