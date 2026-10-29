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

var _ object.Object = &Center{}

const CenterType object.Type = "container.center"

type Center struct {
	instance fyne.CanvasObject
}

func (obj *Center) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Center) Type() object.Type {
	return CenterType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Center) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Center) inspectUI() string {
	return "container.center"
}

func (obj *Center) Interface() interface{} {
	return obj.instance
}

func (obj *Center) IsTruthy() bool {
	return true
}

func (obj *Center) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", CenterType, opType)
}

func (obj *Center) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Center) Attrs() []object.AttrSpec {
	return centerAttrs.Specs()
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Center) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Center) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", CenterType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Center) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var centerAttrs = object.NewAttrRegistry[*Center](string(CenterType))

func (obj *Center) getAttrUI(name string) (object.Object, bool) {
	return centerAttrs.GetAttr(obj, name)

}

func NewCenter(objects ...fyne.CanvasObject) *Center {
	return &Center{
		instance: container.NewCenter(objects...),
	}
}
func init() {
	objattrs.ContainerOps(centerAttrs)
}
