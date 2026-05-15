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

func (obj *SelectEntry) Inspect() string {
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
	err := errors.New("eval error: unsupported operation for " + string(SelectEntryType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", SelectEntryType, opType)
	return errObj, err
}

func (obj *SelectEntry) Equals(other object.Object) bool {
	return obj == other
}

func (obj *SelectEntry) Attrs() []object.AttrSpec {
	return nil
}

func (obj *SelectEntry) SetAttr(name string, value object.Object) error {
	switch name {
	case "Text":
		text, err := object.AsString(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		fyne.Do(func() {
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

func (obj *SelectEntry) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Text":
		return object.NewString(obj.instance.Text), true

	case "PlaceHolder":
		return object.NewString(obj.instance.PlaceHolder), true

	case "SetText":
		return object.NewBuiltin("widget.SelectEntry.SetText", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}

			text, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}

			fyne.Do(func() {
				obj.instance.SetText(text)
			})

			return object.Nil, nil
		}), true

	case "SetOptions":
		return object.NewBuiltin("widget.SelectEntry.SetOptions", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
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

			fyne.Do(func() {
				obj.instance.SetOptions(options)
			})

			return object.Nil, nil
		}), true

	case "OnChanged":
		return object.NewBuiltin("widget.SelectEntry.OnChanged", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}

			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}

			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("selectentry.OnChanged: unable to get call function"), nil
			}

			obj.instance.OnChanged = func(text string) {
				obj.w.Do(func() {
					callFunc(ctx, fn, []object.Object{object.NewString(text)})
				})
			}

			return object.Nil, nil
		}), true
	}

	return nil, false
}
