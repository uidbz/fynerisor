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

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Center) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Center) inspectUI() string {
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

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Center) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Center) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", CenterType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Center) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *Center) getAttrUI(name string) (object.Object, bool) {
	if o, ok := containerOps("Center", obj.instance, name); ok {
		return o, true
	}
	switch name {
	case "Hide":
		return object.NewBuiltin("container.Center.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("container.Center.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
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

func NewCenter(objects ...fyne.CanvasObject) *Center {
	return &Center{
		instance: container.NewCenter(objects...),
	}
}
