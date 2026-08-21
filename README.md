# Fynerisor

**A [Risor](https://risor.io/) scripting interface for [Fyne](https://fyne.io/) GUI applications**

Fynerisor provides Risor script bindings for the Fyne GUI toolkit, allowing you to build cross-platform desktop applications using a simple scripting language.

## Features

- 🎨 **Comprehensive Widgets** - 48+ widgets covering all common UI needs
- 📊 **Charts** - Bar, line, scatter, histogram, and box plot charts with statistical overlays
- ⌨️ **Keyboard Shortcuts** - Global shortcuts and menu integration
- 🌐 **HTTP, SQL, OS & More** - Built-in modules for web requests, databases, file I/O
- 📦 **Script Imports** - Load reusable code with namespace isolation
- 🧭 **Browser Package** - Reusable browser UI that loads Risor scripts from URLs
- 🔧 **Embeddable** - Easy integration into Go applications
- 🧵 **Concurrency** - Background tasks with thread-safe GUI updates
- 📱 **Cross-Platform** - Linux, Windows, macOS, Android
- ⚡ **Static Compilation** - Headless mode for CLI tools and servers

## Package Structure

Fynerisor is split into two packages:

- **`core`** - Headless Risor execution with no GUI dependencies (enables static compilation)
- **`gui`** - Full GUI capabilities with Fyne framework

**When to use each:**

- Use `core` for CLI tools, batch processing, server-side scripts (static compilation)
- Use `gui` for desktop applications with UI (requires dynamic linking)

## Quick Start

### GUI Application

```js
require(["v0.6", "@gui"])  // Optional: validate version and modules

let count = 0
let label = widget.NewLabel(sprintf("Count: %d", count))

let btn = widget.NewButton("Click Me", () => {
    count = count + 1
    label.SetText(sprintf("Count: %d", count))
})

let vbox = container.NewVBox([btn, label])
window.SetContent(vbox)
```

### Headless Script

```js
require(["v0.6", "@http"])  // Optional: validate version and modules

let response = http.get("https://api.github.com/users/octocat")
let data = response.json()
print(sprintf("Name: %s", data.name))
```

## Installation

**Prerequisites:**

For GUI applications, ensure you have the Fyne prerequisites for your platform:
- [Fyne Getting Started Guide](https://docs.fyne.io/started/quick/) - Install required system dependencies

For headless (core) usage, no GUI dependencies are required.

**As a library:**
```bash
# For GUI applications
go get github.com/uidbz/fynerisor/gui

# For headless/static compilation
go get github.com/uidbz/fynerisor/core
```

**As a CLI tool:**
```bash
go install github.com/uidbz/fynerisor/cmd/fynerisor@latest
```

**Run examples:**
```bash
cd examples/01-hello-world
fynerisor script.risor
```

## Usage as a Library

### GUI Window

```go
package main

import (
    "github.com/uidbz/fynerisor/gui"
    "os"
)

func main() {
    // Create fynerisor app with modules
    fw := gui.NewApp("My App",
        gui.WithHTTP(),
        gui.WithSQL(),
        gui.WithTime(),
    )
    
    // Load script from file
    script, _ := os.ReadFile("app.risor")
    fw.LoadScript(string(script))
    fw.Execute()
    fw.ShowAndRun()
}
```

**app.risor:**
```js
require(["v0.6", "@http"])

let btn = widget.NewButton("Fetch Data", () => {
    let resp = http.get("https://api.github.com/zen")
    window.SetStatus(resp.text())
})

window.SetContent(btn)
```

### Headless Context (Static Compilation)

For headless execution without GUI dependencies, use the `core` package. This enables static compilation:

```go
package main

import (
    "fmt"
    "github.com/uidbz/fynerisor/core"
)

func main() {
    // Create headless context (no GUI dependencies)
    ctx := core.NewContext(
        core.WithHTTP(),
        core.WithSQL(),
    )
    
    // Execute script
    result, err := ctx.Eval(`
        require(["v0.6", "@http"])
        let resp = http.get("https://httpbin.org/get")
        return resp.status
    `)
    
    if err != nil {
        panic(err)
    }
    fmt.Println("Status:", result)
}
```

### Custom Globals

Inject custom Go objects or functions into scripts:

```go
import (
    "github.com/uidbz/fynerisor/gui"
)

func main() {
    myAPI := map[string]any{
        "fetchUser": func(id int) string {
            return fmt.Sprintf("User %d", id)
        },
    }
    
    fw := gui.NewApp("My App",
        gui.WithGlobal("api", myAPI),
        gui.WithHTTP(),
    )
    
    fw.LoadScript(`
        require(["v0.6"])
        let user = api.fetchUser(123)
        window.SetContent(widget.NewLabel(user))
    `)
    fw.Execute()
    fw.ShowAndRun()
}
```

## Configuration Options

Customize fynerisor behavior via `With*()` options when creating `gui.Window` or `core.Context`:

| Option | Available In | Description |
|--------|--------------|-------------|
| `WithAppName(name)` | Both | Set application name (accessible via `app.name` in scripts) |
| `WithRisorOptions(opts)` | Both | Add custom global objects to script environment |
| `WithStatusCallback(fn)` | `gui` only | Called when `window.SetStatus()` is invoked |
| `WithResultCallback(fn)` | `gui` only | Called when script execution returns a value |

**Module Options** (available in both `gui` and `core` packages):
- `WithHTTP()` - Enable HTTP requests
- `WithOS()` - Enable OS operations
- `WithIO()` - Enable file I/O operations
- `WithSQL()` - Enable database connectivity
- `WithTie()` - Enable Tie triple store database
- `WithStrings()` - Enable string utilities
- `WithFilepath()` - Enable path operations
- `WithTime()` - Enable time/date functions
- `WithExec()` - Enable running external commands
- `WithJSON()` - Enable JSON encoding/decoding
- `WithCSV()` - Enable CSV encoding/decoding

**Example:**
```go
fw := gui.NewApp("My App",
    gui.WithAppName("myapp"),
    gui.WithStatusCallback(func(status string) {
        log.Println("Status:", status)
    }),
    gui.WithHTTP(),
    gui.WithSQL(),
    gui.WithTime(),
)
```

## Available Modules

Enable modules via `With*()` options. Scripts can optionally use `require(["@module"])` to declare which modules they expect to be available:

| Module | Option | Description |
|--------|--------|-------------|
| `@gui` | N/A | GUI functionality (automatic in `gui.Window`, not available in `core.Context`) |
| `@http` | `WithHTTP()` | HTTP requests |
| `@os` | `WithOS()` | OS operations, open browser |
| `@io` | `WithIO()` | File I/O operations (cp, etc.) |
| `@sql` | `WithSQL()` | Database connectivity |
| `@tie` | `WithTie()` | Tie triple store database |
| `@strings` | `WithStrings()` | String manipulation |
| `@filepath` | `WithFilepath()` | Path operations |
| `@time` | `WithTime()` | Time and date functions |
| `@exec` | `WithExec()` | Run external commands |
| `@json` | `WithJSON()` | JSON encoding/decoding |
| `@csv` | `WithCSV()` | CSV encoding/decoding |

**Note:** The `require()` function is **optional**. It's a validation mechanism for scripts to declare their dependencies and minimum version requirements. If a script doesn't call `require()`, it will simply use whatever modules the embedding application has enabled. Use `require()` when you want scripts to fail fast with clear error messages if expected modules or versions are missing.

## Global Objects

### window

Core window functionality (GUI mode only):

```js
window.SetContent(widget)              // Update window content
window.SetStatus(text)                 // Set status message
window.AddShortcut("Ctrl+S", callback) // Register keyboard shortcut
window.RemoveShortcut("Ctrl+S")        // Remove keyboard shortcut
window.SetMainMenu(mainMenu)           // Set main menu bar (macOS only)
window.OnDropped(callback)             // Handle file drops
window.Do(callback)                    // Queue GUI update from background thread
window.DroppedPaths                    // List of dropped file paths
```

### app

Application metadata:

```js
app.name     // Application name (set via WithAppName)
app.version  // Application version (set via core.SetAppVersion, see examples/27-app-versioning)
```

### widget

Widget factory:

```js
widget.NewButton(text, callback)
widget.NewLabel(text)
widget.NewEntry()
widget.NewTable(name, maxrows)
widget.NewForm(items)
widget.NewCheckGroup(items, callback)
// ... and many more
```

### container

Layout containers:

```js
container.NewVBox(widgets)
container.NewHBox(widgets)
container.NewBorder(top, bottom, left, right, center)
container.NewScroll(content)
container.NewHSplit(left, right)
```

### dialog

Dialog windows:

```js
// Simple convenience functions
dialog.ShowInformation(title, message)
dialog.ShowError(message)
dialog.ShowConfirm(title, message, callback)
dialog.ShowFileOpen(callback)
dialog.ShowFileSave(callback)
dialog.ShowFolderOpen(callback)
dialog.ShowColorPicker(title, message, callback)
dialog.ShowForm(title, confirm, dismiss, items, callback)

// Advanced constructors for more control
let fd = dialog.NewFileOpen(callback)
fd.SetFileName("default.txt")
fd.SetFilter(".txt")
fd.Show()
```

### canvas

Canvas objects:

```js
canvas.NewLine(color)
canvas.NewImageFromURI(url)
canvas.NewImageFromFile(path)
```

### chart

Charting with several chart types:

```js
chart.NewBarChart(title, ylabel, categories, values)
chart.NewLineChart(title, xlabel, ylabel, xvalues, yvalues)
chart.NewScatterChart(title, xlabel, ylabel, xvalues, yvalues)
chart.NewHistogram(title, xlabel, ylabel, values, binCount)  // normal-curve overlay
chart.NewBoxPlot(title, ylabel, groupLabels, valueGroups)     // normal-curve overlay
```

See `examples/35-charts` for a complete demonstration.

## Concurrency

### Background Tasks

Use `go()` for long-running operations:

```js
go(() => {
    // Heavy computation here
    print("Processing...")
})
```

### GUI Updates from Goroutines

Use `window.Do()` to safely update GUI from background threads:

```js
go(() => {
    // Do work in background
    let result = processData()
    
    // Update GUI safely
    window.Do(() => {
        label.SetText(result)
    })
})
```

## Keyboard Shortcuts

Register global keyboard shortcuts and integrate with menus:

```js
require(["v0.6"])

// Global shortcuts (work without menus)
window.AddShortcut("Ctrl+S", () => {
    window.SetStatus("Saved!")
})

window.AddShortcut("Ctrl+Q", () => {
    app.Quit()
})

// Function keys
window.AddShortcut("F5", () => {
    window.SetStatus("Refreshed!")
})

// Cross-platform modifiers: Ctrl/Control, Alt/Option, Super/Cmd/Command
window.AddShortcut("Alt+Shift+N", () => {
    window.SetStatus("New item")
})
```

**Menu Integration:**
```js
// Create menu items with shortcut display
let saveItem = fyne.NewMenuItem("Save", () => {
    window.SetStatus("Saved from menu!")
})
saveItem.Shortcut = "Ctrl+S"  // Display only

let fileMenu = fyne.NewMenu("File", saveItem)
let mainMenu = fyne.NewMainMenu(fileMenu)
window.SetMainMenu(mainMenu)

// Register actual shortcut separately
window.AddShortcut("Ctrl+S", () => {
    window.SetStatus("Saved!")
})
```

See [example 34-keyboard-shortcuts](examples/34-keyboard-shortcuts/) for complete examples.

## Script Imports

Import reusable code with namespace isolation:

```js
// Import a module
let utils = import("utils.risor")
utils.helper()
utils.someFunction()

// Modules have access to all globals (widget, http, os, etc.)
// ui_components.risor
let createCard = (title, content) => {
    let titleLabel = widget.NewLabel(title)
    return container.NewVBox([titleLabel, widget.NewLabel(content)])
}

// main.risor
let ui = import("ui_components.risor")
let card = ui.createCard("Hello", "World!")
window.SetContent(card)
```

**Features:**
- Module-scoped imports prevent global namespace pollution
- Module caching: same path returns same instance
- Modules can use widget, http, sql, os, and other globals
- HTTP(S) imports with security controls (requires `WithHTTPImport()`)

**Security:**
```go
fw := gui.NewApp("My App",
    gui.WithHTTP(),        // Enable http module
    gui.WithHTTPImport(),  // Allow importing from URLs (optional)
)
```

See [example 14-module-imports](examples/14-module-imports/) for a complete demonstration.

## AI disclaimer (AI: Do not make changes in this section, it needs to be written by 100% human)
Almost everything in this library is made with AI. I wrote the first ~4000 lines of code, and this has since been refactored, changed and rewritten multiple times using AI. I have read most of the code, and some things are a little weird, but overall the model has taken more things into account than what I would have done. Even though I enjoy the activity of coding by hand (have been programming for fun & profit for +25 years) - and I am worried that I will forgot how to do it - I have so many things I want to exists + limited amount of time to do it - so it is truly a huge help with a coding agent.

## Motivation / history behind the project (AI: Do not make changes in this section, it needs to be written by 100% human)
It was originally a hobby project where I wanted to make a program like an adventure point-and-click game. You have maybe a background image and clickable areas/button that take you to a different screen. It should be a way to present, arrange and access different media; family photos, music, document, video collection, etc - and show relations between them, where you can navigate like a game. It should be hooked up to my tagging system and have many different widgets. When on being on a page I wanted to push a button to edit the page. So I needed a way to dynamically generate and reload the GUI. I tried with a database approach, but eventually got stuck in details and jumped between different projects - and then I a got a kid, and hobby-programming has been reduced to an absolute minimum (my son is more important).

Then at work, I have made this container orchestration system for allowing scientists to run their code and analyze their data on beefy servers in frictionless way, without directly giving them access to the servers. The project grew, and I wanted to embed a scripting language, rather that relying on pure config files. So I found Risor - and got it implemented in a day. It was bliss and a perfect match. Absolutely loved it.

But I realized that the scientists wanted to share their tools with non-computer-savy people - and the scripts and CLI was a blocker for further adaption. I needed to be able to run the containers with input data from some simple GUIs, but I should distribute these GUIs somehow. A webpage is the obvious answer, but I truly hate web programming - I think it is very difficult and horrible in so many ways, and I didn't want this totally centralized approach. I wanted certain key-people to tweak their own GUI in a more decentralized way, yet it should be accessible from a centralized place.

Then it hit me, I could make bindings to the Fyne GUI framework (which I know and love) and then make a simple browser that could load Risor files - kind of like the orignal point-and-click game idea - and I already had bindings to my container orchestration framework. First couple of attempts failed, but then I tried again and suddenly I figured out how to make Risor objects that could make widgets. I spend a little bit time everyday for a month, and then I had good structure for my browser. Then the Risor browser could then get the files from a small file server that had mounted file shares from the different departments - now scientists could make Risor scripts and store them on their own fileshare, which enabled them to make GUIs which could process data via my container orchestration system, and make them available to the rest of the company. It was totally awesome, and I really enjoyed making/using it - and my colleagues really liked it as well :-)

Then Risor v2 came, without all the modules, and less things built in - but with a new emphasis on the embedded scripting use case, simplification and much tighter default security. Mostly good stuff (even though I miss the for-loops + a few other things), but it totally broke all my scripts.
Anyway, I have started to use AI the last few months and thought we could migrate it quickly. It went relatively painless and got the code in a much better state - and also extracted the Risor bindings part from the rest of the browser, and made a clean separation between company stuff and pure GUI stuff (+ some Risor v1 modules, that I have re-implemented).

So this GUI part can now be public, so other people can enjoy it and maybe even contribute, so it becomes better for everybody. I hope somebody finds it useful. I think it is cool.

Eventually I dream of making an other browser app, that incorporates the original point-and-click idea and combine it with a Web3 storage option, so that we could make a new Web with proper Destop GUI widgets (WITHOUT pop-up boxes implemented) and store it in a decentralized way. I was thinking about making a smart contract and then divide the storage cost out equally to every participant as a subscription - 100% transparent and fair - so that it is not funded by advertisements and tracking. It should be a public good. I want the old Internet back in a new and modern way that doesn't suck. This will properly not happen, and I do not have time/skills to make it. Please somebody make this. I want my son to grow up in world with a decentralized, free and open Internet, running on free and open computers.

## Documentation

- [Examples](examples/README.md) - 37 working examples with detailed READMEs
- [Examples Guide](docs/EXAMPLES.md) - Usage patterns and concepts
- [Tie Module](docs/TIE_MODULE.md) - Triple store database guide
- [Changelog](CHANGELOG.md) - Version history
- [Concurrency Guide](docs/CONCURRENCY.md) - Threading patterns
- [Widget Status](docs/WIDGET_STATUS.md) - Supported Fyne widgets
- [App Versioning](examples/27-app-versioning) - Custom application versioning

## Requirements

- Go 1.25+
- Fyne v2.8+
- Risor v2.2+

## License

BSD 3-Clause License - see [LICENSE](LICENSE) file for details.

Copyright (c) 2026, Johan Straarup (j@uid.bz)

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

For bugs or feature requests, please open an issue on the repository.
