package widget

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/vmexec"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"

	timemodule "github.com/uidbz/fynerisor/modules/time"
)

var _ object.Object = &Calendar{}

const CalendarType object.Type = "widget.Calendar"

// Calendar wraps fyne's Calendar widget
type Calendar struct {
	instance *widget.Calendar
	w        WindowInterface
}

func NewCalendar(w WindowInterface, startDate time.Time, onChanged func(time.Time)) *Calendar {
	// If no callback provided, use a no-op function to avoid nil panics in Fyne
	if onChanged == nil {
		onChanged = func(time.Time) {}
	}

	instance := widget.NewCalendar(startDate, onChanged)

	return &Calendar{
		instance: instance,
		w:        w,
	}
}

func (obj *Calendar) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Calendar) Type() object.Type {
	return CalendarType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Calendar) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Calendar) inspectUI() string {
	return "widget.Calendar"
}

func (obj *Calendar) Interface() interface{} {
	return obj.instance
}

func (obj *Calendar) IsTruthy() bool {
	return true
}

func (obj *Calendar) Cost() int {
	return 0
}

func (obj *Calendar) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'widget.Calendar'")
}

func (obj *Calendar) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", CalendarType, opType)
}

func (obj *Calendar) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Calendar) Attrs() []object.AttrSpec {
	return nil
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Calendar) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Calendar) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", CalendarType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Calendar) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *Calendar) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "OnSelected":
		return object.NewBuiltin("widget.Calendar.OnSelected", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}

			cb, err := vmexec.NewCallback(ctx, "widget.Calendar.OnSelected", args[0])
			if err != nil {
				return nil, err
			}

			// Event: queued on the script's VM goroutine.
			obj.instance.OnChanged = func(t time.Time) {
				cb.Fire(timemodule.NewTimeObject(t))
			}

			return object.Nil, nil
		}), true
	case "Hide":
		return object.NewBuiltin("widget.Calendar.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("widget.Calendar.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
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
