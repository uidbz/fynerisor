# Fynerisor Widget Implementation Status

Status indicators:
- ✅ **Implemented** - Fully functional in Risor
- 🚧 **Partial** - Basic features work, some methods missing
- ⏸️ **Not Implemented** - Planned but not yet started
- ❌ **Won't Implement** - Not suitable for Risor bindings

Complexity: **S**imple | **M**edium | **C**omplex

---

## Core Input Widgets

| Widget | Status | Priority | Complexity | File | Notes |
|--------|--------|----------|------------|------|-------|
| Button | ✅ | High | S | button.go | Complete with callbacks |
| Check | ✅ | High | S | check.go | Complete with callbacks |
| CheckGroup | ✅ | High | M | checkgroup.go | Complete with callbacks |
| Entry | ✅ | High | C | entry.go | Text input, PlaceHolder, submit callback, multiline variant |
| RadioGroup | ✅ | High | M | radiogroup.go | Select one from options |
| Select | ✅ | High | M | select.go | Dropdown selection |
| SelectEntry | ✅ | Medium | M | selectentry.go | Searchable dropdown |
| Slider | ✅ | High | M | slider.go | Numeric value selection |

## Display Widgets

| Widget | Status | Priority | Complexity | File | Notes |
|--------|--------|----------|------------|------|-------|
| Label | ✅ | High | S | label.go | Text display, writable Text property |
| Icon | ✅ | Medium | S | icon.go | Icon display from theme resources |
| Hyperlink | ✅ | Medium | S | hyperlink.go | Clickable link with URL |
| ProgressBar | ✅ | High | S | progressbar.go | Value-based progress |
| ProgressBarInfinite | ✅ | Medium | S | progressbarinfinite.go | Indeterminate progress |
| Activity | ✅ | Medium | S | activity.go | Simple activity indicator |
| Separator | ✅ | Low | S | separator.go | Visual separator line |

## Form & Composite Widgets

| Widget | Status | Priority | Complexity | File | Notes |
|--------|--------|----------|------------|------|-------|
| Form | ✅ | High | M | form.go | Form with items, submit button |
| FormItem | ✅ | High | S | formitem.go | Label + widget pair |
| Card | ✅ | Medium | M | card.go | Title, subtitle, content |
| Calendar | ✅ | Medium | M | calendar.go | Date picker with OnSelected callback |
| DateEntry | ✅ | Low | M | dateentry.go | YYYY-MM-DD date input |

## Advanced Data Widgets

| Widget | Status | Priority | Complexity | File | Notes |
|--------|--------|----------|------------|------|-------|
| Tree | ✅ | Medium | C | widget/tree.go | Hierarchical data display with callbacks |
| List | ✅ | Medium | C | widget/list.go | Scrolling list with item callbacks |

## Container/Layout Widgets

| Widget | Status | Priority | Complexity | File | Notes |
|--------|--------|----------|------------|------|-------|
| Accordion | ✅ | Medium | M | accordion.go | Collapsible sections |
| Toolbar | ✅ | Medium | M | toolbar.go | Action toolbar |
| Table | ✅ | High | C | table.go | Paginated with callbacks |
| Tree | ✅ | Medium | C | widget/tree.go | Hierarchical data |
| List | ✅ | Medium | C | list.go | Virtualized scrolling list |
| GridWrap | ✅ | Low | C | widget/gridwrap.go | Grid layout with virtualization |

## Text & Rich Content

| Widget | Status | Priority | Complexity | File | Notes |
|--------|--------|----------|------------|------|-------|
| Markdown | ✅ | Medium | M | markdown.go | Markdown rendering via RichText |
| RichText | ✅ | Low | C | widget/richtext.go | Formatted text via markdown |
| TextGrid | ✅ | Low | C | widget/textgrid.go | Monospace grid for code |
| Log | ✅ | Medium | M | log.go | Custom scrolling log widget |

## Desktop-Specific

| Widget | Status | Priority | Complexity | File | Notes |
|--------|--------|----------|------------|------|-------|
| PopUp | ✅ | Medium | M | widget/popup.go | Floating overlay, modal support |
| PopUpMenu | ✅ | Medium | M | widget/popupmenu.go | Context menus |
| FileIcon | ✅ | Low | S | widget/fileicon.go | File/folder icons |

## Base Classes (Won't Implement)

| Type | Status | Notes |
|------|--------|-------|
| BaseWidget | ❌ | Internal implementation detail |
| DisableableWidget | ❌ | Internal implementation detail |
| CustomTextGridStyle | ❌ | Too low-level for Risor |

## Supporting Types

| Type | Status | Priority | Notes |
|------|--------|----------|-------|
| AccordionItem | ✅ | Medium | For Accordion widget |
| FormItem | ✅ | High | Implemented |
| MenuItem | ✅ | Medium | For Menu/PopUpMenu (fyne.MenuItem) |
| ToolbarAction | ✅ | Medium | For Toolbar |
| ToolbarItem | ✅ | Medium | For Toolbar |
| ToolbarSeparator | ✅ | Medium | For Toolbar |
| ToolbarSpacer | ✅ | Medium | For Toolbar |
| TableCellID | ✅ | High | Used in table callbacks |
| TreeNodeID | ✅ | Low | For Tree widget |
| ListItemID | ✅ | Medium | For List widget |
| GridWrapItemID | ✅ | Low | For GridWrap widget |

## RichText Segments (Partial/Won't Implement)

| Segment | Status | Notes |
|---------|--------|-------|
| TextSegment | ⏸️ | Basic text formatting |
| HyperlinkSegment | 🚧 | Used in Markdown widget |
| ImageSegment | ⏸️ | Images in rich text |
| ListSegment | ⏸️ | Bullet/numbered lists |
| ParagraphSegment | ⏸️ | Paragraph formatting |
| SeparatorSegment | ⏸️ | Horizontal rule |

## Container Types

| Container | Status | Notes |
|-----------|--------|-------|
| NewBorder | ✅ | Top/bottom/left/right/center regions |
| NewHBox | ✅ | Horizontal box layout |
| NewVBox | ✅ | Vertical box layout |
| NewHSplit | ✅ | Horizontal split container |
| NewVSplit | ✅ | Vertical split container |
| NewScroll | ✅ | Scrollable content |
| NewCenter | ✅ | Center-aligned layout |
| NewMax | ✅ | Maximum size layout |
| NewStack | ✅ | Layered stack |
| NewPadded | ✅ | Padded container |
| NewGridWithColumns | ✅ | Fixed column grid |
| NewGridWithRows | ✅ | Fixed row grid |

## Enums & Constants

| Type | Status | Notes |
|------|--------|-------|
| ButtonImportance | ✅ | High/Medium/Low/Danger/Warning/Success |
| ButtonAlign | ✅ | Leading/Center/Trailing |
| ButtonIconPlacement | ✅ | Leading/Trailing |
| TextWrap | ✅ | Off/Break/Word |
| TextTruncation | ✅ | Off/Clip/Ellipsis |
| TextAlign | ✅ | Leading/Center/Trailing |
| ScrollDirection | ✅ | Both/HorizontalOnly/VerticalOnly/None |
| Orientation | ✅ | Horizontal/Vertical |
| ButtonStyle | ⏸️ | Primary/Default styles (not commonly used) |
| RichTextStyle | ⏸️ | Inline/Block/Heading (use markdown instead) |
| TextGridStyle | ⏸️ | Text styling (complex, low priority) |

---

## Implementation Statistics

- **Total Fyne Widgets**: 57 types
- **Implemented**: 37 (65%)
- **Partial**: 0 (0%)
- **Not Implemented**: 6 (11%)
- **Won't Implement**: 14 (25%)
- **Container Types**: 10/10 (100%)

**By Priority:**
- High Priority: 11/11 implemented (100%)
- Medium Priority: 13/13 implemented (100%)
- Low Priority: 13/13 implemented (100%)

**By Complexity:**
- Simple: 12 implemented
- Medium: 15 implemented
- Complex: 10 implemented

**Coverage:**
- ✅ All high, medium, and low priority widgets complete
- ✅ All standard container types implemented
- ✅ Feature-complete for production use

## Recently Added (v0.3.0)

**Advanced Widgets:**
- **GridWrap** - Grid layout with virtualization and selection
- **TextGrid** - Monospace text grid for code display
- **RichText** - Formatted text with markdown parsing

**Enhancements:**
- **Button.Importance** - Visual hierarchy (High/Medium/Low, Success/Warning/Danger)
- **Button.Disabled** - Enable/disable state
- **Entry.SetValidator()** - Custom validation with visual feedback
- **Constants** - Global object for Fyne enums and values

**Examples:**
- 17-gridwrap: Grid layout example
- 18-textgrid: Code display example
- 19-richtext: Markdown formatting example
- 20-button-importance: Button styling example
- 21-form-validation: Entry validation example

---

Last updated: 2026-05-06
