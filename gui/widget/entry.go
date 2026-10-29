package widget

import (
	"context"
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"
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
	return entryAttrs.

		// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
		Specs()
}

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

var entryAttrs = object.NewAttrRegistry[*Entry](string(EntryType))

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Entry) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *Entry) getAttrUI(name string) (object.Object, bool) {
	return entryAttrs.GetAttr(obj, name)

	// Event: queued on the script's VM goroutine.

	// Change event: coalesced on the run's queue — a burst of
	// keystrokes while the VM is busy delivers the latest text and
	// never fills the queue (issue #51). Order between distinct
	// changes is kept; only superseded values are skipped.

	// Getter: Fyne calls the validator synchronously on the UI
	// thread. While the VM is busy keep the previous verdict.

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
func init() {
	objattrs.Add(entryAttrs)
	entryAttrs.Define("Text").
		Doc("The text.").Getter(func(obj *Entry) object.
		Object {
		return object.NewString(obj.instance.Text)
	})
	entryAttrs.Define("PlaceHolder").Getter(func(obj *Entry) object.Object {
		return object.NewString(obj.instance.
			PlaceHolder)
	})
	entryAttrs.Define("Disabled").Doc("Whether the object is disabled.").Getter(func(obj *Entry) object.Object {
		return object.NewBool(
			obj.instance.Disabled())
	})
	entryAttrs.Define("SetText").Doc("Set the text.").Args("arg1").Returns("nil").Impl(func(obj *Entry, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {
		val, err := object.AsString(args[0])
		if err !=

			nil {
			return nil, err
		}
		guithread.Do(func() {
			obj.instance.SetText(val)
		})
		return object.
				Nil,

			nil
	})
	entryAttrs.Define("SetPlaceHolder").
		Args("arg1").Returns("nil").Impl(func(obj *Entry, ctx context.Context, args ...object.Object) (
		object.Object, error) {
		val, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}
		obj.instance.PlaceHolder = val
		guithread.
			Do(func() { obj.instance.Refresh() })

		return object.Nil, nil
	})
	entryAttrs.Define("Append").Args("arg1").Returns("nil").Impl(func(obj *Entry, ctx context.Context, args ...object.Object) (object.Object, error) {
		val, err := object.AsString(
			args[0])
		if err != nil {
			return nil, err

		}
		obj.instance.Append(val)
		return object.Nil,
			nil
	})
	entryAttrs.Define("OnSubmitted").Args("arg1").
		Returns(
			"nil").Impl(func(obj *Entry, ctx context.Context, args ...object.Object) (object.
		Object, error) {
		cb, err := vmexec.NewCallback(ctx, "widget.Entry.OnSubmitted", args[0])
		if err != nil {
			return nil, err
		}
		obj.
			instance.OnSubmitted = func(text string) {
			cb.
				Fire(object.NewString(text))
		}

		return object.Nil, nil
	})
	entryAttrs.Define("OnChanged").Doc("Set the change handler; fires on the script VM goroutine.").Args("arg1").Returns("nil").Impl(func(obj *Entry, ctx context.Context, args ...object.Object) (object.Object, error) {
		cb,

			err := vmexec.NewCallback(ctx, "widget.Entry.OnChanged",
			args[0])
		if err != nil {
			return nil,

				err
		}
		obj.instance.OnChanged = func(text string) {
			cb.FireCoalesced(object.
				NewString(text))
		}
		return object.Nil, nil
	})
	entryAttrs.Define("SetValidator").Args("arg1").
		Returns("nil").Impl(func(obj *Entry, ctx context.Context,
		args ...object.Object) (object.
		Object, error) {
		cb, err := vmexec.NewCallback(ctx, "widget.Entry.SetValidator",

			args[0])
		if err != nil {
			return nil, err
		}
		var last error

		obj.instance.
			Validator = func(text string) error {
			s := object.
				NewString(text)
			result, err := cb.Call(s)
			if err != nil {
				if vmexec.
					Quiet(err) {
					return last
				}
				last = err
				return err
			}
			last = validatorVerdict(result)
			return last
		}
		return object.
			Nil, nil
	})
	entryAttrs.Define("Disable").Doc("Disable the object.").Args().Returns("nil").Impl(func(obj *Entry, ctx context.
		Context, args ...object.
		Object) (object.Object, error) {
		guithread.
			Do(func() { obj.instance.Disable() })
		return object.Nil, nil
	})
	entryAttrs.Define("Enable").Doc("Re-enable the object.").Args().Returns("nil").Impl(func(obj *Entry, ctx context.Context, args ...object.
		Object) (object.Object, error) {
		guithread.
			Do(func() { obj.instance.Enable() })
		return object.Nil, nil
	})
}
