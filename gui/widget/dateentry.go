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
)

var _ object.Object = &DateEntry{}

const DateEntryType object.Type = "widget.DateEntry"

// DateEntry wraps fyne's Entry widget with date picker (Fyne v2.6+)
type DateEntry struct {
	instance *widget.Entry
	w        WindowInterface
}

func NewDateEntry(w WindowInterface) *DateEntry {
	// Note: Fyne v2.6+ has widget.NewDateEntry(), but for compatibility
	// we'll use a regular entry for now. Apps can upgrade to use native DateEntry
	// when Fyne v2.6+ is available
	instance := widget.NewEntry()
	instance.PlaceHolder = "YYYY-MM-DD"

	return &DateEntry{
		instance: instance,
		w:        w,
	}
}

func (obj *DateEntry) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *DateEntry) Type() object.Type {
	return DateEntryType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *DateEntry) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *DateEntry) inspectUI() string {
	return "widget.DateEntry"
}

func (obj *DateEntry) Interface() interface{} {
	return obj.instance
}

func (obj *DateEntry) IsTruthy() bool {
	return true
}

func (obj *DateEntry) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", DateEntryType, opType)
}

func (obj *DateEntry) Equals(other object.Object) bool {
	return obj == other
}

func (obj *DateEntry) Attrs() []object.AttrSpec {
	return nil
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *DateEntry) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *DateEntry) setAttrUI(name string, value object.Object) error {
	switch name {
	case "Text":
		text, err := object.AsString(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
			obj.instance.SetText(text)
		})
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", DateEntryType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *DateEntry) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *DateEntry) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "Text":
		return object.NewString(obj.instance.Text), true

	case "SetText":
		return object.NewBuiltin("widget.DateEntry.SetText", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}

			text, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}

			guithread.Do(func() {
				obj.instance.SetText(text)
			})

			return object.Nil, nil
		}), true

	case "SetDate":
		return object.NewBuiltin("widget.DateEntry.SetDate", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}

			// Accept string in YYYY-MM-DD format
			dateStr, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}

			// Parse and validate
			_, parseErr := time.Parse("2006-01-02", dateStr)
			if parseErr != nil {
				return nil, object.Errorf("invalid date format, expected YYYY-MM-DD: %v", parseErr)
			}

			guithread.Do(func() {
				obj.instance.SetText(dateStr)
			})

			return object.Nil, nil
		}), true

	case "OnChanged":
		return object.NewBuiltin("widget.DateEntry.OnChanged", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}

			cb, err := vmexec.NewCallback(ctx, "widget.DateEntry.OnChanged", args[0])
			if err != nil {
				return nil, err
			}

			// Change event: coalesced, latest text wins (issue #51).
			obj.instance.OnChanged = func(text string) {
				cb.FireCoalesced(object.NewString(text))
			}

			return object.Nil, nil
		}), true
	case "Hide":
		return object.NewBuiltin("widget.DateEntry.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("widget.DateEntry.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
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
