package risorchart

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/vmexec"
)

// --- hover (desktop.Hoverable) ----------------------------------------------

func (c *chartCore) MouseIn(*desktop.MouseEvent) {}

func (c *chartCore) MouseMoved(e *desktop.MouseEvent) {
	c.hoverAt(e.Position)
}

func (c *chartCore) MouseOut() {
	if c.hoverIdx >= 0 {
		c.hoverIdx = -1
		c.tipText = ""
		c.refresh()
	}
}

// hoverAt hit-tests the pointer and updates highlight + tooltip. The scene is
// only rebuilt when the hovered element changes, so pointer movement inside a
// bar (or over empty space) costs a single integer comparison.
func (c *chartCore) hoverAt(pt fyne.Position) {
	h, ok := c.hit(pt.X, pt.Y)
	if !ok {
		if c.hoverIdx != -1 {
			c.hoverIdx = -1
			c.tipText = ""
			c.refresh()
		}
		return
	}
	if h.Index == c.hoverIdx {
		return
	}
	c.hoverIdx = h.Index
	c.hoverPt = pt
	c.tipText = c.tooltipFor(h)
	if c.onHoverFn != nil {
		c.onHoverFn(h)
	}
	c.refresh()
}

// tooltipFor renders the hover tooltip text for a hit. Chart types may
// override it for richer summaries.
func (c *chartCore) tooltipFor(h hitResult) string {
	if h.Label == "" {
		return h.ValueStr
	}
	return h.Label + ": " + h.ValueStr
}

// --- tap / right-click / double-click ----------------------------------------

func (c *chartCore) Tapped(e *fyne.PointEvent) {
	if c.onTapFn == nil {
		return
	}
	if h, ok := c.hit(e.Position.X, e.Position.Y); ok && c.onTapFn != nil {
		c.onTapFn(h)
	}
}

func (c *chartCore) TappedSecondary(e *fyne.PointEvent) {
	if c.onSecondaryFn == nil {
		return
	}
	h, _ := c.hit(e.Position.X, e.Position.Y)
	c.onSecondaryFn(h)
}

func (c *chartCore) DoubleTapped(*fyne.PointEvent) {
	c.ResetZoom()
}

// --- wheel zoom (fyne.Scrollable) ---------------------------------------------

func (c *chartCore) Scrolled(e *fyne.ScrollEvent) {
	if !c.zoomEnabled {
		return
	}
	dy := float64(e.Scrolled.DY)
	if dy == 0 {
		return
	}
	if math.Abs(dy) > 10 {
		dy /= 100 // platforms reporting pixel-like deltas
	}
	dy = math.Max(-2, math.Min(2, dy))
	factor := math.Pow(1/1.2, dy) // wheel up (dy>0) zooms in

	c.xScale.zoomAt(factor, c.xScale.fromPixel(e.Position.X))
	c.yScale.zoomAt(factor, c.yScale.fromPixel(e.Position.Y))
	c.refresh()
}

// --- drag pan (fyne.Draggable) -----------------------------------------------

func (c *chartCore) Dragged(e *fyne.DragEvent) {
	if !c.zoomEnabled || (!c.xScale.zoomed() && !c.yScale.zoomed()) {
		return
	}
	plotW := math.Abs(float64(c.xScale.hi - c.xScale.lo))
	plotH := math.Abs(float64(c.yScale.hi - c.yScale.lo))
	if plotW < 1 || plotH < 1 {
		return
	}
	c.xScale.pan(-float64(e.Dragged.DX) * (c.xScale.max - c.xScale.min) / plotW)
	c.yScale.pan(float64(e.Dragged.DY) * (c.yScale.max - c.yScale.min) / plotH)
	c.refresh()
}

func (c *chartCore) DragEnd() {}

// ResetZoom restores the full data domain on both axes.
func (c *chartCore) ResetZoom() {
	c.xScale.reset()
	c.yScale.reset()
	c.refresh()
}

// clearHover drops hover state; call after data changes so stale tooltips and
// highlights never reference removed elements.
func (c *chartCore) clearHover() {
	c.hoverIdx = -1
	c.tipText = ""
}

// --- shared Risor method surface ---------------------------------------------
//
// Chart types handle their own data methods in GetAttr and fall back to this
// for the shared interactive verbs.

func (c *chartCore) getAttr(name string) (object.Object, bool) {
	switch name {
	case "OnTap":
		return c.callbackAttr(name, &c.onTapFn), true
	case "OnHover":
		return c.callbackAttr(name, &c.onHoverFn), true
	case "OnSecondaryTap":
		return c.callbackAttr(name, &c.onSecondaryFn), true
	case "ShowGrid":
		return c.boolAttr(name, func(on bool) {
			c.showGrid = on
			c.refresh()
		}), true
	case "ShowValues":
		return c.boolAttr(name, func(on bool) {
			c.showValues = on
			c.refresh()
		}), true
	case "SetZoomEnabled":
		return c.boolAttr(name, func(on bool) {
			c.zoomEnabled = on
			if !on {
				c.ResetZoom()
			}
		}), true
	case "ResetZoom":
		return object.NewBuiltin(c.typeName+"."+name, func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, fmt.Errorf("wrong number of arguments. got=%d, want=0", len(args))
			}
			guithread.Do(c.ResetZoom)
			return object.Nil, nil
		}), true
	case "SetTitle":
		return c.stringAttr(name, func(s string) {
			c.title = s
			c.refresh()
		}), true
	case "WriteSVG":
		return object.NewBuiltin(c.typeName+"."+name, func(ctx context.Context, args ...object.Object) (object.Object, error) {
			w, h, err := c.sizeArgs(name, args)
			if err != nil {
				return nil, err
			}
			var svg string
			guithread.DoAndWait(func() { svg = c.writeSVG(w, h) })
			return object.NewString(svg), nil
		}), true
	case "SaveImage":
		// OffUI: the SVG is rendered on the UI thread, but the file write
		// runs on the script's goroutine so it cannot stall the UI.
		return guithread.OffUI(object.NewBuiltin(c.typeName+"."+name, func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) < 1 || len(args) > 3 {
				return nil, fmt.Errorf("wrong number of arguments. got=%d, want=1..3", len(args))
			}
			path, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			if !strings.HasSuffix(strings.ToLower(path), ".svg") {
				return nil, fmt.Errorf("chart error: SaveImage supports only .svg files (got %q)", path)
			}
			w, h, err := c.sizeArgs(name, args[1:])
			if err != nil {
				return nil, err
			}
			var svg string
			if err := guithread.DoAndWaitContext(ctx, func() { svg = c.writeSVG(w, h) }); err != nil {
				return nil, err
			}
			if err := os.WriteFile(path, []byte(svg), 0o644); err != nil {
				return nil, fmt.Errorf("SaveImage: %w", err)
			}
			return object.Nil, nil
		})), true
	}
	return nil, false
}

func (c *chartCore) sizeArgs(name string, args []object.Object) (float32, float32, error) {
	var w, h float32
	for i, a := range args {
		f, err := object.AsFloat(a)
		if err != nil {
			return 0, 0, fmt.Errorf("%s: expected number for size, got %s", name, a.Type())
		}
		if i == 0 {
			w = float32(f)
		} else {
			h = float32(f)
		}
	}
	return w, h, nil
}

// callbackAttr builds a registration builtin for an (index, label, value)
// callback, following the widget.Button OnTapped pattern.
func (c *chartCore) callbackAttr(name string, slot *func(hitResult)) *object.Builtin {
	return object.NewBuiltin(c.typeName+"."+name, func(ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("wrong number of arguments. got=%d, want=1", len(args))
		}
		cb, err := vmexec.NewCallback(ctx, c.typeName+"."+name, args[0])
		if err != nil {
			return nil, err
		}
		guithread.Do(func() {
			// Event: hover/click callbacks are queued on the script's VM
			// goroutine, never run on the UI thread.
			*slot = func(h hitResult) {
				cb.Fire(
					object.NewInt(int64(h.Index)),
					object.NewString(h.Label),
					object.NewFloat(h.Value),
				)
			}
		})
		return object.Nil, nil
	})
}

func (c *chartCore) boolAttr(name string, apply func(bool)) *object.Builtin {
	return object.NewBuiltin(c.typeName+"."+name, func(ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("wrong number of arguments. got=%d, want=1", len(args))
		}
		on := args[0].IsTruthy()
		guithread.Do(func() { apply(on) })
		return object.Nil, nil
	})
}

func (c *chartCore) stringAttr(name string, apply func(string)) *object.Builtin {
	return object.NewBuiltin(c.typeName+"."+name, func(ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("wrong number of arguments. got=%d, want=1", len(args))
		}
		s, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}
		guithread.Do(func() { apply(s) })
		return object.Nil, nil
	})
}
