package binding

import (
	"context"
	"fmt"
	"github.com/uidbz/fynerisor/gui/guithread"

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

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *String) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *String) inspectUI() string {
	val, _ := obj.instance.Get()
	return fmt.Sprintf("binding.String(%q)", val)
}

func (obj *String) Interface() interface{} {
	return obj.instance
}

func (obj *String) IsTruthy() bool {
	return true
}

func (obj *String) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", StringType, opType)
}

func (obj *String) Equals(other object.Object) bool {
	return obj == other
}

func (obj *String) Attrs() []object.AttrSpec {
	return nil
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *String) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *String) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", StringType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *String) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *String) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "Get":
		return object.NewBuiltin("binding.String.Get", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			val, err := obj.instance.Get()
			if err != nil {
				return nil, object.Errorf("binding error: %v", err)
			}
			return object.NewString(val), nil
		}), true

	case "Set":
		return object.NewBuiltin("binding.String.Set", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}
			val, err := object.AsString(args[0])
			if err != nil {
				return nil, object.Errorf("type error: %v", err)
			}
			err = obj.instance.Set(val)
			if err != nil {
				return nil, object.Errorf("binding error: %v", err)
			}
			return object.Nil, nil
		}), true

	case "AddListener":
		return object.NewBuiltin("binding.String.AddListener", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}
			cb, err := vmexec.NewCallback(ctx, "binding.String.AddListener", args[0])
			if err != nil {
				return nil, err
			}

			// Change event: coalesced on the run's queue — listeners
			// re-read the binding, so the latest change is enough and a
			// burst cannot fill the queue (issues #11, #51).
			obj.instance.AddListener(binding.NewDataListener(func() { cb.FireCoalesced() }))
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
