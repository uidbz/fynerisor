package container

import (
	"fmt"

	"fyne.io/fyne/v2/container"

	"fyne.io/fyne/v2"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &Stack{}

const StackType object.Type = "container.stack"

type Stack struct {
	instance fyne.CanvasObject
}

func (obj *Stack) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Stack) Type() object.Type {
	return StackType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Stack) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Stack) inspectUI() string {
	return "container.stack"
}

func (obj *Stack) Interface() interface{} {
	return obj.instance
}

func (obj *Stack) IsTruthy() bool {
	return true
}

func (obj *Stack) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", StackType, opType)
}

func (obj *Stack) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Stack) Attrs() []object.AttrSpec {
	return stackAttrs.Specs()
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Stack) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Stack) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", StackType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Stack) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var stackAttrs = objattrs.Register(string(StackType), object.NewAttrRegistry[*Stack](string(StackType)))

func (obj *Stack) getAttrUI(name string) (object.Object, bool) {
	return stackAttrs.GetAttr(obj, name)

}

func NewStack(objects ...fyne.CanvasObject) *Stack {
	return &Stack{
		instance: container.NewStack(objects...),
	}
}
func init() {
	objattrs.ContainerOps(stackAttrs)
}
