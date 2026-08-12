# fynerisor Examples

This directory contains example applications demonstrating how to use the fynerisor package.

## Running the Examples

Each example is a standalone Go program. Navigate to the example directory and run:

```bash
cd 01-hello-world
go run main.go
```

## Prerequisites

Before running the examples, ensure you have:

1. **Go 1.21 or later** installed
2. **Fyne system dependencies** - see [Fyne Prerequisites](https://docs.fyne.io/started/)
   - Linux: `xorg-dev` or `wayland-dev` packages
   - Windows: C compiler (e.g., TDM-GCC or MSYS2)
   - macOS: Xcode command line tools

## Example List

Each example has its own README with detailed explanation. Browse the directories to explore:

- **01-hello-world** - Basic label widget
- **02-button-callback** - Interactive buttons with callbacks
- **03-form-input** - Form with validation
- **04-table-display** - Paginated table
- **05-progress-widgets** - Progress bars, sliders, activity indicators
- **06-icon-hyperlink-card** - Icon, hyperlink, and card widgets
- **07-radiogroup** - Radio button groups
- **08-accordion** - Expandable accordion widget
- **09-toolbar** - Toolbar with actions
- **10-calendar** - Calendar date picker
- **11-list** - Virtualized scrolling list
- **12-tree** - Hierarchical tree widget
- **13-http-fetch** - HTTP requests and JSON
- **14-module-imports** - Module-scoped imports with namespaces
- **15-sql-test** - Database connectivity
- **16-context-builder** - Headless context (core package)
- **17-gridwrap** - Grid layout
- **18-textgrid** - Code display with TextGrid
- **19-richtext** - Markdown formatting
- **20-button-importance** - Button styling with importance levels
- **21-form-validation** - Entry validation
- **22-popup** - Popup menus
- **23-menu** - Menu and menu items
- **24-constants** - Using Fyne constants
- **25-data-binding** - Reactive UI with data binding
- **26-data-binding-types** - All binding types (String, Bool, Int, Float)
- **27-app-versioning** - Application version checking
- **28-custom-struct** - Expose custom Go types
- **30-dialogs** - File, confirm, color picker dialogs
- **31-apptabs** - Tabbed interface
- **32-table-widgets** - Widget mode for table cells
- **33-image-gallery** - Images in table cells
- **34-keyboard-shortcuts** - Global shortcuts and menu integration
- **35-charts** - Various chart visualizations (bar, line, scatter, pie)
- **36-tie-headless** - Triple store client (headless, requires tie-daemon)
- **37-tie-gui** - Triple browser GUI with query and add (requires tie-daemon)

## Example Structure

Each example follows this structure:

```
example-name/
├── main.go          # Go program that creates the window
├── script.risor     # Risor script that builds the UI
└── README.md        # Detailed explanation and concepts
```

## Common Patterns

### Basic Window Setup

All examples follow this pattern:

```go
package main

import (
    "github.com/uidbz/fynerisor/gui"
    "os"
)

func main() {
    // 1. Create fynerisor app
    fw := gui.NewApp("Title")
    
    // 2. Load and execute script
    script, _ := os.ReadFile("script.risor")
    fw.LoadScript(string(script))
    fw.Execute()
    
    // 3. Show window
    fw.ShowAndRun()
}
```

### Risor Script Pattern

Scripts use global factory objects:

```js
// Create widgets
let label = widget.NewLabel("Text")
let button = widget.NewButton("Click", () => {
    // Callback code
})

// Create layout
let layout = container.NewVBox(label, button)

// Set as window content
window.SetContent(layout)
```

## Global Objects Available in Scripts

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
- **require**: Version and module validation
- **import**: Module imports with namespace isolation

## Tips

### Development Workflow

Since scripts are loaded from files, you can:
1. Run the example
2. Edit the `.risor` file
3. Close and re-run to see changes

No recompilation needed when modifying scripts!

### Error Handling

Check the console output for error messages. The status callback (when provided) will show script execution status:

```go
fw := gui.NewApp("My App",
    gui.WithStatusCallback(func(status string) {
        fmt.Printf("Status: %s\n", status)
    }),
)
```

### Custom Data

Pass custom Go data or functions to Risor scripts:

```go
import (
    "github.com/deepnoodle-ai/risor/v2"
    "github.com/uidbz/fynerisor/gui"
)

// Named global (accessible in scripts)
fw := gui.NewApp("My App",
    gui.WithGlobal("myapi", myAPIObject),
)

// Or use Risor options for advanced VM configuration
customOpts := risor.WithEnv(map[string]any{
    "myModule": myModuleObject,
})

fw := gui.NewApp("My App",
    gui.WithRisorOptions(customOpts),
)
```

Then in Risor: `myapi.someMethod()` or `myModule.doSomething("hello")`

## Next Steps

After reviewing these examples:

1. Read the [fynerisor README](../README.md) for complete API reference
2. Explore the [EXAMPLES.md](../docs/EXAMPLES.md) for more complex scenarios
3. Check the [Fyne documentation](https://docs.fyne.io/) for widget details
4. Review [Risor v2 syntax](https://risor.io/docs/) for language features

## Troubleshooting

### "cannot find package" errors

Ensure you're in the example directory when running `go run main.go`.

### Graphics/display errors

Install Fyne system dependencies for your platform. See the [Fyne Getting Started guide](https://docs.fyne.io/started/).

### Script errors

Check console output for error messages. Common issues:
- Syntax errors in `.risor` file (check parentheses, commas)
- Missing widget or container function names
- Incorrect callback signatures

## Contributing

Found a bug or have an example idea? Open an issue or submit a PR!
