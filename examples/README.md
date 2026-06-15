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

### 01-hello-world

**Difficulty:** Beginner  
**Concepts:** Basic setup, file-based scripts, label widget

The simplest possible fynerisor application. Loads a Risor script from a file and displays "Hello World" in a label.

**Learn:** How to create a window, load a script, and display a simple widget.

### 02-button-callback

**Difficulty:** Beginner  
**Concepts:** Interactive widgets, callbacks, state management, VBox layout

Interactive application with buttons that respond to clicks and update the UI.

**Learn:** How to handle button clicks, maintain state in callbacks, and update widget properties dynamically.

### 03-form-input

**Difficulty:** Intermediate  
**Concepts:** Entry widgets, forms, validation, HBox layout

Form application with text input fields, validation logic, and error handling.

**Learn:** How to read user input, validate data, display error messages, and work with multiple widgets in a form.

### 04-table-display

**Difficulty:** Advanced  
**Concepts:** Table widget, pagination, data callbacks, Go-to-Risor data passing

Table widget displaying paginated data with custom data source callbacks.

**Learn:** How to create tables, implement pagination, handle row clicks, and pass data functions from Go to Risor scripts.

### 05-progress-widgets

**Difficulty:** Beginner  
**Concepts:** Progress bars, sliders, activity indicators, separators, state management

Demonstrates progress indicators, sliders, and visual separators with interactive controls.

**Learn:** How to use ProgressBar, ProgressBarInfinite, Slider, Activity, and Separator widgets. Shows linking widgets and managing toggle state.

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
    "github.com/uidbz/fynerisor"
    "os"
)

func main() {
    // 1. Create fynerisor app
    fw := fynerisor.NewApp("Title")
    
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

- **window**: Window control (`SetContent`, `OnDropped`, `CaptureStdout`)
- **widget**: Widget factory (`NewButton`, `NewLabel`, `NewEntry`, `NewTable`, etc.)
- **container**: Layout factory (`NewVBox`, `NewHBox`, `NewBorder`, `NewSplit`, `NewScroll`)
- **canvas**: Canvas objects (`NewLine`, `NewImageFromURI`)
- **chart**: Chart factory (`NewBarChart`)

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
fw := fynerisor.NewApp("My App",
    fynerisor.WithStatusCallback(func(status string) {
        fmt.Printf("Status: %s\n", status)
    }),
)
```

### Custom Data

Pass custom Go functions to Risor scripts using `WithRisorOptions`:

```go
import "github.com/deepnoodle-ai/risor/v2"

customFuncs := risor.WithEnv(map[string]any{
    "myModule": map[string]any{
        "doSomething": func(arg string) string {
            return "Result: " + arg
        },
    },
})

fw := fynerisor.NewApp("My App",
    fynerisor.WithRisorOptions(customFuncs),
)
```

Then in Risor: `myModule.doSomething("hello")`

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
