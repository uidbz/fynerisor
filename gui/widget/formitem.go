package widget

import (
	"fmt"
	"github.com/uidbz/fynerisor/gui/guithread"

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

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *FormItem) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *FormItem) inspectUI() string {
	return "widget.FormItem"
}

func (obj *FormItem) Interface() interface{} {
	return obj.instance
}

func (obj *FormItem) IsTruthy() bool {
	return true
}

func (obj *FormItem) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", FormItemType, opType)
}

func (obj *FormItem) Equals(other object.Object) bool {
	return obj == other
}

func (obj *FormItem) Attrs() []object.AttrSpec {
	return nil
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *FormItem) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *FormItem) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", FormItemType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *FormItem) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *FormItem) getAttrUI(name string) (object.Object, bool) {
	return nil, false
}

func NewFormItem(label string, item fyne.CanvasObject) *FormItem {
	return &FormItem{instance: widget.NewFormItem(label, item)}
}
