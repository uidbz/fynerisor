package widget

import (
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
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
	return nil, false
}

func NewForm(items ...*widget.FormItem) *Form {
	return &Form{instance: widget.NewForm(items...)}
}
