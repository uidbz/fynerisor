package risorcanvas

import (
	"errors"
	"fmt"

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

func (obj *Line) Inspect() string {
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
	err := errors.New("eval error: unsupported operation for " + string(LineType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", LineType, opType)
	return errObj, err
}

func (obj *Line) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Line) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Line) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", LineType, name)
}

func (g *Line) GetAttr(name string) (object.Object, bool) {
	return nil, false
}

func NewLine(instance fyne.CanvasObject) *Line {
	return &Line{instance: instance}
}
