# PopUp Example

Demonstrates PopUp widgets - floating overlays above the main UI.

## Features

- **Standard PopUp**: Non-modal, allows background interaction
- **Modal PopUp**: Blocks background interaction until closed
- **Positioned PopUp**: Show at specific screen coordinates
- **Info Dialogs**: Common use case for notifications

## PopUp Types

**NewPopUp(content, canvas):**
- Non-modal popup
- User can interact with background
- Closes when clicked outside

**NewModalPopUp(content, canvas):**
- Modal popup with shadow
- Blocks all background interaction
- User must close it explicitly

## Methods

- `Show()` - Display the popup
- `Hide()` - Hide the popup
- `ShowAtPosition(x, y)` - Show at coordinates
- `Move(x, y)` - Reposition popup
- `Resize(w, h)` - Resize popup

## Run

```bash
go run main.go
```

Or with fynerisor CLI:

```bash
fynerisor script.risor
```
