package widget

import (
	"context"
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"github.com/uidbz/fynerisor/gui/guithread"
	"fyne.io/fyne/v2/widget"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &TextGrid{}

const TextGridType object.Type = "widget.TextGrid"

// TextGrid wraps fyne's TextGrid widget - a monospace text grid for displaying code or tabular data.
//
// TextGrid displays text in a fixed-width font with optional line numbers and whitespace indicators.
// Perfect for code editors, terminals, and tabular data.
//
// Example usage in Risor:
//
//	let grid = widget.NewTextGrid()
//	grid.ShowLineNumbers = true
//	grid.TabWidth = 4
//	grid.SetText("function hello() {\n    return 42\n}")
//
//	// Modify individual rows
//	grid.SetRow(0, "// Modified first line")
//	let text = grid.Row(0)
//	let count = grid.RowCount()
//
// Properties:
//   - Text: Full text content
//   - TabWidth: Number of spaces per tab (default: 4)
//   - ShowLineNumbers: Display line numbers
//   - ShowWhitespace: Display tabs and spaces
type TextGrid struct {
	instance *widget.TextGrid
}

func NewTextGrid(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 0 {
		return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
	}

	instance := widget.NewTextGrid()

	return &TextGrid{
		instance: instance,
	}, nil
}

func (obj *TextGrid) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *TextGrid) Type() object.Type {
	return TextGridType
}

func (obj *TextGrid) Inspect() string {
	return "widget.TextGrid"
}

func (obj *TextGrid) Interface() interface{} {
	return obj.instance
}

func (obj *TextGrid) IsTruthy() bool {
	return true
}

func (obj *TextGrid) Cost() int {
	return 0
}

func (obj *TextGrid) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'widget.TextGrid'")
}

func (obj *TextGrid) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(TextGridType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", TextGridType, opType)
	return errObj, err
}

func (obj *TextGrid) Equals(other object.Object) bool {
	return obj == other
}

func (obj *TextGrid) Attrs() []object.AttrSpec {
	return nil
}

func (obj *TextGrid) SetAttr(name string, value object.Object) error {
	switch name {
	case "Text":
		s, err := object.AsString(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
			obj.instance.SetText(s)
		})
		return nil
	case "TabWidth":
		i, err := object.AsInt(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
			obj.instance.TabWidth = int(i)
			obj.instance.Refresh()
		})
		return nil
	case "ShowLineNumbers":
		b, err := object.AsBool(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
			obj.instance.ShowLineNumbers = b
			obj.instance.Refresh()
		})
		return nil
	case "ShowWhitespace":
		b, err := object.AsBool(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
			obj.instance.ShowWhitespace = b
			obj.instance.Refresh()
		})
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", TextGridType, name)
}

func (obj *TextGrid) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Text":
		return object.NewString(obj.instance.Text()), true
	case "TabWidth":
		return object.NewInt(int64(obj.instance.TabWidth)), true
	case "ShowLineNumbers":
		return object.NewBool(obj.instance.ShowLineNumbers), true
	case "ShowWhitespace":
		return object.NewBool(obj.instance.ShowWhitespace), true

	case "SetText":
		return object.NewBuiltin("widget.TextGrid.SetText", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}

			text, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}

			guithread.Do(func() {
				obj.instance.SetText(text)
			})

			return object.Nil, nil
		}), true

	case "SetRow":
		return object.NewBuiltin("widget.TextGrid.SetRow", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return object.Errorf("wrong number of arguments. got=%d, want=2", len(args)), nil
			}

			row, err := object.AsInt(args[0])
			if err != nil {
				return nil, err
			}

			text, err := object.AsString(args[1])
			if err != nil {
				return nil, err
			}

			guithread.Do(func() {
				// Create TextGridRow from string
				runes := []rune(text)
				cells := make([]widget.TextGridCell, len(runes))
				for i, r := range runes {
					cells[i] = widget.TextGridCell{Rune: r}
				}
				obj.instance.SetRow(int(row), widget.TextGridRow{Cells: cells})
			})

			return object.Nil, nil
		}), true

	case "Row":
		return object.NewBuiltin("widget.TextGrid.Row", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}

			row, err := object.AsInt(args[0])
			if err != nil {
				return nil, err
			}

			return object.NewString(obj.instance.RowText(int(row))), nil
		}), true

	case "RowCount":
		return object.NewBuiltin("widget.TextGrid.RowCount", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}

			return object.NewInt(int64(len(obj.instance.Rows))), nil
		}), true

	case "Refresh":
		return object.NewBuiltin("widget.TextGrid.Refresh", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}

			guithread.Do(func() {
				obj.instance.Refresh()
			})

			return object.Nil, nil
		}), true
	case "Hide":
		return object.NewBuiltin("widget.TextGrid.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("widget.TextGrid.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Show()
			})
			return object.Nil, nil
		}), true
	}

	return nil, false
}
