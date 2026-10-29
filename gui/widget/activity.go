package widget

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/fynerisor/gui/guithread"

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
	return nil
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Activity) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *Activity) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "Start":
		return object.NewBuiltin("widget.Activity.Start", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			guithread.Do(func() {
				obj.instance.Start()
			})
			return object.Nil, nil
		}), true
	case "Stop":
		return object.NewBuiltin("widget.Activity.Stop", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			guithread.Do(func() {
				obj.instance.Stop()
			})
			return object.Nil, nil
		}), true
	case "Hide":
		return object.NewBuiltin("widget.Activity.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("widget.Activity.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			guithread.Do(func() {
				obj.instance.Show()
			})
			return object.Nil, nil
		}), true
	}
	return nil, false
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Activity) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Activity) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", ActivityType, name)
}
