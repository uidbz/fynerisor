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

var _ object.Object = &Float{}

const FloatType object.Type = "binding.Float"

// Float wraps a Fyne data binding for float64 values.
type Float struct {
	instance binding.Float
}

func (obj *Float) Type() object.Type {
	return FloatType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Float) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Float) inspectUI() string {
	val, _ := obj.instance.Get()
	return fmt.Sprintf("binding.Float(%f)", val)
}

func (obj *Float) Interface() interface{} {
	return obj.instance
}

func (obj *Float) IsTruthy() bool {
	return true
}

func (obj *Float) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", FloatType, opType)
}

func (obj *Float) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Float) Attrs() []object.AttrSpec {
	return floatAttrs.Specs()
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Float) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Float) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", FloatType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Float) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var floatAttrs = object.NewAttrRegistry[*Float](string(FloatType))

func (obj *Float) getAttrUI(name string) (object.Object, bool) {
	// Change event: coalesced on the run's queue — listeners
	// re-read the binding, so the latest change is enough and a
	// burst cannot fill the queue (issues #11, #51).
	return floatAttrs.GetAttr(obj, name)
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
func init() {
	floatAttrs.Define("Get").Args().Returns("nil").Impl(func(obj *Float, ctx context.Context,

		args ...object.Object) (object.
		Object, error) {
		val, err := obj.instance.Get()
		if err != nil {
			return nil, object.Errorf("binding error: %v",

				err)
		}
		return object.
			NewFloat(val), nil
	})
	floatAttrs.Define("Set").Args("arg1").
		Returns("nil").Impl(func(obj *Float, ctx context.Context, args ...object.Object) (object.
		Object, error) {
		val, err := object.AsFloat(args[0])
		if err != nil {
			return nil, object.
				Errorf("type error: %v", err)
		}
		err =
			obj.
				instance.Set(val)
		if err != nil {
			return nil, object.Errorf("binding error: %v",

				err)
		}
		return object.Nil, nil
	})
	floatAttrs.Define("AddListener").Args("arg1").Returns("nil").Impl(func(obj *Float, ctx context.
		Context, args ...object.Object) (object.Object, error) {
		cb, err := vmexec.
			NewCallback(ctx, "binding.Float.AddListener",

				args[0])
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
