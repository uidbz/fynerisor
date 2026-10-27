package container

import (
	"context"
	"errors"
	"fmt"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"

	"fyne.io/fyne/v2"
	"github.com/uidbz/fynerisor/gui/guithread"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &Border{}

const BorderType object.Type = "container.border"

type Border struct {
	instance                         *fyne.Container
	top, bottom, left, right, center fyne.CanvasObject
}

// rebuild re-applies the border layout after an edge/center change. The Fyne
// border layout keeps its edge objects internally, so the only way to swap
// them is to restate layout + children (same shape container.NewBorder uses).
func (obj *Border) rebuild() {
	obj.instance.Layout = layout.NewBorderLayout(obj.top, obj.bottom, obj.left, obj.right)
	var all []fyne.CanvasObject
	if obj.center != nil {
		all = append(all, obj.center)
	}
	for _, e := range []fyne.CanvasObject{obj.top, obj.bottom, obj.left, obj.right} {
		if e != nil {
			all = append(all, e)
		}
	}
	obj.instance.Objects = all
	obj.instance.Refresh()
}

func (obj *Border) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Border) Type() object.Type {
	return BorderType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Border) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Border) inspectUI() string {
	return "container.border"
}

func (obj *Border) Interface() interface{} {
	return obj.instance
}

func (obj *Border) IsTruthy() bool {
	return true
}

func (obj *Border) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'container.border'")
}

func (obj *Border) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(BorderType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", BorderType, opType)
	return errObj, err
}

func (obj *Border) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Border) Attrs() []object.AttrSpec {
	return nil
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Border) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Border) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", BorderType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Border) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *Border) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "SetTop":
		return borderSetter("SetTop", obj, func(b *Border, co fyne.CanvasObject) { b.top = co }), true
	case "SetBottom":
		return borderSetter("SetBottom", obj, func(b *Border, co fyne.CanvasObject) { b.bottom = co }), true
	case "SetLeft":
		return borderSetter("SetLeft", obj, func(b *Border, co fyne.CanvasObject) { b.left = co }), true
	case "SetRight":
		return borderSetter("SetRight", obj, func(b *Border, co fyne.CanvasObject) { b.right = co }), true
	case "SetCenter":
		return borderSetter("SetCenter", obj, func(b *Border, co fyne.CanvasObject) { b.center = co }), true
	case "Refresh":
		return object.NewBuiltin("container.Border.Refresh", func(_ context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() { obj.instance.Refresh() })
			return object.Nil, nil
		}), true
	case "Hide":
		return object.NewBuiltin("container.Border.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("container.Border.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Show()
			})
			return object.Nil, nil
		}), true
	}
	return nil, false
}

func NewBorder(top, bottom, left, right, center fyne.CanvasObject) *Border {
	var border *fyne.Container
	if center != nil {
		border = container.NewBorder(top, bottom, left, right, center)
	} else {
		border = container.NewBorder(top, bottom, left, right)
	}
	return &Border{instance: border, top: top, bottom: bottom, left: left, right: right, center: center}
}

// borderSetter returns a SetX builtin that replaces one border slot.
func borderSetter(name string, obj *Border, set func(*Border, fyne.CanvasObject)) *object.Builtin {
	return object.NewBuiltin("container.Border."+name, func(_ context.Context, args ...object.Object) (object.Object, error) {
		if len(args) != 1 {
			return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
		}
		var co fyne.CanvasObject
		if args[0].Type() != object.NIL {
			var ok bool
			co, ok = canvasArg(args[0])
			if !ok {
				return object.Errorf("%s: expected canvas object, got %s", name, args[0].Type()), nil
			}
		}
		guithread.Do(func() {
			set(obj, co)
			obj.rebuild()
		})
		return object.Nil, nil
	})
}
