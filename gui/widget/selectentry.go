package widget

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/vmexec"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &SelectEntry{}

const SelectEntryType object.Type = "widget.SelectEntry"

// SelectEntry wraps fyne's SelectEntry widget - a searchable dropdown entry
type SelectEntry struct {
	instance *widget.SelectEntry
	w        WindowInterface
}

func NewSelectEntry(w WindowInterface, options []string, onChanged func(string)) *SelectEntry {
	var instance *widget.SelectEntry
	if onChanged != nil {
		instance = widget.NewSelectEntry(options)
		instance.OnChanged = onChanged
	} else {
		instance = widget.NewSelectEntry(options)
	}

	return &SelectEntry{
		instance: instance,
		w:        w,
	}
}

func (obj *SelectEntry) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *SelectEntry) Type() object.Type {
	return SelectEntryType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *SelectEntry) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *SelectEntry) inspectUI() string {
	return "widget.SelectEntry"
}

func (obj *SelectEntry) Interface() interface{} {
	return obj.instance
}

func (obj *SelectEntry) IsTruthy() bool {
	return true
}

func (obj *SelectEntry) Cost() int {
	return 0
}

func (obj *SelectEntry) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'widget.SelectEntry'")
}

func (obj *SelectEntry) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", SelectEntryType, opType)
}

func (obj *SelectEntry) Equals(other object.Object) bool {
	return obj == other
}

func (obj *SelectEntry) Attrs() []object.AttrSpec {
	return nil
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *SelectEntry) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *SelectEntry) setAttrUI(name string, value object.Object) error {
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
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", SelectEntryType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *SelectEntry) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *SelectEntry) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "Text":
		return object.NewString(obj.instance.Text), true

	case "PlaceHolder":
		return object.NewString(obj.instance.PlaceHolder), true

	case "SetText":
		return object.NewBuiltin("widget.SelectEntry.SetText", func(ctx context.Context, args ...object.Object) (object.Object, error) {
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

	case "SetOptions":
		return object.NewBuiltin("widget.SelectEntry.SetOptions", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}

			list, err := object.AsList(args[0])
			if err != nil {
				return nil, err
			}

			options := make([]string, len(list.Value()))
			for i, item := range list.Value() {
				opt, err := object.AsString(item)
				if err != nil {
					return nil, err
				}
				options[i] = opt
			}

			guithread.Do(func() {
				obj.instance.SetOptions(options)
			})

			return object.Nil, nil
		}), true

	case "OnChanged":
		return object.NewBuiltin("widget.SelectEntry.OnChanged", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}

			cb, err := vmexec.NewCallback(ctx, "widget.SelectEntry.OnChanged", args[0])
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
		return object.NewBuiltin("widget.SelectEntry.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("widget.SelectEntry.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
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
