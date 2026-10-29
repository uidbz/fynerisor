package widget

import (
	"context"
	"fmt"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/vmexec"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &List{}

const ListType object.Type = "widget.List"

// List wraps fyne's List widget - a scrolling list with virtualized rendering
type List struct {
	instance *widget.List
	w        WindowInterface
	lastLen  atomic.Int64 // last length returned by the Length callback, reused during VM contention
	// createCb is the registered template callback. The builtins that
	// register it and the render getters that call it all run on Fyne's UI
	// thread, so it needs no locking (issue #49).
	createCb *vmexec.Callback
}

func NewList(w WindowInterface) *List {
	instance := widget.NewList(
		func() int { return 0 }, // length - will be set via Length()
		func() fyne.CanvasObject { return widget.NewLabel("template") }, // createItem - will be set via CreateItem()
		func(widget.ListItemID, fyne.CanvasObject) {},                   // updateItem - will be set via UpdateItem()
	)

	return &List{
		instance: instance,
		w:        w,
	}
}

// refresh re-renders the list with fresh getter results; scheduled by the
// getters' callbacks when they hit VM contention.
func (obj *List) refresh() {
	guithread.Do(obj.instance.Refresh)
}

func (obj *List) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *List) Type() object.Type {
	return ListType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *List) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *List) inspectUI() string {
	return "widget.List"
}

func (obj *List) Interface() interface{} {
	return obj.instance
}

func (obj *List) IsTruthy() bool {
	return true
}

func (obj *List) Cost() int {
	return 0
}

func (obj *List) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'widget.List'")
}

func (obj *List) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", ListType, opType)
}

func (obj *List) Equals(other object.Object) bool {
	return obj == other
}

func (obj *List) Attrs() []object.AttrSpec {
	return nil
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *List) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *List) setAttrUI(name string, value object.Object) error {
	switch name {
	case "HideSeparators":
		b, err := object.AsBool(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		obj.instance.HideSeparators = b
		guithread.Do(func() {
			obj.instance.Refresh()
		})
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", ListType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *List) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *List) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "Length":
		return object.NewBuiltin("widget.List.Length", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}

			cb, err := vmexec.NewCallback(ctx, "widget.List.Length", args[0])
			if err != nil {
				return nil, err
			}
			cb.RefreshWhenBusy(obj.refresh)

			// Getter: Fyne calls it synchronously during layout on the UI
			// thread. While the VM is busy (ErrBusy) keep the last-known
			// length so the list doesn't collapse; the list is refreshed
			// once the VM is idle again.
			obj.instance.Length = func() int {
				result, err := cb.Call()
				if err != nil {
					cb.Report(err)
					return int(obj.lastLen.Load())
				}
				count, err := object.AsInt(result)
				if err != nil {
					cb.Report(fmt.Errorf("expected int, got %s", result.Type()))
					return int(obj.lastLen.Load())
				}
				obj.lastLen.Store(count)
				return int(obj.lastLen.Load())
			}

			return object.Nil, nil
		}), true

	case "CreateItem":
		return object.NewBuiltin("widget.List.CreateItem", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}

			cb, err := vmexec.NewCallback(ctx, "widget.List.CreateItem", args[0])
			if err != nil {
				return nil, err
			}
			cb.RefreshWhenBusy(obj.refresh)
			obj.createCb = cb

			obj.instance.CreateItem = func() fyne.CanvasObject {
				inner, ok := callTemplate(cb)
				if !ok {
					// VM busy or run stopped: never pool a fake. The slot
					// is invisible for now and takes the script's real
					// widget the first UpdateItem can run (issue #50).
					return newTemplateSlot()
				}
				return inner
			}

			return object.Nil, nil
		}), true

	case "UpdateItem":
		return object.NewBuiltin("widget.List.UpdateItem", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}

			cb, err := vmexec.NewCallback(ctx, "widget.List.UpdateItem", args[0])
			if err != nil {
				return nil, err
			}
			cb.RefreshWhenBusy(obj.refresh)

			obj.instance.UpdateItem = func(id widget.ListItemID, item fyne.CanvasObject) {
				// Getter-like: runs synchronously during render. Materialize
				// any slot left by a busy-VM template call so the script
				// always gets its own widget, then wrap the canvas object so
				// Risor can access it (issue #50).
				item, ok := resolveTemplateItem(item, func() (fyne.CanvasObject, bool) {
					return callTemplate(obj.createCb)
				})
				if !ok {
					return // VM still busy; a later refresh retries
				}
				_, err := cb.Call(object.NewInt(int64(id)), wrapCanvasObject(item))
				cb.Report(err)
			}

			return object.Nil, nil
		}), true

	case "OnSelected":
		return object.NewBuiltin("widget.List.OnSelected", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}

			cb, err := vmexec.NewCallback(ctx, "widget.List.OnSelected", args[0])
			if err != nil {
				return nil, err
			}

			// Event: queued on the script's VM goroutine, never run on (or
			// blocking) the UI thread.
			obj.instance.OnSelected = func(id widget.ListItemID) {
				cb.Fire(object.NewInt(int64(id)))
			}

			return object.Nil, nil
		}), true

	case "Refresh":
		return object.NewBuiltin("widget.List.Refresh", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}

			guithread.Do(func() {
				obj.instance.Refresh()
			})

			return object.Nil, nil
		}), true

	case "Select":
		return object.NewBuiltin("widget.List.Select", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}

			id, err := object.AsInt(args[0])
			if err != nil {
				return nil, err
			}

			guithread.Do(func() {
				obj.instance.Select(widget.ListItemID(id))
			})

			return object.Nil, nil
		}), true

	case "Unselect":
		return object.NewBuiltin("widget.List.Unselect", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}

			id, err := object.AsInt(args[0])
			if err != nil {
				return nil, err
			}

			guithread.Do(func() {
				obj.instance.Unselect(widget.ListItemID(id))
			})

			return object.Nil, nil
		}), true

	case "UnselectAll":
		return object.NewBuiltin("widget.List.UnselectAll", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}

			guithread.Do(func() {
				obj.instance.UnselectAll()
			})

			return object.Nil, nil
		}), true

	case "ScrollTo":
		return object.NewBuiltin("widget.List.ScrollTo", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}

			id, err := object.AsInt(args[0])
			if err != nil {
				return nil, err
			}

			guithread.Do(func() {
				obj.instance.ScrollTo(widget.ListItemID(id))
			})

			return object.Nil, nil
		}), true
	case "Hide":
		return object.NewBuiltin("widget.List.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("widget.List.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			guithread.Do(func() {
				obj.instance.Show()
			})
			return object.Nil, nil
		}), true
	}

	return nil, false
}

// wrapCanvasObject wraps a fyne.CanvasObject so Risor scripts can access it
// This is a simplified wrapper that exposes common widget types
func wrapCanvasObject(obj fyne.CanvasObject) object.Object {
	switch v := obj.(type) {
	case *widget.Label:
		return &Label{instance: v}
	case *widget.Card:
		return &Card{instance: v}
	case *widget.Button:
		return &Button{instance: v}
	default:
		// For now, return a generic wrapper that can't be modified
		// This allows UpdateItem to receive the item but limits what can be done with it
		return &GenericCanvasObject{obj: obj}
	}
}

// GenericCanvasObject is a minimal wrapper for canvas objects we don't have full bindings for
type GenericCanvasObject struct {
	obj fyne.CanvasObject
}

func (g *GenericCanvasObject) Type() object.Type      { return "canvas_object" }
func (g *GenericCanvasObject) Inspect() string        { return "canvas_object" }
func (g *GenericCanvasObject) Interface() interface{} { return g.obj }
func (g *GenericCanvasObject) IsTruthy() bool         { return true }
func (g *GenericCanvasObject) Cost() int              { return 0 }
func (g *GenericCanvasObject) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("cannot marshal canvas_object")
}
func (g *GenericCanvasObject) RunOperation(op.BinaryOpType, object.Object) (object.Object, error) {
	return object.Nil, fmt.Errorf("no operations")
}
func (g *GenericCanvasObject) Equals(other object.Object) bool           { return g == other }
func (g *GenericCanvasObject) GetAttr(name string) (object.Object, bool) { return nil, false }

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (g *GenericCanvasObject) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, g.setAttrUI)
}

func (g *GenericCanvasObject) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("no attributes")
}
func (g *GenericCanvasObject) Attrs() []object.AttrSpec { return nil }
