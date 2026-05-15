package binding

import (
	"context"
	"errors"
	"fmt"

	"fyne.io/fyne/v2/data/binding"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &Float{}

const FloatType object.Type = "binding.Float"

// Float wraps a Fyne data binding for float64 values.
type Float struct {
	instance binding.Float
}

func (obj *Float) Type() object.Type {
	return FloatType
}

func (obj *Float) Inspect() string {
	val, _ := obj.instance.Get()
	return fmt.Sprintf("binding.Float(%f)", val)
}

func (obj *Float) Interface() interface{} {
	return obj.instance
}

func (obj *Float) IsTruthy() bool {
	return true
}

func (obj *Float) Cost() int {
	return 0
}

func (obj *Float) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'binding.Float'")
}

func (obj *Float) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(FloatType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", FloatType, opType)
	return errObj, err
}

func (obj *Float) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Float) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Float) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", FloatType, name)
}

func (obj *Float) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Get":
		return object.NewBuiltin("binding.Float.Get", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			val, err := obj.instance.Get()
			if err != nil {
				return object.Errorf("binding error: %v", err), nil
			}
			return object.NewFloat(val), nil
		}), true

	case "Set":
		return object.NewBuiltin("binding.Float.Set", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			val, err := object.AsFloat(args[0])
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
		return object.NewBuiltin("binding.Float.AddListener", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}

			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("binding.Float.AddListener: unable to get call function"), nil
			}

			listener := binding.NewDataListener(func() {
				go func() {
					_, err := callFunc(ctx, fn, []object.Object{})
					if err != nil {
						fmt.Printf("ERROR: binding listener error: %v\n", err)
					}
				}()
			})

			obj.instance.AddListener(listener)
			return object.Nil, nil
		}), true
	}
	return nil, false
}

func NewFloat() *Float {
	return &Float{
		instance: binding.NewFloat(),
	}
}

func BindFloat(value float64) *Float {
	f := binding.NewFloat()
	f.Set(value)
	return &Float{
		instance: f,
	}
}
