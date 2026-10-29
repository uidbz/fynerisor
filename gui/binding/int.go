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

var _ object.Object = &Int{}

const IntType object.Type = "binding.Int"

// Int wraps a Fyne data binding for int values.
type Int struct {
	instance binding.Int
}

func (obj *Int) Type() object.Type {
	return IntType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Int) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Int) inspectUI() string {
	val, _ := obj.instance.Get()
	return fmt.Sprintf("binding.Int(%d)", val)
}

func (obj *Int) Interface() interface{} {
	return obj.instance
}

func (obj *Int) IsTruthy() bool {
	return true
}

func (obj *Int) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", IntType, opType)
}

func (obj *Int) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Int) Attrs() []object.AttrSpec {
	return intAttrs.Specs()
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Int) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Int) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", IntType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Int) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var intAttrs = object.NewAttrRegistry[*Int](string(IntType))

func (obj *Int) getAttrUI(name string) (object.Object, bool) {
	return intAttrs.GetAttr(obj, name)

	// Change event: coalesced on the run's queue — listeners
	// re-read the binding, so the latest change is enough and a
	// burst cannot fill the queue (issues #11, #51).

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
func init() {
	intAttrs.Define(
		"Get",
	).Args().Returns("nil").
		Impl(func(obj *Int, ctx context.Context, args ...object.Object) (object.Object, error) {
			val, err := obj.instance.Get()
			if err !=
				nil {
				return nil, object.Errorf("binding error: %v",

					err)
			}
			return object.NewInt(
				int64(val)), nil
		})
	intAttrs.Define(
		"Set",
	).Args("arg1").Returns(
		"nil").Impl(func(obj *Int, ctx context.Context, args ...object.Object) (object.
		Object, error) {
		val, err := object.AsInt(args[0])
		if err != nil {
			return nil, object.Errorf("type error: %v", err)
		}
		err = obj.instance.
			Set(int(val))
		if err !=
			nil {
			return nil, object.Errorf("binding error: %v",

				err)
		}
		return object.Nil, nil
	})
	intAttrs.Define(
		"AddListener",
	).Args("arg1").
		Returns("nil").Impl(func(obj *Int, ctx context.
		Context, args ...object.Object) (object.Object, error) {
		cb, err := vmexec.NewCallback(ctx, "binding.Int.AddListener", args[0])
		if err != nil {
			return nil, err
		}
		obj.
			instance.AddListener(binding.
			NewDataListener(func() {
				cb.FireCoalesced()
			}))
		return object.Nil, nil
	})
}
