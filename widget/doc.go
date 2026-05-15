// Package widget provides Risor bindings for Fyne GUI widgets.
//
// This package wraps Fyne v2 widgets for use in Risor v2 scripts, enabling
// cross-platform desktop GUI development with a simple scripting interface.
//
// # Available Widgets
//
// Input Widgets:
//   - Button: Clickable button with callback
//   - Entry: Single-line text input with optional submit callback
//   - MultiLineEntry: Multi-line text input
//   - Check: Checkbox with changed callback
//   - CheckGroup: Multiple checkboxes with changed callback
//   - Select: Dropdown selection with changed callback
//   - RadioGroup: Radio button group with changed callback
//   - SelectEntry: Searchable dropdown with changed callback
//   - Slider: Numeric value slider with changed callback
//
// Display Widgets:
//   - Label: Text display with writable Text property
//   - Icon: Icon display from theme resources
//   - Hyperlink: Clickable link with URL
//   - ProgressBar: Value-based progress indicator
//   - ProgressBarInfinite: Indeterminate progress spinner
//   - Activity: Simple activity indicator
//   - Separator: Visual separator line
//   - Markdown: Markdown content renderer
//
// Data Widgets:
//   - Table: Paginated table with callbacks for data fetching
//   - Tree: Hierarchical data display with callbacks
//   - List: Scrolling list with item callbacks
//   - Log: Read-only scrollable log display
//
// Composite Widgets:
//   - Form: Form with items and optional submit button
//   - FormItem: Label + widget pair for forms
//   - Card: Title, subtitle, and content card
//   - Accordion: Collapsible sections with items
//   - AccordionItem: Individual accordion section
//   - Toolbar: Action toolbar with items
//
// # Risor v2 Syntax
//
// All callbacks must use arrow function syntax:
//
//	let btn = widget.NewButton("Click", () => {
//	    window.SetStatus("Clicked!")
//	})
//
// # Thread Safety
//
// All widget callbacks execute synchronously in the UI thread. The Risor VM
// is single-threaded and not thread-safe. Never spawn goroutines or use
// async patterns inside callbacks.
//
// See CONCURRENCY.md for detailed information about thread safety.
package widget
