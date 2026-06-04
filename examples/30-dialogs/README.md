# Dialog Examples

This example demonstrates all dialog types available in fynerisor.

## Running the Examples

### Basic Dialogs (Show* functions)

Run the basic example showing simple convenience functions:

```bash
go run main.go
```

Or from the repository root:
```bash
go run ./examples/30-dialogs
```

This demonstrates:
- ShowInformation - simple info messages
- ShowError - error notifications
- ShowConfirm - yes/no confirmations
- ShowFileOpen - file picker
- ShowFileSave - save file dialog
- ShowFolderOpen - folder picker
- ShowColorPicker - color selection
- ShowForm - forms with validation
- ShowCustom - custom content dialogs
- ShowCustomConfirm - custom content with confirmation

### Advanced Dialogs (New* constructors)

Run the advanced example showing dialog objects with more control:

```bash
go run main.go advanced
```

This demonstrates:
- NewFileOpen with SetFilter, SetLocation, SetFileName
- NewFileSave with pre-filled filename
- NewConfirm with custom button text and importance
- NewColorPicker with Advanced mode
- NewCustom with complex widget content
- NewForm with programmatic submit
- NewCustomConfirm with custom content

## Dialog Types

### Simple Show* Functions
Use these for quick, one-off dialogs:
- `dialog.ShowInformation(title, message)`
- `dialog.ShowError(message)`
- `dialog.ShowConfirm(title, message, callback)`
- `dialog.ShowFileOpen(callback)`
- `dialog.ShowColorPicker(title, message, callback)`

### Advanced New* Constructors
Use these when you need more control:
- Create dialog object: `let fd = dialog.NewFileOpen(callback)`
- Customize: `fd.SetFilter(".txt")`, `fd.SetLocation("/tmp")`
- Show when ready: `fd.Show()`

## Key Features

- All callbacks are automatically marshalled to GUI thread
- File dialogs return paths as strings (not URI objects)
- Color picker returns RGB map: `{R: 255, G: 128, B: 0, A: 255}`
- Form dialogs include automatic validation
- All dialogs support Show(), Hide(), Refresh()
