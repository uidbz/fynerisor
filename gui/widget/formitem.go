package widget

import (
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &FormItem{}

const FormItemType object.Type = "widget.FormItem"

type FormItem struct {
	instance *widget.FormItem
}

func (obj *FormItem) Type() object.Type {
	return FormItemType
}

func (obj *FormItem) Inspect() string {
	return "widget.FormItem"
}

func (obj *FormItem) Interface() interface{} {
	return obj.instance
}

func (obj *FormItem) IsTruthy() bool {
	return true
}

func (obj *FormItem) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'widget.FormItem'")
}

func (obj *FormItem) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(FormItemType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", FormItemType, opType)
	return errObj, err
}

func (obj *FormItem) Equals(other object.Object) bool {
	return obj == other
}

func (obj *FormItem) Attrs() []object.AttrSpec {
	return nil
}

func (obj *FormItem) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", FormItemType, name)
}

func (obj *FormItem) GetAttr(name string) (object.Object, bool) {
	return nil, false
}

func NewFormItem(label string, item fyne.CanvasObject) *FormItem {
	return &FormItem{instance: widget.NewFormItem(label, item)}
}
