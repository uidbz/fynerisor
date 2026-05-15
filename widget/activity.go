package widget

import (
	"context"
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

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

func (obj *Activity) Inspect() string {
	return "widget.Activity()"
}

func (obj *Activity) Interface() interface{} {
	return obj.instance
}

func (obj *Activity) IsTruthy() bool {
	return true
}

func (obj *Activity) Cost() int {
	return 0
}

func (obj *Activity) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Activity) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'widget.Activity'")
}

func (obj *Activity) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(ActivityType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", ActivityType, opType)
	return errObj, err
}

func (obj *Activity) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Activity) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Activity) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Start":
		return object.NewBuiltin("widget.Activity.Start", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Start()
			})
			return object.Nil, nil
		}), true
	case "Stop":
		return object.NewBuiltin("widget.Activity.Stop", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Stop()
			})
			return object.Nil, nil
		}), true
	}
	return nil, false
}

func (obj *Activity) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", ActivityType, name)
}
