package container

import (
	"errors"
	"fmt"

	"fyne.io/fyne/v2/container"

	"fyne.io/fyne/v2"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &Scroll{}

const ScrollType object.Type = "container.scroll"

type Scroll struct {
	instance *container.Scroll
}

func (obj *Scroll) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Scroll) Type() object.Type {
	return ScrollType
}

func (obj *Scroll) Inspect() string {
	return "container.scroll"
}

func (obj *Scroll) Interface() interface{} {
	return obj.instance
}

func (obj *Scroll) IsTruthy() bool {
	return true
}

func (obj *Scroll) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'container.scroll'")
}

func (obj *Scroll) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(ScrollType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", ScrollType, opType)
	return errObj, err
}

func (obj *Scroll) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Scroll) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Scroll) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", ScrollType, name)
}

func (obj *Scroll) GetAttr(name string) (object.Object, bool) {
	return nil, false
}

func NewScroll(content fyne.CanvasObject) *Scroll {
	return &Scroll{
		instance: container.NewScroll(content),
	}
}
