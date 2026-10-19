# Chart Examples

This example demonstrates the various chart types available in fynerisor:

- **Bar Chart** - Display categorical data with rectangular bars
- **Line Chart** - Show trends over continuous data
- **Scatter Plot** - Visualize individual data points
- **Histogram** - Show distribution of numerical data with normal distribution overlay
- **Box Plot** - Compare distributions across groups (Tukey boxes with whiskers and outliers)

## Features Demonstrated

- Creating different chart types using the `chart` global object
- Using tabs to organize multiple visualizations
- Setting custom titles and axis labels
- Working with data arrays

## Running

```bash
go run main.go
```

## Chart API

### Bar Chart
```javascript
chart.NewBarChart(title, ylabel, labels, values)
```
- `title` (string): Chart title
- `ylabel` (string): Y-axis label
- `labels` (array of strings): Category labels
- `values` (array of numbers): Bar heights

### Line Chart
```javascript
chart.NewLineChart(title, xlabel, ylabel, xvalues, yvalues)
```
- `title` (string): Chart title
- `xlabel` (string): X-axis label
- `ylabel` (string): Y-axis label
- `xvalues` (array of numbers): X coordinates
- `yvalues` (array of numbers): Y coordinates

### Scatter Chart
```javascript
chart.NewScatterChart(title, xlabel, ylabel, xvalues, yvalues)
```
- `title` (string): Chart title
- `xlabel` (string): X-axis label
- `ylabel` (string): Y-axis label
- `xvalues` (array of numbers): X coordinates
- `yvalues` (array of numbers): Y coordinates

### Histogram
```javascript
chart.NewHistogram(title, xlabel, ylabel, values, bins)
```
- `title` (string): Chart title
- `xlabel` (string): X-axis label
- `ylabel` (string): Y-axis label
- `values` (array of numbers): Data values
- `bins` (integer): Number of bins for grouping data

The histogram automatically overlays a fitted normal distribution curve (red line) for comparison.

### Box Plot
```javascript
chart.NewBoxPlot(title, ylabel, labels, datasets)
```
- `title` (string): Chart title
- `ylabel` (string): Y-axis label
- `labels` (array of strings): Group labels
- `datasets` (array of arrays): Each inner array contains the data for one group

The box plot shows a Tukey summary for each group: IQR box, median line,
whiskers at 1.5*IQR, mean marker, and outlier points.

## Notes

All charts are live widgets rendered from Fyne canvas primitives. Hovering
highlights elements with a tooltip; mouse wheel zooms, dragging pans, and
double-clicking resets the view. Every chart also supports event callbacks
(`OnTap`, `OnHover`, `OnSecondaryTap`), live data updates, and SVG export
(`SaveImage`). See `examples/42-interactive-charts` for those features.

Statistical features:
- **Histogram**: Overlays the normal distribution curve fitted to the data's
  mean and standard deviation (toggle with `ShowNormal(bool)`)
- **Box Plot**: Five-number summary with Tukey whiskers and outlier markers
