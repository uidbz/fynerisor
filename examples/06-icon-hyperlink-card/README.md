# Example 06: Icon, Hyperlink & Card Widgets

Demonstrates theme icons, clickable hyperlinks, and card layouts.

## What It Does

- Shows **Icon** widgets with 40+ built-in theme icons
- Demonstrates dynamic icon changes with SetResource()
- Shows **Hyperlink** widgets that open URLs in browser
- Demonstrates custom OnTapped callbacks for links
- Shows dynamic hyperlink URL/text updates
- Demonstrates **Card** widgets for grouped content
- Shows cards with various content (labels, buttons, icons)
- Demonstrates dynamic card property updates

## Running

```bash
cd examples/06-icon-hyperlink-card
go run main.go
```

## Widgets Demonstrated

### Icon

Display theme icons:

```js
let icon = widget.NewIcon("search")

// Change icon dynamically
icon.SetResource("settings")
```

**Available Icon Names** (40+ built-in):

*Navigation:*
- cancel, confirm, delete, search, searchReplace
- navigateBack, navigateNext, arrowUp, arrowDown

*Actions:*
- check, add, remove, cut, copy, paste, undo, redo

*Media:*
- mediaPlay, mediaPause, mediaStop, mediaRecord, mediaReplay
- mediaSkipNext, mediaSkipPrevious

*Files:*
- file, folder, folderOpen, document, documentCreate
- documentPrint, documentSave

*Info:*
- info, question, warning, error

*Misc:*
- settings, home, help, history, mail
- visibility, visibilityOff

**Methods:**
- `SetResource(iconName)` - Change the displayed icon

### Hyperlink

Clickable links that open in the default browser:

```js
let link = widget.NewHyperlink("Fyne Website", "https://fyne.io")
```

**Constructor:**
- `NewHyperlink(text, url)` - Creates link with text and URL

**Properties:**
- `Text` - Link text (read/write)
- `URL` - Link URL (read/write)

**Methods:**
- `SetText(text)` - Update link text
- `SetURL(url)` - Update link URL
- `OnTapped(callback)` - Custom action (overrides browser open)

**Custom Action Example:**

```js
let link = widget.NewHyperlink("Count Clicks", "https://example.com")

let clicks = 0
link.OnTapped(() => {
    clicks = clicks + 1
    print(`Clicked ${clicks} times`)
    // Browser won't open when OnTapped is set
})
```

### Card

Container widget with title, subtitle, and content:

```js
let card = widget.NewCard(
    "Card Title",
    "Card Subtitle",
    widget.NewLabel("Card content here")
)
```

**Constructor:**
- `NewCard(title, subtitle, content)` - All parameters required
- Pass empty string `""` for no title/subtitle
- Pass `nil` or any widget for content

**Properties:**
- `Title` - Card title (read/write)
- `Subtitle` - Card subtitle (read/write)

**Methods:**
- `SetTitle(text)` - Update title
- `SetSubTitle(text)` - Update subtitle
- `SetContent(widget)` - Replace card content

**Complex Content Example:**

```js
let cardContent = container.NewVBox(
    widget.NewLabel("Multiple widgets"),
    widget.NewButton("Action", () => print("Clicked")),
    container.NewHBox(
        widget.NewIcon("info"),
        widget.NewLabel("With icons too")
    )
)

let card = widget.NewCard("Rich Content", "Multiple elements", cardContent)
```

## Key Concepts

### Theme Icons

Icons automatically adapt to the current theme (light/dark):

```js
let icon = widget.NewIcon("file")
// Automatically uses theme-appropriate color
```

Icons are vector-based and scale cleanly.

### Hyperlink Behavior

Default behavior opens URL in system browser:

```js
let link = widget.NewHyperlink("Google", "https://google.com")
// Clicking opens browser
```

Override with OnTapped for custom actions:

```js
link.OnTapped(() => {
    // Custom action - browser won't open
    print("Custom action instead of opening browser")
})
```

### Card Organization

Cards provide visual grouping:

```js
// Group related settings
let settingsCard = widget.NewCard(
    "Settings",
    "Application preferences",
    container.NewVBox(
        widget.NewCheck("Enable feature", (checked) => {}),
        widget.NewCheck("Show tooltips", (checked) => {})
    )
)

// Group status information
let statusCard = widget.NewCard(
    "Status",
    "Current state",
    widget.NewLabel("All systems operational")
)
```

### Dynamic Updates

All three widgets support dynamic updates:

```js
// Icon
icon.SetResource("warning")

// Hyperlink
link.SetText("New Text")
link.SetURL("https://new-url.com")

// Card
card.SetTitle("Updated Title")
card.SetSubTitle("Updated Subtitle")
card.SetContent(widget.NewLabel("New content"))
```

### Icon in Cards

Combine widgets for rich layouts:

```js
let icon = widget.NewIcon("check")
let label = widget.NewLabel("Task completed")

let content = container.NewHBox(icon, label)
let card = widget.NewCard("Success", "", content)
```

### Scrollable Content

This example uses a scrollable container for long content:

```js
let longContent = container.NewVBox(
    card1,
    card2,
    card3,
    // ... many cards
)

let scrollable = container.NewScroll(longContent)
window.SetContent(scrollable)
```

## Icon Resource Names

The icon widget supports these built-in theme resources:

**Most Common:**
- `"cancel"`, `"confirm"`, `"delete"`
- `"search"`, `"settings"`, `"home"`
- `"add"`, `"remove"`
- `"file"`, `"folder"`, `"document"`
- `"info"`, `"warning"`, `"error"`

See the example for a complete list of available icons.

## Files

- `main.go`: Go program that creates the window
- `widgets.risor`: Risor script demonstrating widgets
- `README.md`: This file
