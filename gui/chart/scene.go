package risorchart

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
)

// sceneKind identifies the primitive type of a scene item.
type sceneKind int

const (
	kindRect sceneKind = iota
	kindLine
	kindText
)

// sceneItem is one drawing primitive in device pixels (widget-local).
// It is intentionally plain data so the same scene can be synced to pooled
// Fyne canvas objects for display and serialized to SVG for export.
type sceneItem struct {
	kind sceneKind

	// Rect: x,y = top-left, w,h = extent.
	// Line: x,y = point 1, w,h = point 2.
	// Text: x,y = box top-left, w,h = box extent (w may be 0 = natural).
	x, y, w, h float32

	col    color.Color // fill for rects/text, stroke for lines
	lineW  float32     // stroke width for lines
	radius float32     // corner radius for rects
	text   string
	fontSz float32
	align  fyne.TextAlign
	bold   bool
}

// scene is an ordered list of primitives; later items draw on top.
type scene struct {
	items []sceneItem
}

func (s *scene) clear() {
	s.items = s.items[:0]
}

func (s *scene) rect(x, y, w, h float32, fill color.Color) {
	s.items = append(s.items, sceneItem{kind: kindRect, x: x, y: y, w: w, h: h, col: fill})
}

func (s *scene) roundRect(x, y, w, h, radius float32, fill color.Color) {
	s.items = append(s.items, sceneItem{kind: kindRect, x: x, y: y, w: w, h: h, col: fill, radius: radius})
}

func (s *scene) line(x1, y1, x2, y2 float32, col color.Color, width float32) {
	s.items = append(s.items, sceneItem{kind: kindLine, x: x1, y: y1, w: x2, h: y2, col: col, lineW: width})
}

func (s *scene) text(x, y, w, h float32, str string, col color.Color, fontSz float32, align fyne.TextAlign) {
	s.items = append(s.items, sceneItem{kind: kindText, x: x, y: y, w: w, h: h, text: str, col: col, fontSz: fontSz, align: align})
}

func (s *scene) textBold(x, y, w, h float32, str string, col color.Color, fontSz float32, align fyne.TextAlign) {
	s.items = append(s.items, sceneItem{kind: kindText, x: x, y: y, w: w, h: h, text: str, col: col, fontSz: fontSz, align: align, bold: true})
}

// plotPainter draws scene items clamped to the plot area so that zoomed or
// panned series never bleed into the axes chrome.
type plotPainter struct {
	s                        *scene
	left, top, right, bottom float32
}

func newPlotPainter(s *scene, left, top, right, bottom float32) *plotPainter {
	return &plotPainter{s: s, left: left, top: top, right: right, bottom: bottom}
}

func (p *plotPainter) rect(x, y, w, h float32, fill color.Color) {
	if w < 0 {
		x, w = x+w, -w
	}
	if h < 0 {
		y, h = y+h, -h
	}
	x0, y0 := mathFMax(x, p.left), mathFMax(y, p.top)
	x1, y1 := mathFMin(x+w, p.right), mathFMin(y+h, p.bottom)
	if x1 <= x0 || y1 <= y0 {
		return
	}
	p.s.rect(x0, y0, x1-x0, y1-y0, fill)
}

func (p *plotPainter) line(x1, y1, x2, y2 float32, col color.Color, width float32) {
	// Liang-Barsky segment clipping to the plot rect.
	dx, dy := x2-x1, y2-y1
	t0, t1 := float32(0), float32(1)
	pairs := [...]struct{ p, q float32 }{
		{-dx, x1 - p.left}, {dx, p.right - x1},
		{-dy, y1 - p.top}, {dy, p.bottom - y1},
	}
	for _, pc := range pairs {
		if pc.p == 0 {
			if pc.q < 0 {
				return
			}
			continue
		}
		t := pc.q / pc.p
		if pc.p < 0 {
			if t > t1 {
				return
			}
			if t > t0 {
				t0 = t
			}
		} else {
			if t < t0 {
				return
			}
			if t < t1 {
				t1 = t
			}
		}
	}
	if t0 > t1 {
		return
	}
	p.s.line(x1+t0*dx, y1+t0*dy, x1+t1*dx, y1+t1*dy, col, width)
}

// polyline draws a connected set of clipped segments from pixel points.
func (p *plotPainter) polyline(xs, ys []float32, col color.Color, width float32) {
	for i := 1; i < len(xs); i++ {
		p.line(xs[i-1], ys[i-1], xs[i], ys[i], col, width)
	}
}

// text draws text only if its anchor lies inside the plot area.
func (p *plotPainter) text(x, y, w, h float32, str string, col color.Color, fontSz float32, align fyne.TextAlign) {
	if x < p.left-2 || x > p.right+2 || y < p.top-2 || y > p.bottom+2 {
		return
	}
	p.s.text(x, y, w, h, str, col, fontSz, align)
}

func (p *plotPainter) contains(x, y float32) bool {
	return x >= p.left && x <= p.right && y >= p.top && y <= p.bottom
}

// --- SVG export -------------------------------------------------------------

func (s *scene) toSVG(width, height float32) string {
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%s" height="%s" viewBox="0 0 %s %s">`,
		fmtFloat(width), fmtFloat(height), fmtFloat(width), fmtFloat(height))
	for _, it := range s.items {
		switch it.kind {
		case kindRect:
			fmt.Fprintf(&b, `<rect x="%s" y="%s" width="%s" height="%s" rx="%s" fill="%s"/>`,
				fmtFloat(it.x), fmtFloat(it.y), fmtFloat(it.w), fmtFloat(it.h),
				fmtFloat(it.radius), colorToSVG(it.col))
		case kindLine:
			fmt.Fprintf(&b, `<line x1="%s" y1="%s" x2="%s" y2="%s" stroke="%s" stroke-width="%s"/>`,
				fmtFloat(it.x), fmtFloat(it.y), fmtFloat(it.w), fmtFloat(it.h),
				colorToSVG(it.col), fmtFloat(maxF(it.lineW, 1)))
		case kindText:
			anchor := "start"
			tx := it.x
			switch it.align {
			case fyne.TextAlignCenter:
				anchor = "middle"
				tx = it.x + it.w/2
			case fyne.TextAlignTrailing:
				anchor = "end"
				tx = it.x + it.w
			}
			baseline := it.y + it.h*0.8
			weight := ""
			if it.bold {
				weight = ` font-weight="bold"`
			}
			fmt.Fprintf(&b, `<text x="%s" y="%s" font-family="sans-serif" font-size="%s" fill="%s" text-anchor="%s"%s>%s</text>`,
				fmtFloat(tx), fmtFloat(baseline), fmtFloat(it.fontSz), colorToSVG(it.col), anchor, weight, xmlEscape(it.text))
		}
	}
	b.WriteString("</svg>")
	return b.String()
}

func xmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&apos;")
	return r.Replace(s)
}

func colorToSVG(c color.Color) string {
	if c == nil {
		return "#000000"
	}
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

func fmtFloat(f float32) string {
	s := strconv.FormatFloat(float64(f), 'f', 2, 32)
	s = strings.TrimRight(s, "0")
	s = strings.TrimSuffix(s, ".")
	return s
}

func maxF(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func mathFMax(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func mathFMin(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
