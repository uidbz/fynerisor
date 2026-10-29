package risorcanvas

import (
	"fmt"
	"github.com/uidbz/fynerisor/gui/guithread"

	"fyne.io/fyne/v2"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &Line{}

const LineType object.Type = "canvas.line"

type Line struct {
	instance fyne.CanvasObject
}

func (obj *Line) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Line) Type() object.Type {
	return LineType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Line) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Line) inspectUI() string {
	return "canvas.line"
}

func (obj *Line) Interface() interface{} {
	return obj.instance
}

func (obj *Line) IsTruthy() bool {
	return true
}

func (obj *Line) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'canvas.line'")
}

func (obj *Line) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", LineType, opType)
}

func (obj *Line) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Line) Attrs() []object.AttrSpec {
	return nil
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Line) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Line) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", LineType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (g *Line) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, g.getAttrUI)
}

func (g *Line) getAttrUI(name string) (object.Object, bool) {
	return nil, false
}

func NewLine(instance fyne.CanvasObject) *Line {
	return &Line{instance: instance}
}
