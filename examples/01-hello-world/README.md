# Example 01: Hello World

The simplest possible fynerisor application. Demonstrates loading a Risor script from a file and creating a basic label widget.

## What It Does

- Loads `hello.risor` from disk
- Creates a label widget with "Hello World" text
- Displays the label in the window

## Running

```bash
cd examples/01-hello-world
go run main.go
```

## Key Concepts

### File-based Scripts

This example loads the Risor script from an external file:

```go
script, _ := os.ReadFile("hello.risor")
fyneWindow.LoadScript(string(script))
fyneWindow.Execute()
```

This pattern allows you to modify the script without recompiling.

### Minimal Window Setup

The window is created with nil callbacks for simplicity:

```go
fyneWindow := fynerisor.NewWindow(w, nil, nil, nil)
```

### Widget Creation

The Risor script uses the global `widget` factory object:

```js
let label = widget.NewLabel("Hello World")
```

### Setting Content

The `window` object (provided by fynerisor) is used to set the window content:

```js
window.SetContent(label)
```

## Files

- `main.go`: Go program that creates the window and loads the script
- `hello.risor`: Risor script that creates the UI
- `README.md`: This file
