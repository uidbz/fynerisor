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

var _ object.Object = &GridWrap{}

const GridWrapType object.Type = "widget.GridWrap"

// GridWrap wraps fyne's GridWrap widget - a grid layout with virtualized rendering.
//
// GridWrap efficiently displays many items in a grid by only rendering visible items.
// It requires three callback functions to provide data:
//   - Length(): Returns total item count
//   - CreateItem(): Creates a template widget
//   - UpdateItem(id, item): Populates a specific item
//
// Example usage in Risor:
//
//	let grid = widget.NewGridWrap()
//	grid.Length(() => 100)
//	grid.CreateItem(() => widget.NewLabel(""))
//	grid.UpdateItem((id, item) => {
//	    item.SetText(sprintf("Item %d", id))
//	})
//	grid.Refresh()
//
// Supports selection with OnSelected/OnUnselected callbacks.
type GridWrap struct {
	instance *widget.GridWrap
	w        WindowInterface
	lastLen  atomic.Int64 // last length returned by the Length callback, reused during VM contention
	// createCb is the registered template callback; UI thread only (see List).
	createCb *vmexec.Callback
}

// refresh re-renders the grid with fresh getter results; scheduled by the
// getters' callbacks when they hit VM contention.
func (obj *GridWrap) refresh() {
	guithread.Do(obj.instance.Refresh)
}

func NewGridWrap(w WindowInterface) *GridWrap {
	instance := widget.NewGridWrap(
		func() int { return 0 }, // length - will be set via Length()
		func() fyne.CanvasObject { return widget.NewLabel("template") }, // createItem - will be set via CreateItem()
		func(widget.GridWrapItemID, fyne.CanvasObject) {},               // updateItem - will be set via UpdateItem()
	)

	return &GridWrap{
		instance: instance,
		w:        w,
	}
}

func (obj *GridWrap) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *GridWrap) Type() object.Type {
	return GridWrapType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *GridWrap) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *GridWrap) inspectUI() string {
	return "widget.GridWrap"
}

func (obj *GridWrap) Interface() interface{} {
	return obj.instance
}

func (obj *GridWrap) IsTruthy() bool {
	return true
}

func (obj *GridWrap) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", GridWrapType, opType)
}

func (obj *GridWrap) Equals(other object.Object) bool {
	return obj == other
}

func (obj *GridWrap) Attrs() []object.AttrSpec {
	return nil
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *GridWrap) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *GridWrap) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", GridWrapType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *GridWrap) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *GridWrap) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "Length":
		return object.NewBuiltin("widget.GridWrap.Length", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}

			cb, err := vmexec.NewCallback(ctx, "widget.GridWrap.Length", args[0])
			if err != nil {
				return nil, err
			}
			cb.RefreshWhenBusy(obj.refresh)

			// Getter: called synchronously by Fyne during layout.
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
		return object.NewBuiltin("widget.GridWrap.CreateItem", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}

			cb, err := vmexec.NewCallback(ctx, "widget.GridWrap.CreateItem", args[0])
			if err != nil {
				return nil, err
			}
			cb.RefreshWhenBusy(obj.refresh)
			obj.createCb = cb

			// Getter: called synchronously by Fyne during layout.
			obj.instance.CreateItem = func() fyne.CanvasObject {
				inner, ok := callTemplate(cb)
				if !ok {
					// VM busy or run stopped: never pool a fake (issue #50).
					return newTemplateSlot()
				}
				return inner
			}

			return object.Nil, nil
		}), true

	case "UpdateItem":
		return object.NewBuiltin("widget.GridWrap.UpdateItem", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}

			cb, err := vmexec.NewCallback(ctx, "widget.GridWrap.UpdateItem", args[0])
			if err != nil {
				return nil, err
			}
			cb.RefreshWhenBusy(obj.refresh)

			// Getter-like: called synchronously by Fyne during render.
			obj.instance.UpdateItem = func(id widget.GridWrapItemID, item fyne.CanvasObject) {
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
		return object.NewBuiltin("widget.GridWrap.OnSelected", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}

			cb, err := vmexec.NewCallback(ctx, "widget.GridWrap.OnSelected", args[0])
			if err != nil {
				return nil, err
			}

			// Event: queued on the script's VM goroutine.
			obj.instance.OnSelected = func(id widget.GridWrapItemID) {
				cb.Fire(object.NewInt(int64(id)))
			}

			return object.Nil, nil
		}), true

	case "OnUnselected":
		return object.NewBuiltin("widget.GridWrap.OnUnselected", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}

			cb, err := vmexec.NewCallback(ctx, "widget.GridWrap.OnUnselected", args[0])
			if err != nil {
				return nil, err
			}

			// Event: queued on the script's VM goroutine.
			obj.instance.OnUnselected = func(id widget.GridWrapItemID) {
				cb.Fire(object.NewInt(int64(id)))
			}

			return object.Nil, nil
		}), true

	case "Select":
		return object.NewBuiltin("widget.GridWrap.Select", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}

			id, err := object.AsInt(args[0])
			if err != nil {
				return nil, err
			}

			guithread.Do(func() {
				obj.instance.Select(widget.GridWrapItemID(id))
			})

			return object.Nil, nil
		}), true

	case "Unselect":
		return object.NewBuiltin("widget.GridWrap.Unselect", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}

			id, err := object.AsInt(args[0])
			if err != nil {
				return nil, err
			}

			guithread.Do(func() {
				obj.instance.Unselect(widget.GridWrapItemID(id))
			})

			return object.Nil, nil
		}), true

	case "UnselectAll":
		return object.NewBuiltin("widget.GridWrap.UnselectAll", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}

			guithread.Do(func() {
				obj.instance.UnselectAll()
			})

			return object.Nil, nil
		}), true

	case "Refresh":
		return object.NewBuiltin("widget.GridWrap.Refresh", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}

			guithread.Do(func() {
				obj.instance.Refresh()
			})

			return object.Nil, nil
		}), true
	case "Hide":
		return object.NewBuiltin("widget.GridWrap.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("widget.GridWrap.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
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
