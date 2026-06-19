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

var _ object.Object = &Form{}

const FormType object.Type = "widget.Form"

type Form struct {
	instance *widget.Form
}

func (obj *Form) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Form) Type() object.Type {
	return FormType
}

func (obj *Form) Inspect() string {
	return "widget.Form"
}

func (obj *Form) Interface() interface{} {
	return obj.instance
}

func (obj *Form) IsTruthy() bool {
	return true
}

func (obj *Form) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'widget.Form'")
}

func (obj *Form) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(FormType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", FormType, opType)
	return errObj, err
}

func (obj *Form) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Form) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Form) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", FormType, name)
}

func (obj *Form) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Hide":
		return object.NewBuiltin("widget.Form.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("widget.Form.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
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

func NewForm(items ...*widget.FormItem) *Form {
	return &Form{instance: widget.NewForm(items...)}
}
