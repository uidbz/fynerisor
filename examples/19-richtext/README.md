# RichText Example

Demonstrates the RichText widget - formatted text with markdown support.

## Features

- **Markdown Parsing**: Full markdown syntax support
- **Text Formatting**: Bold, italic, headers, lists, links
- **Text Wrapping**: Control how text wraps
- **Scrollable**: Large content automatically scrollable
- **Dynamic Content**: Update text at runtime

## Concepts

- Creating RichText from markdown
- Parsing markdown dynamically
- Controlling text wrapping
- Loading different content

## Note

Direct segment manipulation is not exposed in Risor bindings. Use `ParseMarkdown()` for rich formatting instead.

## Run

```bash
go run main.go
```

Or with fynerisor CLI:

```bash
fynerisor script.risor
```
