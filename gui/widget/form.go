package widget

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/fynerisor/gui/guithread"

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

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Form) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Form) inspectUI() string {
	return "widget.Form"
}

func (obj *Form) Interface() interface{} {
	return obj.instance
}

func (obj *Form) IsTruthy() bool {
	return true
}

func (obj *Form) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", FormType, opType)
}

func (obj *Form) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Form) Attrs() []object.AttrSpec {
	return nil
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Form) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Form) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", FormType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Form) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *Form) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "Hide":
		return object.NewBuiltin("widget.Form.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("widget.Form.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
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
