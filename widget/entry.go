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

var _ object.Object = &Entry{}

const EntryType object.Type = "widget.Entry"

type Entry struct {
	instance *widget.Entry
	w WindowInterface // needed for accepting functions as argument (OnSubmitted)
}

func (obj *Entry) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Entry) Type() object.Type {
	return EntryType
}

func (obj *Entry) Inspect() string {
	return "widget.Entry"
}

func (obj *Entry) Interface() interface{} {
	return obj.instance
}

func (obj *Entry) IsTruthy() bool {
	return true
}

func (obj *Entry) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'widget.Entry'")
}

func (obj *Entry) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(EntryType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", EntryType, opType)
	return errObj, err
}

func (obj *Entry) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Entry) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Entry) SetAttr(name string, value object.Object) error {
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
		fyne.Do(func() {
			obj.instance.Refresh()
		})
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", EntryType, name)
}

func (obj *Entry) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Text":
		return object.NewString(obj.instance.Text), true

	case "PlaceHolder":
		return object.NewString(obj.instance.PlaceHolder), true

	case "SetText":
		return object.NewBuiltin("widget.Entry.SetText", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=exactly 1", len(args)), nil
			}

			val, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}

			fyne.Do(func() {
				obj.instance.SetText(val)
			})

			return object.Nil, nil
		}), true

	case "Append":
		return object.NewBuiltin("widget.Entry.Append", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=exactly 1", len(args)), nil
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
				return object.Errorf("wrong number of arguments. got=%d, want=2", len(args)), nil
			}
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("button (%s): unable to get call function", name), nil
			}

			obj.instance.OnSubmitted = func(text string) {
				obj.w.Do(func() {
					s := object.NewString(text)
					callFunc(ctx, fn, []object.Object{s})
				})
			}

			return object.Nil, nil
		}), true

	case "OnChanged":
		return object.NewBuiltin("widget.Entry.OnChanged", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("entry.OnChanged: unable to get call function"), nil
			}

			obj.instance.OnChanged = func(text string) {
				go func() {
					s := object.NewString(text)
					callFunc(ctx, fn, []object.Object{s})
				}()
			}

			return object.Nil, nil
		}), true

	case "SetValidator":
		return object.NewBuiltin("widget.Entry.SetValidator", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("entry.SetValidator: unable to get call function"), nil
			}

			// Validator returns error or nil
			obj.instance.Validator = func(text string) error {
				s := object.NewString(text)
				result, err := callFunc(ctx, fn, []object.Object{s})
				if err != nil {
					return err
				}
				// If result is nil or empty string, validation passes
				if result == object.Nil {
					return nil
				}
				if str, ok := result.(*object.String); ok {
					if str.Value() == "" {
						return nil
					}
					return fmt.Errorf("%s", str.Value())
				}
				return nil
			}

			return object.Nil, nil
		}), true
	}
	return nil, false
}

func NewEntry(w WindowInterface) *Entry {
	return &Entry{instance: widget.NewEntry(), w: w}
}

func NewMultiLineEntry() *Entry {
	return &Entry{instance: widget.NewMultiLineEntry()}
}
