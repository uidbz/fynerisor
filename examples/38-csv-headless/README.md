# 38-csv-headless

Headless (no GUI) example demonstrating the `csv` module. Uses only `core` —
no Fyne/GUI dependencies.

## What it demonstrates

Every function in the `csv` module:

- `csv.parse(text)` — first row is treated as headers, returns a **list of maps**
- `csv.parse(text, {header: false})` — returns a **list of lists** (raw rows)
- `csv.parse(text, {delimiter: ";"})` — custom field delimiter
- `csv.format(rows)` — encode a list of maps; columns are sorted alphabetically
- `csv.format(rows, {columns: [...]})` — control column order / subset
- `csv.format(rows, {header: false})` — omit the header row
- `csv.format(listOfLists, {delimiter: ";"})` — encode raw rows verbatim
- `csv.write(path, rows)` / `csv.read(path)` — file round-trip

Enable the module with `core.WithCSV()` and declare it in scripts with
`require(["@csv"])`.

## Run

```bash
go run main.go
```

The file round-trip writes `people-out.csv` in the current directory.
