# Release Notes - v0.5.1

## Widget Support in Table Cells 🎨

This release adds the ability to use any Fyne widget type in table cells, transforming tables from static data displays into interactive components.

### Key Features

**Widget Mode Tables:**
- Use buttons, icons, labels, entries, checkboxes, dropdowns, or any other Fyne widget in table cells
- Two callback pattern: CreateCell (widget creation) and UpdateCell (data binding)
- Automatic caching and reuse of widgets for optimal performance
- Canvas image support for displaying thumbnails and pictures

**Multi-Format Export:**
- Export table data to CSV, XLSX, or JSON formats
- Select which columns to export
- Export current page or all data
- Configurable export path and filename
- Default exports go to ./exports directory

**Smart Filtering:**
- Filtering works seamlessly with widget mode tables
- Automatic row index mapping (filtered → original data)
- Widget state and interactivity preserved during filtering
- No script changes required - mapping is completely transparent

**OS Module Enhancements:**
- `os.read_dir(path)` - List directory contents
- `os.getwd()` - Get current working directory  
- `os.exec(command, args)` - Execute external commands

### Examples

**Example 32 - Table Widgets:**
Interactive table with buttons and icons demonstrating widget mode:
```bash
cd examples/32-table-widgets
go run main.go
```

Features:
- Icons in first column (different icons per role)
- Delete buttons in last column
- Add/toggle controls
- Full interactivity with filtering

**Example 33 - Image Gallery:**
Browse local image files with thumbnail previews:
```bash
cd examples/33-image-gallery
go run main.go
```

Features:
- Directory scanning with os.read_dir()
- Canvas images displayed in table cells
- File path and name display
- Export capabilities

### Breaking Changes

None - this release is fully backward compatible. String-mode tables continue to work unchanged.

### Widget Mode Pattern

```javascript
// Create widgets once per cell
table.CreateCell((col, row) => {
    if (col == 0) return widget.NewIcon("info")
    if (col == 3) return widget.NewButton("Delete", () => {})
    return widget.NewLabel("")
})

// Update widgets with data
table.UpdateCell((col, row, cell) => {
    let user = state.users[row]  // Row index automatically mapped during filtering
    
    if (col == 0) {
        cell.SetResource("settings")
    } else if (col == 3) {
        cell.SetText("Delete")
        cell.OnTapped(() => {
            deleteUser(user.name)
            table.Refresh()
        })
    }
})
```

### Installation

```bash
go get github.com/uidbz/fynerisor@v0.5.1
```

Or update your go.mod:
```
require github.com/uidbz/fynerisor v0.5.1
```

### What's Next

See the [CHANGELOG](CHANGELOG.md) for complete details and the [Roadmap](docs/ROADMAP.md) for future plans.
