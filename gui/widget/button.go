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

var _ object.Object = &Button{}

const ButtonType object.Type = "widget.Button"

type Button struct {
	instance *widget.Button
}

func (obj *Button) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Button) Type() object.Type {
	return ButtonType
}

func (obj *Button) Inspect() string {
	return "widget.Button"
}

func (obj *Button) Interface() interface{} {
	return obj.instance
}

func (obj *Button) IsTruthy() bool {
	return true
}

func (obj *Button) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'widget.Button'")
}

func (obj *Button) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(ButtonType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", ButtonType, opType)
	return errObj, err
}

func (obj *Button) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Button) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Button) SetAttr(name string, value object.Object) error {
	switch name {
	case "Text":
		text, err := object.AsString(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
			obj.instance.SetText(text)
		})
		return nil
	case "Importance":
		// widget.Importance values: High=1, Medium=2, Low=3, DangerHigh=4, WarningHigh=5, SuccessHigh=6
		i, err := object.AsInt(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
			obj.instance.Importance = widget.Importance(i)
			obj.instance.Refresh()
		})
		return nil
	case "Disabled":
		b, err := object.AsBool(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
			if b {
				obj.instance.Disable()
			} else {
				obj.instance.Enable()
			}
		})
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", ButtonType, name)
}

func (obj *Button) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Text":
		return object.NewString(obj.instance.Text), true
	case "Importance":
		return object.NewInt(int64(obj.instance.Importance)), true
	case "Disabled":
		return object.NewBool(obj.instance.Disabled()), true
	case "SetText":
		return object.NewBuiltin("widget.Button.SetText", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			text, err := object.AsString(args[0])
			if err != nil {
				return object.Errorf("type error: %v", err), nil
			}
			guithread.Do(func() {
				obj.instance.SetText(text)
			})
			return object.Nil, nil
		}), true
	case "Disable":
		return object.NewBuiltin("widget.Button.Disable", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Disable()
			})
			return object.Nil, nil
		}), true
	case "Enable":
		return object.NewBuiltin("widget.Button.Enable", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Enable()
			})
			return object.Nil, nil
		}), true
	case "Hide":
		return object.NewBuiltin("widget.Button.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("widget.Button.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Show()
			})
			return object.Nil, nil
		}), true
	case "OnTapped":
		return object.NewBuiltin("widget.Button.OnTapped", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("button: unable to get call function"), nil
			}

			// Set the button's OnTapped callback
			guithread.Do(func() {
				obj.instance.OnTapped = func() {
					_, err := callFunc(ctx, fn, []object.Object{})
					if err != nil {
						fmt.Println("Button.OnTapped ERROR:", err)
					}
				}
			})
			return object.Nil, nil
		}), true
	}
	return nil, false
}

func NewButton(label string, tapped func()) *Button {
	return &Button{instance: widget.NewButton(label, tapped)}
}
