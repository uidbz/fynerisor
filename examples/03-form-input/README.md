# Example 03: Form Input

Demonstrates form widgets, text entry fields, and input validation.

## What It Does

- Creates a registration form with name and email fields using `widget.NewForm()`
- Validates user input when the submit button is clicked
- Shows error messages for invalid input
- Displays success message when validation passes
- Provides a clear button to reset the form
- Uses separate labels for instructions and results

## Running

```bash
cd examples/03-form-input
go run main.go
```

## Key Concepts

### Entry Widgets

Text input fields are created with `widget.NewEntry()`:

```js
let nameEntry = widget.NewEntry()
let emailEntry = widget.NewEntry()
```

### Reading Entry Values

Access the entered text via the `Text` property:

```js
let name = nameEntry.Text
```

### Input Validation

Perform validation in button callbacks:

```js
if (name == "") {
    resultLabel.Text = "Error: Name is required"
    return
}
```

### Form Widget

The Form widget takes an array of FormItem widgets:

```js
let items = [
    widget.NewFormItem("Name:", nameEntry),
    widget.NewFormItem("Email:", emailEntry),
]

let form = widget.NewForm(items)
```

FormItem widgets provide labels for input fields.

### Horizontal Layout

HBox arranges widgets horizontally:

```js
container.NewHBox(submitButton, clearButton)
```

### Early Return Pattern

Use `return` in callbacks to exit early on validation failure:

```js
if (email == "") {
    resultLabel.Text = "Error: Email is required"
    return  // Stop here, don't continue validation
}
```

### String Methods

Risor provides string methods like `contains()`:

```js
if (!email.contains("@")) {
    resultLabel.Text = "Error: Invalid email"
}
```

## Files

- `main.go`: Go program that loads the form
- `form.risor`: Risor script with form validation logic
- `README.md`: This file
