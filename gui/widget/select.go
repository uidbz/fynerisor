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

var _ object.Object = &Select{}

const SelectType object.Type = "widget.Select"

type Select struct {
	instance *widget.Select
}

func (obj *Select) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Select) Type() object.Type {
	return SelectType
}

func (obj *Select) Inspect() string {
	return "widget.Select"
}

func (obj *Select) Interface() interface{} {
	return obj.instance
}

func (obj *Select) IsTruthy() bool {
	return true
}

func (obj *Select) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'widget.Select'")
}

func (obj *Select) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(SelectType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", SelectType, opType)
	return errObj, err
}

func (obj *Select) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Select) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Select) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", SelectType, name)
}

func (obj *Select) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "SetSelectedIndex":
		return object.NewBuiltin("widget.Select.SetSelectedIndex", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=exactly 1", len(args)), nil
			}

			index, err := object.AsInt(args[0])
			if err != nil {
				return nil, err
			}

			obj.instance.SetSelectedIndex(int(index))

			return object.Nil, nil
		}), true
	case "SetOptions":
		return object.NewBuiltin("widget.Select.SetOptions", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=exactly 1", len(args)), nil
			}

			options, err := object.AsStringSlice(args[0])
			if err != nil {
				return nil, err
			}

			obj.instance.SetOptions(options)

			return object.Nil, nil
		}), true

	case "SetSelected":
		return object.NewBuiltin("widget.Select.SetSelected", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=exactly 1", len(args)), nil
			}

			selected, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}

			obj.instance.SetSelected(selected)

			return object.Nil, nil
		}), true
	case "Selected":
		return object.NewString(obj.instance.Selected), true
	case "SelectedIndex":
		return object.NewInt(int64(obj.instance.SelectedIndex())), true
	case "Hide":
		return object.NewBuiltin("widget.Select.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("widget.Select.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
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

func NewSelect(options []string, changed func(string)) *Select {
	return &Select{instance: widget.NewSelect(options, changed)}
}
