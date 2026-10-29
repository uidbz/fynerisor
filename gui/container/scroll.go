package container

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2/container"

	"fyne.io/fyne/v2"
	"github.com/uidbz/fynerisor/gui/guithread"

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

func (obj *Scroll) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'container.scroll'")
}

func (obj *Scroll) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", ScrollType, opType)
}

func (obj *Scroll) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Scroll) Attrs() []object.AttrSpec {
	return nil
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
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

func (obj *Scroll) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "SetContent":
		return object.NewBuiltin("container.Scroll.SetContent", func(_ context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}
			co, ok := canvasArg(args[0])
			if !ok {
				return nil, object.Errorf("SetContent: expected canvas object, got %s", args[0].Type())
			}
			guithread.Do(func() {
				obj.instance.Content = co
				obj.instance.Refresh()
			})
			return object.Nil, nil
		}), true
	case "Refresh":
		return object.NewBuiltin("container.Scroll.Refresh", func(_ context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			guithread.Do(func() { obj.instance.Refresh() })
			return object.Nil, nil
		}), true
	case "Hide":
		return object.NewBuiltin("container.Scroll.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("container.Scroll.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			guithread.Do(func() {
				obj.instance.Show()
			})
			return object.Nil, nil
		}), true
	case "ScrollToBottom":
		return object.NewBuiltin("container.Scroll.ScrollToBottom", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			guithread.Do(func() {
				obj.instance.ScrollToBottom()
			})
			return object.Nil, nil
		}), true
	case "ScrollToTop":
		return object.NewBuiltin("container.Scroll.ScrollToTop", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			guithread.Do(func() {
				obj.instance.ScrollToTop()
			})
			return object.Nil, nil
		}), true
	}
	return nil, false
}

func NewScroll(content fyne.CanvasObject) *Scroll {
	return &Scroll{
		instance: container.NewScroll(content),
	}
}
