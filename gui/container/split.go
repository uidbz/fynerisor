package container

import (
	"errors"
	"fmt"

	"fyne.io/fyne/v2/container"

	"fyne.io/fyne/v2"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &Split{}

const SplitType object.Type = "container.Split"

type Split struct {
	instance fyne.CanvasObject
}

func (obj *Split) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Split) Type() object.Type {
	return SplitType
}

func (obj *Split) Inspect() string {
	return "container.Split"
}

func (obj *Split) Interface() interface{} {
	return obj.instance
}

func (obj *Split) IsTruthy() bool {
	return true
}

func (obj *Split) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'container.split'")
}

func (obj *Split) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(SplitType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", SplitType, opType)
	return errObj, err
}

func (obj *Split) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Split) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Split) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", SplitType, name)
}

func (obj *Split) GetAttr(name string) (object.Object, bool) {
	return nil, false
}

func NewHSplit(leading, trailing fyne.CanvasObject) *Split {
	return &Split{
		instance: container.NewHSplit(leading, trailing),
	}
}

func NewVSplit(top, botttom fyne.CanvasObject) *Split {
	return &Split{
		instance: container.NewVSplit(top, botttom),
	}
}
