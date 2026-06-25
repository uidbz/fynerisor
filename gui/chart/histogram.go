package risorchart

import (
	"bytes"
	"errors"
	"fmt"
	"image/color"

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

var _ object.Object = &Histogram{}

const HistogramType object.Type = "chart.histogram"

type Histogram struct {
	container *fyne.Container
}

func (obj *Histogram) CanvasObject() fyne.CanvasObject {
	return obj.container
}

func (r *Histogram) Type() object.Type {
	return HistogramType
}

func (r *Histogram) Inspect() string {
	return "chart.histogram"
}

func (r *Histogram) Interface() interface{} {
	return r.container
}

func (r *Histogram) IsTruthy() bool {
	return true
}

func (r *Histogram) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal '" + string(HistogramType) + "'")
}

func (r *Histogram) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(HistogramType))
	obj := object.Errorf("eval error: unsupported operation for %s: %v", HistogramType, opType)

	return obj, err
}

func (r *Histogram) Equals(other object.Object) bool {
	return r == other
}

func (r *Histogram) Attrs() []object.AttrSpec {
	return nil
}

func (r *Histogram) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", HistogramType, name)
}

func (g *Histogram) GetAttr(name string) (object.Object, bool) {
	return nil, false
}

func createHistogram(title, xlabel, ylabel string, values []float64, bins int) ([]byte, error) {
	p := plot.New()
	p.Title.Text = title
	p.Title.TextStyle.Font.Size = 16
	p.X.Label.Text = xlabel
	p.X.Label.TextStyle.Font.Size = 12
	p.Y.Label.Text = ylabel
	p.Y.Label.TextStyle.Font.Size = 12

	v := make(plotter.Values, len(values))
	for i, val := range values {
		v[i] = val
	}

	hist, err := plotter.NewHist(v, bins)
	if err != nil {
		return nil, fmt.Errorf("could not create histogram: %w", err)
	}
	hist.FillColor = color.RGBA{R: 70, G: 130, B: 180, A: 180}

	p.Add(hist)

	// Calculate mean and standard deviation from the data
	mean, stddev := calculateStats(values)

	// Create normal distribution curve overlay
	// Scale the curve to match the histogram's total area
	min, max, _, _ := hist.DataRange()
	binWidth := (max - min) / float64(bins)
	totalArea := float64(len(values)) * binWidth // Total area of histogram

	// Create points for the normal distribution curve
	numPoints := 200
	pts := make(plotter.XYs, numPoints)
	xStep := (max - min) / float64(numPoints-1)

	for i := 0; i < numPoints; i++ {
		x := min + float64(i)*xStep
		// PDF scaled by total area to match histogram height
		y := normalPDF(x, mean, stddev) * totalArea
		pts[i].X = x
		pts[i].Y = y
	}

	// Create line for normal distribution
	line, err := plotter.NewLine(pts)
	if err != nil {
		return nil, fmt.Errorf("could not create normal curve: %w", err)
	}
	line.LineStyle.Color = color.RGBA{R: 220, G: 20, B: 60, A: 255} // Crimson
	line.LineStyle.Width = vg.Points(2.5)

	p.Add(line)

	// Add legend
	p.Legend.Add("Data", hist)
	p.Legend.Add("Normal Distribution", line)
	p.Legend.Top = true
	p.Legend.Left = true

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

// NewHistogram creates a new histogram widget with the given title, axis labels, values, and number of bins.
// Returns an error if the chart cannot be created.
func NewHistogram(title, xlabel, ylabel string, values []float64, bins int) (*Histogram, error) {
	if bins <= 0 {
		return nil, fmt.Errorf("chart error: bins must be greater than 0 (got %d)", bins)
	}

	b, err := createHistogram(title, xlabel, ylabel, values, bins)
	if err != nil {
		return nil, fmt.Errorf("create histogram: %w", err)
	}

	img := canvas.NewImageFromResource(fyne.NewStaticResource("histogram", b))
	img.FillMode = canvas.ImageFillOriginal
	return &Histogram{container: container.NewStack(img)}, nil
}
