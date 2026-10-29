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

var _ object.Object = &VBox{}

const VBoxType object.Type = "container.vbox"

type VBox struct {
	instance fyne.CanvasObject
}

func (obj *VBox) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *VBox) Type() object.Type {
	return VBoxType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *VBox) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *VBox) inspectUI() string {
	return "container.vbox"
}

func (obj *VBox) Interface() interface{} {
	return obj.instance
}

func (obj *VBox) IsTruthy() bool {
	return true
}

func (obj *VBox) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", VBoxType, opType)
}

func (obj *VBox) Equals(other object.Object) bool {
	return obj == other
}

func (obj *VBox) Attrs() []object.AttrSpec {
	return vBoxAttrs.Specs()
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *VBox) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *VBox) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", VBoxType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *VBox) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var vBoxAttrs = object.NewAttrRegistry[*VBox](string(VBoxType))

func (obj *VBox) getAttrUI(name string) (object.Object, bool) {
	return vBoxAttrs.GetAttr(obj, name)

}

func NewVBox(objects ...fyne.CanvasObject) *VBox {
	return &VBox{
		instance: container.NewVBox(objects...),
	}
}
func init() {
	objattrs.ContainerOps(vBoxAttrs)
}
