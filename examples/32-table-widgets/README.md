# Table Widgets Example

This example demonstrates using widgets (buttons, entries, etc.) inside table cells instead of just strings.

## Running the Example

```bash
go run main.go
```

Or from the repository root:
```bash
go run ./examples/04-table-widgets
```

## Features Demonstrated

### Widget Mode vs String Mode

Tables in fynerisor support two modes:
- **String mode** (default): Traditional table with text-only cells
- **Widget mode** (opt-in): Cells can contain any Fyne widget (buttons, entries, selects, etc.)

### Enabling Widget Mode

Use `CreateCell` and `UpdateCell` callbacks instead of the `Data` callback:

```js
// CreateCell - called once per cell to create the widget
table.CreateCell((col, row) => {
    if (col == 3) {
        return widget.NewButton("Delete", () => {})
    }
    return widget.NewLabel("")
})

// UpdateCell - called to populate/update the widget with data
table.UpdateCell((col, row, cell) => {
    let user = users[row]
    if (col == 3) {
        cell.SetText("Delete")
        cell.OnTapped(() => {
            print("Deleting user:", user.name)
        })
    }
})
```

### Key Concepts

1. **CreateCell(col, row)** - Creates the widget for each cell
   - Called once per cell when the table is initialized
   - Return the appropriate widget type for each column
   - Think of it as a factory that creates the cell template

2. **UpdateCell(col, row, cell)** - Populates the widget with data
   - Called whenever the table needs to display or refresh data
   - Configure the widget with the appropriate data for that row/column
   - Can set text, callbacks, visibility, etc.

3. **Dynamic Updates** - Call `table.Refresh()` to update all cells
   - Useful after adding/removing/modifying data
   - Triggers UpdateCell for all visible cells

### Pattern: Action Buttons

```js
table.CreateCell((col, row) => {
    if (col == actionColumn) {
        return widget.NewButton("Action", () => {})
    }
    return widget.NewLabel("")
})

table.UpdateCell((col, row, cell) => {
    if (col == actionColumn) {
        cell.SetText("Delete " + data[row].name)
        cell.OnTapped(() => {
            // Perform action
            data.splice(row, 1)
            table.Refresh()
        })
    }
})
```

### Pattern: Status Indicators

```js
table.CreateCell((col, row) => {
    return widget.NewLabel("")
})

table.UpdateCell((col, row, cell) => {
    if (col == statusColumn) {
        let status = data[row].status
        cell.SetText(status)
        // Could also set colors, importance, etc.
    }
})
```

## Backward Compatibility

The traditional string-based API still works unchanged:

```js
table.Data((offset, limit) => {
    return [["John", "Engineer"], ["Jane", "Manager"]]
})
```

Don't mix `Data()` with `CreateCell()`/`UpdateCell()` - use one approach or the other.

## Tips

- **Performance**: CreateCell is called only once per cell. UpdateCell is called frequently.
- **Widget types**: Any widget that implements `CanvasObject()` can be used
- **Dynamic data**: Store data in a separate array/object, not in the widgets themselves
- **Callbacks**: Set up event handlers in UpdateCell, not CreateCell
- **Refresh**: Always call `table.Refresh()` after modifying the underlying data

## Example Output

The example creates a user management table with:
- Icon column (icon widget - varies by role)
- Name column (label)
- Role column (label)
- Status column (label with checkmark/circle prefix)
- Actions column (button)

Features:
- Role-specific icons (settings for Engineers/Developers, folder for Managers, document for Designers)
- Visual status indicators (✓ for Active, ○ for Inactive)
- Delete buttons for each row
- Control panel with "Add User" and "Toggle First Status" buttons

## Supported Widget Types

All Fyne widget types are supported in table cells:
- Labels, Buttons, Entries, Selects
- Icons, Checks, CheckGroups, RadioGroups
- Sliders, ProgressBars, Hyperlinks
- Cards, Forms, Accordions, Separators
- And more...
