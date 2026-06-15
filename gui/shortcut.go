package gui

import (
	"fmt"
	"strings"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// ParseShortcutString parses strings like "Ctrl+S", "Alt+Shift+A", etc.
func ParseShortcutString(s string) (fyne.Shortcut, error) {
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
		// Handle digits
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
	// Add F1-F12
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

// FormatShortcut converts a CustomShortcut back to string format
func FormatShortcut(cs *desktop.CustomShortcut) string {
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

// ShortcutWrapper wraps a Fyne shortcut for Risor
type ShortcutWrapper struct {
	Shortcut fyne.Shortcut
}

func NewShortcutWrapper(s fyne.Shortcut) *ShortcutWrapper {
	return &ShortcutWrapper{Shortcut: s}
}

func (sw *ShortcutWrapper) Type() object.Type {
	return "shortcut"
}

func (sw *ShortcutWrapper) Inspect() string {
	if cs, ok := sw.Shortcut.(*desktop.CustomShortcut); ok {
		return fmt.Sprintf("shortcut<%s>", FormatShortcut(cs))
	}
	return "shortcut"
}

func (sw *ShortcutWrapper) Interface() interface{} {
	return sw.Shortcut
}

func (sw *ShortcutWrapper) IsTruthy() bool {
	return true
}

func (sw *ShortcutWrapper) Cost() int {
	return 0
}

func (sw *ShortcutWrapper) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal shortcut")
}

func (sw *ShortcutWrapper) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for shortcut: %v", opType), nil
}

func (sw *ShortcutWrapper) Equals(other object.Object) bool {
	if osw, ok := other.(*ShortcutWrapper); ok {
		return sw.Shortcut == osw.Shortcut
	}
	return false
}

func (sw *ShortcutWrapper) GetAttr(name string) (object.Object, bool) {
	return nil, false
}

func (sw *ShortcutWrapper) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: shortcut object has no attribute %q", name)
}

func (sw *ShortcutWrapper) Attrs() []object.AttrSpec {
	return nil
}
