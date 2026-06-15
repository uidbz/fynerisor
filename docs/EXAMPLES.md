# fynerisor Examples

Production-ready examples demonstrating fynerisor widget and container usage.

> **Note:** Compatible with Risor v2 syntax. Key changes from v1:
> - Use `let` for variable declarations
> - Parentheses required in conditionals: `if (x > 0) { ... }`
> - Arrow functions: `x => x * 2` or `(a, b) => a + b`
> - No `for` loops - use `.map()`, `.filter()`, `.each()`, `.reduce()`
> - String formatting: `` `Hello ${name}!` `` or `sprintf("Hello %s!", name)`

## Table of Contents
- [Hello World](#hello-world)
- [Button with State](#button-with-state)
- [Form with Validation](#form-with-validation)
- [Data Table with Pagination](#data-table-with-pagination)
- [Bar Chart](#bar-chart)
- [Complex Layout](#complex-layout)
- [File Drop Handler](#file-drop-handler)
- [Calendar Date Picker](#calendar-date-picker)
- [List with Virtualization](#list-with-virtualization)
- [Code Organization with ImportScript](#code-organization-with-importscript)
- [Background Tasks with go()](#background-tasks-with-go)
- [File Operations with io module](#file-operations-with-io-module)
- [Detecting Application Context](#detecting-application-context)
- [Keyboard Shortcuts](#keyboard-shortcuts)

---

## Hello World

The simplest fynerisor app.

```js
let label = widget.NewLabel("Hello World")
window.SetContent(label)
```

---

## Button with State

Interactive button that tracks click count.

```js
let clickCount = 0

let label = widget.NewLabel("Click the button")

let button = widget.NewButton("Click me!", () => {
    clickCount = clickCount + 1
    label.Text = `Clicked ${clickCount} times`
})

let layout = container.NewVBox(label, button)
window.SetContent(layout)
```

**Key concepts:**
- Arrow functions for callbacks: `() => { ... }`
- State management with closures
- Dynamic property updates: `label.Text = "..."`
- VBox container for vertical layout

---

## Form with Validation

Form with input validation and error messages.

```js
let nameEntry = widget.NewEntry()
let emailEntry = widget.NewEntry()

let titleLabel = widget.NewLabel("Fill out the form and click Submit")
let resultLabel = widget.NewLabel("")

let submitButton = widget.NewButton("Submit", () => {
    let name = nameEntry.Text
    let email = emailEntry.Text

    // Validation
    if (name == "") {
        resultLabel.Text = "Error: Name is required"
        return
    }

    if (email == "") {
        resultLabel.Text = "Error: Email is required"
        return
    }

    if (!email.contains("@")) {
        resultLabel.Text = "Error: Invalid email"
        return
    }

    // Success
    resultLabel.Text = `Welcome, ${name}! Email: ${email}`
})

let clearButton = widget.NewButton("Clear", () => {
    nameEntry.Text = ""
    emailEntry.Text = ""
    resultLabel.Text = ""
})

// Create form with FormItems
let formItems = [
    widget.NewFormItem("Name:", nameEntry),
    widget.NewFormItem("Email:", emailEntry),
]

let content = container.NewVBox(
    titleLabel,
    widget.NewForm(formItems),
    container.NewHBox(submitButton, clearButton),
    resultLabel
)

window.SetContent(content)
```

**Key concepts:**
- Entry widgets for text input
- Separate labels for title and results
- `widget.NewForm()` takes an array of FormItems
- FormItem widgets provide labels: `widget.NewFormItem("Label:", widget)`
- If statements with parentheses: `if (condition) { ... }`
- Early return for validation errors
- Template strings: `` `Welcome, ${name}` ``
- HBox for horizontal button layout

---

## Data Table with Pagination

Table widget displaying paginated data.

```js
let table = widget.NewTable("People", 10)

let allData = [
    ["1", "Alice", "alice@example.com"],
    ["2", "Bob", "bob@example.com"],
    ["3", "Carol", "carol@example.com"],
    ["4", "David", "david@example.com"],
    ["5", "Eve", "eve@example.com"],
    ["6", "Frank", "frank@example.com"],
    ["7", "Grace", "grace@example.com"],
    ["8", "Henry", "henry@example.com"],
    ["9", "Iris", "iris@example.com"],
    ["10", "Jack", "jack@example.com"],
]

table.Columns(() => ["ID", "Name", "Email"])

table.RowCount(() => len(allData))

table.Data((offset, limit) => {
    let end = offset + limit
    if (end > len(allData)) {
        end = len(allData)
    }
    return allData[offset:end]
})

table.SetOnClick((row, col) => {
    print(`Clicked row ${row}, column ${col}`)
})

table.Refresh()
window.SetContent(table)
```

**Key concepts:**
- Table widget with page size (10 rows)
- Arrow function callbacks: `() => ...` and `(offset, limit) => { ... }`
- Pagination with offset/limit parameters
- List slicing: `allData[offset:end]`
- Click handler with row/column indices
- `table.Refresh()` to update display

---

## Bar Chart

Simple data visualization.

```js
let labels = ["Q1", "Q2", "Q3", "Q4"]
let values = [100, 150, 120, 180]

let chart = chart.NewBarChart(
    "Quarterly Sales",  // title
    "Revenue ($K)",     // y-axis label
    labels,            // x-axis labels
    values             // data values
)

window.SetContent(chart)
```

**Key concepts:**
- Chart creation with title and axis labels
- Separate arrays for labels and values
- Simple visualization API

---

## Complex Layout

Multi-region border layout with split panes.

```js
// Top toolbar
let toolbar = container.NewHBox(
    widget.NewButton("New", () => print("New")),
    widget.NewButton("Open", () => print("Open")),
    widget.NewButton("Save", () => print("Save"))
)

// Left sidebar
let sidebar = container.NewVBox(
    widget.NewLabel("Navigation"),
    widget.NewButton("Home", () => print("Home")),
    widget.NewButton("Settings", () => print("Settings"))
)

// Main content area
let content = widget.NewLabel("Main content area")

// Status bar
let statusBar = widget.NewLabel("Ready")

// Create split pane (sidebar + content)
let mainArea = container.NewHSplit(sidebar, content)

// Create border layout
let layout = container.NewBorder(
    toolbar,    // top
    statusBar,  // bottom
    nil,        // left
    nil,        // right
    mainArea    // center
)

window.SetContent(layout)
```

**Key concepts:**
- Border layout with five regions
- HSplit for draggable divider
- Use `nil` for empty regions
- Nested containers for complex UIs
- Multiple VBox/HBox layouts

---

## File Drop Handler

Handle drag-and-drop file operations.

```js
let pathLabel = widget.NewLabel("Drop files here")
let fileList = widget.NewMultiLineEntry()
fileList.Text = "No files dropped yet"

window.OnDropped((paths) => {
    let count = len(paths)
    pathLabel.Text = `Received ${count} file(s)`
    
    let fileNames = []
    paths.each((path) => {
        fileNames.append(filepath.base(path))
    })
    
    fileList.Text = "\n".join(fileNames)
})

let layout = container.NewVBox(
    pathLabel,
    fileList
)

window.SetContent(layout)
```

**Key concepts:**
- `window.OnDropped()` for drag-and-drop
- Arrow function callback: `(paths) => { ... }`
- `.each()` method for iteration
- `filepath.base()` to get filename
- String `.join()` method: `separator.join(list)`
- MultiLineEntry for text area

---

## Tips & Tricks

### Variable Declarations
```js
// Use let for variables
let name = "Alice"
let count = 42
let items = [1, 2, 3]

// Reassignment (no let needed)
count = 100
```

### Widget Property Updates
```js
// Most widgets support property updates
let label = widget.NewLabel("Initial")
label.Text = "Updated"

let entry = widget.NewEntry()
entry.Text = "Default value"
```

### Conditionals
```js
// Parentheses required in Risor v2
if (value > 10) {
    print("Greater")
} else if (value > 5) {
    print("Medium")
} else {
    print("Small")
}
```

### Functions
```js
// Function declaration
function validate(text) {
    return text.length() > 0
}

// Arrow functions
let double = x => x * 2
let add = (a, b) => a + b
```

### Iteration
```js
// Use .each() instead of for loops
items.each((item) => {
    print(item)
})

// Map transforms
let doubled = numbers.map(x => x * 2)

// Filter
let positive = numbers.filter(x => x > 0)

// Reduce
let sum = numbers.reduce(0, (acc, x) => acc + x)
```

### String Manipulation
```js
// Template strings
let greeting = `Hello ${name}!`

// Join list with separator
let text = ", ".join(items)

// Replace
let clean = text.replace_all("'", "''")

// Check contains
if (email.contains("@")) {
    print("Valid email")
}
```

### Lists and Slicing
```js
let items = [1, 2, 3, 4, 5]
let first = items[0]
let last = items[-1]
let slice = items[1:3]  // [2, 3]
items.append(6)
items.extend([7, 8])
```

### Error Handling
```js
// If/else pattern
if (reply.Success) {
    // Handle success
} else {
    print(`Error: ${reply.ErrorMsg}`)
}

// Try/catch/finally
try {
    let result = riskyOperation()
    processResult(result)
} catch (e) {
    print(`Error: ${e.message()}`)
} finally {
    cleanup()
}
```

### Container Patterns
```js
// Vertical stacking
container.NewVBox(widget1, widget2, widget3)

// Horizontal arrangement
container.NewHBox(widget1, widget2)

// Border (top, bottom, left, right, center)
container.NewBorder(
    toolbar,      // top
    statusBar,    // bottom
    sidebar,      // left
    nil,         // right (empty)
    content      // center
)

// Split panes
container.NewHSplit(left, right)
container.NewVSplit(top, bottom)

// Scrollable
container.NewScroll(largeContent)
```

### Callback Patterns
```js
// Button callback
widget.NewButton("Click", () => {
    print("Clicked!")
})

// Entry submit
let entry = widget.NewEntry()
entry.OnSubmitted((text) => {
    print(`Submitted: ${text}`)
})

// Table click
table.SetOnClick((row, col) => {
    print(`Clicked ${row}, ${col}`)
})

// Check callback
widget.NewCheck("Enable", (checked) => {
    print(`Checked: ${checked}`)
})

// Select callback
widget.NewSelect(options, (selected) => {
    print(`Selected: ${selected}`)
})
```

### Dynamic Data
```js
// Table with callback functions
let table = widget.NewTable("Data", 20)

table.Columns(() => ["Col1", "Col2", "Col3"])
table.RowCount(() => dataSource.count())
table.Data((offset, limit) => {
    return dataSource.fetch(offset, limit)
})

table.Refresh()  // Call when data changes
```

### Layout Tips
- Use VBox for vertical stacking
- Use HBox for horizontal grouping
- Use Border for complex layouts with regions
- Use Split for resizable panes
- Use Scroll for large content
- Nest containers for complex UIs
- Use FormItem to add labels to inputs

### Common Patterns
```js
// Form with submit
let input = widget.NewEntry()
let submit = widget.NewButton("Submit", () => {
    process(input.Text)
})

let formItems = [widget.NewFormItem("Input:", input)]
let form = widget.NewForm(formItems)

let layout = container.NewVBox(form, submit)

// Status display
let status = widget.NewLabel("Ready")
// Later update: status.Text = "Processing..."

// Button group
let buttons = container.NewHBox(
    widget.NewButton("OK", handleOK),
    widget.NewButton("Cancel", handleCancel)
)

// List display
let list = widget.NewMultiLineEntry()
list.Text = items | "\n".join
```

---

## Calendar Date Picker

Date selection with time module integration.

```js
let dateLabel = widget.NewLabel("No date selected")

let calendar = widget.NewCalendar(time.now(), (date) => {
    dateLabel.Text = `Selected: ${date.year}-${date.month}-${date.day}`
    let formatted = date.format("Monday, January 2, 2006")
    print(`Date: ${formatted}`)
})

let layout = container.NewVBox(
    widget.NewLabel("Select a date:"),
    calendar,
    dateLabel
)

window.SetContent(layout)
```

**Key concepts:**
- `time.now()` for current date
- `time.date(year, month, day)` for specific dates
- Date object with `year`, `month`, `day` properties
- `date.format()` for custom formatting
- Calendar callback receives TimeObject

---

## List with Virtualization

Efficient scrolling list for large datasets.

```js
let items = ["Apple", "Banana", "Cherry", "Date", "Elderberry"]
let itemCount = len(items)

let statusLabel = widget.NewLabel("Select an item")

let myList = widget.NewList()

myList.Length(() => {
    return itemCount
})

myList.CreateItem(() => {
    return widget.NewLabel("")
})

myList.UpdateItem((id, item) => {
    item.Text = items[id]
})

myList.OnSelected((id) => {
    statusLabel.Text = `Selected: ${items[id]}`
})

let layout = container.NewBorder(
    statusLabel,
    nil,
    nil,
    nil,
    myList
)

window.SetContent(layout)
```

**Key concepts:**
- Virtualized rendering (only visible items rendered)
- `Length()` callback for item count
- `CreateItem()` creates template widget
- `UpdateItem()` populates specific item
- `OnSelected()` for selection events
- Efficient for thousands of items

---

## Code Organization with ImportScript

Split code across multiple files for better organization.

**utils.risor:**
```js
let greet = (name) => {
    return `Hello, ${name}!`
}

let add = (a, b) => { return a + b }
```

**app.risor:**
```js
let result = greet("World")
let sum = add(5, 3)

let label = widget.NewLabel(`${result} Sum: ${sum}`)
window.SetContent(label)
```

**main.go:**
```go
package main

import (
    "os"
    "github.com/uidbz/fynerisor"
)

func main() {
    fw := fynerisor.NewApp("My App")

    // Import utilities before main script
    fw.ImportScript("utils.risor")

    // Load main script
    script, _ := os.ReadFile("app.risor")
    fw.LoadScript(string(script))

    fw.Execute()
    fw.ShowAndRun()
}
```

**Key concepts:**
- `ImportScript()` for loading code from files/URLs
- Scripts are concatenated before execution
- Shared global scope
- Can import multiple files
- Supports HTTP(S) URLs: `ImportScript("https://example.com/lib.risor")`

---

## Dialogs and User Interaction

Show dialogs for user confirmation, file selection, color picking, and custom content.

```js
require(["v0.5"])

// Simple information dialog
dialog.ShowInformation("Success", "Operation completed successfully")

// Error dialog
dialog.ShowError("Something went wrong!")

// Confirmation dialog with callback
dialog.ShowConfirm("Confirm Delete", "Are you sure?", (confirmed) => {
    if (confirmed) {
        print("User confirmed deletion")
    }
})

// File picker
dialog.ShowFileOpen((path, err) => {
    if (path != nil) {
        print("Selected file:", path)
    }
})

// Color picker with RGB callback
dialog.ShowColorPicker("Choose Color", "Select a color", (color) => {
    print(`RGB(${color.R}, ${color.G}, ${color.B})`)
})

// Form dialog with validation
let nameEntry = widget.NewEntry()
let emailEntry = widget.NewEntry()
let nameItem = widget.NewFormItem("Name", nameEntry)
let emailItem = widget.NewFormItem("Email", emailEntry)

dialog.ShowForm("User Info", "Submit", "Cancel", [nameItem, emailItem], (submitted) => {
    if (submitted) {
        print("Form submitted!")
    }
})
```

**Advanced dialog control:**
```js
// Create dialog with more control
let fileDialog = dialog.NewFileOpen((path, err) => {
    if (path != nil) {
        window.SetStatus("Opened: " + path)
    }
})

// Customize before showing
fileDialog.SetFileName("default.txt")
fileDialog.SetFilter(".txt")       // Only show .txt files
fileDialog.SetLocation("/tmp")     // Start in /tmp directory
fileDialog.Show()

// Custom content dialog
let label = widget.NewLabel("This is custom content!")
let entry = widget.NewEntry()
let content = container.NewVBox([label, entry])
dialog.ShowCustom("Custom Dialog", "OK", content)
```

**Key concepts:**
- Simple Show* functions for common cases
- New* constructors for advanced customization
- Callbacks for user responses
- File dialogs return path strings
- Color picker returns RGB map
- Form dialogs include automatic validation

---

## Complete Examples

The `examples/` directory contains complete working applications:

**GUI Examples:**
1. **01-hello-world** - Minimal app with label
2. **02-button-callback** - Interactive button with state
3. **03-form-input** - Form with validation
4. **04-table-display** - Table with paginated data
5. **05-progress-widgets** - Progress bars and sliders
6. **06-icon-hyperlink-card** - Icons, links, and cards
7. **07-radiogroup** - Radio button groups
8. **08-accordion** - Collapsible sections
9. **09-toolbar** - Action toolbar with icons
10. **10-calendar** - Date picker with time module
11. **11-list** - Virtualized scrolling list
12. **12-tree** - Hierarchical tree widget
13. **12-imports** - Code organization with ImportScript
14. **13-http-fetch** - HTTP requests and JSON parsing
15. **15-sql-test** - Database connectivity
30. **30-dialogs** - Dialog windows (info, error, confirm, file picker, color picker, forms)

**Headless Example:**
16. **16-context-builder** - Headless script execution

Each example includes:
- Complete Go program (main.go)
- Risor script (*.risor)
- README with explanations

---

## Background Tasks with go()

Run long operations in background without blocking the GUI.

```js
require(["v0.2", "@gui"])

let statusLabel = widget.NewLabel("Ready")
let startButton = widget.NewButton("Start Processing", () => {
    statusLabel.Text = "Processing..."
    
    // Run in background
    go(() => {
        // Simulate long operation
        let result = 0
        range(1000000).each((i) => {
            result = result + i
        })
        
        // Update GUI safely from background thread
        window.Do(() => {
            statusLabel.Text = sprintf("Complete! Result: %d", result)
        })
    })
})

let layout = container.NewVBox(statusLabel, startButton)
window.SetContent(layout)
```

**Key concepts:**
- `go(() => { ... })` spawns goroutine
- Long operations won't freeze GUI
- Use `window.Do()` for GUI updates from goroutines
- `print()` is thread-safe and works in `go()`

**Important:** Direct widget updates from `go()` are unsafe. Always use `window.Do()`:

```js
// ❌ UNSAFE - crashes or corrupts GUI
go(() => {
    label.SetText("Done")  // Direct update from goroutine
})

// ✅ SAFE - queued to GUI thread
go(() => {
    window.Do(() => {
        label.SetText("Done")  // Wrapped in window.Do()
    })
})
```

---

## File Operations with io module

Copy files and perform I/O operations.

```js
require(["v0.2", "@io"])

let sourceEntry = widget.NewEntry()
sourceEntry.SetPlaceHolder("/path/to/source.txt")

let destEntry = widget.NewEntry()
destEntry.SetPlaceHolder("/path/to/destination.txt")

let statusLabel = widget.NewLabel("Ready")

let copyButton = widget.NewButton("Copy File", () => {
    let src = sourceEntry.Text
    let dst = destEntry.Text
    
    if (src == "" || dst == "") {
        statusLabel.Text = "Error: Both paths required"
        return
    }
    
    try {
        io.cp(src, dst)
        statusLabel.Text = sprintf("Copied %s to %s", src, dst)
    } catch (err) {
        statusLabel.Text = sprintf("Error: %s", err)
    }
})

let formItems = [
    widget.NewFormItem("Source:", sourceEntry),
    widget.NewFormItem("Destination:", destEntry),
]
let form = widget.NewForm(formItems)

let layout = container.NewVBox(form, copyButton, statusLabel)
window.SetContent(layout)
```

**Key concepts:**
- `require(["@io"])` to enable io module
- `io.cp(src, dst)` copies files
- Use `try/catch` for error handling
- Enabled via `WithIO()` option in Go code

---

## Detecting Application Context

Conditional behavior based on which app is running the script.

```js
require(["v0.2"])

let message = ""

if (app.name == "goto") {
    message = "Running in goto - web-like navigation available"
} else if (app.name == "lars") {
    message = "Running in LARS - runtime system context"
} else if (app.name == "fynerisor") {
    message = "Running in standalone fynerisor"
} else {
    message = sprintf("Running in: %s", app.name)
}

let label = widget.NewLabel(message)
window.SetContent(label)
```

**Key concepts:**
- `app.name` contains the application name
- Set via `WithAppName("myapp")` in Go
- Default is `"fynerisor"`
- Enables environment-specific features

**Example use cases:**
```js
// Different behavior per app
if (app.name == "goto") {
    // Use gui.goto() for navigation
    button = widget.NewButton("Next", () => { gui.goto("next.risor") })
} else {
    // Use different navigation
    button = widget.NewButton("Next", () => { loadNextScreen() })
}

// Config path based on app
let configPath = sprintf("/home/user/.%s/config.toml", app.name)

// Feature detection
let hasNavigation = (app.name == "goto")
```

---

## Keyboard Shortcuts

Global keyboard shortcuts and menu integration.

```js
require(["v0.6"])

let statusLabel = widget.NewLabel("Press Ctrl+S to save, Ctrl+Q to quit")

// Register global shortcuts
window.AddShortcut("Ctrl+S", () => {
    window.SetStatus("Saved!")
    statusLabel.SetText("Document saved at " + time.now().format("15:04:05"))
})

window.AddShortcut("Ctrl+Q", () => {
    window.SetStatus("Quitting...")
    app.Quit()
})

// Function keys
window.AddShortcut("F5", () => {
    window.SetStatus("Refreshed!")
})

// Alt combinations
window.AddShortcut("Alt+Shift+N", () => {
    window.SetStatus("New item created")
})

window.SetContent(statusLabel)
```

**Key concepts:**
- `window.AddShortcut(shortcut, callback)` - Register global shortcuts
- `window.RemoveShortcut(shortcut)` - Remove shortcuts
- Shortcuts work without visible menus
- Cross-platform modifiers: `Ctrl`/`Control`, `Alt`/`Option`, `Super`/`Cmd`/`Command`

**Menu integration:**
```js
// Create menu items with shortcut hints
let saveItem = fyne.NewMenuItem("Save", () => {
    window.SetStatus("Saved from menu!")
})
saveItem.Shortcut = "Ctrl+S"  // Display only, register separately

let quitItem = fyne.NewMenuItem("Quit", () => {
    app.Quit()
})
quitItem.Shortcut = "Ctrl+Q"

// Create menu
let fileMenu = fyne.NewMenu("File", saveItem, quitItem)
let mainMenu = fyne.NewMainMenu(fileMenu)
window.SetMainMenu(mainMenu)

// Register actual shortcuts
window.AddShortcut("Ctrl+S", () => { window.SetStatus("Saved!") })
window.AddShortcut("Ctrl+Q", () => { app.Quit() })
```

**Supported keys:**
- Letters: `A-Z`
- Numbers: `0-9`
- Function keys: `F1-F12`
- Special keys: `Escape`, `Tab`, `Return`, `Space`, `Backspace`, `Delete`, `Left`, `Right`, `Up`, `Down`, `Home`, `End`, `PageUp`, `PageDown`

---

For more details, see the [examples/](examples/) directory with complete working applications.
