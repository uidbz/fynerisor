package widget

import (
	"context"
	"errors"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

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

func (obj *Calendar) Inspect() string {
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
	err := errors.New("eval error: unsupported operation for " + string(CalendarType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", CalendarType, opType)
	return errObj, err
}

func (obj *Calendar) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Calendar) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Calendar) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", CalendarType, name)
}

func (obj *Calendar) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "OnSelected":
		return object.NewBuiltin("widget.Calendar.OnSelected", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}

			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}

			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("calendar.OnSelected: unable to get call function"), nil
			}

			obj.instance.OnChanged = func(t time.Time) {
				obj.w.Do(func() {
					timeObj := timemodule.NewTimeObject(t)
					callFunc(ctx, fn, []object.Object{timeObj})
				})
			}

			return object.Nil, nil
		}), true
	case "Hide":
		return object.NewBuiltin("widget.Calendar.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("widget.Calendar.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Show()
			})
			return object.Nil, nil
		}), true
	}

	return nil, false
}
