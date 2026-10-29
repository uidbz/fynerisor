package widget

import (
	"context"
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/vmexec"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &Entry{}

const EntryType object.Type = "widget.Entry"

type Entry struct {
	instance *widget.Entry
	w        WindowInterface
}

// validatorVerdict maps a validator script result to a Fyne validation
// error: nil or "" passes, any other string is the error message.
func validatorVerdict(result object.Object) error {
	if result == nil || result == object.Nil {
		return nil
	}
	if str, ok := result.(*object.String); ok {
		if str.Value() == "" {
			return nil
		}
		return errors.New(str.Value())
	}
	return nil
}

func (obj *Entry) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Entry) Type() object.Type {
	return EntryType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Entry) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Entry) inspectUI() string {
	return "widget.Entry"
}

func (obj *Entry) Interface() interface{} {
	return obj.instance
}

func (obj *Entry) IsTruthy() bool {
	return true
}

func (obj *Entry) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", EntryType, opType)
}

func (obj *Entry) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Entry) Attrs() []object.AttrSpec {
	return nil
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Entry) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Entry) setAttrUI(name string, value object.Object) error {
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

	case "PlaceHolder":
		text, err := object.AsString(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		obj.instance.PlaceHolder = text
		guithread.Do(func() {
			obj.instance.Refresh()
		})
		return nil

	case "Disabled":
		b, err := object.AsBool(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
			if b {
				obj.instance.Disable()
			} else {
				obj.instance.Enable()
			}
		})
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", EntryType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Entry) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *Entry) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "Text":
		return object.NewString(obj.instance.Text), true

	case "PlaceHolder":
		return object.NewString(obj.instance.PlaceHolder), true

	case "Disabled":
		return object.NewBool(obj.instance.Disabled()), true

	case "SetText":
		return object.NewBuiltin("widget.Entry.SetText", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=exactly 1", len(args))
			}

			val, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}

			guithread.Do(func() {
				obj.instance.SetText(val)
			})

			return object.Nil, nil
		}), true

	case "SetPlaceHolder":
		return object.NewBuiltin("widget.Entry.SetPlaceHolder", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}

			val, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}

			obj.instance.PlaceHolder = val
			guithread.Do(func() {
				obj.instance.Refresh()
			})

			return object.Nil, nil
		}), true

	case "Append":
		return object.NewBuiltin("widget.Entry.Append", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=exactly 1", len(args))
			}

			val, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}

			obj.instance.Append(val)

			return object.Nil, nil
		}), true

	case "OnSubmitted":
		return object.NewBuiltin("widget.OnSubmitted", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}
			cb, err := vmexec.NewCallback(ctx, "widget.Entry.OnSubmitted", args[0])
			if err != nil {
				return nil, err
			}

			// Event: queued on the script's VM goroutine.
			obj.instance.OnSubmitted = func(text string) {
				cb.Fire(object.NewString(text))
			}

			return object.Nil, nil
		}), true

	case "OnChanged":
		return object.NewBuiltin("widget.Entry.OnChanged", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}
			cb, err := vmexec.NewCallback(ctx, "widget.Entry.OnChanged", args[0])
			if err != nil {
				return nil, err
			}

			// Change event: coalesced on the run's queue — a burst of
			// keystrokes while the VM is busy delivers the latest text and
			// never fills the queue (issue #51). Order between distinct
			// changes is kept; only superseded values are skipped.
			obj.instance.OnChanged = func(text string) {
				cb.FireCoalesced(object.NewString(text))
			}

			return object.Nil, nil
		}), true

	case "SetValidator":
		return object.NewBuiltin("widget.Entry.SetValidator", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}
			cb, err := vmexec.NewCallback(ctx, "widget.Entry.SetValidator", args[0])
			if err != nil {
				return nil, err
			}

			// Getter: Fyne calls the validator synchronously on the UI
			// thread. While the VM is busy keep the previous verdict.
			var last error
			obj.instance.Validator = func(text string) error {
				s := object.NewString(text)
				result, err := cb.Call(s)
				if err != nil {
					if vmexec.Quiet(err) {
						return last
					}
					last = err
					return err
				}
				last = validatorVerdict(result)
				return last
			}

			return object.Nil, nil
		}), true

	case "Disable":
		return object.NewBuiltin("widget.Entry.Disable", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			guithread.Do(func() {
				obj.instance.Disable()
			})
			return object.Nil, nil
		}), true

	case "Enable":
		return object.NewBuiltin("widget.Entry.Enable", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			guithread.Do(func() {
				obj.instance.Enable()
			})
			return object.Nil, nil
		}), true
	case "Hide":
		return object.NewBuiltin("widget.Entry.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("widget.Entry.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
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

func NewEntry(w WindowInterface) *Entry {
	return &Entry{instance: widget.NewEntry(), w: w}
}

// NewEntryWithData creates a new entry bound to a string data binding (bi-directional)
func NewEntryWithData(data interface{ Get() (string, error) }, w WindowInterface) *Entry {
	// Cast to binding.String
	bindingStr, ok := data.(binding.String)
	if !ok {
		// Fallback: create a regular entry
		return &Entry{instance: widget.NewEntry(), w: w}
	}

	return &Entry{instance: widget.NewEntryWithData(bindingStr), w: w}
}

func NewMultiLineEntry(w WindowInterface) *Entry {
	return &Entry{instance: widget.NewMultiLineEntry(), w: w}
}
