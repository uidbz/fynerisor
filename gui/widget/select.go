package widget

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &Select{}

const SelectType object.Type = "widget.Select"

type Select struct {
	instance *widget.Select
}

func (obj *Select) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Select) Type() object.Type {
	return SelectType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Select) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Select) inspectUI() string {
	return "widget.Select"
}

func (obj *Select) Interface() interface{} {
	return obj.instance
}

func (obj *Select) IsTruthy() bool {
	return true
}

func (obj *Select) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", SelectType, opType)
}

func (obj *Select) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Select) Attrs() []object.AttrSpec {
	return selectAttrs.

		// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
		Specs()
}

func (obj *Select) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Select) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", SelectType, name)
}

var selectAttrs = object.NewAttrRegistry[*Select](string(SelectType))

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Select) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *Select) getAttrUI(name string) (object.Object, bool) {
	return selectAttrs.GetAttr(obj, name)

}

func NewSelect(options []string, changed func(string)) *Select {
	return &Select{instance: widget.NewSelect(options, changed)}
}
func init() {
	objattrs.Add(selectAttrs)
	selectAttrs.Define("SetSelectedIndex").Args(
		"arg1").Returns("nil").Impl(func(obj *Select,
		ctx context.Context, args ...object.Object) (object.Object, error) {
		index, err := object.AsInt(args[0])
		if err != nil {
			return nil, err

		}
		obj.instance.SetSelectedIndex(
			int(index))
		return object.Nil, nil
	})
	selectAttrs.Define("SetOptions").Args("arg1").Returns("nil").Impl(func(obj *Select, ctx context.
		Context, args ...object.Object) (object.
		Object, error) {
		options, err := object.AsStringSlice(args[0])
		if err != nil {
			return nil,

				err
		}
		obj.instance.SetOptions(options)
		return object.Nil, nil
	})
	selectAttrs.Define("SetSelected").Args("arg1").Returns("nil").Impl(func(obj *Select, ctx context.Context, args ...object.Object) (object.
		Object, error) {
		selected, err := object.AsString(args[0])
		if err != nil {
			return nil,
				err

		}
		obj.instance.SetSelected(selected)
		return object.Nil, nil
	})
	selectAttrs.Define("Selected").Doc("The selected index.").Getter(func(obj *Select) object.
		Object {
		return object.NewString(obj.instance.
			Selected)
	})
	selectAttrs.Define("SelectedIndex").
		Getter(
			func(obj *Select) object.
				Object {
				return object.
					NewInt(int64(obj.instance.SelectedIndex()))
			})
}
