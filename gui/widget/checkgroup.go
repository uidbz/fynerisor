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

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *CheckGroup) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *CheckGroup) inspectUI() string {
	return "widget.CheckGroup"
}

func (obj *CheckGroup) Interface() interface{} {
	return obj.instance
}

func (obj *CheckGroup) IsTruthy() bool {
	return true
}

func (obj *CheckGroup) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", CheckGroupType, opType)
}

func (obj *CheckGroup) Equals(other object.Object) bool {
	return obj == other
}

func (obj *CheckGroup) Attrs() []object.AttrSpec {
	return checkGroupAttrs.Specs()
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *CheckGroup) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *CheckGroup) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", CheckGroupType, name)
}

var checkGroupAttrs = objattrs.Register(string(CheckGroupType), object.NewAttrRegistry[*CheckGroup](string(CheckGroupType)))

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *CheckGroup) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *CheckGroup) getAttrUI(name string) (object.Object, bool) {
	return checkGroupAttrs.GetAttr(obj, name)

}

func NewCheckGroup(options []string, checked func([]string)) *CheckGroup {
	return &CheckGroup{instance: widget.NewCheckGroup(options, checked)}
}
func init() {
	objattrs.Add(checkGroupAttrs)
	checkGroupAttrs.
		Define("SetSelected").Args(
		"arg1",
	).Returns("nil").Impl(func(obj *CheckGroup, ctx context.Context, args ...object.Object) (object.Object, error) {
		for _, x := range args {
			o, err := object.AsStringSlice(x)
			if err !=
				nil {
				return nil, err
			}
			guithread.
				Do(func() { obj.instance.SetSelected(o) })
		}
		return object.Nil, nil
	})
}
