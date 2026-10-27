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

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Split) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Split) inspectUI() string {
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

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Split) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Split) setAttrUI(name string, value object.Object) error {
	switch name {
	case "Offset":
		split, ok := obj.instance.(*container.Split)
		if !ok {
			return fmt.Errorf("attribute error: instance is not a Split container")
		}

		offset, err := object.AsFloat(value)
		if err != nil {
			return err
		}

		guithread.Do(func() {
			split.Offset = float64(offset)
			split.Refresh()
		})
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", SplitType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Split) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *Split) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "Hide":
		return object.NewBuiltin("container.Split.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("container.Split.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Show()
			})
			return object.Nil, nil
		}), true
	case "Offset":
		split, ok := obj.instance.(*container.Split)
		if !ok {
			return object.NewFloat(0.5), true // Default if not a Split
		}
		return object.NewFloat(split.Offset), true
	}
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
