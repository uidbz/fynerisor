package gui

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"

	risorwidget "github.com/uidbz/fynerisor/gui/widget"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
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

func (obj *Fyne) Inspect() string {
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
	return object.Errorf("eval error: unsupported operation for %s: %v", FyneType, opType), nil
}

func (obj *Fyne) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Fyne) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", FyneType, name)
}

func (obj *Fyne) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "NewMenuItem":
		return object.NewBuiltin("fyne.NewMenuItem", risorwidget.NewMenuItem(obj.w)), true

	case "NewMenuItemSeparator":
		return object.NewBuiltin("fyne.NewMenuItemSeparator", risorwidget.NewMenuItemSeparator), true

	case "NewMenu":
		return object.NewBuiltin("fyne.NewMenu", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) < 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=at least 1", len(args)), nil
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
					return object.Errorf("argument error: expected MenuItem, got %s", args[i].Type()), nil
				}
				items = append(items, menuItemObj.Instance())
			}

			menu := fyne.NewMenu(label, items...)

			return &Menu{instance: menu}, nil
		}), true

	case "NewMainMenu":
		return object.NewBuiltin("fyne.NewMainMenu", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) < 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=at least 1", len(args)), nil
			}

			// Collect Menus from list
			listObj, ok := args[0].(*object.List)
			if !ok {
				return object.Errorf("argument error: expected list of menus, got %s", args[0].Type()), nil
			}

			menus := make([]*fyne.Menu, 0, len(listObj.Value()))
			for _, item := range listObj.Value() {
				menuObj, ok := item.(*Menu)
				if !ok {
					return object.Errorf("argument error: expected Menu in list, got %s", item.Type()), nil
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

func (obj *Menu) Inspect() string {
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
	return object.Errorf("eval error: unsupported operation for Menu: %v", opType), nil
}

func (obj *Menu) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Menu) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: Menu object has no writable attributes")
}

func (obj *Menu) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Label":
		return object.NewString(obj.instance.Label), true

	case "Refresh":
		return object.NewBuiltin("fyne.Menu.Refresh", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
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

func (obj *MainMenu) Inspect() string {
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
	return object.Errorf("eval error: unsupported operation for MainMenu: %v", opType), nil
}

func (obj *MainMenu) Equals(other object.Object) bool {
	return obj == other
}

func (obj *MainMenu) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: MainMenu object has no writable attributes")
}

func (obj *MainMenu) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Refresh":
		return object.NewBuiltin("fyne.MainMenu.Refresh", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
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
