package container

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"fyne.io/fyne/v2"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/vmguard"
	"fyne.io/fyne/v2/container"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &AppTabs{}

const AppTabsType object.Type = "container.AppTabs"

type AppTabs struct {
	instance *container.AppTabs
}

func (obj *AppTabs) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *AppTabs) Type() object.Type {
	return AppTabsType
}

func (obj *AppTabs) Inspect() string {
	return "container.AppTabs"
}

func (obj *AppTabs) Interface() interface{} {
	return obj.instance
}

func (obj *AppTabs) IsTruthy() bool {
	return true
}

func (obj *AppTabs) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'container.AppTabs'")
}

func (obj *AppTabs) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(AppTabsType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", AppTabsType, opType)
	return errObj, err
}

func (obj *AppTabs) Equals(other object.Object) bool {
	return obj == other
}

func (obj *AppTabs) Attrs() []object.AttrSpec {
	return nil
}

func (obj *AppTabs) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", AppTabsType, name)
}

func (obj *AppTabs) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Append":
		return object.NewBuiltin("container.AppTabs.Append", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			tabItem, ok := args[0].(*TabItem)
			if !ok {
				return object.Errorf("type error: expected TabItem, got %s", args[0].Type()), nil
			}
			guithread.Do(func() {
				obj.instance.Append(tabItem.instance)
			})
			return object.Nil, nil
		}), true

	case "Remove":
		return object.NewBuiltin("container.AppTabs.Remove", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			tabItem, ok := args[0].(*TabItem)
			if !ok {
				return object.Errorf("type error: expected TabItem, got %s", args[0].Type()), nil
			}
			guithread.Do(func() {
				obj.instance.Remove(tabItem.instance)
			})
			return object.Nil, nil
		}), true

	case "RemoveIndex":
		return object.NewBuiltin("container.AppTabs.RemoveIndex", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			index, err := object.AsInt(args[0])
			if err != nil {
				return object.Errorf("type error: %v", err), nil
			}
			guithread.Do(func() {
				obj.instance.RemoveIndex(int(index))
			})
			return object.Nil, nil
		}), true

	case "Select":
		return object.NewBuiltin("container.AppTabs.Select", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			tabItem, ok := args[0].(*TabItem)
			if !ok {
				return object.Errorf("type error: expected TabItem, got %s", args[0].Type()), nil
			}
			guithread.Do(func() {
				obj.instance.Select(tabItem.instance)
			})
			return object.Nil, nil
		}), true

	case "SelectIndex":
		return object.NewBuiltin("container.AppTabs.SelectIndex", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			index, err := object.AsInt(args[0])
			if err != nil {
				return object.Errorf("type error: %v", err), nil
			}
			guithread.Do(func() {
				obj.instance.SelectIndex(int(index))
			})
			return object.Nil, nil
		}), true

	case "OnSelected":
		return object.NewBuiltin("container.AppTabs.OnSelected", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("type error: expected function, got %s", args[0].Type()), nil
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("container.AppTabs.OnSelected: unable to get call function"), nil
			}
			// Fyne invokes OnSelected on the GUI thread when the active tab
			// changes. Run the script callback through the VM guard (with panic
			// recovery, mirroring safeCall) so a faulty callback can't crash the
			// process or corrupt the single-threaded VM. guithread.Do keeps the
			// call on the GUI thread whether Fyne dispatches inline or queued.
			obj.instance.OnSelected = func(item *container.TabItem) {
				guithread.Do(func() {
					defer func() {
						if r := recover(); r != nil {
							fmt.Fprintf(os.Stderr, "recovered panic in OnSelected callback: %v\n%s\n", r, debug.Stack())
						}
					}()
					vmguard.Call(callFunc, ctx, fn, []object.Object{&TabItem{instance: item}})
				})
			}
			return object.Nil, nil
		}), true

	case "Selected":
		return object.NewBuiltin("container.AppTabs.Selected", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			selected := obj.instance.Selected()
			if selected == nil {
				return object.Nil, nil
			}
			return &TabItem{instance: selected}, nil
		}), true

	case "SelectedIndex":
		return object.NewBuiltin("container.AppTabs.SelectedIndex", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			return object.NewInt(int64(obj.instance.SelectedIndex())), nil
		}), true

	case "EnableIndex":
		return object.NewBuiltin("container.AppTabs.EnableIndex", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			index, err := object.AsInt(args[0])
			if err != nil {
				return object.Errorf("type error: %v", err), nil
			}
			guithread.Do(func() {
				obj.instance.EnableIndex(int(index))
			})
			return object.Nil, nil
		}), true

	case "DisableIndex":
		return object.NewBuiltin("container.AppTabs.DisableIndex", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			index, err := object.AsInt(args[0])
			if err != nil {
				return object.Errorf("type error: %v", err), nil
			}
			guithread.Do(func() {
				obj.instance.DisableIndex(int(index))
			})
			return object.Nil, nil
		}), true

	case "EnableItem":
		return object.NewBuiltin("container.AppTabs.EnableItem", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			tabItem, ok := args[0].(*TabItem)
			if !ok {
				return object.Errorf("type error: expected TabItem, got %s", args[0].Type()), nil
			}
			guithread.Do(func() {
				obj.instance.EnableItem(tabItem.instance)
			})
			return object.Nil, nil
		}), true

	case "DisableItem":
		return object.NewBuiltin("container.AppTabs.DisableItem", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			tabItem, ok := args[0].(*TabItem)
			if !ok {
				return object.Errorf("type error: expected TabItem, got %s", args[0].Type()), nil
			}
			guithread.Do(func() {
				obj.instance.DisableItem(tabItem.instance)
			})
			return object.Nil, nil
		}), true

	case "Hide":
		return object.NewBuiltin("container.AppTabs.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Hide()
			})
			return object.Nil, nil
		}), true

	case "Show":
		return object.NewBuiltin("container.AppTabs.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
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

func NewAppTabs(items ...*container.TabItem) *AppTabs {
	return &AppTabs{
		instance: container.NewAppTabs(items...),
	}
}

func NewAppTabsFromItems(items ...*TabItem) *AppTabs {
	tabItems := make([]*container.TabItem, len(items))
	for i, item := range items {
		tabItems[i] = item.instance
	}
	return &AppTabs{
		instance: container.NewAppTabs(tabItems...),
	}
}
