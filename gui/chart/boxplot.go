package risorchart

import (
	"bytes"
	"errors"
	"fmt"
	"image/color"
	"math"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &BoxPlot{}

const BoxPlotType object.Type = "chart.boxplot"

type BoxPlot struct {
	container *fyne.Container
}

func (obj *BoxPlot) CanvasObject() fyne.CanvasObject {
	return obj.container
}

func (r *BoxPlot) Type() object.Type {
	return BoxPlotType
}

func (r *BoxPlot) Inspect() string {
	return "chart.boxplot"
}

func (r *BoxPlot) Interface() interface{} {
	return r.container
}

func (r *BoxPlot) IsTruthy() bool {
	return true
}

func (r *BoxPlot) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal '" + string(BoxPlotType) + "'")
}

func (r *BoxPlot) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(BoxPlotType))
	obj := object.Errorf("eval error: unsupported operation for %s: %v", BoxPlotType, opType)

	return obj, err
}

func (r *BoxPlot) Equals(other object.Object) bool {
	return r == other
}

func (r *BoxPlot) Attrs() []object.AttrSpec {
	return nil
}

func (r *BoxPlot) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", BoxPlotType, name)
}

func (g *BoxPlot) GetAttr(name string) (object.Object, bool) {
	return nil, false
}

func createBoxPlot(title, ylabel string, labels []string, datasets [][]float64) ([]byte, error) {
	p := plot.New()
	p.Title.Text = title
	p.Title.TextStyle.Font.Size = 16
	p.Y.Label.Text = ylabel
	p.Y.Label.TextStyle.Font.Size = 12

	// Colors for box plots and curves
	colors := []color.RGBA{
		{R: 70, G: 130, B: 180, A: 255},  // Steel blue
		{R: 220, G: 20, B: 60, A: 255},   // Crimson
		{R: 50, G: 205, B: 50, A: 255},   // Lime green
		{R: 255, G: 140, B: 0, A: 255},   // Dark orange
		{R: 138, G: 43, B: 226, A: 255},  // Blue violet
	}

	// Find global min/max for scaling the normal distribution
	globalMin, globalMax := math.Inf(1), math.Inf(-1)
	for _, data := range datasets {
		for _, v := range data {
			if v < globalMin {
				globalMin = v
			}
			if v > globalMax {
				globalMax = v
			}
		}
	}

	// Create box plots and normal distribution curves
	for i, data := range datasets {
		if len(data) == 0 {
			continue
		}

		// Create box plot
		values := make(plotter.Values, len(data))
		copy(values, data)

		boxPlot, err := plotter.NewBoxPlot(vg.Points(40), float64(i), values)
		if err != nil {
			return nil, fmt.Errorf("could not create box plot: %w", err)
		}

		colorIdx := i % len(colors)
		fillColor := colors[colorIdx]
		fillColor.A = 100 // Semi-transparent
		boxPlot.FillColor = fillColor

		p.Add(boxPlot)

		// Calculate mean and standard deviation
		mean, stddev := calculateStats(data)

		// Create normal distribution curve
		// Scale the curve to fit next to the box plot
		numPoints := 100
		xOffset := float64(i)
		curveWidth := 0.35 // Half-width of curve relative to box position

		pts := make(plotter.XYs, numPoints)
		yRange := globalMax - globalMin
		yStep := yRange / float64(numPoints-1)

		// Find max PDF value to scale the curve
		maxPDF := normalPDF(mean, mean, stddev)
		if maxPDF == 0 {
			maxPDF = 1 // Avoid division by zero
		}

		for j := 0; j < numPoints; j++ {
			y := globalMin + float64(j)*yStep
			pdfVal := normalPDF(y, mean, stddev)
			// Scale and position the curve
			xScaled := xOffset + (pdfVal/maxPDF)*curveWidth
			pts[j].X = xScaled
			pts[j].Y = y
		}

		// Create line for normal distribution
		line, err := plotter.NewLine(pts)
		if err != nil {
			return nil, fmt.Errorf("could not create normal curve: %w", err)
		}
		line.LineStyle.Color = colors[colorIdx]
		line.LineStyle.Width = vg.Points(2)

		p.Add(line)
	}

	// Set nominal X labels
	p.NominalX(labels...)

	// Fixed resolution for consistent chart quality
	img := vgimg.New(8*vg.Inch, 6*vg.Inch)
	dc := draw.New(img)
	p.Draw(dc)

	var buf bytes.Buffer
	png := vgimg.PngCanvas{Canvas: img}
	if _, err := png.WriteTo(&buf); err != nil {
		return nil, fmt.Errorf("could not write PNG to buffer: %w", err)
	}

	return buf.Bytes(), nil
}

// NewBoxPlot creates a new box plot widget with optional normal distribution overlays.
// Each dataset is shown as a box plot with a normal distribution curve fitted to the data.
// Returns an error if the chart cannot be created.
func NewBoxPlot(title, ylabel string, labels []string, datasets [][]float64) (*BoxPlot, error) {
	if len(labels) != len(datasets) {
		return nil, fmt.Errorf("chart error: labels and datasets must have same length (got %d labels, %d datasets)", len(labels), len(datasets))
	}

	b, err := createBoxPlot(title, ylabel, labels, datasets)
	if err != nil {
		return nil, fmt.Errorf("create box plot: %w", err)
	}

	img := canvas.NewImageFromResource(fyne.NewStaticResource("boxplot", b))
	img.FillMode = canvas.ImageFillOriginal
	return &BoxPlot{container: container.NewStack(img)}, nil
}
