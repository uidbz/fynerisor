package widget

import (
	"context"
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"github.com/uidbz/fynerisor/gui/guithread"
	"fyne.io/fyne/v2/widget"

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
}

func NewGridWrap(w WindowInterface) *GridWrap {
	instance := widget.NewGridWrap(
		func() int { return 0 },                              // length - will be set via Length()
		func() fyne.CanvasObject { return widget.NewLabel("template") }, // createItem - will be set via CreateItem()
		func(widget.GridWrapItemID, fyne.CanvasObject) {}, // updateItem - will be set via UpdateItem()
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

func (obj *GridWrap) Inspect() string {
	return "widget.GridWrap"
}

func (obj *GridWrap) Interface() interface{} {
	return obj.instance
}

func (obj *GridWrap) IsTruthy() bool {
	return true
}

func (obj *GridWrap) Cost() int {
	return 0
}

func (obj *GridWrap) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'widget.GridWrap'")
}

func (obj *GridWrap) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(GridWrapType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", GridWrapType, opType)
	return errObj, err
}

func (obj *GridWrap) Equals(other object.Object) bool {
	return obj == other
}

func (obj *GridWrap) Attrs() []object.AttrSpec {
	return nil
}

func (obj *GridWrap) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", GridWrapType, name)
}

func (obj *GridWrap) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Length":
		return object.NewBuiltin("widget.GridWrap.Length", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}

			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}

			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("gridwrap.Length: unable to get call function"), nil
			}

			obj.instance.Length = func() int {
				result, err := callFunc(ctx, fn, []object.Object{})
				if err != nil {
					fmt.Printf("GridWrap.Length error: %v\n", err)
					return 0
				}
				count, err := object.AsInt(result)
				if err != nil {
					fmt.Printf("GridWrap.Length: expected int, got %s\n", result.Type())
					return 0
				}
				return int(count)
			}

			return object.Nil, nil
		}), true

	case "CreateItem":
		return object.NewBuiltin("widget.GridWrap.CreateItem", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}

			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}

			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("gridwrap.CreateItem: unable to get call function"), nil
			}

			obj.instance.CreateItem = func() fyne.CanvasObject {
				result, err := callFunc(ctx, fn, []object.Object{})
				if err != nil {
					fmt.Printf("GridWrap.CreateItem error: %v\n", err)
					return widget.NewLabel("error")
				}

				canvasObj, ok := result.(interface{ CanvasObject() fyne.CanvasObject })
				if !ok {
					fmt.Printf("GridWrap.CreateItem: expected widget, got %s\n", result.Type())
					return widget.NewLabel("error")
				}

				return canvasObj.CanvasObject()
			}

			return object.Nil, nil
		}), true

	case "UpdateItem":
		return object.NewBuiltin("widget.GridWrap.UpdateItem", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}

			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}

			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("gridwrap.UpdateItem: unable to get call function"), nil
			}

			// UpdateItem must be synchronous to avoid VM concurrency issues
			obj.instance.UpdateItem = func(id widget.GridWrapItemID, item fyne.CanvasObject) {
				wrappedItem := wrapCanvasObject(item)
				_, err := callFunc(ctx, fn, []object.Object{
					object.NewInt(int64(id)),
					wrappedItem,
				})
				if err != nil {
					fmt.Printf("GridWrap.UpdateItem error: %v\n", err)
				}
			}

			return object.Nil, nil
		}), true

	case "OnSelected":
		return object.NewBuiltin("widget.GridWrap.OnSelected", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}

			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}

			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("gridwrap.OnSelected: unable to get call function"), nil
			}

			obj.instance.OnSelected = func(id widget.GridWrapItemID) {
				// Call synchronously to avoid race condition
				callFunc(ctx, fn, []object.Object{object.NewInt(int64(id))})
			}

			return object.Nil, nil
		}), true

	case "OnUnselected":
		return object.NewBuiltin("widget.GridWrap.OnUnselected", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}

			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}

			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("gridwrap.OnUnselected: unable to get call function"), nil
			}

			obj.instance.OnUnselected = func(id widget.GridWrapItemID) {
				// Call synchronously to avoid race condition
				callFunc(ctx, fn, []object.Object{object.NewInt(int64(id))})
			}

			return object.Nil, nil
		}), true

	case "Select":
		return object.NewBuiltin("widget.GridWrap.Select", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
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
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
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
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}

			guithread.Do(func() {
				obj.instance.UnselectAll()
			})

			return object.Nil, nil
		}), true

	case "Refresh":
		return object.NewBuiltin("widget.GridWrap.Refresh", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}

			guithread.Do(func() {
				obj.instance.Refresh()
			})

			return object.Nil, nil
		}), true
	case "Hide":
		return object.NewBuiltin("widget.GridWrap.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("widget.GridWrap.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
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

