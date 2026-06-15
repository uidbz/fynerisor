package widget

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &MenuItem{}

const MenuItemType object.Type = "fyne.MenuItem"

// MenuItem wraps fyne's MenuItem - a single item in a menu.
//
// MenuItem represents one entry in a menu with a label and optional action callback.
//
// Example usage in Risor:
//
//	let item1 = fyne.NewMenuItem("Open", () => { print("Open clicked") })
//	let item2 = fyne.NewMenuItem("Save", () => { print("Save clicked") })
//	let separator = fyne.NewMenuItemSeparator()
//
// Properties:
//   - Label: Item text
//   - Disabled: Whether item is disabled
//   - Checked: Whether item shows checkmark
type MenuItem struct {
	instance *fyne.MenuItem
	w        WindowInterface
}

func NewMenuItem(w WindowInterface) func(ctx context.Context, args ...object.Object) (object.Object, error) {
	return func(ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) != 2 {
			return object.Errorf("wrong number of arguments. got=%d, want=2", len(args)), nil
		}

		label, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}

		fn, ok := args[1].(*object.Closure)
		if !ok {
			return object.Errorf("argument error: expected function, got %s", args[1].Type()), nil
		}

		callFunc, ok := object.GetCallFunc(ctx)
		if !ok {
			return object.Errorf("menuitem: unable to get call function"), nil
		}

		instance := fyne.NewMenuItem(label, func() {
			w.Do(func() {
				_, err := callFunc(ctx, fn, []object.Object{})
				if err != nil {
					fmt.Println("MenuItem action error:", err)
				}
			})
		})

		return &MenuItem{
			instance: instance,
			w:        w,
		}, nil
	}
}

func NewMenuItemSeparator(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 0 {
		return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
	}

	instance := fyne.NewMenuItemSeparator()

	return &MenuItem{
		instance: instance,
	}, nil
}

func (obj *MenuItem) Instance() *fyne.MenuItem {
	return obj.instance
}

func (obj *MenuItem) Type() object.Type {
	return MenuItemType
}

func (obj *MenuItem) Inspect() string {
	return fmt.Sprintf("fyne.MenuItem(%s)", obj.instance.Label)
}

func (obj *MenuItem) Interface() interface{} {
	return obj.instance
}

func (obj *MenuItem) IsTruthy() bool {
	return true
}

func (obj *MenuItem) Cost() int {
	return 0
}

func (obj *MenuItem) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'fyne.MenuItem'")
}

func (obj *MenuItem) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(MenuItemType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", MenuItemType, opType)
	return errObj, err
}

func (obj *MenuItem) Equals(other object.Object) bool {
	return obj == other
}

func (obj *MenuItem) Attrs() []object.AttrSpec {
	return nil
}

func (obj *MenuItem) SetAttr(name string, value object.Object) error {
	switch name {
	case "Label":
		s, err := object.AsString(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		obj.instance.Label = s
		return nil
	case "Disabled":
		b, err := object.AsBool(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		obj.instance.Disabled = b
		return nil
	case "Checked":
		b, err := object.AsBool(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		obj.instance.Checked = b
		return nil
	case "Shortcut":
		if value == object.Nil {
			obj.instance.Shortcut = nil
			return nil
		}

		// Accept string like "Ctrl+S"
		if strObj, ok := value.(*object.String); ok {
			shortcut, err := parseShortcutString(strObj.Value())
			if err != nil {
				return fmt.Errorf("invalid shortcut: %w", err)
			}
			obj.instance.Shortcut = shortcut
			return nil
		}

		return fmt.Errorf("Shortcut must be a string")
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", MenuItemType, name)
}

func (obj *MenuItem) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Label":
		return object.NewString(obj.instance.Label), true
	case "Disabled":
		return object.NewBool(obj.instance.Disabled), true
	case "Checked":
		return object.NewBool(obj.instance.Checked), true
	case "IsSeparator":
		return object.NewBool(obj.instance.IsSeparator), true
	case "Shortcut":
		if obj.instance.Shortcut == nil {
			return object.Nil, true
		}
		// Return string representation
		if cs, ok := obj.instance.Shortcut.(*desktop.CustomShortcut); ok {
			return object.NewString(formatShortcut(cs)), true
		}
		return object.Nil, true
	}
	return nil, false
}

// parseShortcutString parses strings like "Ctrl+S", "Alt+Shift+A", etc.
// Duplicated from gui/shortcut.go to avoid import cycle
func parseShortcutString(s string) (fyne.Shortcut, error) {
	parts := strings.Split(s, "+")
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty shortcut string")
	}

	var mods fyne.KeyModifier
	var keyName string

	// Process all parts except the last (which is the key)
	for i := 0; i < len(parts)-1; i++ {
		mod := strings.TrimSpace(parts[i])
		switch strings.ToLower(mod) {
		case "ctrl", "control":
			mods |= fyne.KeyModifierControl
		case "alt", "option":
			mods |= fyne.KeyModifierAlt
		case "shift":
			mods |= fyne.KeyModifierShift
		case "super", "cmd", "command":
			mods |= fyne.KeyModifierSuper
		default:
			return nil, fmt.Errorf("unknown modifier: %s", mod)
		}
	}

	// Last part is the key
	keyName = strings.TrimSpace(parts[len(parts)-1])
	keyCode, err := parseKeyName(keyName)
	if err != nil {
		return nil, err
	}

	return &desktop.CustomShortcut{
		KeyName:  keyCode,
		Modifier: mods,
	}, nil
}

// parseKeyName converts string key names to fyne.KeyName
func parseKeyName(name string) (fyne.KeyName, error) {
	// Handle single letters
	if len(name) == 1 {
		char := strings.ToUpper(name)[0]
		if char >= 'A' && char <= 'Z' {
			return fyne.KeyName(char), nil
		}
		if char >= '0' && char <= '9' {
			return fyne.KeyName(char), nil
		}
	}

	// Handle special keys
	switch strings.ToLower(name) {
	case "return", "enter":
		return fyne.KeyReturn, nil
	case "escape", "esc":
		return fyne.KeyEscape, nil
	case "tab":
		return fyne.KeyTab, nil
	case "space":
		return fyne.KeySpace, nil
	case "backspace":
		return fyne.KeyBackspace, nil
	case "delete", "del":
		return fyne.KeyDelete, nil
	case "home":
		return fyne.KeyHome, nil
	case "end":
		return fyne.KeyEnd, nil
	case "pageup":
		return fyne.KeyPageUp, nil
	case "pagedown":
		return fyne.KeyPageDown, nil
	case "up":
		return fyne.KeyUp, nil
	case "down":
		return fyne.KeyDown, nil
	case "left":
		return fyne.KeyLeft, nil
	case "right":
		return fyne.KeyRight, nil
	case "f1":
		return fyne.KeyF1, nil
	case "f2":
		return fyne.KeyF2, nil
	case "f3":
		return fyne.KeyF3, nil
	case "f4":
		return fyne.KeyF4, nil
	case "f5":
		return fyne.KeyF5, nil
	case "f6":
		return fyne.KeyF6, nil
	case "f7":
		return fyne.KeyF7, nil
	case "f8":
		return fyne.KeyF8, nil
	case "f9":
		return fyne.KeyF9, nil
	case "f10":
		return fyne.KeyF10, nil
	case "f11":
		return fyne.KeyF11, nil
	case "f12":
		return fyne.KeyF12, nil
	default:
		return "", fmt.Errorf("unknown key: %s", name)
	}
}

// formatShortcut converts a CustomShortcut back to string format
func formatShortcut(cs *desktop.CustomShortcut) string {
	parts := []string{}

	if cs.Modifier&fyne.KeyModifierControl != 0 {
		parts = append(parts, "Ctrl")
	}
	if cs.Modifier&fyne.KeyModifierAlt != 0 {
		parts = append(parts, "Alt")
	}
	if cs.Modifier&fyne.KeyModifierShift != 0 {
		parts = append(parts, "Shift")
	}
	if cs.Modifier&fyne.KeyModifierSuper != 0 {
		parts = append(parts, "Super")
	}

	parts = append(parts, string(cs.KeyName))
	return strings.Join(parts, "+")
}
