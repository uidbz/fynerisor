package widget

import (
	"context"
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"github.com/uidbz/fynerisor/gui/guithread"
	"fyne.io/fyne/v2/widget"

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

func (obj *ProgressBarInfinite) Inspect() string {
	return "widget.ProgressBarInfinite()"
}

func (obj *ProgressBarInfinite) Interface() interface{} {
	return obj.instance
}

func (obj *ProgressBarInfinite) IsTruthy() bool {
	return true
}

func (obj *ProgressBarInfinite) Cost() int {
	return 0
}

func (obj *ProgressBarInfinite) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *ProgressBarInfinite) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'widget.ProgressBarInfinite'")
}

func (obj *ProgressBarInfinite) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(ProgressBarInfiniteType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", ProgressBarInfiniteType, opType)
	return errObj, err
}

func (obj *ProgressBarInfinite) Equals(other object.Object) bool {
	return obj == other
}

func (obj *ProgressBarInfinite) Attrs() []object.AttrSpec {
	return nil
}

func (obj *ProgressBarInfinite) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Start":
		return object.NewBuiltin("widget.ProgressBarInfinite.Start", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Show()
				obj.instance.Start()
			})
			return object.Nil, nil
		}), true
	case "Stop":
		return object.NewBuiltin("widget.ProgressBarInfinite.Stop", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Stop()
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Running":
		return object.NewBool(obj.instance.Running()), true
	case "Hide":
		return object.NewBuiltin("widget.ProgressBarInfinite.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("widget.ProgressBarInfinite.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Show()
			})
			return object.Nil, nil
		}), true
	}
	return nil, false
}

func (obj *ProgressBarInfinite) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", ProgressBarInfiniteType, name)
}
