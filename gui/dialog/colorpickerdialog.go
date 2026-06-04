package dialog

import (
	"context"
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &ColorPickerDialog{}

const ColorPickerDialogType object.Type = "dialog.ColorPickerDialog"

// ColorPickerDialog wraps dialog.ColorPickerDialog for Risor scripting
type ColorPickerDialog struct {
	instance *dialog.ColorPickerDialog
	window   fyne.Window
}

func (obj *ColorPickerDialog) Type() object.Type {
	return ColorPickerDialogType
}

func (obj *ColorPickerDialog) Inspect() string {
	return "dialog.ColorPickerDialog"
}

func (obj *ColorPickerDialog) Interface() interface{} {
	return obj.instance
}

func (obj *ColorPickerDialog) IsTruthy() bool {
	return true
}

func (obj *ColorPickerDialog) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'dialog.ColorPickerDialog'")
}

func (obj *ColorPickerDialog) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for %s: %v", ColorPickerDialogType, opType), nil
}

func (obj *ColorPickerDialog) Equals(other object.Object) bool {
	return obj == other
}

func (obj *ColorPickerDialog) Attrs() []object.AttrSpec {
	return nil
}

func (obj *ColorPickerDialog) SetAttr(name string, value object.Object) error {
	switch name {
	case "Advanced":
		advanced, err := object.AsBool(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		fyne.Do(func() {
			obj.instance.Advanced = advanced
			obj.instance.Refresh()
		})
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", ColorPickerDialogType, name)
}

func (obj *ColorPickerDialog) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Advanced":
		return object.NewBool(obj.instance.Advanced), true
	case "Show":
		return object.NewBuiltin("dialog.ColorPickerDialog.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Show()
			})
			return object.Nil, nil
		}), true
	case "Hide":
		return object.NewBuiltin("dialog.ColorPickerDialog.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Refresh":
		return object.NewBuiltin("dialog.ColorPickerDialog.Refresh", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Refresh()
			})
			return object.Nil, nil
		}), true
	case "SetDismissText":
		return object.NewBuiltin("dialog.ColorPickerDialog.SetDismissText", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			text, err := object.AsString(args[0])
			if err != nil {
				return object.Errorf("type error: %v", err), nil
			}
			fyne.Do(func() {
				obj.instance.SetDismissText(text)
			})
			return object.Nil, nil
		}), true
	case "SetColor":
		return object.NewBuiltin("dialog.ColorPickerDialog.SetColor", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			colorMap, ok := args[0].(*object.Map)
			if !ok {
				return object.Errorf("type error: color must be a map with R, G, B, A keys"), nil
			}

			r, _ := object.AsInt(colorMap.Value()["R"])
			g, _ := object.AsInt(colorMap.Value()["G"])
			b, _ := object.AsInt(colorMap.Value()["B"])
			a := int64(255)
			if aVal, ok := colorMap.Value()["A"]; ok {
				a, _ = object.AsInt(aVal)
			}

			c := color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}

			fyne.Do(func() {
				obj.instance.SetColor(c)
			})
			return object.Nil, nil
		}), true
	}
	return nil, false
}

func NewColorPicker(title, message string, callback func(color.Color), parent fyne.Window) *ColorPickerDialog {
	return &ColorPickerDialog{
		instance: dialog.NewColorPicker(title, message, callback, parent),
		window:   parent,
	}
}
