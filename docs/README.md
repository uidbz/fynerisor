# fynerisor

Risor language bindings for the Fyne GUI framework.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Installation](#installation)
  - [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Non-GUI Usage](#non-gui-usage)
- [Global Objects](#global-objects)
- [Widgets](#widgets)
  - [Button](#button)
  - [Label](#label)
  - [Entry](#entry)
  - [MultiLineEntry](#multilineentry)
  - [Check](#check)
  - [CheckGroup](#checkgroup)
  - [Select](#select)
  - [Form](#form)
  - [Table](#table)
  - [Log](#log)
  - [Markdown](#markdown)
  - [ProgressBar](#progressbar)
  - [ProgressBarInfinite](#progressbarinfinite)
  - [Activity](#activity)
  - [Slider](#slider)
  - [Icon](#icon)
  - [Hyperlink](#hyperlink)
  - [Card](#card)
  - [RadioGroup](#radiogroup)
  - [Accordion](#accordion)
  - [Toolbar](#toolbar)
  - [Separator](#separator)
- [Containers](#containers)
  - [VBox](#vbox)
  - [HBox](#hbox)
  - [Border](#border)
  - [HSplit](#hsplit)
  - [VSplit](#vsplit)
  - [Scroll](#scroll)
- [Canvas Objects](#canvas-objects)
  - [Image](#image)
  - [Line](#line)
- [Charts](#charts)
  - [BarChart](#barchart)
- [Window Object](#window-object)
  - [SetContent](#setcontent)
  - [OnDropped](#ondropped)
  - [CaptureStdout](#capturestdout)
  - [Properties](#properties)
- [Callbacks](#callbacks)
- [Passing Custom Data](#passing-custom-data)
- [Examples](#examples)
- [Testing](#testing)
  - [Running Tests](#running-tests)
  - [Test Coverage](#test-coverage)
  - [Writing Tests](#writing-tests)
- [Architecture](#architecture)
  - [Script Execution](#script-execution)
  - [Widget Wrapping](#widget-wrapping)
  - [Callback Execution](#callback-execution)
- [Risor v2 Syntax](#risor-v2-syntax)
- [License](#license)
- [Contributing](#contributing)
- [Support](#support)

## Overview

fynerisor enables building cross-platform desktop applications using Risor scripts with full access to Fyne's widget and container system. Create interactive GUIs with buttons, forms, tables, charts, and more through simple factory functions.

## Version

**Current Version**: 0.8.2  
**Risor Compatibility**: v2.2+ (arrow functions required)  
**Fyne Compatibility**: v2.8+

### What's New in v0.8.2

- **Multi-row (hierarchical) table headers**: `table.HeaderLevels(fn)` draws a stacked header above the columns, repeating a parent once across the run it covers so a merged source cell survives — see `examples/41-table-header-levels`
- **Tie Module**: `insert_table` accepts header *rows* as well as a flat label list, and `read_table` returns them as `header_levels` — see [TIE_MODULE.md](TIE_MODULE.md)
- **Tie v0.5.1**: Upgraded the tie dependency, which stores the header levels alongside the table

### What's New in v0.8.1

- **Fyne v2.8.1**: Upgraded the Fyne GUI framework
- **Dialog sizing**: `Resize(width, height)` on custom/confirm/form dialogs, plus a `window.Size` getter (`{width, height}`) for computing relative sizes
- **ProgressBarInfinite**: Now starts idle — visible only while running (`Start()`/`Stop()`)
- **Tie Module**: `delete_table`, plus `sort_by_value` and `descending` options on `query`
- **Stability**: Table/List getters stay resilient under VM contention; `ErrConcurrentAccess` is now exported; FlexTable resets scroll offset on `SetData`

### What's New in v0.8.0

- **Tie Module**: Triple store client (`tie.connect`, then `add`/`get`/`set`/`update`/`delete`, `query`, `expand`, `batch`, `dump_stream`, `insert_table`/`read_table`, etc.) — see [TIE_MODULE.md](TIE_MODULE.md) (enable with `require(["@tie"])`)
- **JSON Module**: `json.parse`, `json.marshal`, `json.marshal_indent`, `json.valid`, `json.read`, `json.write` (enable with `require(["@json"])`)
- **CSV Module**: `csv.parse`, `csv.format`, `csv.read`, `csv.write` — header rows map to lists of maps by default, with `{header: false}`, `delimiter`, and `columns` options (enable with `require(["@csv"])`)
- **Risor v2.2.0**: Upgraded the Risor language runtime

### What's New in v0.7.0

- **Browser Package**: Reusable `browser/` UI for building browser-style apps that load Risor scripts from HTTP(S)/`file://` URLs, with navigation, history, source view, plugins, and programmatic navigation from scripts (`browser.Open`, `browser.params`, etc.)
- **Reference Browser + Android**: `cmd/fynerisor-browser` can be packaged as an Android APK (`fyne package -os android`)
- **Charts**: Bar, line, scatter, histogram, and box plot charts with statistical helpers
- **Concurrency Guard**: `vmguard` detects concurrent Risor VM access and recovers from callback panics
- **time.unix(seconds)**: Construct a time value from a Unix timestamp
- **Table Performance**: Column resize no longer does a per-cell VM round-trip
- **Browser Security**: IO module disabled by default in the reference browser

### What's New in v0.6.0

- **Keyboard Shortcuts**: Global shortcuts with `window.AddShortcut("Ctrl+S", callback)` and menu integration via `MenuItem.Shortcut`
- **48+ Widget Bindings**: Complete widget coverage across all priority levels
- **Data Binding System**: Reactive UI with String, Bool, Int, and Float bindings
- **Dialog Support**: File pickers, confirmations, color pickers, forms, and custom dialogs
- **Module-Scoped Imports**: Namespace isolation with `import()` function
- **SQL Module**: Database connectivity with MySQL, PostgreSQL, SQLite, and SQL Server support
- **HTTP Module**: REST API calls with JSON parsing
- **Risor v2 Migration**: Full migration to Risor v2 with arrow function syntax
- **Advanced Widgets**: Tree and List widgets for hierarchical and scrolling data
- **Critical Bug Fixes**: Fixed race condition in widget callbacks (see CONCURRENCY.md)
- **CLI Enhancements**: Error messages now display without --verbose flag
- **Iterator Support**: SQL row iterators with `.map()` and `.each()` for streaming
- **Enhanced require()**: List syntax and module requirements (`require(["v0.6", "@sql"])`)
- **Requirement Analysis**: `AnalyzeRequirements()` function for external tools

## Features

- **27 Fyne Widget Bindings** (47% coverage): Buttons, labels, entries, forms, tables, calendars, lists, progress bars, sliders, icons, cards, accordions, toolbars, and more
- **Layout Containers**: VBox, HBox, Border, Split, Scroll layouts
- **Interactive Callbacks**: Handle button clicks, form submissions, table interactions, date selections
- **Script-based UI**: Build and modify interfaces using Risor scripts
- **Risor v2 Support**: Built for Risor v2 with modern syntax (arrow functions, template strings, .each())
- **Standard Modules**: Time, HTTP, OS, Strings, Filepath modules included
- **Code Organization**: ImportScript() for splitting code across files and URLs
- **Cross-platform**: Works on Windows, Linux, and macOS
- **CLI Tool**: `fynerisor-cli` for running scripts directly with watch mode, stdin, headless testing
- **See [WIDGET_STATUS.md](WIDGET_STATUS.md)** for complete widget implementation status

## Installation

```bash
go get github.com/uidbz/fynerisor
```

### Prerequisites

fynerisor requires Fyne's system dependencies:

- **Linux**: `xorg-dev` or `wayland-dev` packages
- **Windows**: C compiler (TDM-GCC or MSYS2)
- **macOS**: Xcode command line tools

See [Fyne Prerequisites](https://docs.fyne.io/started/) for detailed setup instructions.

## Quick Start

### Simple Approach (Recommended)

**Go Program** (`main.go`):
```go
package main

import "github.com/uidbz/fynerisor/gui"

func main() {
	// NewApp creates both Fyne app and window in one call
	fw := gui.NewApp("Hello World",
		gui.WithHTTP(),
	)

	script := `
		require(["v0.6", "@http"])
		let btn = widget.NewButton("Click Me", () => {
			window.SetStatus("Clicked!")
		})
		window.SetContent(btn)
	`

	fw.LoadScript(script)
	fw.Execute()
	fw.ShowAndRun()
}
```

### Advanced Approach (More Control)

**Go Program** (`main.go`):
```go
package main

import (
	"fyne.io/fyne/v2/app"
	"github.com/uidbz/fynerisor"
	"os"
)

func main() {
	// Create Fyne app with custom settings
	a := app.New()
	w := a.NewWindow("Hello World")

	// Wrap with fynerisor
	fw := gui.NewWindow(w,
		gui.WithHTTP(),
	)

	// Load and execute script
	script, _ := os.ReadFile("app.risor")
	fw.LoadScript(string(script))
	fw.Execute()

	// Show window
	fw.ShowAndRun()
}
```

**Risor Script** (`app.risor`):
```js
let label = widget.NewLabel("Hello, World!")
let button = widget.NewButton("Click me", () => {
    label.Text = "Button clicked!"
})

let layout = container.NewVBox(label, button)
window.SetContent(layout)
```

Run with: `go run main.go`

## Non-GUI Usage

Use fynerisor's module system (HTTP, SQL, OS, etc.) and import functionality in headless scripts, CLI tools, or server applications without GUI dependencies.

**Example:**

```go
package main

import (
	"fmt"
	"log"
	"github.com/uidbz/fynerisor/core"
)

func main() {
	// Create context with desired modules (same options as Window)
	ctx := core.NewContext(
		core.WithHTTP(),
		core.WithSQL(),
		core.WithOS(),
	)

	script := `
		require(["v0.6", "@http", "@sql", "@os"])

		let platform = os.goos()
		print("Running on:", platform)

		let data = http.get("https://httpbin.org/get").json()
		print("Fetched:", data)

		// Use SQL
		let conn = sql.connect("sqlite3::memory:")
		conn.exec("CREATE TABLE users (id INT, name TEXT)")
		conn.exec("INSERT INTO users VALUES (?, ?)", 1, "Alice")
		
		let rows = conn.query("SELECT * FROM users")
		print("Users:", rows)
	`

	result, err := ctx.Eval(script)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Result:", result)
}
```

**With imports:**

```go
import "os"

// Fetch function to load imports
fetchFunc := func(path string) (string, error) {
	data, err := os.ReadFile(path)
	return string(data), err
}

script := `
	import("utils.risor")
	require(["v0.6", "@http"])
	
	let result = myUtilFunction(42)
`

result, err := ctx.EvalWithImports(script, fetchFunc)
```

**Available options (same as Window):**
- `WithHTTP()` - HTTP module
- `WithSQL()` - SQL module  
- `WithTie()` - Tie triple store module
- `WithOS()` - OS module
- `WithIO()` - File I/O module
- `WithExec()` - Exec module
- `WithStrings()` - Strings module
- `WithFilepath()` - Filepath module
- `WithTime()` - Time module
- `WithJSON()` - JSON module
- `WithCSV()` - CSV module
- `WithRisorOptions()` - Custom globals

Note: Window-specific options like `WithStatusCallback()` and `WithResultCallback()` are silently ignored when used with ContextBuilder.

See [example 16-context-builder](examples/16-context-builder) for complete usage.

## Global Objects

Risor scripts have access to these global factory objects:

- **window**: Window control (`SetContent`, `SetStatus`, `AddShortcut`, `RemoveShortcut`, `SetMainMenu`, `OnDropped`, `CaptureStdout`)
- **widget**: Widget factory (`NewButton`, `NewLabel`, `NewEntry`, `NewTable`, etc.)
- **container**: Layout factory (`NewVBox`, `NewHBox`, `NewBorder`, `NewSplit`, `NewScroll`, `NewGrid`, `NewStack`)
- **canvas**: Canvas objects (`NewLine`, `NewImageFromURI`, `NewRectangle`, `NewCircle`, `NewText`)
- **chart**: Chart factory (`NewBarChart`, `NewLineChart`, `NewPieChart`)
- **dialog**: Dialog functions (`ShowFileOpen`, `ShowFileSave`, `ShowConfirm`, `ShowColorPicker`, `ShowForm`, `ShowCustom`)
- **binding**: Data binding factory (`NewString`, `NewBool`, `NewInt`, `NewFloat`)
- **constants**: Fyne constants (`ImportanceHigh`, `ImportanceLow`, etc.)
- **fyne**: Fyne utilities (`NewMenu`, `NewMenuItem`, `NewMainMenu`, `NewSize`, `NewPos`)
- **app**: Application metadata (`app.name`, `app.version`)
- **print**: Output function
- **require**: Version and module validation (`require(["v0.6", "@sql", "@http"])`)
- **import**: Module imports with namespace isolation (`import("math", "mymath")`)

## Widgets

### Button
Clickable button with callback function.

```js
let btn = widget.NewButton("Click Me", () => {
    print("Button clicked!")
})
```

### Label
Display static or dynamic text. Text property is read/write.

```js
let label = widget.NewLabel("Initial text")
label.Text = "Updated text"
```

### Entry
Single-line text input field.

```js
let entry = widget.NewEntry()
entry.Text = "Default value"
entry.OnSubmitted((text) => {
    print("Submitted:", text)
})
```

### MultiLineEntry
Multi-line text input area.

```js
let textArea = widget.NewMultiLineEntry()
```

### Check
Checkbox with change callback.

```js
let check = widget.NewCheck("Enable feature", (checked) => {
    print("Checked:", checked)
})
check.SetChecked(true)
let isChecked = check.Checked
```

### CheckGroup
Multiple checkboxes with multi-selection.

```js
let options = ["Option 1", "Option 2", "Option 3"]
let group = widget.NewCheckGroup(options, (selected) => {
    print("Selected:", selected)
})
```

### Select
Dropdown selection widget.

```js
let options = ["Red", "Green", "Blue"]
let dropdown = widget.NewSelect(options, (selected) => {
    print("Selected:", selected)
})
```

### Form
Form layout with labeled fields. Takes an array of FormItem widgets.

```js
let nameEntry = widget.NewEntry()
let emailEntry = widget.NewEntry()

let formItems = [
    widget.NewFormItem("Name:", nameEntry),
    widget.NewFormItem("Email:", emailEntry),
]

let form = widget.NewForm(formItems)
```

### Table
Advanced table widget with pagination and data callbacks.

```js
let table = widget.NewTable("My Table", 20)

table.Columns(() => {
    return ["ID", "Name", "Status"]
})

table.RowCount(() => {
    return 100
})

table.Data((offset, limit) => {
    // Return rows for current page
    return [
        ["1", "Alice", "Active"],
        ["2", "Bob", "Inactive"]
    ]
})

table.SetOnClick((row, col) => {
    print(`Clicked row ${row}, column ${col}`)
})
```

#### Multi-row (hierarchical) headers

`HeaderLevels` draws several header rows above the columns, for data where the top
row groups the columns below it. The list is row-major — `levels[i][j]` is level
`i` of column `j` — which is the shape `db.read_table` reports `header_levels` in,
so a table read from tie can be wired straight through:

```js
table.Columns(() => ["Sample", "20°C\x1fRep 1", "20°C\x1fRep 2", "37°C\x1fRep 1"])
table.HeaderLevels(() => [
    ["Sample", "20°C",  "20°C",  "37°C"],
    ["",       "Rep 1", "Rep 2", "Rep 1"],
])
```

A parent that covers several columns is repeated across them (as `20°C` is above),
and the widget draws it once across the run — so pass the levels forward-filled,
with blanks explicit, and do not try to pre-merge them yourself.

`HeaderLevels` is display-only: `Columns` stays the column identity used for
sorting, filtering and export, and the header row height follows the number of
levels. Returning an empty list restores the single-row default, where each column
shows its own name.

See `examples/41-table-header-levels`.

### Log
Scrolling log widget with item limit.

```js
let log = widget.NewLog(100)
log.Append("Log entry 1")
log.Append("Log entry 2")
log.Clear()
```

### Markdown
Render CommonMark markdown.

```js
let md = widget.NewMarkdown("# Title\n\nSome **bold** text")
```

### ProgressBar
Display determinate progress (0.0 to 1.0).

```js
let progress = widget.NewProgressBar()
progress.SetValue(0.5)  // 50%
```

### ProgressBarInfinite
Animated indeterminate progress indicator.

```js
let spinner = widget.NewProgressBarInfinite()
spinner.Start()
spinner.Stop()
```

### Activity
Simple circular activity indicator.

```js
let activity = widget.NewActivity()
activity.Start()
activity.Stop()
```

### Slider
Numeric value selection with range.

```js
let slider = widget.NewSlider(0.0, 100.0)
slider.SetValue(50.0)
slider.OnChanged((value) => {
    print('Value: {value}')
})
```

### Icon
Display theme icons that adapt to light/dark themes.

```js
let icon = widget.NewIcon("search")
icon.SetResource("settings")  // Change icon
```

### Hyperlink
Clickable link that opens URLs in browser.

```js
let link = widget.NewHyperlink("Fyne Docs", "https://docs.fyne.io")
link.OnTapped(() => {
    print("Custom action instead of browser")
})
```

### Card
Container with title, subtitle, and content.

```js
let card = widget.NewCard(
    "Card Title",
    "Subtitle",
    widget.NewLabel("Content")
)
card.SetTitle("New Title")
```

### RadioGroup
Single-selection radio buttons.

```js
let radio = widget.NewRadioGroup(
    ["Option 1", "Option 2", "Option 3"],
    (selected) => {
        print('Selected: {selected}')
    }
)
radio.SetSelected("Option 2")
radio.Horizontal = true  // Horizontal layout
```

### SelectEntry
Searchable dropdown entry - combines text entry with dropdown selection.

```js
let options = ["Apple", "Banana", "Cherry", "Date", "Elderberry"]
let selectEntry = widget.NewSelectEntry(options)

selectEntry.PlaceHolder = "Search fruits..."
selectEntry.OnChanged((text) => {
    print("Typed/selected: " + text)
})

// Update options dynamically
selectEntry.SetOptions(["New", "Options", "List"])

// Set text programmatically
selectEntry.SetText("Apple")

// Read current text
let current = selectEntry.Text
```

### DateEntry
Date entry field with YYYY-MM-DD format.

```js
let dateEntry = widget.NewDateEntry()

dateEntry.OnChanged((date) => {
    print("Date entered: " + date)
})

// Set date programmatically
dateEntry.SetDate("2024-12-25")

// Or set text directly
dateEntry.SetText("2024-01-15")

// Read current date
let currentDate = dateEntry.Text
```

### Accordion
Collapsible sections.

```js
let item1 = widget.NewAccordionItem("Section 1", "", content1)
let item2 = widget.NewAccordionItem("Section 2", "", content2)
let accordion = widget.NewAccordion(item1, item2)

accordion.Open(0)        // Open first item
accordion.Close(0)       // Close first item
accordion.MultiOpen = true  // Allow multiple open
```

### Toolbar
Action toolbar with icon buttons.

```js
let save = widget.NewToolbarAction("documentSave", () => {
    print("Save clicked")
})
let open = widget.NewToolbarAction("folderOpen", () => {
    print("Open clicked")
})

let toolbar = widget.NewToolbar(
    save, open,
    widget.NewToolbarSeparator(),
    widget.NewToolbarSpacer(),  // Push to right
    widget.NewToolbarAction("settings", () => {})
)
```

### Separator
Visual horizontal or vertical line.

```js
let separator = widget.NewSeparator()
```

## Containers

### VBox
Vertical box - stacks widgets vertically.

```js
let vbox = container.NewVBox(
    widget.NewLabel("Top"),
    widget.NewLabel("Middle"),
    widget.NewLabel("Bottom")
)
```

### HBox
Horizontal box - arranges widgets horizontally.

```js
let hbox = container.NewHBox(
    widget.NewLabel("Left"),
    widget.NewLabel("Center"),
    widget.NewLabel("Right")
)
```

### Border
Border layout with five regions. Use `nil` for empty regions.

```js
let border = container.NewBorder(
    widget.NewLabel("Top"),
    widget.NewLabel("Bottom"),
    widget.NewLabel("Left"),
    widget.NewLabel("Right"),
    widget.NewLabel("Center")
)
```

### HSplit
Horizontal split with draggable divider.

```js
let split = container.NewHSplit(
    widget.NewLabel("Left pane"),
    widget.NewLabel("Right pane")
)
```

### VSplit
Vertical split with draggable divider.

```js
let split = container.NewVSplit(
    widget.NewLabel("Top pane"),
    widget.NewLabel("Bottom pane")
)
```

### Scroll
Scrollable container for large content.

```js
let content = container.NewVBox(
    widget.NewLabel("Item 1"),
    // ... many items
    widget.NewLabel("Item 100")
)

let scrollable = container.NewScroll(content)
window.SetContent(scrollable)
```

## Canvas Objects

### Image
Display images from URIs.

```js
let img = canvas.NewImageFromURI("file:///path/to/image.png")
img.SetImageFromURI("file:///other/image.png")
```

### Line
Draw colored lines.

```js
let line = canvas.NewLine("red")
// Colors: red, green, blue, black, white, yellow, cyan, magenta
```

## Charts

### BarChart
Bar chart visualization.

```js
let chart = chart.NewBarChart(
    "Sales Report",           // title
    "Revenue",               // y-axis label
    ["Q1", "Q2", "Q3", "Q4"], // x-axis labels
    [100, 150, 120, 180]     // values
)
```

## Window Object

The `window` object provides window control functions.

### SetContent
Set the window content to a widget.

```js
window.SetContent(myWidget)
```

### SetStatus
Set the window status message (calls status callback if provided).

```js
window.SetStatus("Loading...")
window.SetStatus("Ready!")
```

### OnDropped
Handle file drag-and-drop events.

```js
window.OnDropped((paths) => {
    print("Dropped files:", paths)
})
```

### AddShortcut
Register a global keyboard shortcut.

```js
window.AddShortcut("Ctrl+S", () => {
    print("Save triggered!")
})

// Cross-platform modifiers: Ctrl/Control, Alt/Option, Super/Cmd/Command
window.AddShortcut("Alt+Shift+N", () => {
    print("New item")
})

// Function keys
window.AddShortcut("F5", () => {
    print("Refresh")
})
```

### RemoveShortcut
Remove a previously registered shortcut.

```js
window.RemoveShortcut("Ctrl+S")
```

### SetMainMenu
Set the window's main menu bar (macOS only, ignored on other platforms).

```js
let fileMenu = fyne.NewMenu("File",
    fyne.NewMenuItem("Save", saveCallback),
    fyne.NewMenuItem("Quit", quitCallback)
)
let mainMenu = fyne.NewMainMenu(fileMenu)
window.SetMainMenu(mainMenu)
```

### CaptureStdout
Capture stdout in a callback.

```js
window.CaptureStdout((line) => {
    print("Captured:", line)
})
```

### Properties
- `window.DroppedPaths` - List of dropped file paths

## Callbacks

Callbacks in Go code allow monitoring script execution:

```go
fyneWindow := gui.NewWindow(w,
    gui.WithStatusCallback(func(status string) {
        fmt.Printf("Status: %s\n", status)
    }),
    gui.WithResultCallback(func(result string) {
        fmt.Printf("Result: %s\n", result)
    }),
)
```

Callbacks are optional - omit the options if not needed.

## Passing Custom Data

Pass Go functions to Risor scripts using `risor.WithEnv`:

```go
import "github.com/deepnoodle-ai/risor/v2"

dataFuncs := map[string]any{
    "data": map[string]any{
        "getUsers": func() []string {
            return []string{"Alice", "Bob", "Carol"}
        },
    },
}

fyneWindow := gui.NewWindow(w,
    gui.WithRisorOptions(risor.WithEnv(dataFuncs)),
)
```

In Risor script:
```js
let users = data.getUsers()
```

## Modules

Fynerisor includes standard modules that can be enabled via options:

### HTTP Module
Enable with `gui.WithHTTP()`:

```js
let response = http.get("https://api.example.com/data")
print(response.status)    // HTTP status code
print(response.ok)         // true if 2xx status

// Parse JSON response
let data = response.json()
print(data["key"])

// Other methods
http.post(url, headers, body)
http.put(url, headers, body)
http.delete(url, headers)
```

### OS Module
Enable with `gui.WithOS()`:

```js
let platform = os.goos()           // "linux", "windows", "darwin"
let user = os.current_user()       // Map with username, uid, home_dir, etc.
os.open_browser("https://example.com")  // Open URL in default browser
```

### Strings Module
Enable with `gui.WithStrings()`:

```js
strings.replace_all(str, old, new)
strings.trim_prefix(str, prefix)
strings.split(str, sep)
strings.join(list, sep)
strings.to_lower(str)
strings.to_upper(str)
```

### Filepath Module
Enable with `gui.WithFilepath()`:

```js
filepath.join(parts...)
filepath.base(path)
filepath.dir(path)
filepath.ext(path)
filepath.abs(path)
```

## CLI Tool

The `fynerisor-cli` command runs Risor scripts directly:

```bash
# Run a script
fynerisor-cli script.risor

# With options
fynerisor-cli --title "My App" --width 1024 --height 768 script.risor

# Watch mode (auto-reload on file change)
fynerisor-cli --watch script.risor

# Verbose mode (show status messages)
fynerisor-cli --verbose script.risor

# Headless mode (for testing)
fynerisor-cli --headless script.risor

# Read from stdin
cat script.risor | fynerisor-cli -

# Pass script arguments
fynerisor-cli script.risor arg1 arg2
# Access in script via: args[0], args[1]

# Custom globals from JSON file
fynerisor-cli --globals data.json script.risor
```

## Examples

See the [examples/](examples/) directory for complete working applications:

- **01-hello-world**: Basic setup with file-based script
- **02-button-callback**: Interactive widgets and callbacks
- **03-table**: Dynamic table with data loading
- **04-http-fetch**: HTTP requests with JSON parsing
- **03-form-input**: Form validation and input handling
- **04-table-display**: Table widget with paginated data
- **05-progress-slider**: Progress bars, sliders, and activity indicators
- **06-icon-hyperlink-card**: Theme icons, hyperlinks, and card layouts
- **07-radiogroup**: Single-selection radio button groups
- **08-accordion**: Collapsible accordion sections
- **09-toolbar**: Action toolbars with icons

Each example includes a README with detailed explanations and usage patterns.

## Testing

The fynerisor package includes a comprehensive automated test suite covering all widgets, containers, properties, and callbacks.

### Running Tests

Run all tests:

```bash
cd fynerisor
go test
```

Run tests with verbose output:

```bash
go test -v
```

Generate coverage report:

```bash
go test -cover
```

Generate detailed HTML coverage report:

```bash
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Coverage

The test suite (`fynerisor_test.go`) includes 9 comprehensive test categories:

**TestWidgetCreation** - Verifies all widget types can be created:
- Button, Label, Entry, Check, RadioGroup, Slider, ProgressBar, Icon, Card, Accordion, Toolbar

**TestContainerCreation** - Tests container creation:
- VBox, HBox, Border, Scroll

**TestWidgetProperties** - Tests property access and modification:
- Label.Text (read/write)
- Slider.Value (read/write)
- ProgressBar.Value (read/write)

**TestComplexLayout** - Tests nested containers:
- Border container with HBox header, center content, and footer
- Multiple levels of nesting

**TestCallbacks** - Tests widget callback execution:
- Button OnTapped callback
- Check OnChanged callback
- RadioGroup OnChanged callback with SetSelected()

**TestTableWidget** - Tests table widget functionality:
- Columns() callback for column headers
- RowCount() callback for total row count
- Data() callback for paginated data retrieval

**TestAccordionWidget** - Tests accordion functionality:
- Multiple AccordionItem instances
- MultiOpen property (allow multiple sections open)
- Open() method for expanding sections

**TestToolbarWidget** - Tests toolbar components:
- ToolbarAction with icon and callback
- ToolbarSeparator for visual separation
- ToolbarSpacer for layout control

**TestFormWidget** - Tests form functionality:
- FormItem with label and widget
- Form container with multiple items

### Writing Tests

All tests use Fyne's test application driver to simulate GUI operations without requiring a display:

```go
func TestMyWidget(t *testing.T) {
    a := test.NewApp()
    defer a.Quit()
    w := a.NewWindow("Test")
    fw := NewWindow(w)

    script := `
        let myWidget = widget.NewLabel("Test")
        window.SetContent(myWidget)
    `

    fw.LoadScript(script)
    fw.Execute()
    time.Sleep(50 * time.Millisecond)

    if fw.Status != "Ready!" {
        t.Errorf("Expected Ready!, got %s", fw.Status)
    }
}
```

Key points:
- Use `test.NewApp()` to create a headless test application
- Create a `NewWindow` wrapper for Risor script execution
- Use `time.Sleep()` to allow async script execution to complete
- Check `fw.Status` for script execution success

## Architecture

### Script Execution

1. Create a Fyne window with `app.NewWindow()`
2. Wrap it with `gui.NewWindow()`
3. Set the `Script` field to Risor code
4. Call `RunCode()` to execute (runs in goroutine)
5. Call `ShowAndRun()` to display the window

### Widget Wrapping

Each Fyne widget is wrapped in a Risor object that:
- Implements `object.Object` interface
- Implements `IsCanvasObject` interface with `CanvasObject()` method
- Exposes properties and methods via `GetAttr()`/`SetAttr()`
- Handles callbacks by queueing on the GUI thread

### Callback Execution

Widget callbacks execute on the GUI thread via a function queue. This ensures thread-safe UI updates from Risor scripts.

## Risor v2 Syntax

Key syntax notes for Risor v2:

- Variable declarations: `let x = value`
- Conditionals: `if (condition) { ... }` (parentheses required)
- Arrow functions: `() => { ... }` or `x => x * 2`
- Template strings: `` `Hello ${name}` ``
- No traditional for loops: use `.each()`, `.map()`, `.filter()`, `.reduce()`

See [Risor documentation](https://risor.io/docs/) for full language reference.

## Version Requirements in Scripts

To enforce version requirements in your Risor scripts, use the `require()` function:

```js
// Require minimum fynerisor version (v0.2 or higher)
require("v0.2")

let calendar = widget.NewCalendar(time.now(), (date) => {
    print(`Selected: ${date.year}-${date.month}-${date.day}`)
})

window.SetContent(calendar)
```

**The `require()` Function**:
```js
// Minimum version (>= operator) - accepts v0.2.0, v0.2.1, v0.3, v1.0, etc.
require("v0.2")

// Exact version match (== operator) - only accepts v0.2.0
require("==v0.2.0")

// Explicit patch version
require("v0.2.0")

// Major version requirement
require("v1")
```

**Version Comparison**:
- `require("v0.1")` - OK if running v0.1.0 or higher
- `require("v0.2")` - OK if running v0.2.0 or higher
- `require("v0.3")` - FAILS if running v0.2.0 (generates error)
- `require("v1.0")` - FAILS if running v0.2.0 (generates error)
- `require("==v0.2.0")` - OK only if running exactly v0.2.0 (fails on v0.2.1 or v0.3.0)

**Error Messages**:
If the version requirement is not met, script execution fails with:
```
ERROR: fynerisor version v0.3 or higher required, but running v0.2.0
ERROR: fynerisor version ==v0.2.0 required (exact match), but running v0.2.1
```

**Optional: Documentation Comments**:
For human-readable documentation, you can also add comments:
```js
// @fynerisor v0.2+
// @requires calendar time
require("v0.2")
```

**Checking Version in Go Code**:
```go
// In your Go code
import "github.com/uidbz/fynerisor/core"

info := core.GetVersion()
fmt.Printf("fynerisor %s (Risor %s, Fyne %s)\n", 
    info.Version, info.RisorCompat, info.FyneCompat)
```

## License

[Add license information]

## Contributing

[Add contribution guidelines]

## Support

For issues or questions about fynerisor, open an issue on the repository.
