# Example 08: Accordion Widget

Demonstrates collapsible sections with expandable/collapsible accordion items.

## What It Does

- Shows **Accordion** widget for collapsible sections
- Demonstrates single-open mode (default) vs multi-open mode
- Shows programmatic Open/Close/CloseAll control
- Demonstrates dynamic item management (Append, Remove, RemoveIndex)
- Shows rich content in accordion items (forms, buttons, other widgets)
- Demonstrates practical use cases (FAQ, settings panels)

## Running

```bash
cd examples/08-accordion
go run main.go
```

## Widgets

### Accordion

Container widget that displays collapsible sections. Users can expand/collapse items by clicking their titles.

**Constructor:**
```js
let accordion = widget.NewAccordion(item1, item2, item3, ...)
```

Takes zero or more AccordionItem objects.

**Properties:**

`MultiOpen` (read/write, boolean):
```js
accordion.MultiOpen = true   // Allow multiple items open simultaneously
accordion.MultiOpen = false  // Only one item open at a time (default)
```

**Methods:**

**Open(index)** - Open item at index:
```js
accordion.Open(0)   // Open first item
```

**Close(index)** - Close item at index:
```js
accordion.Close(1)  // Close second item
```

**CloseAll()** - Close all items:
```js
accordion.CloseAll()
```

**Append(item)** - Add item to end:
```js
let newItem = widget.NewAccordionItem("Title", "", content)
accordion.Append(newItem)
```

**Remove(item)** - Remove specific item:
```js
accordion.Remove(item1)
```

**RemoveIndex(index)** - Remove item at index:
```js
accordion.RemoveIndex(0)  // Remove first item
```

**Refresh()** - Force visual refresh:
```js
accordion.Refresh()
```

### AccordionItem

Individual collapsible section within an Accordion.

**Constructor:**
```js
let item = widget.NewAccordionItem(title, detail, content)
```

**Parameters:**
- `title` (string) - Item title shown in header
- `detail` (string) - Currently unused, pass empty string ""
- `content` (CanvasObject) - Widget(s) to display when expanded, or nil

**Properties:**

`Title` (read/write, string):
```js
item.Title = "New Title"
let currentTitle = item.Title
```

`Open` (read/write, boolean):
```js
item.Open = true   // Expand the item
item.Open = false  // Collapse the item
```

## Key Concepts

### Single-Open vs Multi-Open Mode

**Single-Open (default):**
```js
let accordion = widget.NewAccordion(item1, item2, item3)
// Opening one item automatically closes others
```

**Multi-Open:**
```js
let accordion = widget.NewAccordion(item1, item2, item3)
accordion.MultiOpen = true
// Multiple items can be open simultaneously
```

### Creating Items with Complex Content

Simple content:
```js
let item = widget.NewAccordionItem(
    "Title",
    "",
    widget.NewLabel("Simple text content")
)
```

Complex content with multiple widgets:
```js
let content = container.NewVBox(
    widget.NewLabel("Form fields:"),
    widget.NewEntry(),
    widget.NewButton("Submit", () => { print("Submitted") })
)

let item = widget.NewAccordionItem("Form", "", content)
```

### Programmatic Control

Open/close items from code:
```js
accordion.Open(0)        // Open first item
accordion.Close(1)       // Close second item
accordion.CloseAll()     // Close everything
```

Check if item is open:
```js
if item.Open {
    print("Item is expanded")
}
```

Set item state directly:
```js
item.Open = true   // Expand
item.Open = false  // Collapse
```

### Dynamic Item Management

Start with empty accordion:
```js
let accordion = widget.NewAccordion()
```

Add items dynamically:
```js
let item = widget.NewAccordionItem("New", "", widget.NewLabel("Content"))
accordion.Append(item)
```

Remove items:
```js
accordion.Remove(item)           // Remove specific item
accordion.RemoveIndex(0)         // Remove first item
```

**Pattern for clearing all items:**
```js
let itemCount = 3  // Track your items
while itemCount > 0 {
    accordion.RemoveIndex(0)
    itemCount = itemCount - 1
}
```

### Updating Item Content

You cannot change the content of an AccordionItem after creation. To update content:

1. Remove the old item
2. Create new item with updated content
3. Append the new item

```js
// Remove old item
accordion.RemoveIndex(0)

// Create new item with updated content
let newItem = widget.NewAccordionItem(
    "Updated Title",
    "",
    widget.NewLabel("New content")
)

// Add it back
accordion.Append(newItem)
```

### Use Cases

**FAQ / Help Sections:**
```js
let faq = widget.NewAccordion(
    widget.NewAccordionItem("How do I...?", "", widget.NewLabel("Answer...")),
    widget.NewAccordionItem("What is...?", "", widget.NewLabel("Explanation...")),
    widget.NewAccordionItem("Where can I...?", "", widget.NewLabel("Location..."))
)
```

**Settings Panels:**
```js
let settings = widget.NewAccordion(
    widget.NewAccordionItem("Account", "", accountForm),
    widget.NewAccordionItem("Privacy", "", privacyOptions),
    widget.NewAccordionItem("Advanced", "", advancedSettings)
)
settings.MultiOpen = true  // Allow editing multiple sections
```

**Wizard Steps:**
```js
let wizard = widget.NewAccordion(
    widget.NewAccordionItem("Step 1: Setup", "", step1Content),
    widget.NewAccordionItem("Step 2: Configure", "", step2Content),
    widget.NewAccordionItem("Step 3: Review", "", step3Content)
)
// Open only current step
wizard.Open(0)
```

### Styling & Behavior

- Item headers are always visible
- Click header to toggle open/closed
- Open item shows content below header
- Smooth animation on expand/collapse
- In single-open mode, opening new item closes current
- In multi-open mode, items stay open until explicitly closed

### Limitations

- Cannot change item content after creation (must remove and re-add)
- Detail parameter is currently unused in constructor
- No direct access to item list (use Append/Remove instead)
- Cannot reorder items (must remove and re-add in desired order)

## Example Sections

1. **Basic Accordion** - Simple three-item accordion with text/widget content
2. **Multi-Open** - Multiple sections can be open simultaneously
3. **Dynamic Control** - Open/Close items programmatically with buttons
4. **Add/Remove Items** - Dynamically add and remove items
5. **FAQ Example** - Practical FAQ use case with links
6. **Rich Content** - Complex forms and widgets inside accordion items

## Files

- `main.go`: Go program that creates the window
- `widgets.risor`: Risor script demonstrating Accordion
- `README.md`: This file
