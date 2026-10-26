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

var _ object.Object = &Int{}

const IntType object.Type = "binding.Int"

// Int wraps a Fyne data binding for int values.
type Int struct {
	instance binding.Int
}

func (obj *Int) Type() object.Type {
	return IntType
}

func (obj *Int) Inspect() string {
	val, _ := obj.instance.Get()
	return fmt.Sprintf("binding.Int(%d)", val)
}

func (obj *Int) Interface() interface{} {
	return obj.instance
}

func (obj *Int) IsTruthy() bool {
	return true
}

func (obj *Int) Cost() int {
	return 0
}

func (obj *Int) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'binding.Int'")
}

func (obj *Int) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(IntType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", IntType, opType)
	return errObj, err
}

func (obj *Int) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Int) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Int) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", IntType, name)
}

func (obj *Int) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Get":
		return object.NewBuiltin("binding.Int.Get", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			val, err := obj.instance.Get()
			if err != nil {
				return object.Errorf("binding error: %v", err), nil
			}
			return object.NewInt(int64(val)), nil
		}), true

	case "Set":
		return object.NewBuiltin("binding.Int.Set", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			val, err := object.AsInt(args[0])
			if err != nil {
				return object.Errorf("type error: %v", err), nil
			}
			err = obj.instance.Set(int(val))
			if err != nil {
				return object.Errorf("binding error: %v", err), nil
			}
			return object.Nil, nil
		}), true

	case "AddListener":
		return object.NewBuiltin("binding.Int.AddListener", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			cb, err := vmexec.NewCallback(ctx, "binding.Int.AddListener", args[0])
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

func NewInt() *Int {
	return &Int{
		instance: binding.NewInt(),
	}
}

func BindInt(value int) *Int {
	i := binding.NewInt()
	i.Set(value)
	return &Int{
		instance: i,
	}
}
