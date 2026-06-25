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

var _ object.Object = &LineChart{}

const LineChartType object.Type = "chart.linechart"

type LineChart struct {
	container *fyne.Container
}

func (obj *LineChart) CanvasObject() fyne.CanvasObject {
	return obj.container
}

func (r *LineChart) Type() object.Type {
	return LineChartType
}

func (r *LineChart) Inspect() string {
	return "chart.linechart"
}

func (r *LineChart) Interface() interface{} {
	return r.container
}

func (r *LineChart) IsTruthy() bool {
	return true
}

func (r *LineChart) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal '" + string(LineChartType) + "'")
}

func (r *LineChart) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(LineChartType))
	obj := object.Errorf("eval error: unsupported operation for %s: %v", LineChartType, opType)

	return obj, err
}

func (r *LineChart) Equals(other object.Object) bool {
	return r == other
}

func (r *LineChart) Attrs() []object.AttrSpec {
	return nil
}

func (r *LineChart) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", LineChartType, name)
}

func (g *LineChart) GetAttr(name string) (object.Object, bool) {
	return nil, false
}

func createLineChart(title, xlabel, ylabel string, xvalues, yvalues []float64) ([]byte, error) {
	p := plot.New()
	p.Title.Text = title
	p.Title.TextStyle.Font.Size = 16
	p.X.Label.Text = xlabel
	p.X.Label.TextStyle.Font.Size = 12
	p.Y.Label.Text = ylabel
	p.Y.Label.TextStyle.Font.Size = 12
	p.Add(plotter.NewGrid())

	pts := make(plotter.XYs, len(xvalues))
	for i := range xvalues {
		pts[i].X = xvalues[i]
		pts[i].Y = yvalues[i]
	}

	line, err := plotter.NewLine(pts)
	if err != nil {
		return nil, fmt.Errorf("could not create line: %w", err)
	}
	line.LineStyle.Width = vg.Points(2)
	line.LineStyle.Color = color.RGBA{R: 70, G: 130, B: 180, A: 255}

	p.Add(line)

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

// NewLineChart creates a new line chart widget with the given title, axis labels, and x/y values.
// Returns an error if the chart cannot be created (e.g., mismatched x and y lengths).
func NewLineChart(title, xlabel, ylabel string, xvalues, yvalues []float64) (*LineChart, error) {
	if len(xvalues) != len(yvalues) {
		return nil, fmt.Errorf("chart error: x and y values must have same length (got %d x values, %d y values)", len(xvalues), len(yvalues))
	}

	b, err := createLineChart(title, xlabel, ylabel, xvalues, yvalues)
	if err != nil {
		return nil, fmt.Errorf("create line chart: %w", err)
	}

	img := canvas.NewImageFromResource(fyne.NewStaticResource("linechart", b))
	img.FillMode = canvas.ImageFillOriginal
	return &LineChart{container: container.NewStack(img)}, nil
}
