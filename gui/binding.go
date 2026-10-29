package gui

import (
	"context"
	"fmt"
	"github.com/uidbz/fynerisor/gui/guithread"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
	risorbinding "github.com/uidbz/fynerisor/gui/binding"
)

var _ object.Object = &Binding{}

const BindingType object.Type = "binding"

// Binding is a factory object that provides data binding creation functions to Risor scripts.
// Data bindings allow widgets to automatically sync with data sources.
//
// Available in Risor scripts as the global 'binding' object.
//
// Example usage:
//
//	let data = binding.NewString()
//	data.Set("Hello")
//	let label = widget.NewLabelWithData(data)
//	data.Set("Updated!")  // Label automatically updates
type Binding struct{}

func (obj *Binding) Type() object.Type {
	return BindingType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Binding) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Binding) inspectUI() string {
	return "binding"
}

func (obj *Binding) Interface() interface{} {
	return nil
}

func (obj *Binding) IsTruthy() bool {
	return true
}

func (obj *Binding) Cost() int {
	return 0
}

func (obj *Binding) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'binding'")
}

func (obj *Binding) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", BindingType, opType)
}

func (obj *Binding) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Binding) Attrs() []object.AttrSpec {
	return nil
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Binding) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Binding) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", BindingType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Binding) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *Binding) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "NewString":
		return object.NewBuiltin("binding.NewString", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) > 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0 or 1", len(args))
			}
			if len(args) == 0 {
				return risorbinding.NewString(), nil
			}
			// Optional initial value
			val, err := object.AsString(args[0])
			if err != nil {
				return nil, object.Errorf("type error: %v", err)
			}
			return risorbinding.BindString(val), nil
		}), true

	case "NewBool":
		return object.NewBuiltin("binding.NewBool", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) > 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0 or 1", len(args))
			}
			if len(args) == 0 {
				return risorbinding.NewBool(), nil
			}
			// Optional initial value
			val, err := object.AsBool(args[0])
			if err != nil {
				return nil, object.Errorf("type error: %v", err)
			}
			return risorbinding.BindBool(val), nil
		}), true

	case "NewInt":
		return object.NewBuiltin("binding.NewInt", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) > 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0 or 1", len(args))
			}
			if len(args) == 0 {
				return risorbinding.NewInt(), nil
			}
			// Optional initial value
			val, err := object.AsInt(args[0])
			if err != nil {
				return nil, object.Errorf("type error: %v", err)
			}
			return risorbinding.BindInt(int(val)), nil
		}), true

	case "NewFloat":
		return object.NewBuiltin("binding.NewFloat", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) > 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0 or 1", len(args))
			}
			if len(args) == 0 {
				return risorbinding.NewFloat(), nil
			}
			// Optional initial value
			val, err := object.AsFloat(args[0])
			if err != nil {
				return nil, object.Errorf("type error: %v", err)
			}
			return risorbinding.BindFloat(val), nil
		}), true
	}
	return nil, false
}

func NewBinding() *Binding {
	return &Binding{}
}
