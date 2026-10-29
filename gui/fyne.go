package gui

import (
	"context"
	"fmt"
	"github.com/uidbz/fynerisor/gui/guithread"

	"fyne.io/fyne/v2"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
	risorwidget "github.com/uidbz/fynerisor/gui/widget"
)

const FyneType object.Type = "fyne"

// Fyne provides access to fyne package types like Menu and MenuItem.
// Available in scripts as the global 'fyne' object.
type Fyne struct {
	w *Window
}

func (obj *Fyne) Type() object.Type {
	return FyneType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Fyne) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Fyne) inspectUI() string {
	return "fyne"
}

func (obj *Fyne) Interface() interface{} {
	return nil
}

func (obj *Fyne) IsTruthy() bool {
	return true
}

func (obj *Fyne) Cost() int {
	return 0
}

func (obj *Fyne) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'fyne'")
}

func (obj *Fyne) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", FyneType, opType)
}

func (obj *Fyne) Equals(other object.Object) bool {
	return obj == other
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Fyne) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Fyne) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", FyneType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Fyne) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *Fyne) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "NewMenuItem":
		return object.NewBuiltin("fyne.NewMenuItem", risorwidget.NewMenuItem(obj.w)), true

	case "NewMenuItemSeparator":
		return object.NewBuiltin("fyne.NewMenuItemSeparator", risorwidget.NewMenuItemSeparator), true

	case "NewMenu":
		return object.NewBuiltin("fyne.NewMenu", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) < 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=at least 1", len(args))
			}

			label, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}

			// Collect MenuItems
			items := make([]*fyne.MenuItem, 0, len(args)-1)
			for i := 1; i < len(args); i++ {
				menuItemObj, ok := args[i].(*risorwidget.MenuItem)
				if !ok {
					return nil, object.Errorf("argument error: expected MenuItem, got %s", args[i].Type())
				}
				items = append(items, menuItemObj.Instance())
			}

			menu := fyne.NewMenu(label, items...)

			return &Menu{instance: menu}, nil
		}), true

	case "NewMainMenu":
		return object.NewBuiltin("fyne.NewMainMenu", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) < 1 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=at least 1", len(args))
			}

			// Collect Menus from list
			listObj, ok := args[0].(*object.List)
			if !ok {
				return nil, object.Errorf("argument error: expected list of menus, got %s", args[0].Type())
			}

			menus := make([]*fyne.Menu, 0, len(listObj.Value()))
			for _, item := range listObj.Value() {
				menuObj, ok := item.(*Menu)
				if !ok {
					return nil, object.Errorf("argument error: expected Menu in list, got %s", item.Type())
				}
				menus = append(menus, menuObj.instance)
			}

			mainMenu := fyne.NewMainMenu(menus...)

			return &MainMenu{instance: mainMenu}, nil
		}), true
	}
	return nil, false
}

func (obj *Fyne) Attrs() []object.AttrSpec {
	return []object.AttrSpec{
		{Name: "NewMenuItem"},
		{Name: "NewMenuItemSeparator"},
		{Name: "NewMenu"},
		{Name: "NewMainMenu"},
	}
}

// Menu wraps fyne.Menu
type Menu struct {
	instance *fyne.Menu
}

func (obj *Menu) Instance() *fyne.Menu {
	return obj.instance
}

func (obj *Menu) Type() object.Type {
	return "fyne.Menu"
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Menu) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Menu) inspectUI() string {
	return fmt.Sprintf("fyne.Menu(%s)", obj.instance.Label)
}

func (obj *Menu) Interface() interface{} {
	return obj.instance
}

func (obj *Menu) IsTruthy() bool {
	return true
}

func (obj *Menu) Cost() int {
	return 0
}

func (obj *Menu) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'fyne.Menu'")
}

func (obj *Menu) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for Menu: %v", opType)
}

func (obj *Menu) Equals(other object.Object) bool {
	return obj == other
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Menu) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Menu) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: Menu object has no writable attributes")
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Menu) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *Menu) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "Label":
		return object.NewString(obj.instance.Label), true

	case "Refresh":
		return object.NewBuiltin("fyne.Menu.Refresh", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			obj.instance.Refresh()
			return object.Nil, nil
		}), true
	}
	return nil, false
}

func (obj *Menu) Attrs() []object.AttrSpec {
	return []object.AttrSpec{
		{Name: "Label"},
		{Name: "Refresh"},
	}
}

// MainMenu wraps fyne.MainMenu
type MainMenu struct {
	instance *fyne.MainMenu
}

func (obj *MainMenu) Instance() *fyne.MainMenu {
	return obj.instance
}

func (obj *MainMenu) Type() object.Type {
	return "fyne.MainMenu"
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *MainMenu) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *MainMenu) inspectUI() string {
	return "fyne.MainMenu"
}

func (obj *MainMenu) Interface() interface{} {
	return obj.instance
}

func (obj *MainMenu) IsTruthy() bool {
	return true
}

func (obj *MainMenu) Cost() int {
	return 0
}

func (obj *MainMenu) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'fyne.MainMenu'")
}

func (obj *MainMenu) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for MainMenu: %v", opType)
}

func (obj *MainMenu) Equals(other object.Object) bool {
	return obj == other
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *MainMenu) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *MainMenu) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: MainMenu object has no writable attributes")
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *MainMenu) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *MainMenu) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "Refresh":
		return object.NewBuiltin("fyne.MainMenu.Refresh", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			obj.instance.Refresh()
			return object.Nil, nil
		}), true
	}
	return nil, false
}

func (obj *MainMenu) Attrs() []object.AttrSpec {
	return []object.AttrSpec{
		{Name: "Refresh"},
	}
}
