package gui

import (
	"context"
	"fmt"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"

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

var fyneAttrs = objattrs.Register(string(FyneType), object.NewAttrRegistry[*Fyne](string(FyneType)))

func (obj *Fyne) getAttrUI(name string) (object.Object, bool) {
	if o, ok := fyneAttrs.GetAttr(obj, name); ok {
		return o, true
	}
	switch name {

	case "NewMenuItem":
		return object.NewBuiltin("fyne.NewMenuItem", risorwidget.NewMenuItem(obj.w)), true

	case "NewMenuItemSeparator":
		return object.NewBuiltin("fyne.NewMenuItemSeparator", risorwidget.NewMenuItemSeparator), true
	}
	return nil, false

	// Collect MenuItems

	// Collect Menus from list

}

func (obj *Fyne) Attrs() []object.AttrSpec {
	return append(fyneAttrs.Specs(),
		object.AttrSpec{Name: "NewMenuItem", Doc: "Create a menu item: NewMenuItem(label, action)", Args: []string{"label", "action"}},
		object.AttrSpec{Name: "NewMenuItemSeparator", Doc: "Create a menu separator item"},
	)
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

var menuAttrs = objattrs.Register("fyne.Menu", object.NewAttrRegistry[*Menu]("fyne.Menu"))

func (obj *Menu) getAttrUI(name string) (object.Object, bool) {
	return menuAttrs.GetAttr(obj, name)

}

func (obj *Menu) Attrs() []object.AttrSpec {
	return menuAttrs.Specs()

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

var mainMenuAttrs = objattrs.Register("fyne.MainMenu", object.NewAttrRegistry[*MainMenu]("fyne.MainMenu"))

func (obj *MainMenu) getAttrUI(name string) (object.Object, bool) {
	return mainMenuAttrs.GetAttr(obj, name)

}

func (obj *MainMenu) Attrs() []object.AttrSpec {
	return mainMenuAttrs.Specs()

}
func init() {
	menuAttrs.Define("Label").Doc("The menu's label.").Getter(
		func(obj *Menu) object.Object {
			return object.NewString(obj.instance.Label)
		})
	menuAttrs.Define("Refresh").Doc("Redraw the menu.").Args().Returns("nil").Impl(func(obj *Menu, ctx context.Context, args ...object.
		Object) (object.Object, error) {
		obj.instance.
			Refresh()
		return object.Nil, nil
	})
}
func init() {
	mainMenuAttrs.
		Define("Refresh").Doc("Redraw the main menu.").Args().Returns("nil").Impl(func(obj *MainMenu, ctx context.
		Context, args ...object.Object) (object.Object, error) {
		obj.instance.
			Refresh()
		return object.Nil, nil
	})
}
func init() {
	fyneAttrs.Define("NewMenu").Doc("Create a menu: NewMenu(label, item...)").Variadic("args").Returns("nil").Impl(func(obj *Fyne, ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) < 1 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=at least 1",

				len(args))
		}
		label, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}
		items := make([]*fyne.
			MenuItem, 0, len(args)-1)
		for i := 1; i < len(args); i++ {
			menuItemObj,
				ok := args[i].(*risorwidget.MenuItem)
			if !ok {
				return nil, object.
					Errorf("argument error: expected MenuItem, got %s", args[i].Type())
			}
			items = append(items, menuItemObj.Instance())
		}
		menu := fyne.NewMenu(label,
			items...)
		return &Menu{instance: menu},
			nil
	})
	fyneAttrs.Define("NewMainMenu").Doc("Create a main menu: NewMainMenu(menu...)").Variadic("args").Returns("nil").Impl(func(obj *Fyne,

		ctx context.Context,

		args ...object.Object) (object.Object, error) {
		if len(args) < 1 {
			return nil, object.Errorf("wrong number of arguments. got=%d, want=at least 1",

				len(args))
		}
		listObj, ok := args[0].(*object.List)
		if !ok {
			return nil, object.Errorf("argument error: expected list of menus, got %s",
				args[0].Type())
		}
		menus := make([]*fyne.
			Menu, 0,
			len(listObj.Value()))
		for _, item := range listObj.Value() {
			menuObj, ok := item.(*Menu)
			if !ok {
				return nil, object.Errorf("argument error: expected Menu in list, got %s",

					item.Type())
			}
			menus = append(
				menus,

				menuObj.instance)
		}
		mainMenu := fyne.
			NewMainMenu(menus...)
		return &MainMenu{instance: mainMenu}, nil
	})
}
