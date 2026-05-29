package container

import (
	"errors"
	"fmt"

	"fyne.io/fyne/v2/container"

	"fyne.io/fyne/v2"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &Border{}

const BorderType object.Type = "container.border"

type Border struct {
	instance fyne.CanvasObject
}

func (obj *Border) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Border) Type() object.Type {
	return BorderType
}

func (obj *Border) Inspect() string {
	return "container.border"
}

func (obj *Border) Interface() interface{} {
	return obj.instance
}

func (obj *Border) IsTruthy() bool {
	return true
}

func (obj *Border) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'container.border'")
}

func (obj *Border) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(BorderType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", BorderType, opType)
	return errObj, err
}

func (obj *Border) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Border) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Border) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", BorderType, name)
}

func (obj *Border) GetAttr(name string) (object.Object, bool) {
	// switch name {
	// case "NewBorderLayout":
	// 	return object.NewBuiltin("layout.NewBorderLayout", func(ctx context.Context, args ...object.Object) object.Object {
	// 		if len(args) != 2 {
	// 			return object.Errorf("wrong number of arguments. got=%d, want=2", len(args))
	// 		}

	// 		borderlayout := layout.NewBorderLayout(args[0].Interface().(fyne.CanvasObject), args[1].Interface().(fyne.CanvasObject), nil, nil)

	// 		return NewFyneBorder(borderlayout)
	// 	}), true
	// }
	return nil, false
}

func NewBorder(top, bottom, left, right, center fyne.CanvasObject) *Border {
	var border *fyne.Container
	if center != nil {
		border = container.NewBorder(top, bottom, left, right, center)
	} else {
		border = container.NewBorder(top, bottom, left, right)
	}
	return &Border{instance: border}
}
