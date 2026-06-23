# Widget Implementation Status

This document lists the Fyne widgets and features available in fynerisor.

## Implemented Widgets

### Input Widgets
- **Button** - Clickable button with text, icon, importance levels, callbacks
- **Check** - Checkbox with label and OnChanged callback
- **CheckGroup** - Multiple checkboxes with OnChanged callback
- **Entry** - Text input with PlaceHolder, OnChanged, OnSubmitted, multiline support
- **RadioGroup** - Radio button group with OnChanged callback
- **Select** - Dropdown selection with OnChanged callback
- **SelectEntry** - Searchable dropdown with OnChanged callback
- **Slider** - Numeric slider with Min, Max, Step, OnChanged callback

### Display Widgets
- **Label** - Text display with writable Text property, alignment, styling
- **Icon** - Icon display from theme resources
- **Hyperlink** - Clickable link with URL
- **ProgressBar** - Determinate progress indicator with Value property
- **ProgressBarInfinite** - Indeterminate progress animation
- **Activity** - Simple activity/loading indicator
- **Separator** - Visual separator line (horizontal/vertical)

### Form Widgets
- **Form** - Form container with items, submit/cancel buttons, OnSubmit/OnCancel
- **FormItem** - Label + widget pair for forms
- **Card** - Card with title, subtitle, image, content
- **Calendar** - Date picker with OnSelected callback
- **DateEntry** - Date input field (YYYY-MM-DD format)

### Data Widgets
- **Table** - Paginated table with filtering, sorting, export (CSV/XLSX/JSON)
  - Widget mode: use any widget type in cells (buttons, icons, images, etc.)
  - CreateCell and UpdateCell callbacks for custom rendering
- **Tree** - Hierarchical data display with expand/collapse
- **List** - Virtualized scrolling list with CreateItem/UpdateItem callbacks
- **GridWrap** - Grid layout with virtualization and selection

### Layout Widgets
- **Accordion** - Collapsible sections with AccordionItem
- **Toolbar** - Action toolbar with ToolbarAction, ToolbarSeparator, ToolbarSpacer

### Text Widgets
- **Markdown** - Markdown rendering (uses RichText internally)
- **RichText** - Formatted text display via markdown
- **TextGrid** - Monospace text grid for code display with line numbers
- **Log** - Custom scrolling log widget

### Desktop Widgets
- **PopUp** - Floating overlay window (modal/non-modal)
- **PopUpMenu** - Context menus
- **FileIcon** - File/folder icons with file type detection

## Container Types

All standard Fyne containers are supported:

- **NewBorder** - Border layout (top, bottom, left, right, center)
- **NewHBox** - Horizontal box layout
- **NewVBox** - Vertical box layout
- **NewHSplit** - Horizontal split container with draggable divider
- **NewVSplit** - Vertical split container with draggable divider
- **NewScroll** - Scrollable content wrapper
- **NewCenter** - Center-aligned layout
- **NewMax** - Maximum size layout
- **NewStack** - Layered stack layout
- **NewPadded** - Padded container
- **NewGridWithColumns** - Fixed column grid
- **NewGridWithRows** - Fixed row grid

## Canvas Objects

- **Image** - Display images from files or data
- **Rectangle** - Colored rectangles
- **Circle** - Colored circles
- **Line** - Lines between points
- **Text** - Basic text rendering

## Dialogs

- **dialog.ShowInformation** - Information dialog
- **dialog.ShowError** - Error dialog
- **dialog.ShowConfirm** - Confirmation dialog with Yes/No
- **dialog.ShowFileOpen** - File picker (single/multiple)
- **dialog.ShowFileSave** - File save dialog
- **dialog.ShowFolderOpen** - Folder picker
- **dialog.ShowCustom** - Custom dialog with any widget
- **dialog.ShowCustomConfirm** - Custom dialog with confirm buttons

## Window Features

- **window.SetContent()** - Set window content
- **window.Resize()** - Set window size
- **window.SetTitle()** - Set window title
- **window.ShowAndRun()** - Display window and start event loop
- **window.Close()** - Close window
- **window.AddShortcut()** - Register keyboard shortcuts (Ctrl+S, etc.)
- **window.RemoveShortcut()** - Remove keyboard shortcuts
- **window.SetMainMenu()** - Set application menu bar

## Data Binding

- **binding.NewString()** - String data binding
- **binding.NewFloat()** - Float data binding
- **binding.NewInt()** - Integer data binding
- **binding.NewBool()** - Boolean data binding
- Widget properties can be bound to data sources for reactive updates

## Charts (via go-echarts)

- **Line charts** - Line plots with multiple series
- **Bar charts** - Bar/column charts
- **Pie charts** - Pie/donut charts
- **Scatter plots** - XY scatter plots
- Interactive features: zoom, pan, tooltips, legends

## Modules Available

Enable with `gui.With*()` options in your application:

- **http** - HTTP client (WithHTTP)
- **os** - File system operations (WithOS)
- **sql** - Database access (WithSQL)
- **strings** - String utilities (WithStrings)
- **filepath** - Path manipulation (WithFilepath)
- **time** - Time operations (WithTime)
- **io** - I/O operations (WithIO)
- **exec** - Execute external commands (WithExec)

## Constants & Enums

Available via the global `constants` object:

- **ButtonImportance** - High, Medium, Low, Success, Warning, Danger
- **ButtonAlign** - Leading, Center, Trailing
- **ButtonIconPlacement** - Leading, Trailing
- **TextWrap** - Off, Break, Word
- **TextTruncation** - Off, Clip, Ellipsis
- **TextAlign** - Leading, Center, Trailing
- **ScrollDirection** - Both, HorizontalOnly, VerticalOnly, None
- **Orientation** - Horizontal, Vertical

## Not Implemented

Low-level or internal Fyne types not suitable for Risor scripting:

- BaseWidget (internal implementation detail)
- DisableableWidget (internal implementation detail)
- CustomTextGridStyle (too low-level)
- Individual RichText segments (use Markdown widget instead)

---

**See [examples/](../examples/) directory for usage examples of all widgets.**

Last updated: 2026-06-23
