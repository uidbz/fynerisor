# Example 02: Button with Callback

Demonstrates interactive widgets with callback functions that update the UI in response to user actions.

## What It Does

- Creates a label and two buttons
- Tracks button click count
- Updates the label text when buttons are clicked
- Shows how to use closures to maintain state

## Running

```bash
cd examples/02-button-callback
go run main.go
```

## Key Concepts

### Callbacks

Buttons accept callback functions that execute when clicked:

```js
let button = widget.NewButton("Click me!", () => {
    // This code runs when the button is clicked
    label.Text = "Updated text"
})
```

### State Management

Variables in the script maintain state across callback executions:

```js
let clickCount = 0

let button = widget.NewButton("Click me!", () => {
    clickCount = clickCount + 1  // Increments each time
})
```

### Widget Property Updates

Widget properties can be modified after creation:

```js
let label = widget.NewLabel("Initial text")
// Later in a callback:
label.Text = "New text"  // Updates the displayed text
```

### Container Layouts

The VBox container arranges widgets vertically:

```js
let layout = container.NewVBox(label, button1, button2)
```

### Status and Result Callbacks

This example demonstrates using callbacks in the Go code to monitor execution:

```go
statusCallback := func(status string) {
    fmt.Printf("Status: %s\n", status)
}

fyneWindow := fynerisor.NewWindow(w, nil, statusCallback, resultCallback)
```

## Files

- `main.go`: Go program with status monitoring
- `button.risor`: Risor script with interactive buttons
- `README.md`: This file
