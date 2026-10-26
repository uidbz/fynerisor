package binding

import (
	"context"
	"errors"
	"fmt"

	"fyne.io/fyne/v2/data/binding"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"

	"github.com/uidbz/fynerisor/gui/vmexec"
)

var _ object.Object = &String{}

const StringType object.Type = "binding.String"

// String wraps a Fyne data binding for string values.
// It allows widgets to automatically sync with data sources.
type String struct {
	instance binding.String
}

func (obj *String) CanvasObject() interface{} {
	return obj.instance
}

func (obj *String) Type() object.Type {
	return StringType
}

func (obj *String) Inspect() string {
	val, _ := obj.instance.Get()
	return fmt.Sprintf("binding.String(%q)", val)
}

func (obj *String) Interface() interface{} {
	return obj.instance
}

func (obj *String) IsTruthy() bool {
	return true
}

func (obj *String) Cost() int {
	return 0
}

func (obj *String) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'binding.String'")
}

func (obj *String) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(StringType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", StringType, opType)
	return errObj, err
}

func (obj *String) Equals(other object.Object) bool {
	return obj == other
}

func (obj *String) Attrs() []object.AttrSpec {
	return nil
}

func (obj *String) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", StringType, name)
}

func (obj *String) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Get":
		return object.NewBuiltin("binding.String.Get", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			val, err := obj.instance.Get()
			if err != nil {
				return object.Errorf("binding error: %v", err), nil
			}
			return object.NewString(val), nil
		}), true

	case "Set":
		return object.NewBuiltin("binding.String.Set", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			val, err := object.AsString(args[0])
			if err != nil {
				return object.Errorf("type error: %v", err), nil
			}
			err = obj.instance.Set(val)
			if err != nil {
				return object.Errorf("binding error: %v", err), nil
			}
			return object.Nil, nil
		}), true

	case "AddListener":
		return object.NewBuiltin("binding.String.AddListener", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			cb, err := vmexec.NewCallback(ctx, "binding.String.AddListener", args[0])
			if err != nil {
				return nil, err
			}

			// Event: each change is queued on the script's VM goroutine
			// (never a goroutine per change, issue #11).
			obj.instance.AddListener(binding.NewDataListener(func() { cb.Fire() }))
			return object.Nil, nil
		}), true
	}
	return nil, false
}

func NewString() *String {
	return &String{
		instance: binding.NewString(),
	}
}

func BindString(value string) *String {
	str := binding.NewString()
	str.Set(value)
	return &String{
		instance: str,
	}
}
