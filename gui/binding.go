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

func (obj *Binding) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", BindingType, opType)
}

func (obj *Binding) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Binding) Attrs() []object.AttrSpec {
	return bindingAttrs.Specs()
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

var bindingAttrs = object.NewAttrRegistry[*Binding](string(BindingType))

func (obj *Binding) getAttrUI(name string) (object.Object, bool) {
	return bindingAttrs.GetAttr(obj, name)

	// Optional initial value

	// Optional initial value

	// Optional initial value

	// Optional initial value

}

func NewBinding() *Binding {
	return &Binding{}
}
func init() {
	bindingAttrs.
		Define("NewString").
		Variadic("args").Returns("nil").Impl(func(obj *Binding, ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) > 1 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=0 or 1",

				len(args))
		}
		if len(args) == 0 {
			return risorbinding.NewString(), nil
		}
		val, err := object.AsString(args[0])
		if err != nil {
			return nil, object.Errorf("type error: %v", err)
		}
		return risorbinding.BindString(val), nil
	})
	bindingAttrs.
		Define("NewBool").Variadic(
		"args").Returns("nil").Impl(func(obj *Binding, ctx context.Context, args ...object.Object) (object.
		Object, error) {
		if len(args) > 1 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=0 or 1",

				len(args))
		}
		if len(args) == 0 {
			return risorbinding.NewBool(), nil
		}
		val, err := object.AsBool(args[0])
		if err != nil {
			return nil, object.Errorf("type error: %v",
				err)
		}
		return risorbinding.BindBool(val), nil
	})
	bindingAttrs.
		Define("NewInt").Variadic("args").Returns("nil").Impl(func(obj *Binding,

		ctx context.Context, args ...object.Object) (object.
		Object, error) {
		if len(args) > 1 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=0 or 1",

				len(args))
		}
		if len(args) == 0 {
			return risorbinding.NewInt(), nil
		}
		val, err := object.AsInt(args[0])
		if err != nil {
			return nil, object.Errorf("type error: %v",
				err)
		}
		return risorbinding.BindInt(int(val)), nil
	})
	bindingAttrs.
		Define("NewFloat").Variadic("args").Returns("nil").Impl(func(obj *Binding, ctx context.Context, args ...object.Object) (
		object.Object, error) {
		if len(args) > 1 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=0 or 1",

				len(args))
		}
		if len(args) == 0 {
			return risorbinding.NewFloat(), nil
		}
		val, err := object.AsFloat(args[0])
		if err != nil {
			return nil, object.Errorf(
				"type error: %v", err)
		}
		return risorbinding.BindFloat(val), nil
	})
}
