package container

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2/container"

	"fyne.io/fyne/v2"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &Scroll{}

const ScrollType object.Type = "container.scroll"

type Scroll struct {
	instance *container.Scroll
}

func (obj *Scroll) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *Scroll) Type() object.Type {
	return ScrollType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Scroll) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Scroll) inspectUI() string {
	return "container.scroll"
}

func (obj *Scroll) Interface() interface{} {
	return obj.instance
}

func (obj *Scroll) IsTruthy() bool {
	return true
}

func (obj *Scroll) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", ScrollType, opType)
}

func (obj *Scroll) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Scroll) Attrs() []object.AttrSpec {
	return scrollAttrs.

		// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
		Specs()
}

func (obj *Scroll) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Scroll) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", ScrollType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Scroll) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var scrollAttrs = object.NewAttrRegistry[*Scroll](string(ScrollType))

func (obj *Scroll) getAttrUI(name string) (object.Object, bool) {
	return scrollAttrs.GetAttr(obj, name)

}

func NewScroll(content fyne.CanvasObject) *Scroll {
	return &Scroll{
		instance: container.NewScroll(content),
	}
}
func init() {
	objattrs.Add(scrollAttrs)
	scrollAttrs.Define(
		"SetContent",
	).Args("arg1").
		Returns("nil").Impl(func(obj *Scroll, ctx context.
		Context, args ...object.Object) (object.Object, error) {
		co, ok := canvasArg(args[0])
		if !ok {
			return nil, object.Errorf("SetContent: expected canvas object, got %s",

				args[0].Type())
		}
		guithread.Do(func() {
			obj.instance.
				Content = co
			obj.instance.
				Refresh()
		})
		return object.
			Nil, nil
	})
	scrollAttrs.Define(
		"ScrollToBottom",
	).Args().Returns("nil").Impl(func(obj *Scroll, ctx context.Context, args ...object.Object) (object.Object,

		error) {
		guithread.Do(func() {
			obj.instance.ScrollToBottom()
		})
		return object.Nil, nil
	})
	scrollAttrs.Define(
		"ScrollToTop",
	).Args().Returns("nil").Impl(func(obj *Scroll, ctx context.Context, args ...object.Object) (object.Object, error) {
		guithread.Do(func() {
			obj.instance.ScrollToTop()
		})
		return object.Nil, nil
	})
}
