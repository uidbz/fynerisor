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

var _ object.Object = &Bool{}

const BoolType object.Type = "binding.Bool"

// Bool wraps a Fyne data binding for bool values.
type Bool struct {
	instance binding.Bool
}

func (obj *Bool) Type() object.Type {
	return BoolType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Bool) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Bool) inspectUI() string {
	val, _ := obj.instance.Get()
	return fmt.Sprintf("binding.Bool(%v)", val)
}

func (obj *Bool) Interface() interface{} {
	return obj.instance
}

func (obj *Bool) IsTruthy() bool {
	return true
}

func (obj *Bool) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", BoolType, opType)
}

func (obj *Bool) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Bool) Attrs() []object.AttrSpec {
	return boolAttrs.Specs()
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Bool) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Bool) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", BoolType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Bool) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var boolAttrs = object.NewAttrRegistry[*Bool](string(BoolType))

func (obj *Bool) getAttrUI(name string) (object.Object, bool) {
	// Change event: coalesced on the run's queue — listeners
	// re-read the binding, so the latest change is enough and a
	// burst cannot fill the queue (issues #11, #51).
	return boolAttrs.GetAttr(obj, name)
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
func init() {
	boolAttrs.Define(
		"Get").Args().Returns("nil").Impl(func(obj *Bool, ctx context.Context,
		args ...object.Object) (object.Object, error) {
		val, err := obj.instance.Get()
		if err != nil {
			return nil, object.Errorf("binding error: %v",

				err)
		}
		return object.NewBool(val), nil
	})
	boolAttrs.Define(
		"Set").Args("arg1").
		Returns("nil").Impl(func(obj *Bool, ctx context.Context, args ...object.Object) (object.
		Object, error) {
		val, err := object.AsBool(args[0])
		if err != nil {
			return nil, object.
				Errorf("type error: %v", err)
		}
		err = obj.
			instance.
			Set(val)
		if err != nil {
			return nil, object.Errorf("binding error: %v",

				err)
		}
		return object.Nil, nil
	})
	boolAttrs.Define(
		"AddListener",
	).Args(
		"arg1").
		Returns("nil").Impl(func(obj *Bool, ctx context.
		Context, args ...object.Object) (object.Object, error) {
		cb, err := vmexec.NewCallback(ctx, "binding.Bool.AddListener", args[0])
		if err != nil {
			return nil, err
		}

		obj.instance.AddListener(binding.
			NewDataListener(func() {
				cb.FireCoalesced()
			}))
		return object.Nil, nil
	})
}
