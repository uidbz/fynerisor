# Calendar Widget Example

This example demonstrates the Calendar widget with the time module for date selection.

## Features Demonstrated

- **Calendar with Callback**: Shows how to handle date selection events
- **Time Module**: Using `time.now()`, `time.date()`, and `time.parse()`
- **Date Properties**: Accessing `year`, `month`, `day` from TimeObject
- **Date Formatting**: Using the `format()` method with Go time layouts
- **Multiple Calendars**: Creating calendars with different start dates
- **Calendar without Callback**: Simple display-only calendar

## Running the Example

```bash
cd examples/10-calendar
go mod tidy
go run main.go
```

Or use the CLI:

```bash
cd examples/10-calendar
../../cmd/fynerisor/fynerisor-cli app.risor
```

## Key Concepts

### Creating a Calendar

```js
// Calendar with current date and callback
let calendar = widget.NewCalendar(time.now(), (date) => {
    print(`Selected: ${date.year}-${date.month}-${date.day}`)
})

// Calendar with specific start date
let specificDate = time.date(2026, 1, 1)
let calendar2 = widget.NewCalendar(specificDate, (date) => {
    print(`Selected date`)
})

// Calendar without callback
let calendar3 = widget.NewCalendar(time.now())

// Calendar with default date (today)
let calendar4 = widget.NewCalendar()
```

### Time Module Functions

```js
// Get current time
let now = time.now()

// Create specific date (year, month, day)
let date = time.date(2026, 12, 25)

// Parse date string (YYYY-MM-DD format)
let parsed = time.parse("2026-01-15")
```

### TimeObject Properties

```js
let date = time.date(2026, 5, 1)

// Access properties
let year = date.year      // 2026
let month = date.month    // 5
let day = date.day        // 1

// Format with Go time layout
let formatted = date.format("2006-01-02")           // "2026-05-01"
let readable = date.format("Monday, January 2, 2006")  // "Thursday, May 1, 2026"
```

### Common Date Formats

```js
// ISO 8601
date.format("2006-01-02")

// US format
date.format("01/02/2006")

// Long format
date.format("Monday, January 2, 2006")

// Short format
date.format("Jan 2, 2006")

// With time
date.format("2006-01-02 15:04:05")
```

## What This Example Shows

1. **Interactive Date Selection**: Click dates to see callback execution
2. **Multiple Calendar Instances**: Each calendar can have different start dates and callbacks
3. **Date Object Usage**: Working with TimeObject properties and methods
4. **Layout Integration**: Combining calendars with other widgets using containers
5. **Scrollable Content**: Using scroll container for multiple sections

## Notes

- Calendar widget wraps Fyne's `widget.Calendar`
- Callback passes a `TimeObject` (not a string)
- TimeObject provides `year`, `month`, `day` properties
- Use `format()` method for custom date formatting with Go time layouts
- Month is 1-indexed (1 = January, 12 = December)
- Calendar without callback is display-only
