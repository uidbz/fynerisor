# Keyboard Shortcuts Example

This example demonstrates keyboard shortcut support in fynerisor.

## Features Demonstrated

- **Global shortcuts**: Register keyboard shortcuts with `window.AddShortcut()`
- **Works without menus**: Shortcuts work immediately, no menu required
- **Multi-modifier shortcuts**: Use combinations like "Ctrl+Shift+N"
- **Special keys**: Function keys like F1
- **Dynamic management**: Add and remove shortcuts at runtime
- **Cross-platform**: Works on all platforms with standard modifier names

## How It Works

### Register Shortcuts

```javascript
window.AddShortcut("Ctrl+S", () => {
    print("Save triggered!")
})
```

Shortcuts are registered on the window canvas and work immediately.

### Supported Modifiers

- `Ctrl` / `Control` - Control key
- `Alt` / `Option` - Alt/Option key
- `Shift` - Shift key
- `Super` / `Cmd` / `Command` - Super/Command/Windows key

### Supported Keys

- Single letters: `A-Z`, `0-9`
- Special keys: `Return`, `Enter`, `Escape`, `Tab`, `Space`, `Backspace`, `Delete`
- Arrow keys: `Up`, `Down`, `Left`, `Right`
- Function keys: `F1-F12`
- Navigation: `Home`, `End`, `PageUp`, `PageDown`

### Remove Shortcuts

```javascript
window.RemoveShortcut("Ctrl+S")
```

## Running the Example

Basic example (no menu):
```bash
cd examples/34-keyboard-shortcuts
go run main.go
```

With menu display:
```bash
cd examples/34-keyboard-shortcuts
go run main.go -menu
```

## Try It

1. Press **Ctrl+S** - Triggers save action
2. Press **Ctrl+O** - Triggers open action
3. Press **Ctrl+Shift+N** - Multi-modifier shortcut
4. Press **F1** - Function key shortcut
5. Click **Remove Ctrl+S Shortcut** - Disables Ctrl+S
6. Press **Ctrl+S** - No longer works
7. Click **Add Ctrl+S Back** - Re-enables Ctrl+S
8. Press **Ctrl+S** - Works again!

## Notes

- Shortcuts are global to the window
- No visible menu required for shortcuts to work
- Use `MenuItem.Shortcut = "Ctrl+S"` to display shortcuts in menus (display only)
- Shortcuts can be added/removed dynamically at any time
