package container

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"
	"github.com/uidbz/fynerisor/gui/vmexec"

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

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *AppTabs) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *AppTabs) inspectUI() string {
	return "container.AppTabs"
}

func (obj *AppTabs) Interface() interface{} {
	return obj.instance
}

func (obj *AppTabs) IsTruthy() bool {
	return true
}

func (obj *AppTabs) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", AppTabsType, opType)
}

func (obj *AppTabs) Equals(other object.Object) bool {
	return obj == other
}

func (obj *AppTabs) Attrs() []object.AttrSpec {
	return appTabsAttrs.

		// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
		Specs()
}

func (obj *AppTabs) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *AppTabs) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", AppTabsType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *AppTabs) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var appTabsAttrs = object.NewAttrRegistry[*AppTabs](string(AppTabsType))

func (obj *AppTabs) getAttrUI(name string) (object.Object, bool) {
	return appTabsAttrs.GetAttr(obj, name)

	// Event: Fyne invokes OnSelected on the GUI thread; the script
	// callback is queued on the script's VM goroutine (recovered,
	// serialized with every other VM entry).

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
func init() {
	objattrs.Add(appTabsAttrs)
	appTabsAttrs.Define("Append").Args("arg1").Returns("nil").Impl(func(obj *AppTabs, ctx context.Context, args ...object.Object) (object.Object, error) {
		tabItem, ok := args[0].(*TabItem)
		if !ok {
			return nil, object.Errorf("type error: expected TabItem, got %s",
				args[0].Type(),
			)
		}
		guithread.Do(func() {
			obj.instance.
				Append(tabItem.instance)
		})
		return object.Nil,

			nil
	})
	appTabsAttrs.Define("Remove").Args("arg1").Returns("nil").Impl(func(obj *AppTabs, ctx context.Context, args ...object.Object) (object.Object, error) {
		tabItem, ok := args[0].(*TabItem)
		if !ok {
			return nil, object.Errorf("type error: expected TabItem, got %s",
				args[0].Type(),
			)
		}
		guithread.Do(func() {
			obj.instance.
				Remove(tabItem.instance)
		})
		return object.Nil,

			nil
	})
	appTabsAttrs.Define("RemoveIndex").Args("arg1").
		Returns("nil").Impl(func(obj *AppTabs, ctx context.
		Context, args ...object.Object) (object.
		Object, error) {
		index, err := object.AsInt(args[0])
		if err != nil {
			return nil, object.
				Errorf("type error: %v", err)
		}
		guithread.Do(
			func() {
				obj.instance.RemoveIndex(
					int(index))
			})
		return object.Nil, nil
	})
	appTabsAttrs.Define("Select").Doc("Select an index.").Args("arg1").
		Returns("nil").Impl(func(obj *AppTabs, ctx context.Context, args ...object.
		Object) (object.Object, error) {
		tabItem,

			ok := args[0].(*TabItem)
		if !ok {
			return nil,
				object.Errorf("type error: expected TabItem, got %s",

					args[0].Type())
		}
		guithread.
			Do(func() {
				obj.instance.Select(tabItem.instance)
			})
		return object.Nil,
			nil
	})
	appTabsAttrs.Define("SelectIndex").Args("arg1").
		Returns("nil").Impl(func(obj *AppTabs, ctx context.
		Context, args ...object.Object) (object.
		Object, error) {
		index, err := object.AsInt(args[0])
		if err != nil {
			return nil, object.
				Errorf("type error: %v", err)
		}
		guithread.Do(
			func() {
				obj.instance.SelectIndex(
					int(index))
			})
		return object.Nil, nil
	})
	appTabsAttrs.Define("OnSelected").Args(
		"arg1").
		Returns("nil").Impl(func(obj *AppTabs, ctx context.
		Context, args ...object.Object) (object.Object, error) {
		cb, err := vmexec.NewCallback(ctx, "container.AppTabs.OnSelected", args[0])
		if err != nil {
			return nil, err
		}
		obj.instance.
			OnSelected = func(item *container.
			TabItem) {
			cb.Fire(&TabItem{instance: item})
		}

		return object.Nil, nil
	})
	appTabsAttrs.Define("Selected").Doc("The selected index.").Args().
		Returns("nil").Impl(func(obj *AppTabs, ctx context.Context, args ...object.
		Object) (object.Object, error) {
		selected := obj.instance.Selected()
		if selected == nil {
			return object.Nil, nil
		}
		return &TabItem{
			instance: selected}, nil
	})
	appTabsAttrs.Define("SelectedIndex").Args().Returns("nil").Impl(func(obj *AppTabs, ctx context.
		Context, args ...object.Object) (object.Object, error) {
		return object.NewInt(int64(obj.
			instance.SelectedIndex())), nil
	})
	appTabsAttrs.Define("EnableIndex").Args("arg1").
		Returns("nil").Impl(func(obj *AppTabs, ctx context.
		Context, args ...object.Object) (object.
		Object, error) {
		index, err := object.AsInt(args[0])
		if err != nil {
			return nil, object.
				Errorf("type error: %v", err)
		}
		guithread.Do(
			func() {
				obj.instance.EnableIndex(
					int(index))
			})
		return object.Nil, nil
	})
	appTabsAttrs.Define("DisableIndex").Args("arg1").Returns("nil").Impl(func(obj *AppTabs, ctx context.
		Context, args ...object.Object) (object.
		Object, error) {
		index, err := object.AsInt(args[0])
		if err != nil {
			return nil, object.
				Errorf("type error: %v", err)
		}
		guithread.
			Do(
				func() {
					obj.instance.DisableIndex(int(index))
				})
		return object.Nil, nil
	})
	appTabsAttrs.Define("EnableItem").Args(
		"arg1").
		Returns("nil").Impl(func(obj *AppTabs, ctx context.
		Context, args ...object.Object) (object.Object, error) {
		tabItem, ok := args[0].(*TabItem)
		if !ok {
			return nil, object.Errorf("type error: expected TabItem, got %s",
				args[0].
					Type())
		}
		guithread.Do(func() {
			obj.
				instance.EnableItem(tabItem.instance)
		})
		return object.
			Nil, nil
	})
	appTabsAttrs.Define("DisableItem").Args("arg1").
		Returns("nil").Impl(func(obj *AppTabs, ctx context.
		Context, args ...object.Object) (object.
		Object, error) {
		tabItem, ok := args[0].(*TabItem)
		if !ok {
			return nil, object.Errorf("type error: expected TabItem, got %s",
				args[0].
					Type())
		}
		guithread.Do(func() {
			obj.
				instance.DisableItem(tabItem.instance)
		})
		return object.Nil, nil
	})
}
