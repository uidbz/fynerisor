package risorchart

import (
	"bytes"
	"errors"
	"fmt"
	"image/color"
	"log"

	"gonum.org/v1/plot/vg/vgimg"

	"gonum.org/v1/plot/vg/draw"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"

	"fyne.io/fyne/v2/container"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &BarChart{}

const BarChartType object.Type = "chart.barchart"

type BarChart struct {
	container *fyne.Container
}

func (obj *BarChart) CanvasObject() fyne.CanvasObject {
	return obj.container
}

func (r *BarChart) Type() object.Type {
	return BarChartType
}

func (r *BarChart) Inspect() string {
	return "chart.barchart"
}

func (r *BarChart) Interface() interface{} {
	return r.container
}

func (r *BarChart) IsTruthy() bool {
	return true
}

func (r *BarChart) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal '" + string(BarChartType) + "'")
}

func (r *BarChart) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(BarChartType))
	obj := object.Errorf("eval error: unsupported operation for %s: %v", BarChartType, opType)

	return obj, err
}

func (r *BarChart) Equals(other object.Object) bool {
	return r == other
}

func (r *BarChart) Attrs() []object.AttrSpec {
	return nil
}

func (r *BarChart) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", BarChartType, name)
}

func (g *BarChart) GetAttr(name string) (object.Object, bool) {
	return nil, false
}

func createBarChart(title, ylabel string, labels []string, values []float64) ([]byte, error) {
	p := plot.New()
	p.Title.Text = title
	p.Title.TextStyle.Font.Size = 16
	p.Y.Label.Text = ylabel
	p.Y.Label.TextStyle.Font.Size = 12
	p.Add(plotter.NewGrid())

	w := vg.Points(30)
	bars := make(plotter.Values, len(values))
	for i, v := range values {
		bars[i] = v
	}

	barChart, err := plotter.NewBarChart(bars, w)
	if err != nil {
		log.Fatalf("could not create bar chart: %v", err)
	}
	barChart.LineStyle.Width = vg.Length(0)
	barChart.Color = color.RGBA{R: 70, G: 130, B: 180, A: 255}

	p.Add(barChart)
	p.NominalX(labels...)

	formatFloat := func(f float64) string {
		return fmt.Sprintf("%.1f", f)
	}

	// Add value labels on top of bars
	for i, v := range values {
		label, _ := plotter.NewLabels(plotter.XYLabels{
			XYs:    []plotter.XY{{X: float64(i), Y: v + 0.3}},
			Labels: []string{formatFloat(v)},
		})
		p.Add(label)
	}

	labelCount := font.Length(len(labels))

	// Render to in-memory buffer
	img := vgimg.New(labelCount*vg.Inch, labelCount*0.75*vg.Inch)
	dc := draw.New(img)
	p.Draw(dc)

	var buf bytes.Buffer
	png := vgimg.PngCanvas{Canvas: img}
	if _, err := png.WriteTo(&buf); err != nil {
		return nil, fmt.Errorf("could not write PNG to buffer: %w", err)
	}

	return buf.Bytes(), nil
}

// NewBarChart creates a new bar chart widget with the given title, y-axis label, x-axis labels, and values.
// Returns an error if the chart cannot be created (e.g., mismatched labels and values lengths).
func NewBarChart(title, ylabel string, labels []string, values []float64) (*BarChart, error) {
	if len(labels) != len(values) {
		return nil, fmt.Errorf("chart error: labels and values must have same length (got %d labels, %d values)", len(labels), len(values))
	}

	b, err := createBarChart(title, ylabel, labels, values)
	if err != nil {
		return nil, fmt.Errorf("create bar chart: %w", err)
	}

	img := canvas.NewImageFromResource(fyne.NewStaticResource("barchart", b))
	img.FillMode = canvas.ImageFillOriginal
	return &BarChart{container: container.NewStack(img)}, nil
}
