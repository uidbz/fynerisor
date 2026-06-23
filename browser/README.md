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

To allow Risor scripts to navigate programmatically, you have two options:

### Option 1: Use the provided RisorBrowser wrapper

```go
// During window creation
browserWrapper := browser.NewRisorBrowser(nil) // nil for now
window := gui.NewWindow(fyneWindow,
    gui.WithGlobal("browser", browserWrapper),
    // ... other options
)

// After creating the actual browser
actualBrowser := browser.New(window, config)
// Update the wrapper (requires modifying RisorBrowser or using a different approach)
```

### Option 2: Create your own wrapper (recommended)

```go
// Create a custom object that wraps the browser
type MyAppGlobal struct {
    browser *browser.Browser
}

func (m *MyAppGlobal) GetAttr(name string) (object.Object, bool) {
    if name == "Open" {
        return object.NewBuiltin("myapp.Open", func(ctx context.Context, args ...object.Object) (object.Object, error) {
            url, _ := object.AsString(args[0])
            m.browser.Navigate(url)
            return object.Nil, nil
        }), true
    }
    return nil, false
}
// ... implement other object.Object methods

// Register during window creation
gui.WithGlobal("myapp", myAppGlobalInstance)
```

Scripts can then navigate:
```risor
myapp.Open("https://example.com/page")
// or with the browser wrapper:
browser.Open("https://example.com/page")
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

## History Navigation

- **Back()** - Navigate to previous page
- **Forward()** - Navigate to next page
- **Visit(url)** - Add page to history
- **Current()** - Get current URL

History automatically tracks visited pages and supports back/forward navigation with proper stack management.
