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

var _ object.Object = &Max{}

const MaxType object.Type = "container.max"

type Max struct {
	instance fyne.CanvasObject
}

func (obj *Max) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Max) Type() object.Type {
	return MaxType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Max) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Max) inspectUI() string {
	return "container.max"
}

func (obj *Max) Interface() interface{} {
	return obj.instance
}

func (obj *Max) IsTruthy() bool {
	return true
}

func (obj *Max) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", MaxType, opType)
}

func (obj *Max) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Max) Attrs() []object.AttrSpec {
	return maxAttrs.Specs()
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Max) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Max) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", MaxType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Max) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var maxAttrs = objattrs.Register(string(MaxType), object.NewAttrRegistry[*Max](string(MaxType)))

func (obj *Max) getAttrUI(name string) (object.Object, bool) {
	return maxAttrs.GetAttr(obj, name)

}

func NewMax(objects ...fyne.CanvasObject) *Max {
	return &Max{
		instance: container.NewMax(objects...),
	}
}
func init() {
	objattrs.ContainerOps(maxAttrs)
}
