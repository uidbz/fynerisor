# Data Binding Example

Demonstrates Fyne's data binding system which allows widgets to automatically sync with data sources.

## Features

- **Automatic Updates**: Widgets bound to data automatically update when data changes
- **Two-Way Binding**: Entry changes update the binding, which updates the label
- **Listeners**: Get notified when bound data changes
- **Programmatic Updates**: Change data from code and see UI update

## Concepts

- Creating data bindings with `binding.NewString()`
- Binding widgets with `widget.NewLabelWithData()`
- Getting and setting bound values
- Adding change listeners

## Data Binding Types

Currently supported:
- `binding.NewString()` - String data binding
- `binding.NewString("initial")` - String with initial value

More types coming soon:
- Bool, Int, Float bindings
- List bindings for List/Table widgets

## Run

```bash
go run main.go
```

Or with fynerisor CLI:

```bash
fynerisor script.risor
```
