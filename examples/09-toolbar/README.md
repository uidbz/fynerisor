# Example 09: Toolbar Widget

Demonstrates action toolbars with icons, separators, and spacers.

## What It Does

- Shows **Toolbar** widget for action buttons with icons
- Demonstrates **ToolbarAction** for clickable toolbar buttons
- Shows **ToolbarSeparator** for visual grouping
- Demonstrates **ToolbarSpacer** to push items to the right
- Shows dynamic toolbar management (Append/Prepend)
- Demonstrates comprehensive toolbar layouts

## Running

```bash
cd examples/09-toolbar
go run main.go
```

## Widgets

### Toolbar

Container widget that displays a horizontal bar of action items.

**Constructor:**
```js
let toolbar = widget.NewToolbar(item1, item2, item3, ...)
```

Takes zero or more ToolbarItem objects (ToolbarAction, ToolbarSeparator, ToolbarSpacer).

**Methods:**

**Append(item)** - Add item to end:
```js
let action = widget.NewToolbarAction("add", () => {})
toolbar.Append(action)
```

**Prepend(item)** - Add item to beginning:
```js
toolbar.Prepend(action)
```

**Refresh()** - Force visual refresh:
```js
toolbar.Refresh()
```

### ToolbarAction

Action button in a toolbar with an icon and callback.

**Constructor:**
```js
let action = widget.NewToolbarAction(iconName, onActivated)
```

**Parameters:**
- `iconName` (string) - Theme icon name (same icons as widget.NewIcon)
- `onActivated` (function) - Callback `() => {}` when button clicked

**Example:**
```js
let saveAction = widget.NewToolbarAction("documentSave", () => {
    print("Save clicked")
})
```

**Available Icons:**
Same 40+ theme icons as Icon widget: documentSave, folderOpen, cut, copy, paste, undo, redo, search, settings, help, etc. See example 06 for complete list.

### ToolbarSeparator

Visual separator between toolbar item groups.

**Constructor:**
```js
let separator = widget.NewToolbarSeparator()
```

No parameters, no methods. Just adds a vertical line between toolbar items.

**Example:**
```js
let toolbar = widget.NewToolbar(
    action1, action2,
    widget.NewToolbarSeparator(),
    action3, action4
)
```

### ToolbarSpacer

Flexible space that pushes subsequent items to the right.

**Constructor:**
```js
let spacer = widget.NewToolbarSpacer()
```

No parameters, no methods. Expands to fill available space.

**Example:**
```js
let toolbar = widget.NewToolbar(
    leftAction1, leftAction2,
    widget.NewToolbarSpacer(),
    rightAction1, rightAction2
)
// leftActions on left, rightActions on right
```

## Key Concepts

### Basic Toolbar Layout

Simple toolbar with actions:
```js
let save = widget.NewToolbarAction("documentSave", () => {
    print("Save")
})
let open = widget.NewToolbarAction("folderOpen", () => {
    print("Open")
})
let toolbar = widget.NewToolbar(save, open)
```

Actions appear left-to-right in order specified.

### Grouping with Separators

Separate logical groups of actions:
```js
let toolbar = widget.NewToolbar(
    fileNew, fileOpen, fileSave,
    widget.NewToolbarSeparator(),
    editCut, editCopy, editPaste,
    widget.NewToolbarSeparator(),
    viewZoomIn, viewZoomOut
)
```

Each separator adds a vertical line between groups.

### Left/Right Layout with Spacer

Push items to opposite sides:
```js
let toolbar = widget.NewToolbar(
    // Left side
    action1, action2, action3,
    
    // Spacer takes all remaining space
    widget.NewToolbarSpacer(),
    
    // Right side
    settings, help
)
```

Only one spacer is typically used per toolbar.

### Dynamic Toolbar Management

Start empty and add items:
```js
let toolbar = widget.NewToolbar()

// Add to end
toolbar.Append(widget.NewToolbarAction("add", () => {}))

// Add to beginning
toolbar.Prepend(widget.NewToolbarAction("home", () => {}))
```

**Note:** Cannot remove individual items - you'd need to recreate the toolbar.

### Action Callbacks

Each action has its own callback:
```js
let undoAction = widget.NewToolbarAction("undo", () => {
    print("Undo clicked")
    // Perform undo operation
})

let redoAction = widget.NewToolbarAction("redo", () => {
    print("Redo clicked")
    // Perform redo operation
})
```

Callbacks run when the user clicks the toolbar button.

### Icon Selection

Use descriptive icon names:
```js
// File operations
widget.NewToolbarAction("file", () => {})
widget.NewToolbarAction("folderOpen", () => {})
widget.NewToolbarAction("documentSave", () => {})

// Edit operations
widget.NewToolbarAction("cut", () => {})
widget.NewToolbarAction("copy", () => {})
widget.NewToolbarAction("paste", () => {})

// Navigation
widget.NewToolbarAction("navigateBack", () => {})
widget.NewToolbarAction("navigateNext", () => {})
widget.NewToolbarAction("search", () => {})

// Common actions
widget.NewToolbarAction("add", () => {})
widget.NewToolbarAction("remove", () => {})
widget.NewToolbarAction("settings", () => {})
```

Icons automatically adapt to light/dark themes.

### Toolbar Positioning

Toolbars are typically placed at the top of a window:
```js
let toolbar = widget.NewToolbar(...)
let content = widget.NewLabel("Main content area")

let layout = container.NewBorder(
    toolbar,  // top
    nil,      // bottom
    nil,      // left
    nil,      // right
    content   // center
)

window.SetContent(layout)
```

Or in a VBox:
```js
let layout = container.NewVBox(
    toolbar,
    widget.NewSeparator(),
    content
)
```

### Common Toolbar Patterns

**File Toolbar:**
```js
widget.NewToolbar(
    widget.NewToolbarAction("document", () => {}),      // New
    widget.NewToolbarAction("folderOpen", () => {}),    // Open
    widget.NewToolbarAction("documentSave", () => {})   // Save
)
```

**Edit Toolbar:**
```js
widget.NewToolbar(
    widget.NewToolbarAction("cut", () => {}),
    widget.NewToolbarAction("copy", () => {}),
    widget.NewToolbarAction("paste", () => {}),
    widget.NewToolbarSeparator(),
    widget.NewToolbarAction("undo", () => {}),
    widget.NewToolbarAction("redo", () => {})
)
```

**Media Toolbar:**
```js
widget.NewToolbar(
    widget.NewToolbarAction("mediaSkipPrevious", () => {}),
    widget.NewToolbarAction("mediaPlay", () => {}),
    widget.NewToolbarAction("mediaPause", () => {}),
    widget.NewToolbarAction("mediaStop", () => {}),
    widget.NewToolbarAction("mediaSkipNext", () => {})
)
```

### Limitations

- Cannot remove items after adding (must recreate toolbar)
- Cannot reorder items (must recreate toolbar)
- Cannot disable individual actions
- No tooltips on toolbar items
- Icons only - no text labels on actions

## Example Sections

1. **Basic Toolbar** - Simple toolbar with three actions
2. **With Separators** - Grouped actions with visual separators
3. **With Spacer** - Left-aligned and right-aligned actions
4. **Dynamic Toolbar** - Add actions and separators at runtime
5. **Comprehensive** - Full-featured toolbar with multiple groups

## Files

- `main.go`: Go program that creates the window
- `widgets.risor`: Risor script demonstrating Toolbar
- `README.md`: This file
