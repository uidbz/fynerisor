# Example 05: Progress & Slider Widgets

Demonstrates progress indicators, activity indicators, sliders, and separators.

## What It Does

- Shows a **ProgressBar** with adjustable value (0-100%)
- Uses a **Slider** to control the progress bar value
- Demonstrates **ProgressBarInfinite** with Start/Stop control
- Shows an **Activity** indicator with Start/Stop control
- Uses **Separator** widgets for visual organization
- Shows interactive slider with real-time value display

## Running

```bash
cd examples/05-progress-widgets
go run main.go
```

## Widgets Demonstrated

### ProgressBar

A progress bar that shows completion percentage:

```js
let progressBar = widget.NewProgressBar()
progressBar.Min = 0
progressBar.Max = 100
progressBar.Value = 50  // 50% complete
```

**Properties:**
- `Value` - Current value (read/write)
- `Min` - Minimum value (read/write)
- `Max` - Maximum value (read/write)

The progress bar automatically displays as a percentage between Min and Max.

### Slider

Interactive slider for numeric value selection:

```js
let slider = widget.NewSlider(0, 100)  // min, max
slider.Value = 50
slider.Step = 1  // Increment by 1

slider.OnChanged((value) => {
    print(`Slider value: ${value}`)
})
```

**Constructor:**
- `NewSlider(min, max)` - Creates slider with min/max bounds

**Properties:**
- `Value` - Current value (read/write)
- `Min` - Minimum value (read/write)
- `Max` - Maximum value (read/write)
- `Step` - Increment step size (read/write)

**Methods:**
- `OnChanged(callback)` - Called when value changes

### ProgressBarInfinite

Indeterminate progress indicator (animated):

```js
let infinite = widget.NewProgressBarInfinite()

infinite.Start()  // Begin animation
infinite.Stop()   // Stop animation

let running = infinite.Running  // Check if running
```

**Methods:**
- `Start()` - Start animation
- `Stop()` - Stop animation

**Properties:**
- `Running` - Boolean indicating if animating (read-only)

### Activity

Simple spinning activity indicator:

```js
let activity = widget.NewActivity()

activity.Start()  // Show spinning indicator
activity.Stop()   // Hide indicator
```

**Methods:**
- `Start()` - Show and animate
- `Stop()` - Hide

### Separator

Visual separator line:

```js
let separator = widget.NewSeparator()
```

Creates a horizontal line for visual separation between sections.

## Key Concepts

### Linking Widgets

The example shows how to link widgets together:

```js
slider.OnChanged((value) => {
    progressBar.Value = value  // Update progress bar
    label.Text = `Progress: ${value}%`  // Update label
})
```

### Toggle State Management

Track widget state for toggle buttons:

```js
let isRunning = false

button.OnTapped(() => {
    if (isRunning) {
        activity.Stop()
        button.Text = "Start"
        isRunning = false
    } else {
        activity.Start()
        button.Text = "Stop"
        isRunning = true
    }
})
```

### Property Updates

Widget properties can be updated after creation:

```js
let bar = widget.NewProgressBar()
bar.Min = 0       // Set minimum
bar.Max = 200     // Set maximum
bar.Value = 100   // Set current value (50%)
```

### Visual Organization

Use separators to organize UI sections:

```js
let layout = container.NewVBox(
    section1,
    widget.NewSeparator(),  // Visual divider
    section2,
    widget.NewSeparator(),
    section3
)
```

## Number Formatting

Risor provides number methods for display:

```js
let value = 42.7891
value.toFixed(1)   // "42.8"
value.toFixed(2)   // "42.79"
math.round(value)  // 43
```

## Files

- `main.go`: Go program that creates the window
- `progress.risor`: Risor script demonstrating widgets
- `README.md`: This file
