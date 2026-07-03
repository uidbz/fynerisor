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

- `-home <url>` - Set home URL (default: https://fynerisor.com/app)

## Android

The browser can be packaged as an Android app. Because mobile apps receive no
command-line arguments, the startup (home) URL is baked in at build time as
custom app metadata.

### Requirements

- The `fyne` command: `go install fyne.io/fyne/v2/cmd/fyne@latest`
- Android SDK and NDK installed, with `adb` and the NDK on your `PATH`
  (`ANDROID_HOME` / `ANDROID_NDK_HOME` set). See the
  [Fyne mobile docs](https://docs.fyne.io/started/mobile).

### Build

Run from this directory (where `FyneApp.toml` lives):

```bash
# Uses the HomeURL from FyneApp.toml
fyne package -os android -app-id com.fynerisor.browser -icon Icon.png

# Bake in a different startup URL without editing FyneApp.toml
fyne package -os android -app-id com.fynerisor.browser -icon Icon.png \
    --metadata HomeURL=https://your.server/app

# Release build (for distribution)
fyne package -os android -app-id com.fynerisor.browser -icon Icon.png --release
```

`fyne package` auto-generates the `AndroidManifest.xml` (including the
`INTERNET` permission required to fetch scripts over HTTP/HTTPS), so no manifest
file is needed in the repo. The result is a `.apk` in this directory. Install it
with:

```bash
adb install -r "Fynerisor Browser.apk"
```

### How the home URL is resolved

`main.go` picks the startup URL in this order:

1. `-home` flag or positional argument (desktop only)
2. `HomeURL` custom metadata (from `FyneApp.toml` or `--metadata HomeURL=...`)
3. Built-in `fallbackHomeURL`

Custom metadata is read at runtime via `fyne.CurrentApp().Metadata().Custom`.
During development (`go run` / `go build`) it comes from the `[Development]`
table in `FyneApp.toml`; a `--release` build uses the `[Release]` table.

### Packaging files

- `FyneApp.toml` - App metadata (name, ID, icon, version, and the `HomeURL`
  custom metadata under `[Development]` / `[Release]`)
- `Icon.png` - Launcher icon

> **Note:** The table widget's right-click "copy to clipboard" uses a native
> clipboard library that requires initialization on mobile. Clipboard copy from
> tables is a no-op on Android; all other functionality works.

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
