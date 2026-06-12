# AppTabs Example

This example demonstrates the AppTabs container, which provides tabbed navigation for your application.

## Running the Example

```bash
go run main.go
```

Or from the repository root:
```bash
go run ./examples/31-apptabs
```

## Features Demonstrated

### Basic Tab Creation
- Creating tab items with `container.NewTabItem(text, content)`
- Building AppTabs with `container.NewAppTabs(tab1, tab2, tab3)`
- Multiple tabs with different content

### Tab Navigation
- `SelectIndex(index)` - Switch to a specific tab by index
- `SelectedIndex()` - Get the currently selected tab index
- `Selected()` - Get the currently selected TabItem

### Dynamic Tab Management
- `Append(tabItem)` - Add new tabs at runtime
- `Remove(tabItem)` - Remove a specific tab
- `RemoveIndex(index)` - Remove tab by index

### Tab State Control
- `EnableIndex(index)` - Enable a disabled tab
- `DisableIndex(index)` - Disable a tab (grayed out, not selectable)
- `EnableItem(tabItem)` - Enable by TabItem reference
- `DisableItem(tabItem)` - Disable by TabItem reference

### Visibility Control
- `Hide()` - Hide the entire tab container
- `Show()` - Show the tab container

## Usage Pattern

```js
// Create tab content
let content1 = widget.NewLabel("Tab 1 content")
let content2 = widget.NewLabel("Tab 2 content")

// Create tab items
let tab1 = container.NewTabItem("First", content1)
let tab2 = container.NewTabItem("Second", content2)

// Create tabs container
let tabs = container.NewAppTabs(tab1, tab2)

// Control tabs programmatically
tabs.SelectIndex(0)           // Select first tab
let current = tabs.SelectedIndex()  // Get current index
tabs.DisableIndex(1)          // Disable second tab
tabs.Append(newTab)           // Add another tab
```

## Key Concepts

- **TabItem** - Represents a single tab with text label and content
- **AppTabs** - Container that manages multiple TabItem instances
- **Dynamic Management** - Tabs can be added, removed, enabled, and disabled at runtime
- **Callbacks** - Tab selection can trigger status updates or other actions
- **Layout Integration** - Tabs work with other containers (Border, VBox, etc.)

## Tips

- Tab indices are zero-based (first tab is index 0)
- Removing a tab will shift indices of tabs after it
- Disabled tabs appear grayed out and cannot be selected
- Each tab can contain any Fyne widget or container
- Use descriptive tab names to help users navigate
