# Interactive Charts

Demonstrates the interactive chart engine: charts are live widgets built
from Fyne canvas primitives, not static images.

## Features Demonstrated

- **Hover** - hovering a bar/point/bin highlights it and shows a
  label/value tooltip automatically (no script code needed)
- **Tap events** - `OnTap` on the bar chart and scatter plot writes to the
  status bar
- **Right-click menus** - `OnSecondaryTap` + `widget.NewPopUp` gives bars a
  context menu (recolor the bar, close)
- **Feature toggles** - checkboxes call `ShowGrid`, `ShowValues`,
  `ShowNormal`; a slider calls `SetBins`
- **Live data** - "Randomize data" calls `SetValue` per category
- **Zoom & pan** - scroll the wheel over the scatter plot to zoom around
  the cursor, drag to pan, double-click to reset (`ResetZoom` button too)
- **SVG export** - "Export bar chart to SVG" writes `sales-chart.svg`
  containing exactly what is on screen (useful for reports)

## Interactions

| Gesture | Effect |
|---|---|
| Hover an element | highlight + tooltip; fires `OnHover` |
| Left click | fires `OnTap` |
| Right click | fires `OnSecondaryTap` (used here for a menu) |
| Mouse wheel | zoom centered at cursor (double-click resets) |
| Drag | pan (only while zoomed) |

## Running

```bash
go run main.go
```

## Notes

- Constructors are identical to the static charts in `examples/35-charts`;
  the interactive behavior and methods are additive.
- Risor v2 does not implicitly mix ints and floats in arithmetic, hence the
  explicit `float(i)` conversions in the sample data generators.
