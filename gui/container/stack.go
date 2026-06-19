package container

import (
	"context"
	"errors"
	"fmt"

	"fyne.io/fyne/v2/container"

	"fyne.io/fyne/v2"
	"github.com/uidbz/fynerisor/gui/guithread"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &Stack{}

const StackType object.Type = "container.stack"

type Stack struct {
	instance fyne.CanvasObject
}

func (obj *Stack) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Stack) Type() object.Type {
	return StackType
}

func (obj *Stack) Inspect() string {
	return "container.stack"
}

func (obj *Stack) Interface() interface{} {
	return obj.instance
}

func (obj *Stack) IsTruthy() bool {
	return true
}

func (obj *Stack) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'container.stack'")
}

func (obj *Stack) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(StackType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", StackType, opType)
	return errObj, err
}

func (obj *Stack) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Stack) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Stack) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", StackType, name)
}

func (obj *Stack) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Hide":
		return object.NewBuiltin("container.Stack.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("container.Stack.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
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

func NewStack(objects ...fyne.CanvasObject) *Stack {
	return &Stack{
		instance: container.NewStack(objects...),
	}
}
