# GridWrap Example

Demonstrates the GridWrap widget - a grid layout with virtualized rendering for efficiently displaying many items.

## Features

- **Grid Layout**: Automatically arranges items in a grid
- **Virtualized Rendering**: Only renders visible items for performance
- **Selection**: Click to select/deselect items
- **Callbacks**: Length(), CreateItem(), UpdateItem()
- **Events**: OnSelected(), OnUnselected()

## Concepts

- Creating a GridWrap widget
- Setting data source callbacks
- Handling item selection
- Updating status based on events

## Run

```bash
go run main.go
```

Or with fynerisor CLI:

```bash
fynerisor script.risor
```
