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

var _ object.Object = &VBox{}

const VBoxType object.Type = "container.vbox"

type VBox struct {
	instance fyne.CanvasObject
}

func (obj *VBox) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *VBox) Type() object.Type {
	return VBoxType
}

func (obj *VBox) Inspect() string {
	return "container.vbox"
}

func (obj *VBox) Interface() interface{} {
	return obj.instance
}

func (obj *VBox) IsTruthy() bool {
	return true
}

func (obj *VBox) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'container.vbox'")
}

func (obj *VBox) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(VBoxType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", VBoxType, opType)
	return errObj, err
}

func (obj *VBox) Equals(other object.Object) bool {
	return obj == other
}

func (obj *VBox) Attrs() []object.AttrSpec {
	return nil
}

func (obj *VBox) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", VBoxType, name)
}

func (obj *VBox) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Hide":
		return object.NewBuiltin("container.VBox.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("container.VBox.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
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

func NewVBox(objects ...fyne.CanvasObject) *VBox {
	return &VBox{
		instance: container.NewVBox(objects...),
	}
}
