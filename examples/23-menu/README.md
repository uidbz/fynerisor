# Menu & PopUpMenu Example

Demonstrates Menu and PopUpMenu widgets - context menus and menu systems.

## Features

- **MenuItem Creation**: fyne.NewMenuItem(label, callback)
- **Menu Separators**: Visual grouping with fyne.NewMenuItemSeparator()
- **PopUpMenu**: Display menu in overlay
- **Item States**: Disabled and Checked properties
- **Dynamic Updates**: Toggle item states at runtime

## MenuItem Properties

- `Label` - Item text
- `Disabled` - Whether item can be clicked
- `Checked` - Shows checkmark indicator
- `IsSeparator` - Whether this is a separator

## Creating Menus

```js
// Create menu items
let item1 = fyne.NewMenuItem("Copy", () => { print("Copy") })
let item2 = fyne.NewMenuItem("Paste", () => { print("Paste") })
let sep = fyne.NewMenuItemSeparator()

// Create menu
let menu = fyne.NewMenu("Edit", item1, item2, sep)

// Show as popup
let popupMenu = widget.NewPopUpMenu(menu, window.Canvas)
popupMenu.ShowAtPosition(100, 100)
```

## Run

```bash
go run main.go
```

Or with fynerisor CLI:

```bash
fynerisor script.risor
```
