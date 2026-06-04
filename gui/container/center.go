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

var _ object.Object = &Center{}

const CenterType object.Type = "container.center"

type Center struct {
	instance fyne.CanvasObject
}

func (obj *Center) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Center) Type() object.Type {
	return CenterType
}

func (obj *Center) Inspect() string {
	return "container.center"
}

func (obj *Center) Interface() interface{} {
	return obj.instance
}

func (obj *Center) IsTruthy() bool {
	return true
}

func (obj *Center) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'container.center'")
}

func (obj *Center) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(CenterType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", CenterType, opType)
	return errObj, err
}

func (obj *Center) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Center) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Center) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", CenterType, name)
}

func (obj *Center) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Hide":
		return object.NewBuiltin("container.Center.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("container.Center.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
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

func NewCenter(objects ...fyne.CanvasObject) *Center {
	return &Center{
		instance: container.NewCenter(objects...),
	}
}
