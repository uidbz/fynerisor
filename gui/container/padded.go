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

var _ object.Object = &Padded{}

const PaddedType object.Type = "container.padded"

type Padded struct {
	instance fyne.CanvasObject
}

func (obj *Padded) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Padded) Type() object.Type {
	return PaddedType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Padded) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Padded) inspectUI() string {
	return "container.padded"
}

func (obj *Padded) Interface() interface{} {
	return obj.instance
}

func (obj *Padded) IsTruthy() bool {
	return true
}

func (obj *Padded) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", PaddedType, opType)
}

func (obj *Padded) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Padded) Attrs() []object.AttrSpec {
	return paddedAttrs.Specs()
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Padded) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Padded) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", PaddedType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Padded) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var paddedAttrs = object.NewAttrRegistry[*Padded](string(PaddedType))

func (obj *Padded) getAttrUI(name string) (object.Object, bool) {
	return paddedAttrs.GetAttr(obj, name)

}

func NewPadded(objects ...fyne.CanvasObject) *Padded {
	return &Padded{
		instance: container.NewPadded(objects...),
	}
}
func init() {
	objattrs.ContainerOps(paddedAttrs)
}
