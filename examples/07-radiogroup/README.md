# Example 07: RadioGroup Widget

Demonstrates single-selection radio button groups with various configurations and behaviors.

## What It Does

- Shows **RadioGroup** widget for single-selection from options
- Demonstrates vertical and horizontal layouts
- Shows required validation with `Required` property
- Demonstrates dynamic option management (Append, changing Options)
- Shows programmatic selection with SetSelected()
- Demonstrates Enable/Disable functionality
- Shows RadioGroup integration in forms

## Running

```bash
cd examples/07-radiogroup
go run main.go
```

## Widget: RadioGroup

Single-selection widget that displays a group of radio buttons. Only one option can be selected at a time.

### Constructor

```js
let radio = widget.NewRadioGroup(options, onChanged)
```

**Parameters:**
- `options` - Array of strings for the radio options
- `onChanged` - Callback function `(selected) => {}` called when selection changes

### Properties

**Read/Write:**
- `Selected` - Currently selected option (string, empty if none)
- `Options` - Array of available options (can be reassigned)
- `Horizontal` - Boolean, true for horizontal layout (default: false/vertical)
- `Required` - Boolean, true if selection is required (shows validation)

**Example:**
```js
radio.Selected = "Option2"
radio.Options = ["New1", "New2", "New3"]
radio.Horizontal = true
radio.Required = true
```

### Methods

**SetSelected(option)**
```js
radio.SetSelected("Option2")  // Programmatically select an option
radio.SetSelected("")         // Clear selection
```

**Append(option)**
```js
radio.Append("New Option")    // Add new option to end
```

**Disable() / Enable()**
```js
radio.Disable()               // Disable all radio buttons
radio.Enable()                // Re-enable radio buttons
```

**Refresh()**
```js
radio.Refresh()               // Force visual refresh
```

## Key Concepts

### Single Selection

RadioGroup enforces single selection - selecting a new option automatically deselects the previous one:

```js
let radio = widget.NewRadioGroup(
    ["Option 1", "Option 2", "Option 3"],
    (selected) => {
        print(`User selected: ${selected}`)
    }
)
```

### Vertical vs Horizontal Layout

Default is vertical (stacked):
```js
let radio = widget.NewRadioGroup(["A", "B", "C"], (s) => {})
// Displays vertically
```

Horizontal layout:
```js
let radio = widget.NewRadioGroup(["Small", "Medium", "Large"], (s) => {})
radio.Horizontal = true
// Displays horizontally in a row
```

### Required Validation

Mark a RadioGroup as required for form validation:

```js
let radio = widget.NewRadioGroup(["Yes", "No"], (s) => {})
radio.Required = true
// Shows validation state if empty
```

Check in form submission:
```js
if radio.Selected == "" {
    print("Error: Selection required")
}
```

### Dynamic Options

Add options dynamically:
```js
radio.Append("New Option")
```

Replace all options:
```js
radio.Options = ["Completely", "New", "List"]
```

**Note:** When changing options, clear selection if needed:
```js
radio.Options = ["New", "Options"]
radio.Selected = ""  // Clear old selection
```

### Programmatic Selection

Set selection from code:
```js
radio.SetSelected("Option 2")  // Select specific option
```

Clear selection:
```js
radio.SetSelected("")          // No selection
```

Read current selection:
```js
let current = radio.Selected
print(`Currently selected: ${current}`)
```

### Callback Behavior

Callback is triggered:
- When user clicks a radio button
- When SetSelected() is called with a different value
- NOT when setting Selected property directly (must call Refresh)

```js
let radio = widget.NewRadioGroup(["A", "B"], (selected) => {
    print(`Changed to: ${selected}`)
})

radio.SetSelected("B")     // Triggers callback
radio.Selected = "A"       // Does NOT trigger callback
radio.Refresh()            // Updates UI
```

### Form Integration

RadioGroups work well in forms:

```js
let paymentRadio = widget.NewRadioGroup(
    ["Credit Card", "PayPal", "Bank Transfer"],
    (selected) => {
        // Update payment method
    }
)
paymentRadio.Required = true

let form = container.NewVBox(
    widget.NewLabel("Payment Method:"),
    paymentRadio,
    widget.NewButton("Submit", () => {
        if paymentRadio.Selected == "" {
            print("Error: Please select payment method")
        } else {
            print(`Processing with: ${paymentRadio.Selected}`)
        }
    })
)
```

### Enable/Disable

Disable to prevent user interaction:
```js
radio.Disable()  // Gray out, no interaction
radio.Enable()   // Re-enable
```

Useful for conditional forms:
```js
let enableShipping = widget.NewCheck("Enable shipping", (checked) => {
    if checked {
        shippingRadio.Enable()
    } else {
        shippingRadio.Disable()
    }
})
```

## Comparison with Similar Widgets

**RadioGroup vs Select:**
- RadioGroup: All options visible, single selection
- Select: Dropdown, single selection, saves space

**RadioGroup vs CheckGroup:**
- RadioGroup: Single selection (one at a time)
- CheckGroup: Multiple selections allowed

**When to use RadioGroup:**
- 2-7 options (for more, consider Select)
- Options should be visible at all times
- Single selection required
- Visual comparison of options helpful

**When NOT to use RadioGroup:**
- Many options (>7) - use Select dropdown
- Multiple selections needed - use CheckGroup
- Space is limited - use Select

## Example Sections

1. **Basic Vertical** - Default vertical layout with animal selection
2. **Horizontal** - Horizontal layout for size selection
3. **Required** - Required field with validation indicator
4. **Dynamic Options** - Add/remove options at runtime
5. **Programmatic Control** - Set selection, enable/disable from code
6. **Form Integration** - Complete form with payment and shipping selection

## Files

- `main.go`: Go program that creates the window
- `widgets.risor`: Risor script demonstrating RadioGroup
- `README.md`: This file
