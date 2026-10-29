package dialog

import (
	"context"
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"github.com/uidbz/fynerisor/gui/guithread"

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

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *ColorPickerDialog) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *ColorPickerDialog) inspectUI() string {
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
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", ColorPickerDialogType, opType)
}

func (obj *ColorPickerDialog) Equals(other object.Object) bool {
	return obj == other
}

func (obj *ColorPickerDialog) Attrs() []object.AttrSpec {
	return nil
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *ColorPickerDialog) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *ColorPickerDialog) setAttrUI(name string, value object.Object) error {
	switch name {
	case "Advanced":
		advanced, err := object.AsBool(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
			obj.instance.Advanced = advanced
			obj.instance.Refresh()
		})
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", ColorPickerDialogType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *ColorPickerDialog) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *ColorPickerDialog) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "Advanced":
		return object.NewBool(obj.instance.Advanced), true
	case "Show":
		return object.NewBuiltin("dialog.ColorPickerDialog.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			guithread.Do(func() {
				obj.instance.Show()
			})
			return object.Nil, nil
		}), true
	case "Hide":
		return object.NewBuiltin("dialog.ColorPickerDialog.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Refresh":
		return object.NewBuiltin("dialog.ColorPickerDialog.Refresh", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			guithread.Do(func() {
				obj.instance.Refresh()
			})
			return object.Nil, nil
		}), true
	case "SetDismissText":
		return object.NewBuiltin("dialog.ColorPickerDialog.SetDismissText", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}
			text, err := object.AsString(args[0])
			if err != nil {
				return nil, object.Errorf("type error: %v", err)
			}
			guithread.Do(func() {
				obj.instance.SetDismissText(text)
			})
			return object.Nil, nil
		}), true
	case "SetColor":
		return object.NewBuiltin("dialog.ColorPickerDialog.SetColor", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}
			colorMap, ok := args[0].(*object.Map)
			if !ok {
				return nil, object.Errorf("type error: color must be a map with R, G, B, A keys")
			}

			r, _ := object.AsInt(colorMap.Value()["R"])
			g, _ := object.AsInt(colorMap.Value()["G"])
			b, _ := object.AsInt(colorMap.Value()["B"])
			a := int64(255)
			if aVal, ok := colorMap.Value()["A"]; ok {
				a, _ = object.AsInt(aVal)
			}

			c := color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}

			guithread.Do(func() {
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
