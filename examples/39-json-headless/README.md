# 39-json-headless

Headless (no GUI) example demonstrating the `json` module. Uses only `core` —
no Fyne/GUI dependencies.

## What it demonstrates

Every function in the `json` module:

- `json.parse(text)` — decode a JSON string into Risor values (maps, lists, scalars)
- `json.marshal(obj)` — encode a Risor value to a compact JSON string
- `json.marshal_indent(obj)` — pretty-printed JSON (default 2-space indent)
- `json.marshal_indent(obj, "    ")` — pretty-printed with a custom indent
- `json.valid(text)` — report whether a string is valid JSON
- `json.write(path, obj, indent?)` / `json.read(path)` — file round-trip

Enable the module with `core.WithJSON()` and declare it in scripts with
`require(["@json"])`.

## Run

```bash
go run main.go
```

The file round-trip writes `record-out.json` in the current directory.
