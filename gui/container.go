package gui

import (
	"context"
	"errors"
	"fmt"

	"fyne.io/fyne/v2"

	risorcontainer "github.com/uidbz/fynerisor/gui/container"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &Container{}

const ContainerType object.Type = "container"

// Container is a factory object that provides container/layout creation functions to Risor scripts.
// It implements object.Object and exposes methods via GetAttr() for creating various
// Fyne containers like VBox, HBox, Border, HSplit, VSplit, and Scroll.
//
// Available in Risor scripts as the global 'container' object.
//
// Example usage in Risor:
//
//	let vbox = container.NewVBox(widget1, widget2)
//	let hbox = container.NewHBox(widget1, widget2)
//	let border = container.NewBorder(top, bottom, left, right, center)

// Container is a factory object that provides layout container creation functions to Risor scripts.
// It implements object.Object and exposes methods via GetAttr() for creating
// layout containers that organize widgets (VBox, HBox, Border, Split, Scroll).
//
// Available in Risor scripts as the global 'container' object.
//
// Example usage in Risor:
//
//	let vbox = container.NewVBox(widget1, widget2, widget3)
//	let hbox = container.NewHBox(left, center, right)
//	let border = container.NewBorder(top, bottom, left, right, center)
//	let scroll = container.NewScroll(content)
//
// Thread Safety: All container creation is thread-safe.
type Container struct {
}

func (obj *Container) Type() object.Type {
	return ContainerType
}

func (obj *Container) Inspect() string {
	return "container"
}

func (obj *Container) Interface() interface{} {
	return nil
}

func (obj *Container) IsTruthy() bool {
	return true
}

func (obj *Container) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'container'")
}

func (obj *Container) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(ContainerType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", ContainerType, opType)
	return errObj, err
}

func (obj *Container) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Container) Attrs() []object.AttrSpec {
	return nil
}

func (obj *Container) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", ContainerType, name)
}

func (obj *Container) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "NewBorder":
		return object.NewBuiltin("container.NewBorder", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) < 2 {
				return object.Errorf("wrong number of arguments. got=%d, want=2", len(args)), nil
			}

			var top, bottom, left, right, center fyne.CanvasObject
			var ok bool

			if len(args) > 0 && args[0].Type() != object.NIL {
				top, ok = args[0].Interface().(fyne.CanvasObject)
				if !ok {
					return object.Errorf("NewBorder: Expected CanvasObject at argument: %d", 0), nil
				}
			}
			if len(args) > 1 && args[1].Type() != object.NIL {
				bottom, ok = args[1].Interface().(fyne.CanvasObject)
				if !ok {
					return object.Errorf("NewBorder: Wrong type, expected CanvasObject at argument: %d", 1), nil
				}
			}
			if len(args) > 2 && args[2].Type() != object.NIL {
				left, ok = args[2].Interface().(fyne.CanvasObject)
				if !ok {
					return object.Errorf("NewBorder: Wrong type, expected CanvasObject at argument: %d", 2), nil
				}
			}
			if len(args) > 3 && args[3].Type() != object.NIL {
				right, ok = args[3].Interface().(fyne.CanvasObject)
				if !ok {
					return object.Errorf("NewBorder: Wrong type, expected CanvasObject at argument: %d", 3), nil
				}
			}
			if len(args) > 4 && args[4].Type() != object.NIL {
				center, ok = args[4].Interface().(fyne.CanvasObject)
				if !ok {
					return object.Errorf("NewBorder: Wrong type, expected CanvasObject at argument: %d", 4), nil
				}
			}

			return risorcontainer.NewBorder(top, bottom, left, right, center), nil
		}), true

	case "NewHSplit":
		return object.NewBuiltin("container.NewHSplit", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return object.Errorf("wrong number of arguments. got=%d, want=2", len(args)), nil
			}

			leading, ok := args[0].Interface().(fyne.CanvasObject)
			if !ok {
				return object.Errorf("NewHSplit: Wrong type, expected CanvasObject at argument: %d", 0), nil
			}

			trailing, ok := args[1].Interface().(fyne.CanvasObject)
			if !ok {
				return object.Errorf("NewHSplit: Wrong type, expected CanvasObject at argument: %d", 1), nil
			}

			return risorcontainer.NewHSplit(leading, trailing), nil
		}), true

	case "NewVSplit":
		return object.NewBuiltin("container.NewVSplit", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return object.Errorf("wrong number of arguments. got=%d, want=2", len(args)), nil
			}

			top, ok := args[0].Interface().(fyne.CanvasObject)
			if !ok {
				return object.Errorf("NewVSplit: Wrong type, expected CanvasObject at argument: %d", 0), nil
			}

			bottom, ok := args[1].Interface().(fyne.CanvasObject)
			if !ok {
				return object.Errorf("NewVSplit: Wrong type, expected CanvasObject at argument: %d", 1), nil
			}

			return risorcontainer.NewVSplit(top, bottom), nil
		}), true

	case "NewHBox":
		return object.NewBuiltin("container.NewHBox", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) == 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=at least 1", len(args)), nil
			}

			var objects []fyne.CanvasObject

			for i, x := range args {
				if x.Type() == object.LIST {
					items, err := object.AsList(args[0])
					if err != nil {
						return nil, err
					}
					for j, y := range items.Value() {
						o, ok := y.Interface().(fyne.CanvasObject)
						if !ok {
							return object.Errorf("NewHBox: Wrong type, expected CanvasObject within list at position: %d", j), nil
						}
						objects = append(objects, o)
					}
				} else {
					o, ok := x.Interface().(fyne.CanvasObject)
					if !ok {
						return object.Errorf("NewHBox: Wrong type, expected CanvasObject at argument: %d", i), nil
					}
					objects = append(objects, o)
				}
			}

			return risorcontainer.NewHBox(objects...), nil
		}), true

	case "NewVBox":
		return object.NewBuiltin("container.NewVBox", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) == 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=at least 1", len(args)), nil
			}

			var objects []fyne.CanvasObject

			for i, x := range args {
				if x.Type() == object.LIST {
					items, err := object.AsList(args[0])
					if err != nil {
						return nil, err
					}
					for j, y := range items.Value() {
						o, ok := y.Interface().(fyne.CanvasObject)
						if !ok {
							return object.Errorf("NewVBox: Wrong type, expected CanvasObject within list at position: %d", j), nil
						}
						objects = append(objects, o)
					}
				} else {
					o, ok := x.Interface().(fyne.CanvasObject)
					if !ok {
						return object.Errorf("NewVBox: Wrong type, expected CanvasObject at argument: %d", i), nil
					}
					objects = append(objects, o)
				}
			}

			return risorcontainer.NewVBox(objects...), nil
		}), true

	case "NewScroll":
		return object.NewBuiltin("container.NewScroll", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}

			content, ok := args[0].Interface().(fyne.CanvasObject)
			if !ok {
				return object.Errorf("NewScroll: Wrong type, expected CanvasObject at argument: %d", 0), nil
			}

			return risorcontainer.NewScroll(content), nil
		}), true

	case "NewCenter":
		return object.NewBuiltin("container.NewCenter", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) == 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=at least 1", len(args)), nil
			}

			var objects []fyne.CanvasObject
			for i, x := range args {
				o, ok := x.Interface().(fyne.CanvasObject)
				if !ok {
					return object.Errorf("NewCenter: Wrong type, expected CanvasObject at argument: %d", i), nil
				}
				objects = append(objects, o)
			}

			return risorcontainer.NewCenter(objects...), nil
		}), true

	case "NewMax":
		return object.NewBuiltin("container.NewMax", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) == 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=at least 1", len(args)), nil
			}

			var objects []fyne.CanvasObject
			for i, x := range args {
				o, ok := x.Interface().(fyne.CanvasObject)
				if !ok {
					return object.Errorf("NewMax: Wrong type, expected CanvasObject at argument: %d", i), nil
				}
				objects = append(objects, o)
			}

			return risorcontainer.NewMax(objects...), nil
		}), true

	case "NewStack":
		return object.NewBuiltin("container.NewStack", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) == 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=at least 1", len(args)), nil
			}

			var objects []fyne.CanvasObject
			for i, x := range args {
				o, ok := x.Interface().(fyne.CanvasObject)
				if !ok {
					return object.Errorf("NewStack: Wrong type, expected CanvasObject at argument: %d", i), nil
				}
				objects = append(objects, o)
			}

			return risorcontainer.NewStack(objects...), nil
		}), true

	case "NewPadded":
		return object.NewBuiltin("container.NewPadded", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) == 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=at least 1", len(args)), nil
			}

			var objects []fyne.CanvasObject
			for i, x := range args {
				o, ok := x.Interface().(fyne.CanvasObject)
				if !ok {
					return object.Errorf("NewPadded: Wrong type, expected CanvasObject at argument: %d", i), nil
				}
				objects = append(objects, o)
			}

			return risorcontainer.NewPadded(objects...), nil
		}), true

	case "NewGridWithColumns":
		return object.NewBuiltin("container.NewGridWithColumns", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) < 2 {
				return object.Errorf("wrong number of arguments. got=%d, want=at least 2", len(args)), nil
			}

			cols, err := object.AsInt(args[0])
			if err != nil {
				return object.Errorf("NewGridWithColumns: first argument must be int (columns)"), nil
			}

			var objects []fyne.CanvasObject
			for i, x := range args[1:] {
				o, ok := x.Interface().(fyne.CanvasObject)
				if !ok {
					return object.Errorf("NewGridWithColumns: Wrong type, expected CanvasObject at argument: %d", i+1), nil
				}
				objects = append(objects, o)
			}

			return risorcontainer.NewGridWithColumns(int(cols), objects...), nil
		}), true

	case "NewGridWithRows":
		return object.NewBuiltin("container.NewGridWithRows", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) < 2 {
				return object.Errorf("wrong number of arguments. got=%d, want=at least 2", len(args)), nil
			}

			rows, err := object.AsInt(args[0])
			if err != nil {
				return object.Errorf("NewGridWithRows: first argument must be int (rows)"), nil
			}

			var objects []fyne.CanvasObject
			for i, x := range args[1:] {
				o, ok := x.Interface().(fyne.CanvasObject)
				if !ok {
					return object.Errorf("NewGridWithRows: Wrong type, expected CanvasObject at argument: %d", i+1), nil
				}
				objects = append(objects, o)
			}

			return risorcontainer.NewGridWithRows(int(rows), objects...), nil
		}), true

	}
	return nil, false
}

func NewFyneContainer() *Container {
	return &Container{}
}
