package widget

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"

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

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Separator) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Separator) inspectUI() string {
	return "widget.Separator()"
}

func (obj *Separator) Interface() interface{} {
	return obj.instance
}

func (obj *Separator) IsTruthy() bool {
	return true
}

func (obj *Separator) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Separator) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", SeparatorType, opType)
}

func (obj *Separator) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Separator) Attrs() []object.AttrSpec {
	return separatorAttrs.Specs()
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
var separatorAttrs = object.NewAttrRegistry[*Separator](string(SeparatorType))

func (obj *Separator) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *Separator) getAttrUI(name string) (object.Object, bool) {
	return separatorAttrs.GetAttr(obj, name)

}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Separator) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Separator) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", SeparatorType, name)
}
func init() {
	objattrs.Add(separatorAttrs)
}
