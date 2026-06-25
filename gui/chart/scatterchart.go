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

var _ object.Object = &ScatterChart{}

const ScatterChartType object.Type = "chart.scatterchart"

type ScatterChart struct {
	container *fyne.Container
}

func (obj *ScatterChart) CanvasObject() fyne.CanvasObject {
	return obj.container
}

func (r *ScatterChart) Type() object.Type {
	return ScatterChartType
}

func (r *ScatterChart) Inspect() string {
	return "chart.scatterchart"
}

func (r *ScatterChart) Interface() interface{} {
	return r.container
}

func (r *ScatterChart) IsTruthy() bool {
	return true
}

func (r *ScatterChart) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal '" + string(ScatterChartType) + "'")
}

func (r *ScatterChart) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(ScatterChartType))
	obj := object.Errorf("eval error: unsupported operation for %s: %v", ScatterChartType, opType)

	return obj, err
}

func (r *ScatterChart) Equals(other object.Object) bool {
	return r == other
}

func (r *ScatterChart) Attrs() []object.AttrSpec {
	return nil
}

func (r *ScatterChart) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", ScatterChartType, name)
}

func (g *ScatterChart) GetAttr(name string) (object.Object, bool) {
	return nil, false
}

func createScatterChart(title, xlabel, ylabel string, xvalues, yvalues []float64) ([]byte, error) {
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

	scatter, err := plotter.NewScatter(pts)
	if err != nil {
		return nil, fmt.Errorf("could not create scatter: %w", err)
	}
	scatter.GlyphStyle.Color = color.RGBA{R: 70, G: 130, B: 180, A: 255}
	scatter.GlyphStyle.Radius = vg.Points(4)
	scatter.GlyphStyle.Shape = draw.CircleGlyph{}

	p.Add(scatter)

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

// NewScatterChart creates a new scatter plot widget with the given title, axis labels, and x/y values.
// Returns an error if the chart cannot be created (e.g., mismatched x and y lengths).
func NewScatterChart(title, xlabel, ylabel string, xvalues, yvalues []float64) (*ScatterChart, error) {
	if len(xvalues) != len(yvalues) {
		return nil, fmt.Errorf("chart error: x and y values must have same length (got %d x values, %d y values)", len(xvalues), len(yvalues))
	}

	b, err := createScatterChart(title, xlabel, ylabel, xvalues, yvalues)
	if err != nil {
		return nil, fmt.Errorf("create scatter chart: %w", err)
	}

	img := canvas.NewImageFromResource(fyne.NewStaticResource("scatterchart", b))
	img.FillMode = canvas.ImageFillOriginal
	return &ScatterChart{container: container.NewStack(img)}, nil
}
