# Constants & Enums Example

Comprehensive demonstration of all Fyne constants and enums available through the `constants` global object.

## Features

### Button Importance
- **Standard**: ImportanceHigh, ImportanceMedium, ImportanceLow
- **Semantic**: SuccessImportance, WarningImportance, DangerImportance

### Text Wrapping
- **TextWrapOff** (0): No wrapping, text extends beyond container
- **TextWrapBreak** (1): Break at any character
- **TextWrapWord** (2): Break at word boundaries (recommended)

### Text Truncation
- **TextTruncateOff** (0): No truncation
- **TextTruncateClip** (1): Hard clip at edge
- **TextTruncateEllipsis** (2): Show ... for overflow

### Scroll Direction
- **ScrollBoth** (0): Horizontal and vertical
- **ScrollHorizontalOnly** (1): Left/right only
- **ScrollVerticalOnly** (2): Up/down only
- **ScrollNone** (3): No scrolling

### Button Alignment
- **ButtonAlignCenter** (0): Center alignment
- **ButtonAlignLeading** (1): Left alignment
- **ButtonAlignTrailing** (2): Right alignment

### Button Icon Placement
- **ButtonIconLeadingText** (0): Icon before text
- **ButtonIconTrailingText** (1): Icon after text

### Orientation
- **Horizontal** (0): Horizontal orientation
- **Vertical** (1): Vertical orientation

## Usage

All constants are accessed via the global `constants` object:

```risor
btn.Importance = constants.ImportanceHigh
label.Wrapping = constants.TextWrapWord
label.Truncation = constants.TextTruncateEllipsis
```

## Run

```bash
go run main.go
```

Or with fynerisor CLI:

```bash
fynerisor script.risor
```
