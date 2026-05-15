// Package time provides time and date functions for Risor scripts.
package time

import (
	"context"
	"fmt"
	"time"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

// Module returns the time module for risor
func Module() *object.Module {
	return object.NewBuiltinsModule("time", map[string]object.Object{
		"now":   object.NewBuiltin("time.now", now),
		"date":  object.NewBuiltin("time.date", date),
		"parse": object.NewBuiltin("time.parse", parse),
	})
}

// now returns the current time as a Time object
func now(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 0 {
		return object.Errorf("time.now: expected 0 arguments, got %d", len(args)), nil
	}
	return NewTimeObject(time.Now()), nil
}

// date creates a time object from year, month, day
func date(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 3 {
		return object.Errorf("time.date: expected 3 arguments (year, month, day), got %d", len(args)), nil
	}

	year, err := object.AsInt(args[0])
	if err != nil {
		return nil, err
	}

	month, err := object.AsInt(args[1])
	if err != nil {
		return nil, err
	}

	day, err := object.AsInt(args[2])
	if err != nil {
		return nil, err
	}

	t := time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, time.UTC)
	return NewTimeObject(t), nil
}

// parse parses a date string in YYYY-MM-DD format
func parse(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return object.Errorf("time.parse: expected 1 argument, got %d", len(args)), nil
	}

	dateStr, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	t, parseErr := time.Parse("2006-01-02", dateStr)
	if parseErr != nil {
		return nil, parseErr
	}

	return NewTimeObject(t), nil
}

// TimeObject wraps a Go time.Time for Risor
type TimeObject struct {
	value time.Time
}

func NewTimeObject(t time.Time) *TimeObject {
	return &TimeObject{value: t}
}

func (t *TimeObject) Type() object.Type {
	return "time"
}

func (t *TimeObject) Inspect() string {
	return t.value.Format("2006-01-02")
}

func (t *TimeObject) Interface() interface{} {
	return t.value
}

func (t *TimeObject) IsTruthy() bool {
	return true
}

func (t *TimeObject) Cost() int {
	return 0
}

func (t *TimeObject) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.value.Format(time.RFC3339) + `"`), nil
}

func (t *TimeObject) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for time: %v", opType), nil
}

func (t *TimeObject) Equals(other object.Object) bool {
	if otherTime, ok := other.(*TimeObject); ok {
		return t.value.Equal(otherTime.value)
	}
	return false
}

func (t *TimeObject) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "year":
		return object.NewInt(int64(t.value.Year())), true
	case "month":
		return object.NewInt(int64(t.value.Month())), true
	case "day":
		return object.NewInt(int64(t.value.Day())), true
	case "format":
		return object.NewBuiltin("time.format", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("format: expected 1 argument, got %d", len(args)), nil
			}
			layout, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			return object.NewString(t.value.Format(layout)), nil
		}), true
	}
	return nil, false
}

func (t *TimeObject) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("time objects are immutable")
}

func (t *TimeObject) Attrs() []object.AttrSpec {
	return []object.AttrSpec{
		{Name: "year", Doc: "Year component"},
		{Name: "month", Doc: "Month component (1-12)"},
		{Name: "day", Doc: "Day component"},
	}
}

// GetTime extracts the time.Time value from a TimeObject
func GetTime(obj object.Object) (time.Time, bool) {
	if timeObj, ok := obj.(*TimeObject); ok {
		return timeObj.value, true
	}
	return time.Time{}, false
}
