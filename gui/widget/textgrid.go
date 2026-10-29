package widget

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"

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
		return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
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

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *TextGrid) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *TextGrid) inspectUI() string {
	return "widget.TextGrid"
}

func (obj *TextGrid) Interface() interface{} {
	return obj.instance
}

func (obj *TextGrid) IsTruthy() bool {
	return true
}

func (obj *TextGrid) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", TextGridType, opType)
}

func (obj *TextGrid) Equals(other object.Object) bool {
	return obj == other
}

func (obj *TextGrid) Attrs() []object.AttrSpec {
	return textGridAttrs.Specs()
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *TextGrid) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *TextGrid) setAttrUI(name string, value object.Object) error {
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

var textGridAttrs = objattrs.Register(string(TextGridType), object.NewAttrRegistry[*TextGrid](string(TextGridType)))

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *TextGrid) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *TextGrid) getAttrUI(name string) (object.Object, bool) {
	// Create TextGridRow from string
	return textGridAttrs.GetAttr(obj, name)
}
func init() {
	objattrs.Add(textGridAttrs)
	textGridAttrs.Define("Text").Doc("The text.").Getter(func(obj *TextGrid) object.Object {
		return object.NewString(obj.instance.Text())

	})
	textGridAttrs.Define("TabWidth").Getter(func(obj *TextGrid) object.
		Object {
		return object.
			NewInt(int64(obj.instance.TabWidth))
	})
	textGridAttrs.Define("ShowLineNumbers").Getter(func(obj *TextGrid) object.Object {
		return object.NewBool(obj.instance.ShowLineNumbers)
	})
	textGridAttrs.Define("ShowWhitespace").Getter(func(obj *TextGrid) object.Object {
		return object.NewBool(obj.instance.ShowWhitespace)
	},
	)
	textGridAttrs.Define("SetText").Doc(
		"Set the text.",
	).Args("arg1").Returns("nil").Impl(func(obj *TextGrid, ctx context.Context, args ...object.Object) (object.Object, error) {
		text, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}
		guithread.Do(
			func() {
				obj.instance.SetText(text)
			})
		return object.
			Nil, nil
	})
	textGridAttrs.Define("SetRow").Args(
		"arg1",

		"arg2").Returns("nil").Impl(func(obj *TextGrid, ctx context.Context, args ...object.Object) (object.Object, error) {
		row, err := object.AsInt(args[0])
		if err != nil {
			return nil, err

		}
		text, err := object.AsString(args[1])
		if err != nil {
			return nil, err
		}
		guithread.
			Do(func() {
				runes := []rune(text)
				cells := make([]widget.TextGridCell, len(runes))
				for i, r := range runes {
					cells[i] = widget.TextGridCell{Rune: r}
				}
				obj.instance.SetRow(int(row), widget.TextGridRow{Cells: cells})
			})
		return object.Nil, nil
	})
	textGridAttrs.Define("Row").Args("arg1").Returns("nil").Impl(func(obj *TextGrid, ctx context.
		Context, args ...object.Object) (object.
		Object, error) {
		row, err := object.AsInt(args[0])
		if err != nil {
			return nil, err
		}
		return object.
			NewString(obj.instance.RowText(int(row))), nil
	})
	textGridAttrs.Define("RowCount").Args().Returns("nil").Impl(func(obj *TextGrid, ctx context.
		Context, args ...object.Object) (object.Object, error) {
		return object.NewInt(int64(len(obj.
			instance.Rows))), nil
	})
}
