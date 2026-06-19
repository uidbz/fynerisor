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

var _ object.Object = &HBox{}

const HBoxType object.Type = "container.hbox"

type HBox struct {
	instance fyne.CanvasObject
}

func (obj *HBox) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *HBox) Type() object.Type {
	return HBoxType
}

func (obj *HBox) Inspect() string {
	return "container.hbox"
}

func (obj *HBox) Interface() interface{} {
	return obj.instance
}

func (obj *HBox) IsTruthy() bool {
	return true
}

func (obj *HBox) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'container.hbox'")
}

func (obj *HBox) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(HBoxType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", HBoxType, opType)
	return errObj, err
}

func (obj *HBox) Equals(other object.Object) bool {
	return obj == other
}

func (obj *HBox) Attrs() []object.AttrSpec {
	return nil
}

func (obj *HBox) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", HBoxType, name)
}

func (obj *HBox) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Hide":
		return object.NewBuiltin("container.HBox.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("container.HBox.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
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

func NewHBox(objects ...fyne.CanvasObject) *HBox {
	return &HBox{
		instance: container.NewHBox(objects...),
	}
}
