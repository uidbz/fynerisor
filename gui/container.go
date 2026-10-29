package gui

import (
	"context"
	"fmt"
	"github.com/uidbz/fynerisor/gui/guithread"

	"fyne.io/fyne/v2"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
	risorcontainer "github.com/uidbz/fynerisor/gui/container"
)

var _ object.Object = &Container{}

const ContainerType object.Type = "container"

// Container is a factory object that provides container/layout creation functions to Risor scripts.
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

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Container) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Container) inspectUI() string {
	return "container"
}

func (obj *Container) Interface() interface{} {
	return nil
}

func (obj *Container) IsTruthy() bool {
	return true
}

func (obj *Container) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", ContainerType, opType)
}

func (obj *Container) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Container) Attrs() []object.AttrSpec {
	return nil
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Container) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Container) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", ContainerType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Container) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *Container) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "NewBorder":
		return object.NewBuiltin("container.NewBorder", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) < 2 || len(args) > 5 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=2 to 5", len(args))
			}

			var top, bottom, left, right, center fyne.CanvasObject
			var ok bool

			if len(args) > 0 && args[0].Type() != object.NIL {
				top, ok = args[0].Interface().(fyne.CanvasObject)
				if !ok {
					return nil, object.Errorf("NewBorder: Expected CanvasObject at argument: %d", 0)
				}
			}
			if len(args) > 1 && args[1].Type() != object.NIL {
				bottom, ok = args[1].Interface().(fyne.CanvasObject)
				if !ok {
					return nil, object.Errorf("NewBorder: Wrong type, expected CanvasObject at argument: %d", 1)
				}
			}
			if len(args) > 2 && args[2].Type() != object.NIL {
				left, ok = args[2].Interface().(fyne.CanvasObject)
				if !ok {
					return nil, object.Errorf("NewBorder: Wrong type, expected CanvasObject at argument: %d", 2)
				}
			}
			if len(args) > 3 && args[3].Type() != object.NIL {
				right, ok = args[3].Interface().(fyne.CanvasObject)
				if !ok {
					return nil, object.Errorf("NewBorder: Wrong type, expected CanvasObject at argument: %d", 3)
				}
			}
			if len(args) > 4 && args[4].Type() != object.NIL {
				center, ok = args[4].Interface().(fyne.CanvasObject)
				if !ok {
					return nil, object.Errorf("NewBorder: Wrong type, expected CanvasObject at argument: %d", 4)
				}
			}

			return risorcontainer.NewBorder(top, bottom, left, right, center), nil
		}), true

	case "NewHSplit":
		return object.NewBuiltin("container.NewHSplit", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=2", len(args))
			}

			leading, ok := args[0].Interface().(fyne.CanvasObject)
			if !ok {
				return nil, object.Errorf("NewHSplit: Wrong type, expected CanvasObject at argument: %d", 0)
			}

			trailing, ok := args[1].Interface().(fyne.CanvasObject)
			if !ok {
				return nil, object.Errorf("NewHSplit: Wrong type, expected CanvasObject at argument: %d", 1)
			}

			return risorcontainer.NewHSplit(leading, trailing), nil
		}), true

	case "NewVSplit":
		return object.NewBuiltin("container.NewVSplit", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=2", len(args))
			}

			top, ok := args[0].Interface().(fyne.CanvasObject)
			if !ok {
				return nil, object.Errorf("NewVSplit: Wrong type, expected CanvasObject at argument: %d", 0)
			}

			bottom, ok := args[1].Interface().(fyne.CanvasObject)
			if !ok {
				return nil, object.Errorf("NewVSplit: Wrong type, expected CanvasObject at argument: %d", 1)
			}

			return risorcontainer.NewVSplit(top, bottom), nil
		}), true

	case "NewHBox":
		return object.NewBuiltin("container.NewHBox", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) == 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=at least 1", len(args))
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
							return nil, object.Errorf("NewHBox: Wrong type, expected CanvasObject within list at position: %d", j)
						}
						objects = append(objects, o)
					}
				} else {
					o, ok := x.Interface().(fyne.CanvasObject)
					if !ok {
						return nil, object.Errorf("NewHBox: Wrong type, expected CanvasObject at argument: %d", i)
					}
					objects = append(objects, o)
				}
			}

			return risorcontainer.NewHBox(objects...), nil
		}), true

	case "NewVBox":
		return object.NewBuiltin("container.NewVBox", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) == 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=at least 1", len(args))
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
							return nil, object.Errorf("NewVBox: Wrong type, expected CanvasObject within list at position: %d", j)
						}
						objects = append(objects, o)
					}
				} else {
					o, ok := x.Interface().(fyne.CanvasObject)
					if !ok {
						return nil, object.Errorf("NewVBox: Wrong type, expected CanvasObject at argument: %d", i)
					}
					objects = append(objects, o)
				}
			}

			return risorcontainer.NewVBox(objects...), nil
		}), true

	case "NewScroll":
		return object.NewBuiltin("container.NewScroll", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=1", len(args))
			}

			content, ok := args[0].Interface().(fyne.CanvasObject)
			if !ok {
				return nil, object.Errorf("NewScroll: Wrong type, expected CanvasObject at argument: %d", 0)
			}

			return risorcontainer.NewScroll(content), nil
		}), true

	case "NewCenter":
		return object.NewBuiltin("container.NewCenter", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) == 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=at least 1", len(args))
			}

			var objects []fyne.CanvasObject
			for i, x := range args {
				o, ok := x.Interface().(fyne.CanvasObject)
				if !ok {
					return nil, object.Errorf("NewCenter: Wrong type, expected CanvasObject at argument: %d", i)
				}
				objects = append(objects, o)
			}

			return risorcontainer.NewCenter(objects...), nil
		}), true

	case "NewMax":
		return object.NewBuiltin("container.NewMax", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) == 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=at least 1", len(args))
			}

			var objects []fyne.CanvasObject
			for i, x := range args {
				o, ok := x.Interface().(fyne.CanvasObject)
				if !ok {
					return nil, object.Errorf("NewMax: Wrong type, expected CanvasObject at argument: %d", i)
				}
				objects = append(objects, o)
			}

			return risorcontainer.NewMax(objects...), nil
		}), true

	case "NewStack":
		return object.NewBuiltin("container.NewStack", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) == 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=at least 1", len(args))
			}

			var objects []fyne.CanvasObject
			for i, x := range args {
				o, ok := x.Interface().(fyne.CanvasObject)
				if !ok {
					return nil, object.Errorf("NewStack: Wrong type, expected CanvasObject at argument: %d", i)
				}
				objects = append(objects, o)
			}

			return risorcontainer.NewStack(objects...), nil
		}), true

	case "NewPadded":
		return object.NewBuiltin("container.NewPadded", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) == 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=at least 1", len(args))
			}

			var objects []fyne.CanvasObject
			for i, x := range args {
				o, ok := x.Interface().(fyne.CanvasObject)
				if !ok {
					return nil, object.Errorf("NewPadded: Wrong type, expected CanvasObject at argument: %d", i)
				}
				objects = append(objects, o)
			}

			return risorcontainer.NewPadded(objects...), nil
		}), true

	case "NewGridWithColumns":
		return object.NewBuiltin("container.NewGridWithColumns", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) < 2 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=at least 2", len(args))
			}

			cols, err := object.AsInt(args[0])
			if err != nil {
				return nil, object.Errorf("NewGridWithColumns: first argument must be int (columns)")
			}

			var objects []fyne.CanvasObject
			for i, x := range args[1:] {
				if x.Type() == object.LIST {
					items, err := object.AsList(x)
					if err != nil {
						return nil, err
					}
					for j, y := range items.Value() {
						o, ok := y.Interface().(fyne.CanvasObject)
						if !ok {
							return nil, object.Errorf("NewGridWithColumns: Wrong type, expected CanvasObject within list at position: %d", j)
						}
						objects = append(objects, o)
					}
					continue
				}
				o, ok := x.Interface().(fyne.CanvasObject)
				if !ok {
					return nil, object.Errorf("NewGridWithColumns: Wrong type, expected CanvasObject at argument: %d", i+1)
				}
				objects = append(objects, o)
			}

			return risorcontainer.NewGridWithColumns(int(cols), objects...), nil
		}), true

	case "NewGridWithRows":
		return object.NewBuiltin("container.NewGridWithRows", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) < 2 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=at least 2", len(args))
			}

			rows, err := object.AsInt(args[0])
			if err != nil {
				return nil, object.Errorf("NewGridWithRows: first argument must be int (rows)")
			}

			var objects []fyne.CanvasObject
			for i, x := range args[1:] {
				if x.Type() == object.LIST {
					items, err := object.AsList(x)
					if err != nil {
						return nil, err
					}
					for j, y := range items.Value() {
						o, ok := y.Interface().(fyne.CanvasObject)
						if !ok {
							return nil, object.Errorf("NewGridWithRows: Wrong type, expected CanvasObject within list at position: %d", j)
						}
						objects = append(objects, o)
					}
					continue
				}
				o, ok := x.Interface().(fyne.CanvasObject)
				if !ok {
					return nil, object.Errorf("NewGridWithRows: Wrong type, expected CanvasObject at argument: %d", i+1)
				}
				objects = append(objects, o)
			}

			return risorcontainer.NewGridWithRows(int(rows), objects...), nil
		}), true

	case "NewAppTabs":
		return object.NewBuiltin("container.NewAppTabs", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			var tabItems []*risorcontainer.TabItem
			for i, arg := range args {
				tabItem, ok := arg.(*risorcontainer.TabItem)
				if !ok {
					return nil, object.Errorf("NewAppTabs: expected TabItem at argument %d, got %s", i, arg.Type())
				}
				tabItems = append(tabItems, tabItem)
			}
			return risorcontainer.NewAppTabsFromItems(tabItems...), nil
		}), true

	case "NewTabItem":
		return object.NewBuiltin("container.NewTabItem", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=2", len(args))
			}
			text, err := object.AsString(args[0])
			if err != nil {
				return nil, object.Errorf("NewTabItem: first argument must be string (text)")
			}
			content, ok := args[1].Interface().(fyne.CanvasObject)
			if !ok {
				return nil, object.Errorf("NewTabItem: second argument must be CanvasObject")
			}
			return risorcontainer.NewTabItem(text, content), nil
		}), true

	case "NewTabItemWithIcon":
		return object.NewBuiltin("container.NewTabItemWithIcon", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 3 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=3", len(args))
			}
			text, err := object.AsString(args[0])
			if err != nil {
				return nil, object.Errorf("NewTabItemWithIcon: first argument must be string (text)")
			}
			icon, ok := args[1].Interface().(fyne.Resource)
			if !ok {
				return nil, object.Errorf("NewTabItemWithIcon: second argument must be Resource (icon)")
			}
			content, ok := args[2].Interface().(fyne.CanvasObject)
			if !ok {
				return nil, object.Errorf("NewTabItemWithIcon: third argument must be CanvasObject")
			}
			return risorcontainer.NewTabItemWithIcon(text, icon, content), nil
		}), true

	}
	return nil, false
}

func NewFyneContainer() *Container {
	return &Container{}
}
