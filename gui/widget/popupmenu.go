package widget

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/fynerisor/gui/guithread"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &PopUpMenu{}

const PopUpMenuType object.Type = "widget.PopUpMenu"

// PopUpMenu wraps fyne's PopUpMenu widget - a menu displayed in an overlay.
//
// PopUpMenu displays a menu in a floating overlay, commonly used for context menus.
//
// Example usage in Risor:
//
//	let item1 = fyne.NewMenuItem("Copy", () => { print("Copy") })
//	let item2 = fyne.NewMenuItem("Paste", () => { print("Paste") })
//	let menu = fyne.NewMenu("Edit", item1, item2)
//	let popupMenu = widget.NewPopUpMenu(menu, window.Canvas)
//	popupMenu.ShowAtPosition(100, 100)
//
// Methods:
//   - Show(): Display the menu
//   - Hide(): Hide the menu
//   - ShowAtPosition(x, y): Show at specific coordinates
type PopUpMenu struct {
	instance *widget.PopUpMenu
}

func NewPopUpMenu(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 2 {
		return nil, object.Errorf("wrong number of arguments. got=%d, want=2", len(args))
	}

	// First arg should be a Menu (from fyne.NewMenu)
	menuObj, ok := args[0].(interface{ Instance() *fyne.Menu })
	if !ok {
		return nil, object.Errorf("argument error: expected Menu, got %s", args[0].Type())
	}

	canvasObj, ok := args[1].(interface{ Canvas() fyne.Canvas })
	if !ok {
		return nil, object.Errorf("argument error: expected canvas (use window.Canvas), got %s", args[1].Type())
	}
	canvas := canvasObj.Canvas()
	if canvas == nil {
		return nil, object.Errorf("argument error: no canvas available (pass gui.WithHostWindow when embedding with gui.NewCanvasObject)")
	}

	instance := widget.NewPopUpMenu(menuObj.Instance(), canvas)

	return &PopUpMenu{
		instance: instance,
	}, nil
}

func (obj *PopUpMenu) CanvasObject() fyne.CanvasObject {
	return obj.instance
}

func (obj *PopUpMenu) Type() object.Type {
	return PopUpMenuType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *PopUpMenu) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *PopUpMenu) inspectUI() string {
	return "widget.PopUpMenu"
}

func (obj *PopUpMenu) Interface() interface{} {
	return obj.instance
}

func (obj *PopUpMenu) IsTruthy() bool {
	return true
}

func (obj *PopUpMenu) Cost() int {
	return 0
}

func (obj *PopUpMenu) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'widget.PopUpMenu'")
}

func (obj *PopUpMenu) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", PopUpMenuType, opType)
}

func (obj *PopUpMenu) Equals(other object.Object) bool {
	return obj == other
}

func (obj *PopUpMenu) Attrs() []object.AttrSpec {
	return nil
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *PopUpMenu) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *PopUpMenu) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", PopUpMenuType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *PopUpMenu) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *PopUpMenu) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "Show":
		return object.NewBuiltin("widget.PopUpMenu.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}

			guithread.Do(func() {
				obj.instance.Show()
			})

			return object.Nil, nil
		}), true

	case "Hide":
		return object.NewBuiltin("widget.PopUpMenu.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}

			guithread.Do(func() {
				obj.instance.Hide()
			})

			return object.Nil, nil
		}), true

	case "ShowAtPosition":
		return object.NewBuiltin("widget.PopUpMenu.ShowAtPosition", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=2", len(args))
			}

			x, err := object.AsFloat(args[0])
			if err != nil {
				return nil, err
			}

			y, err := object.AsFloat(args[1])
			if err != nil {
				return nil, err
			}

			guithread.Do(func() {
				obj.instance.ShowAtPosition(fyne.NewPos(float32(x), float32(y)))
			})

			return object.Nil, nil
		}), true

	case "Move":
		return object.NewBuiltin("widget.PopUpMenu.Move", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=2", len(args))
			}

			x, err := object.AsFloat(args[0])
			if err != nil {
				return nil, err
			}

			y, err := object.AsFloat(args[1])
			if err != nil {
				return nil, err
			}

			guithread.Do(func() {
				obj.instance.Move(fyne.NewPos(float32(x), float32(y)))
			})

			return object.Nil, nil
		}), true

	case "Resize":
		return object.NewBuiltin("widget.PopUpMenu.Resize", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return nil, object.Errorf("wrong number of arguments. got=%d, want=2", len(args))
			}

			w, err := object.AsFloat(args[0])
			if err != nil {
				return nil, err
			}

			h, err := object.AsFloat(args[1])
			if err != nil {
				return nil, err
			}

			guithread.Do(func() {
				obj.instance.Resize(fyne.NewSize(float32(w), float32(h)))
			})

			return object.Nil, nil
		}), true
	}

	return nil, false
}
