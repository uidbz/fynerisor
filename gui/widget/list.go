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

var _ object.Object = &List{}

const ListType object.Type = "widget.List"

// List wraps fyne's List widget - a scrolling list with virtualized rendering
type List struct {
	instance *widget.List
	w        WindowInterface
}

func NewList(w WindowInterface) *List {
	instance := widget.NewList(
		func() int { return 0 },              // length - will be set via Length()
		func() fyne.CanvasObject { return widget.NewLabel("template") }, // createItem - will be set via CreateItem()
		func(widget.ListItemID, fyne.CanvasObject) {}, // updateItem - will be set via UpdateItem()
	)

	return &List{
		instance: instance,
		w:        w,
	}
}

func (obj *List) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *List) Type() object.Type {
	return ListType
}

func (obj *List) Inspect() string {
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
	err := errors.New("eval error: unsupported operation for " + string(ListType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", ListType, opType)
	return errObj, err
}

func (obj *List) Equals(other object.Object) bool {
	return obj == other
}

func (obj *List) Attrs() []object.AttrSpec {
	return nil
}

func (obj *List) SetAttr(name string, value object.Object) error {
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

func (obj *List) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Length":
		return object.NewBuiltin("widget.List.Length", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}

			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}

			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("list.Length: unable to get call function"), nil
			}

			obj.instance.Length = func() int {
				result, err := callFunc(ctx, fn, []object.Object{})
				if err != nil {
					fmt.Printf("List.Length error: %v\n", err)
					return 0
				}
				count, err := object.AsInt(result)
				if err != nil {
					fmt.Printf("List.Length: expected int, got %s\n", result.Type())
					return 0
				}
				return int(count)
			}

			return object.Nil, nil
		}), true

	case "CreateItem":
		return object.NewBuiltin("widget.List.CreateItem", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}

			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}

			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("list.CreateItem: unable to get call function"), nil
			}

			obj.instance.CreateItem = func() fyne.CanvasObject {
				result, err := callFunc(ctx, fn, []object.Object{})
				if err != nil {
					fmt.Printf("List.CreateItem error: %v\n", err)
					return widget.NewLabel("error")
				}

				canvasObj, ok := result.(interface{ CanvasObject() fyne.CanvasObject })
				if !ok {
					fmt.Printf("List.CreateItem: expected widget, got %s\n", result.Type())
					return widget.NewLabel("error")
				}

				return canvasObj.CanvasObject()
			}

			return object.Nil, nil
		}), true

	case "UpdateItem":
		return object.NewBuiltin("widget.List.UpdateItem", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}

			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}

			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("list.UpdateItem: unable to get call function"), nil
			}

			obj.instance.UpdateItem = func(id widget.ListItemID, item fyne.CanvasObject) {
				// Call synchronously to avoid race condition with Length/CreateItem
				// Wrap the canvas object so Risor can access it
				wrappedItem := wrapCanvasObject(item)
				_, err := callFunc(ctx, fn, []object.Object{
					object.NewInt(int64(id)),
					wrappedItem,
				})
				if err != nil {
					fmt.Printf("List.UpdateItem error: %v\n", err)
				}
			}

			return object.Nil, nil
		}), true

	case "OnSelected":
		return object.NewBuiltin("widget.List.OnSelected", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}

			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}

			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("list.OnSelected: unable to get call function"), nil
			}

			obj.instance.OnSelected = func(id widget.ListItemID) {
				// Call synchronously to avoid race condition
				callFunc(ctx, fn, []object.Object{object.NewInt(int64(id))})
			}

			return object.Nil, nil
		}), true

	case "Refresh":
		return object.NewBuiltin("widget.List.Refresh", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}

			guithread.Do(func() {
				obj.instance.Refresh()
			})

			return object.Nil, nil
		}), true

	case "Select":
		return object.NewBuiltin("widget.List.Select", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
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
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
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
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}

			guithread.Do(func() {
				obj.instance.UnselectAll()
			})

			return object.Nil, nil
		}), true

	case "ScrollTo":
		return object.NewBuiltin("widget.List.ScrollTo", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
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
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("widget.List.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
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

func (g *GenericCanvasObject) Type() object.Type                                     { return "canvas_object" }
func (g *GenericCanvasObject) Inspect() string                                       { return "canvas_object" }
func (g *GenericCanvasObject) Interface() interface{}                                { return g.obj }
func (g *GenericCanvasObject) IsTruthy() bool                                        { return true }
func (g *GenericCanvasObject) Cost() int                                             { return 0 }
func (g *GenericCanvasObject) MarshalJSON() ([]byte, error)                          { return nil, fmt.Errorf("cannot marshal canvas_object") }
func (g *GenericCanvasObject) RunOperation(op.BinaryOpType, object.Object) (object.Object, error) { return object.Nil, fmt.Errorf("no operations") }
func (g *GenericCanvasObject) Equals(other object.Object) bool                      { return g == other }
func (g *GenericCanvasObject) GetAttr(name string) (object.Object, bool)            { return nil, false }
func (g *GenericCanvasObject) SetAttr(name string, value object.Object) error       { return fmt.Errorf("no attributes") }
func (g *GenericCanvasObject) Attrs() []object.AttrSpec                             { return nil }
