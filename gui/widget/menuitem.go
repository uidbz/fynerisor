package widget

import (
	"context"
	"errors"
	"fmt"

	"fyne.io/fyne/v2"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &MenuItem{}

const MenuItemType object.Type = "fyne.MenuItem"

// MenuItem wraps fyne's MenuItem - a single item in a menu.
//
// MenuItem represents one entry in a menu with a label and optional action callback.
//
// Example usage in Risor:
//
//	let item1 = fyne.NewMenuItem("Open", () => { print("Open clicked") })
//	let item2 = fyne.NewMenuItem("Save", () => { print("Save clicked") })
//	let separator = fyne.NewMenuItemSeparator()
//
// Properties:
//   - Label: Item text
//   - Disabled: Whether item is disabled
//   - Checked: Whether item shows checkmark
type MenuItem struct {
	instance *fyne.MenuItem
	w        WindowInterface
}

func NewMenuItem(w WindowInterface) func(ctx context.Context, args ...object.Object) (object.Object, error) {
	return func(ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) != 2 {
			return object.Errorf("wrong number of arguments. got=%d, want=2", len(args)), nil
		}

		label, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}

		fn, ok := args[1].(*object.Closure)
		if !ok {
			return object.Errorf("argument error: expected function, got %s", args[1].Type()), nil
		}

		callFunc, ok := object.GetCallFunc(ctx)
		if !ok {
			return object.Errorf("menuitem: unable to get call function"), nil
		}

		instance := fyne.NewMenuItem(label, func() {
			w.Do(func() {
				_, err := callFunc(ctx, fn, []object.Object{})
				if err != nil {
					fmt.Println("MenuItem action error:", err)
				}
			})
		})

		return &MenuItem{
			instance: instance,
			w:        w,
		}, nil
	}
}

func NewMenuItemSeparator(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 0 {
		return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
	}

	instance := fyne.NewMenuItemSeparator()

	return &MenuItem{
		instance: instance,
	}, nil
}

func (obj *MenuItem) Instance() *fyne.MenuItem {
	return obj.instance
}

func (obj *MenuItem) Type() object.Type {
	return MenuItemType
}

func (obj *MenuItem) Inspect() string {
	return fmt.Sprintf("fyne.MenuItem(%s)", obj.instance.Label)
}

func (obj *MenuItem) Interface() interface{} {
	return obj.instance
}

func (obj *MenuItem) IsTruthy() bool {
	return true
}

func (obj *MenuItem) Cost() int {
	return 0
}

func (obj *MenuItem) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'fyne.MenuItem'")
}

func (obj *MenuItem) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(MenuItemType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", MenuItemType, opType)
	return errObj, err
}

func (obj *MenuItem) Equals(other object.Object) bool {
	return obj == other
}

func (obj *MenuItem) Attrs() []object.AttrSpec {
	return nil
}

func (obj *MenuItem) SetAttr(name string, value object.Object) error {
	switch name {
	case "Label":
		s, err := object.AsString(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		obj.instance.Label = s
		return nil
	case "Disabled":
		b, err := object.AsBool(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		obj.instance.Disabled = b
		return nil
	case "Checked":
		b, err := object.AsBool(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		obj.instance.Checked = b
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", MenuItemType, name)
}

func (obj *MenuItem) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Label":
		return object.NewString(obj.instance.Label), true
	case "Disabled":
		return object.NewBool(obj.instance.Disabled), true
	case "Checked":
		return object.NewBool(obj.instance.Checked), true
	case "IsSeparator":
		return object.NewBool(obj.instance.IsSeparator), true
	}
	return nil, false
}
