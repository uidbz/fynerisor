# HTTP Fetch Example

This example demonstrates using the HTTP module to fetch data from a REST API and display it in a Fyne GUI.

## Features

- HTTP GET request to GitHub API
- JSON response parsing
- Status code display
- Dynamic content display
- Standalone Go application using fynerisor library

## Running

### Build and run the standalone app:

```bash
cd examples/04-http-fetch
go mod tidy
go run main.go
```

### Or with the fynerisor CLI:

```bash
cd cmd/fynerisor
go build
./fynerisor --title "HTTP Fetch" ../../examples/04-http-fetch/main.risor
```

## HTTP Module API

The HTTP module provides the following functions:

### `http.get(url, headers?)`
Performs an HTTP GET request.

```javascript
let response = http.get("https://api.example.com/data")
let response = http.get("https://api.example.com/data", {"Authorization": "Bearer token"})
```

### `http.post(url, headers?, body?)`
Performs an HTTP POST request.

```javascript
let response = http.post("https://api.example.com/data", {}, {"key": "value"})
```

### `http.put(url, headers?, body?)`
Performs an HTTP PUT request.

```javascript
let response = http.put("https://api.example.com/data/1", {}, {"key": "updated"})
```

### `http.delete(url, headers?)`
Performs an HTTP DELETE request.

```javascript
let response = http.delete("https://api.example.com/data/1")
```

### `http.fetch(url, options?)`
Performs a customizable HTTP request.

```javascript
let response = http.fetch("https://api.example.com/data", {
    method: "POST",
    headers: {"Content-Type": "application/json"},
    body: {"key": "value"}
})
```

## Response Object

All HTTP functions return a response object with the following fields:

- `status` (int): HTTP status code (e.g., 200, 404)
- `statusText` (string): HTTP status text (e.g., "200 OK")
- `headers` (map): Response headers
- `body` (string): Response body as string
- `ok` (bool): True if status is 2xx

## Enabling HTTP Module

To use the HTTP module in your application, enable it with `WithHTTP()`:

```go
import "github.com/uidbz/fynerisor"

window := fynerisor.NewApp("My App",
    fynerisor.WithHTTP(),
)
```

Or with the CLI:

```bash
# HTTP module is automatically enabled in the CLI
fynerisor script.risor
```
