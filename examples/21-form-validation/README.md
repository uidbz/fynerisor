# Form Validation Example

Demonstrates entry validation with visual feedback and custom validation rules.

## Features

- **Entry Validation**: Real-time validation with error messages
- **Visual Feedback**: Invalid fields show red borders
- **Custom Rules**: Email format, age range, username length
- **Button Styling**: Success and low importance buttons

## Validation Rules

**Email:**
- Required field
- Must contain "@"
- Must contain "."

**Age:**
- Required field
- Must be 18 or older
- Must be 120 or younger

**Username:**
- Required field
- Minimum 3 characters
- Maximum 20 characters

## Concepts

- `SetValidator()` on Entry widgets
- Return error string for invalid input
- Return `nil` for valid input
- Fyne shows visual feedback automatically

## Run

```bash
go run main.go
```

Or with fynerisor CLI:

```bash
fynerisor script.risor
```
