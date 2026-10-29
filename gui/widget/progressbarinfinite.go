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

var _ object.Object = &ProgressBarInfinite{}

const ProgressBarInfiniteType object.Type = "widget.ProgressBarInfinite"

// ProgressBarInfinite wraps widget.ProgressBarInfinite for Risor scripting
type ProgressBarInfinite struct {
	instance *widget.ProgressBarInfinite
}

// NewProgressBarInfinite creates a new infinite progress bar. Fyne's
// ProgressBarInfinite force-starts its animation inside CreateRenderer (which
// runs the moment the bar is first shown), so a pre-render Stop() cannot keep
// it idle. We create it hidden instead; a hidden widget never builds its
// renderer, so it stays idle until an explicit Start(). Start()/Stop() below
// show/hide the bar, so it is visible only while spinning.
func NewProgressBarInfinite() *ProgressBarInfinite {
	instance := widget.NewProgressBarInfinite()
	instance.Hide()
	return &ProgressBarInfinite{instance: instance}
}

func (obj *ProgressBarInfinite) Type() object.Type {
	return ProgressBarInfiniteType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *ProgressBarInfinite) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *ProgressBarInfinite) inspectUI() string {
	return "widget.ProgressBarInfinite()"
}

func (obj *ProgressBarInfinite) Interface() interface{} {
	return obj.instance
}

func (obj *ProgressBarInfinite) IsTruthy() bool {
	return true
}

func (obj *ProgressBarInfinite) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *ProgressBarInfinite) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", ProgressBarInfiniteType, opType)
}

func (obj *ProgressBarInfinite) Equals(other object.Object) bool {
	return obj == other
}

func (obj *ProgressBarInfinite) Attrs() []object.AttrSpec {
	return progressBarInfiniteAttrs.

		// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
		Specs()
}

var progressBarInfiniteAttrs = object.NewAttrRegistry[*ProgressBarInfinite](string(ProgressBarInfiniteType))

func (obj *ProgressBarInfinite) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *ProgressBarInfinite) getAttrUI(name string) (object.Object, bool) {
	return progressBarInfiniteAttrs.GetAttr(obj, name)

}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *ProgressBarInfinite) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *ProgressBarInfinite) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", ProgressBarInfiniteType, name)
}
func init() {
	objattrs.Add(progressBarInfiniteAttrs)
	progressBarInfiniteAttrs.
		Define("Start").Args().Returns("nil").
		Impl(func(obj *ProgressBarInfinite, ctx context.Context, args ...object.
			Object) (object.Object, error) {
			guithread.Do(func() {
				obj.instance.Show()
				obj.instance.Start()
			})
			return object.Nil, nil
		})
	progressBarInfiniteAttrs.
		Define("Stop").Args().Returns("nil").Impl(func(obj *ProgressBarInfinite, ctx context.Context, args ...object.
		Object) (object.Object, error) {
		guithread.Do(func() {
			obj.instance.Stop()
			obj.instance.Hide()
		})
		return object.Nil, nil
	})
	progressBarInfiniteAttrs.
		Define("Running").
		Getter(func(obj *ProgressBarInfinite) object.Object {
			return object.NewBool(obj.instance.Running())
		})
}
