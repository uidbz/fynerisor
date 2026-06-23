# fynerisor-browser

Reference implementation of a browser using the fynerisor browser package.

## Overview

A minimal browser demonstrating how to use `github.com/uidbz/fynerisor/browser` to build browser-style applications. This serves as:

- **Reference implementation** - Shows best practices for using the browser package
- **Starting point** - Copy and modify for your own projects
- **Example** - Demonstrates plugin architecture and integration

## Features

- Load Risor scripts from HTTP(S) or `file://` URLs
- Browser navigation (address bar, back/forward, refresh, home)
- Source view toggle (split view showing script source)
- Basic authentication (username/password)
- Simple menu (View Source, About, Quit)

## Usage

### Install

```bash
go install github.com/uidbz/fynerisor/cmd/fynerisor-browser@latest
```

### Run

```bash
# Load from URL
fynerisor-browser https://example.com/app

# Load from local file (absolute path)
fynerisor-browser /path/to/script.risor
fynerisor-browser C:\path\to\script.risor  # Windows

# Load from local file (relative path)
fynerisor-browser ./script.risor
fynerisor-browser ../other/script.risor

# Or use file:// URL explicitly
fynerisor-browser file:///path/to/script.risor

# Use default home URL
fynerisor-browser
```

File paths are automatically detected and converted to `file://` URLs. Both Unix-style (`/home/user/script.risor`) and Windows-style (`C:\Users\user\script.risor`) paths are supported.

### Command-line Flags

- `-home <url>` - Set home URL (default: https://example.com)

## Architecture

The reference browser demonstrates the three plugin interfaces:

### MenuProvider

Provides menu items for the side menu:

```go
type BrowserMenuProvider struct {
    app *BrowserApp
}

func (p *BrowserMenuProvider) GetMenuItems() []browser.MenuItem {
    return []browser.MenuItem{
        {Label: "View Source", Action: p.app.browser.ToggleSourceView},
        {Label: "About", Action: p.app.showAbout},
        {Label: "Quit", Action: p.app.app.Quit},
    }
}
```

### AuthProvider

Handles HTTP Basic Authentication:

```go
type BrowserAuthProvider struct {
    app *BrowserApp
}

func (p *BrowserAuthProvider) ShowAuthDialog(url string, container *fyne.Container) (string, string, bool) {
    // Create login form
    // Display in container
    // Return credentials and submitted status
}
```

### SourceViewProvider

Supplies source code for the source view feature:

```go
func (b *BrowserApp) GetCurrentSource() string {
    return b.source
}
```

## Customization

To build your own browser:

1. **Copy this implementation** as a starting point
2. **Customize plugins**:
   - Add your own menu items
   - Implement custom authentication UI
   - Add logo/branding to auth form
3. **Add features**:
   - Script validation
   - Version checking
   - Import handling
   - Custom error pages
4. **Extend functionality**:
   - Add more modules via `gui.With*()` options
   - Implement custom protocol handlers
   - Add bookmarks/favorites
   - Theme switching

## Example Script

Create a simple Risor script to test:

```risor
// hello.risor
let title = widget.NewLabel("Hello, fynerisor-browser!")
title.TextStyle.Bold = true
title.Alignment = 1

let description = widget.NewLabel("This is a Risor script loaded in the browser")
description.Alignment = 1

let button = widget.NewButton("Click Me", func() {
    print("Button clicked!")
})

let content = container.NewVBox(
    title,
    description,
    widget.NewSeparator(),
    button
)

window.SetContent(content)
```

Then load it:

```bash
fynerisor-browser file:///path/to/hello.risor
```

## See Also

- [fynerisor](https://github.com/uidbz/fynerisor) - Risor v2 bindings for Fyne
- [browser package](../../browser/) - Generic browser UI package
- [Risor](https://github.com/risor-io/risor) - Fast and flexible scripting language
- [Fyne](https://fyne.io) - Cross-platform GUI toolkit
