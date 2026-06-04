package container

import (
	"context"
	"errors"
	"fmt"

	"fyne.io/fyne/v2/container"

	"fyne.io/fyne/v2"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &Max{}

const MaxType object.Type = "container.max"

type Max struct {
	instance fyne.CanvasObject
}

func (obj *Max) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Max) Type() object.Type {
	return MaxType
}

func (obj *Max) Inspect() string {
	return "container.max"
}

func (obj *Max) Interface() interface{} {
	return obj.instance
}

func (obj *Max) IsTruthy() bool {
	return true
}

func (obj *Max) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'container.max'")
}

func (obj *Max) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(MaxType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", MaxType, opType)
	return errObj, err
}

func (obj *Max) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Max) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Max) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", MaxType, name)
}

func (obj *Max) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Hide":
		return object.NewBuiltin("container.Max.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("container.Max.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Show()
			})
			return object.Nil, nil
		}), true
	}
	return nil, false
}

func NewMax(objects ...fyne.CanvasObject) *Max {
	return &Max{
		instance: container.NewMax(objects...),
	}
}
