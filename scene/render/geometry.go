package render

import "fyne.io/fyne/v2"

// Rect is a rectangle in local coordinates. Fyne has no rectangle value type
// of its own, and the designer needs one for hit-testing and chrome.
type Rect struct {
	X, Y, W, H float32
}

// NewRect returns a rectangle.
func NewRect(x, y, w, h float32) Rect { return Rect{X: x, Y: y, W: w, H: h} }

// Position returns the top-left corner.
func (r Rect) Position() fyne.Position { return fyne.NewPos(r.X, r.Y) }

// Size returns the rectangle size.
func (r Rect) Size() fyne.Size { return fyne.NewSize(r.W, r.H) }

// Width returns the width.
func (r Rect) Width() float32 { return r.W }

// Height returns the height.
func (r Rect) Height() float32 { return r.H }

// Center returns the center point.
func (r Rect) Center() fyne.Position { return fyne.NewPos(r.X+r.W/2, r.Y+r.H/2) }

// Contains reports whether p lies inside the rectangle (inclusive edges).
func (r Rect) Contains(p fyne.Position) bool {
	return p.X >= r.X && p.X <= r.X+r.W && p.Y >= r.Y && p.Y <= r.Y+r.H
}

// Intersects reports whether two rectangles overlap.
func (r Rect) Intersects(o Rect) bool {
	return r.X < o.X+o.W && o.X < r.X+r.W && r.Y < o.Y+o.H && o.Y < r.Y+r.H
}

// Union returns the bounding rectangle of r and o.
func (r Rect) Union(o Rect) Rect {
	x1 := min32(r.X, o.X)
	y1 := min32(r.Y, o.Y)
	x2 := max32(r.X+r.W, o.X+o.W)
	y2 := max32(r.Y+r.H, o.Y+o.H)
	return Rect{X: x1, Y: y1, W: x2 - x1, H: y2 - y1}
}

// Intersection returns the overlapping rectangle, if any.
func (r Rect) Intersection(o Rect) (Rect, bool) {
	x1 := max32(r.X, o.X)
	y1 := max32(r.Y, o.Y)
	x2 := min32(r.X+r.W, o.X+o.W)
	y2 := min32(r.Y+r.H, o.Y+o.H)
	if x2 <= x1 || y2 <= y1 {
		return Rect{}, false
	}
	return Rect{X: x1, Y: y1, W: x2 - x1, H: y2 - y1}, true
}

// Inset shrinks the rectangle by d on each side.
func (r Rect) Inset(d float32) Rect {
	return Rect{X: r.X + d, Y: r.Y + d, W: r.W - 2*d, H: r.H - 2*d}
}

func min32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func max32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
