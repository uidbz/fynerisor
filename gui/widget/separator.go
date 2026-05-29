package widget

import (
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &Separator{}

const SeparatorType object.Type = "widget.Separator"

// Separator wraps widget.Separator for Risor scripting
type Separator struct {
	instance *widget.Separator
}

// NewSeparator creates a new horizontal separator
func NewSeparator() *Separator {
	return &Separator{instance: widget.NewSeparator()}
}

func (obj *Separator) Type() object.Type {
	return SeparatorType
}

func (obj *Separator) Inspect() string {
	return "widget.Separator()"
}

func (obj *Separator) Interface() interface{} {
	return obj.instance
}

func (obj *Separator) IsTruthy() bool {
	return true
}

func (obj *Separator) Cost() int {
	return 0
}

func (obj *Separator) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Separator) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'widget.Separator'")
}

func (obj *Separator) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(SeparatorType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", SeparatorType, opType)
	return errObj, err
}

func (obj *Separator) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Separator) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Separator) GetAttr(name string) (object.Object, bool) {
	return nil, false
}

func (obj *Separator) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", SeparatorType, name)
}
