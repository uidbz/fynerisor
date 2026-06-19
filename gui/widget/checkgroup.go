package widget

import (
	"context"
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"github.com/uidbz/fynerisor/gui/guithread"
	"fyne.io/fyne/v2/widget"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &CheckGroup{}

const CheckGroupType object.Type = "widget.CheckGroup"

type CheckGroup struct {
	instance *widget.CheckGroup
}

func (obj *CheckGroup) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *CheckGroup) Type() object.Type {
	return CheckGroupType
}

func (obj *CheckGroup) Inspect() string {
	return "widget.CheckGroup"
}

func (obj *CheckGroup) Interface() interface{} {
	return obj.instance
}

func (obj *CheckGroup) IsTruthy() bool {
	return true
}

func (obj *CheckGroup) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'widget.CheckGroup'")
}

func (obj *CheckGroup) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(CheckGroupType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", CheckGroupType, opType)
	return errObj, err
}

func (obj *CheckGroup) Equals(other object.Object) bool {
	return obj == other
}

func (obj *CheckGroup) Attrs() []object.AttrSpec {
	return nil
}

func (obj *CheckGroup) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", CheckGroupType, name)
}

func (obj *CheckGroup) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "SetSelected":
		return object.NewBuiltin("widget.CheckGroup.SetSelected", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=exactly 1", len(args)), nil
			}

			for _, x := range args {
				o, err := object.AsStringSlice(x)
				if err != nil {
					return nil, err
				}
				guithread.Do(func() {
					obj.instance.SetSelected(o)
				})
			}

			return object.Nil, nil
		}), true
	case "Hide":
		return object.NewBuiltin("widget.CheckGroup.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("widget.CheckGroup.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
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

func NewCheckGroup(options []string, checked func([]string)) *CheckGroup {
	return &CheckGroup{instance: widget.NewCheckGroup(options, checked)}
}
