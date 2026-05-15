# Example 04: Table Display

Demonstrates the table widget with paginated data, custom data callbacks, and row click handling.

## What It Does

- Creates a table widget with 5 rows per page
- Displays a directory of people with ID, Name, Email, and Status columns
- Implements pagination through data callbacks
- Handles row click events
- Shows how to pass custom data functions from Go to Risor

## Running

```bash
cd examples/04-table-display
go run main.go
```

## Key Concepts

### Table Widget

Create a table with a title and page size:

```js
let table = widget.NewTable("People Directory", 5)
```

### Column Definition

Define column headers with a callback:

```js
table.Columns(() => {
    return ["ID", "Name", "Email", "Status"]
})
```

### Row Count Callback

Provide the total number of rows:

```js
table.RowCount(() => {
    return data.getCount()
})
```

### Data Callback

Provide paginated data based on offset and limit:

```js
table.Data((offset, limit) => {
    return data.getPage(offset, limit)
})
```

The table will call this function automatically when pagination changes.

### Click Handling

Handle row and column click events:

```js
table.SetOnClick((row, col) => {
    print(`Clicked row ${row}, column ${col}`)
})
```

### Passing Data from Go

This example shows how to pass data access functions from Go to Risor:

```go
dataFuncs := map[string]any{
    "data": map[string]any{
        "getCount": func() int {
            return len(people)
        },
        "getPage": func(offset, limit int) [][]string {
            // Return data slice for current page
        },
    },
}

globals := []risor.Option{
    risor.WithEnv(dataFuncs),
}

fyneWindow := fynerisor.NewWindow(w, globals, nil, nil)
```

The Risor script can then call `data.getCount()` and `data.getPage(offset, limit)`.

### Data Format

Table data is provided as a list of lists (rows of columns):

```go
[][]string{
    {"1", "Alice", "alice@example.com", "Active"},
    {"2", "Bob", "bob@example.com", "Inactive"},
}
```

## Files

- `main.go`: Go program that provides data access functions
- `table.risor`: Risor script that creates the table widget
- `README.md`: This file
