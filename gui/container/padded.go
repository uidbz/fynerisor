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

var _ object.Object = &Padded{}

const PaddedType object.Type = "container.padded"

type Padded struct {
	instance fyne.CanvasObject
}

func (obj *Padded) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Padded) Type() object.Type {
	return PaddedType
}

func (obj *Padded) Inspect() string {
	return "container.padded"
}

func (obj *Padded) Interface() interface{} {
	return obj.instance
}

func (obj *Padded) IsTruthy() bool {
	return true
}

func (obj *Padded) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'container.padded'")
}

func (obj *Padded) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(PaddedType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", PaddedType, opType)
	return errObj, err
}

func (obj *Padded) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Padded) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Padded) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", PaddedType, name)
}

func (obj *Padded) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Hide":
		return object.NewBuiltin("container.Padded.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("container.Padded.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
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

func NewPadded(objects ...fyne.CanvasObject) *Padded {
	return &Padded{
		instance: container.NewPadded(objects...),
	}
}
