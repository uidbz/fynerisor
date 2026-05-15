# List Widget Example

This example demonstrates the List widget with virtualized rendering for displaying scrollable lists efficiently.

## Features Demonstrated

- **Virtualized Rendering**: List only renders visible items, efficient for large datasets
- **Length Callback**: Dynamic item count
- **CreateItem Callback**: Template widget creation
- **UpdateItem Callback**: Item-specific updates
- **Selection**: OnSelected callback for item clicks
- **Programmatic Control**: Select(), ScrollTo(), UnselectAll()
- **Dynamic Updates**: Adding items and refreshing the list

## Running the Example

```bash
cd examples/11-list
go run main.go
```

Or use the CLI:

```bash
cd examples/11-list
../../cmd/fynerisor/fynerisor-cli app.risor
```

## Key Concepts

### Creating a List

```js
let myList = widget.NewList()
```

### Required Callbacks

The List widget requires three callbacks to function:

**1. Length() - Return item count:**
```js
myList.Length(() => {
    return len(items)
})
```

**2. CreateItem() - Create template widget:**
```js
myList.CreateItem(() => {
    return widget.NewLabel("")  // Returns template widget
})
```

**3. UpdateItem() - Update specific item:**
```js
myList.UpdateItem((id, item) => {
    // id = item index (0-based)
    // item = widget created by CreateItem
    item.Text = items[id]
})
```

### Optional Callbacks

**OnSelected - Handle item selection:**
```js
myList.OnSelected((id) => {
    print(`Selected item ${id}`)
})
```

### Methods

```js
// Select specific item
myList.Select(5)

// Scroll to item
myList.ScrollTo(20)

// Unselect all items
myList.UnselectAll()

// Refresh list after data change
myList.Refresh()
```

### Properties

```js
// Hide separators between rows
myList.HideSeparators = true
```

## How It Works

The List widget uses **virtualized rendering**:

1. **Length()** tells the list how many items exist
2. **CreateItem()** creates a template widget (called once per visible row)
3. **UpdateItem()** populates the template with data for a specific item (called when scrolling)

This approach is efficient because:
- Only visible items are rendered
- Template widgets are reused as you scroll
- Memory usage is constant regardless of list size

## UpdateItem Parameters

```js
myList.UpdateItem((id, item) => {
    // id: int - index of the item (0-based)
    // item: widget - the template widget from CreateItem()
    
    // Update the widget with data for this item
    item.Text = myData[id]
})
```

## Common Patterns

### Simple String List

```js
let data = ["Item 1", "Item 2", "Item 3"]

myList.Length(() => len(data))
myList.CreateItem(() => widget.NewLabel(""))
myList.UpdateItem((id, item) => {
    item.Text = data[id]
})
```

### Dynamic Data Updates

```js
// Add item to data (use + operator for array concatenation in Risor v2)
data = data + ["New Item"]

// Refresh list to show changes
myList.Refresh()
```

### Selection Handling

```js
myList.OnSelected((id) => {
    statusLabel.Text = `Selected: ${data[id]}`
    // Trigger action based on selection
})
```

### Programmatic Navigation

```js
// Jump to specific item
myList.ScrollTo(100)

// Pre-select an item
myList.Select(5)

// Clear selection
myList.UnselectAll()
```

## Limitations

- CreateItem() must return a Fyne widget (Label, Button, etc.)
- UpdateItem() receives the widget as a wrapped object
- Currently only Label widgets are fully accessible in UpdateItem()
- Complex item templates may require additional widget bindings

## Performance Notes

- List widget is designed for thousands of items
- Only visible rows are rendered (virtualization)
- UpdateItem is called frequently during scrolling - keep it fast
- Avoid expensive operations in callbacks

## What This Example Shows

1. **Basic List Setup**: Length, CreateItem, UpdateItem pattern
2. **Selection Handling**: OnSelected callback
3. **Programmatic Control**: Select, ScrollTo, Unselect methods
4. **Dynamic Updates**: Adding items and calling Refresh()
5. **User Interaction**: Buttons to manipulate list state

## Notes

- Variable name `list` conflicts with Risor's built-in list module - use a different name (e.g., `myList`, `fruitList`)
- List items are 0-indexed
- List automatically handles scrolling for large datasets
- UpdateItem callback is marshalled to GUI thread automatically
