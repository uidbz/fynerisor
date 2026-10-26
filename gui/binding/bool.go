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

var _ object.Object = &Bool{}

const BoolType object.Type = "binding.Bool"

// Bool wraps a Fyne data binding for bool values.
type Bool struct {
	instance binding.Bool
}

func (obj *Bool) Type() object.Type {
	return BoolType
}

func (obj *Bool) Inspect() string {
	val, _ := obj.instance.Get()
	return fmt.Sprintf("binding.Bool(%v)", val)
}

func (obj *Bool) Interface() interface{} {
	return obj.instance
}

func (obj *Bool) IsTruthy() bool {
	return true
}

func (obj *Bool) Cost() int {
	return 0
}

func (obj *Bool) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'binding.Bool'")
}

func (obj *Bool) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(BoolType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", BoolType, opType)
	return errObj, err
}

func (obj *Bool) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Bool) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Bool) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", BoolType, name)
}

func (obj *Bool) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Get":
		return object.NewBuiltin("binding.Bool.Get", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			val, err := obj.instance.Get()
			if err != nil {
				return object.Errorf("binding error: %v", err), nil
			}
			return object.NewBool(val), nil
		}), true

	case "Set":
		return object.NewBuiltin("binding.Bool.Set", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			val, err := object.AsBool(args[0])
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
		return object.NewBuiltin("binding.Bool.AddListener", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			cb, err := vmexec.NewCallback(ctx, "binding.Bool.AddListener", args[0])
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

func NewBool() *Bool {
	return &Bool{
		instance: binding.NewBool(),
	}
}

func BindBool(value bool) *Bool {
	b := binding.NewBool()
	b.Set(value)
	return &Bool{
		instance: b,
	}
}
