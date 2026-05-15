# Data Binding Types Example

Comprehensive demonstration of all data binding types supported by Fyne.

## Features

### String Binding
- `binding.NewString()` or `binding.NewString("initial")`
- Used with: Label, Entry
- Demo: Type in entry → label updates automatically

### Bool Binding  
- `binding.NewBool()` or `binding.NewBool(true)`
- Used with: Check widget
- Demo: Toggle checkbox → status label updates

### Float Binding
- `binding.NewFloat()` or `binding.NewFloat(50.0)`
- Used with: Slider widget
- Demo: Move slider → volume label updates

### Int Binding
- `binding.NewInt()` or `binding.NewInt(0)`
- Used with: Labels (via listeners)
- Demo: Click buttons → count label updates

## Binding Methods

All bindings support:
- `Get()` - Get current value
- `Set(value)` - Update value
- `AddListener(callback)` - React to changes

## Widget Support

### With Data Binding
- `widget.NewLabelWithData(binding.String)`
- `widget.NewCheckWithData(label, binding.Bool)`
- `widget.NewSliderWithData(min, max, binding.Float)`

### Without Data Binding
Regular widgets can still interact with bindings using Get/Set:
```risor
let data = binding.NewInt(0)
let btn = widget.NewButton("Click", () => {
    data.Set(data.Get() + 1)
})
```

## Run

```bash
go run main.go
```

Or with fynerisor CLI:

```bash
fynerisor script.risor
```
