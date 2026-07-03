# Browser Package

A generic, reusable browser UI for Fyne/fynerisor applications.

## Features

- **Navigation**: Address bar, back/forward buttons, refresh, home
- **History Management**: Browser-style navigation with back/forward stacks
- **URL Handling**: Automatic URL normalization and relative path resolution
- **Source View**: Toggle between normal view and source code view (split layout)
- **Plugin System**: Hybrid architecture with interfaces and callbacks
  - MenuProvider interface for custom menu items
  - AuthProvider interface for custom authentication UI
  - Callbacks for navigation lifecycle (OnNavigate, OnError, etc.)

## Usage

```go
import (
    "github.com/uidbz/fynerisor/browser"
    "github.com/uidbz/fynerisor/gui"
)

// Create fynerisor window
window := gui.NewWindow(fyneWindow, /* options */)

// Create browser with plugins
b := browser.New(window, browser.Config{
    MenuProvider: myMenuProvider,
    AuthProvider: myAuthProvider,
    LoadContent: func(url, username, password string) (string, error) {
        // Load and return content from URL
        return loadContent(url, username, password)
    },
    OnNavigate: func(url string) error {
        // Called before navigation
        return nil
    },
    OnNavigateError: func(url string, err error) {
        // Handle navigation errors
    },
})

// Set home page and navigate
b.SetHomeURL("https://example.com/app")

// Show window
b.ShowAndRun()
```

## Plugin Interfaces

### MenuProvider

Provides menu items for the side menu:

```go
type MenuProvider interface {
    GetMenuItems() []MenuItem
}

type MenuItem struct {
    Label  string
    Action func()
}
```

Example implementation:

```go
type MyMenuProvider struct {
    app *MyApp
}

func (p *MyMenuProvider) GetMenuItems() []browser.MenuItem {
    return []browser.MenuItem{
        {Label: "Settings", Action: p.app.showSettings},
        {Label: "About", Action: p.app.showAbout},
        {Label: "Help", Action: p.app.showHelp},
    }
}
```

### AuthProvider

Handles authentication when needed (HTTP 401):

```go
type AuthProvider interface {
    ShowAuthDialog(url string, contentContainer *fyne.Container) (username, password string, submitted bool)
}
```

Example implementation:

```go
type MyAuthProvider struct {
    app *MyApp
}

func (p *MyAuthProvider) ShowAuthDialog(url string, container *fyne.Container) (string, string, bool) {
    // Create login form UI
    userEntry := widget.NewEntry()
    passEntry := widget.NewEntry()
    passEntry.Password = true
    
    form := widget.NewForm(
        widget.NewFormItem("Username:", userEntry),
        widget.NewFormItem("Password:", passEntry),
    )
    
    submitted := false
    form.OnSubmit = func() {
        submitted = true
    }
    
    // Display in container
    container.Objects = []fyne.CanvasObject{form}
    container.Refresh()
    
    return userEntry.Text, passEntry.Text, submitted
}
```

### SourceViewProvider

Optional interface for enabling source view toggle:

```go
type SourceViewProvider interface {
    GetCurrentSource() string
}

// Enable source view
b.EnableSourceView(sourceProvider, func(url string) {
    // Optional callback for "Open in Browser" button
    openSystemBrowser(url)
})
```

Example implementation:

```go
type MyApp struct {
    currentSource string
}

func (a *MyApp) GetCurrentSource() string {
    return a.currentSource
}
```

## Architecture

The browser package follows a hybrid plugin architecture:

- **Interfaces** for complex, stateful subsystems (menus, auth, source view)
- **Callbacks** for simple hooks (navigation lifecycle, status updates)

This design allows applications to plug in custom behavior while keeping the browser generic and reusable.

## Programmatic Navigation from Scripts

Scripts can navigate programmatically using the `browser` global object:

```go
// Create browser
b := browser.New(window, config)

// Expose browser to scripts
browserWrapper := browser.NewRisorBrowser(b)
window := gui.NewWindow(fyneWindow,
    gui.WithGlobal("browser", browserWrapper),
    // ... other options
)
```

**Available methods in scripts:**

```js
// Navigate to a URL
browser.Open("https://example.com/page")
browser.Open("file:///path/to/script.risor")

// Get current URL
let currentURL = browser.GetURL()
print("Current URL:", currentURL)

// Set status bar text
browser.SetStatus("Loading...")
browser.SetStatus("Ready")

// Read URL query parameters
// For https://example.com/page?myarg=value&other=value2
let myarg = browser.GetParam("myarg")   // "value"
let other = browser.params["other"]      // "value2"
let missing = browser.GetParam("nope")   // "" (empty string when absent)
```

**Example script:**

```js
// Create navigation buttons
let homeBtn = widget.NewButton("Home", () => {
    browser.SetStatus("Going home...")
    browser.Open("https://example.com/home")
})

let aboutBtn = widget.NewButton("About", () => {
    browser.SetStatus("Loading about page...")
    browser.Open("https://example.com/about")
})

let urlBtn = widget.NewButton("Show URL", () => {
    let url = browser.GetURL()
    browser.SetStatus("Current: " + url)
})

window.SetContent(container.NewVBox([homeBtn, aboutBtn, urlBtn]))
```

## Configuration Options

```go
type Config struct {
    // Required interfaces
    MenuProvider MenuProvider
    AuthProvider AuthProvider
    
    // Required callback
    LoadContent func(url, username, password string) (string, error)
    
    // Optional callbacks
    OnNavigate         func(url string) error
    OnNavigateComplete func(url string)
    OnNavigateError    func(url string, err error)
    OnStatusChange     func(status string)
    
    // Optional settings
    HomeURL string
}
```

## URL Handling

The browser automatically handles:
- Adding `https://` prefix to bare domains
- Appending `/index.risor` to directory URLs
- Resolving relative URLs (e.g., `../other.risor`)
- Both HTTP(S) and `file://` protocols

### Query Parameters

Query parameters on the current URL (e.g. `https://server/page?myarg=value&other=value2`)
are parsed on every navigation and exposed to both scripts and host Go code:

**From Risor scripts** (via the `browser` global):
- `browser.params` — a map of parameter name to value
- `browser.GetParam(name)` — the value for a single parameter (empty string if absent)

**From external Go apps** (via the `*browser.Browser` instance):

```go
params := b.GetParams()        // url.Values (a copy; safe to mutate)
value := b.GetParam("myarg")   // first value, or "" if absent
```

## History Navigation

- **Back()** - Navigate to previous page
- **Forward()** - Navigate to next page
- **Visit(url)** - Add page to history
- **Current()** - Get current URL

History automatically tracks visited pages and supports back/forward navigation with proper stack management.
