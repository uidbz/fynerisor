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

var _ object.Object = &Activity{}

const ActivityType object.Type = "widget.Activity"

// Activity wraps widget.Activity for Risor scripting
type Activity struct {
	instance *widget.Activity
}

// NewActivity creates a new activity indicator
func NewActivity() *Activity {
	return &Activity{instance: widget.NewActivity()}
}

func (obj *Activity) Type() object.Type {
	return ActivityType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Activity) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Activity) inspectUI() string {
	return "widget.Activity()"
}

func (obj *Activity) Interface() interface{} {
	return obj.instance
}

func (obj *Activity) IsTruthy() bool {
	return true
}

func (obj *Activity) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Activity) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", ActivityType, opType)
}

func (obj *Activity) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Activity) Attrs() []object.AttrSpec {
	return activityAttrs.Specs()
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
var activityAttrs = objattrs.Register(string(ActivityType), object.NewAttrRegistry[*Activity](string(ActivityType)))

func (obj *Activity) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *Activity) getAttrUI(name string) (object.Object, bool) {
	return activityAttrs.GetAttr(obj, name)

}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Activity) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Activity) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", ActivityType, name)
}
func init() {
	objattrs.Add(activityAttrs)
	activityAttrs.Define("Start").Args().
		Returns("nil").Impl(func(obj *Activity, ctx context.
		Context, args ...object.Object) (object.Object, error) {
		guithread.Do(func() {
			obj.instance.
				Start()
		})
		return object.Nil, nil
	})
	activityAttrs.Define("Stop").Args().
		Returns(
			"nil").Impl(func(obj *Activity, ctx context.Context, args ...object.Object) (object.Object, error) {
		guithread.Do(func() {
			obj.instance.Stop()
		})
		return object.Nil, nil
	})
}
