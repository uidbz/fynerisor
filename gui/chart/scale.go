package risorchart

import (
	"fmt"
	"math"
	"strings"
)

// linearScale maps a continuous data domain to a pixel range and supports
// zoom/pan with clamping against a base domain. It is the single source of
// truth for data<->pixel conversion in all interactive charts.
type linearScale struct {
	// current (possibly zoomed/panned) data domain
	min, max float64
	// base (unzoomed) data domain used for clamping and reset
	baseMin, baseMax float64
	// pixel range: lo maps to min, hi maps to max (hi may be < lo, e.g. y axis)
	lo, hi float32
}

func newLinearScale(min, max float64) *linearScale {
	if max < min {
		min, max = max, min
	}
	if max == min {
		// Degenerate domain (single value or empty data): widen around it.
		if min == 0 {
			min, max = -1, 1
		} else {
			min, max = min*0.9, max*1.1
		}
		if max == min {
			min, max = min-1, min+1
		}
	}
	return &linearScale{min: min, max: max, baseMin: min, baseMax: max}
}

// setRange sets the pixel extent for this scale.
func (s *linearScale) setRange(lo, hi float32) {
	s.lo, s.hi = lo, hi
}

// toPixel converts a data value to a pixel coordinate.
func (s *linearScale) toPixel(v float64) float32 {
	if s.max == s.min {
		return s.lo
	}
	t := (v - s.min) / (s.max - s.min)
	return s.lo + float32(t)*float32(s.hi-s.lo)
}

// fromPixel converts a pixel coordinate to a data value.
func (s *linearScale) fromPixel(px float32) float64 {
	if s.hi == s.lo {
		return s.min
	}
	t := (float64(px) - float64(s.lo)) / float64(s.hi-s.lo)
	return s.min + t*(s.max-s.min)
}

// zoomAt multiplies the domain width by factor (factor < 1 zooms in) keeping
// the anchor data value at the same pixel position, then clamps.
func (s *linearScale) zoomAt(factor, anchor float64) {
	if factor <= 0 {
		return
	}
	width := s.max - s.min
	baseWidth := s.baseMax - s.baseMin
	newWidth := width * factor

	// Clamp zoom limits: 1x .. 64x of the base domain.
	if newWidth >= baseWidth {
		s.min, s.max = s.baseMin, s.baseMax
		return
	}
	if newWidth < baseWidth/64 {
		newWidth = baseWidth / 64
	}

	t := (anchor - s.min) / width
	s.min = anchor - t*newWidth
	s.max = s.min + newWidth
	s.clamp()
}

// pan shifts the domain by delta (in data units), clamped to the base domain.
func (s *linearScale) pan(delta float64) {
	width := s.max - s.min
	if width >= s.baseMax-s.baseMin {
		s.min, s.max = s.baseMin, s.baseMax
		return
	}
	s.min += delta
	s.max += delta
	s.clamp()
}

// clamp keeps the current domain inside the base domain.
func (s *linearScale) clamp() {
	width := s.max - s.min
	if s.min < s.baseMin {
		s.min = s.baseMin
		s.max = s.min + width
	}
	if s.max > s.baseMax {
		s.max = s.baseMax
		s.min = s.max - width
	}
}

// reset restores the base domain.
func (s *linearScale) reset() {
	s.min, s.max = s.baseMin, s.baseMax
}

// zoomed reports whether the scale is currently not at its base domain.
func (s *linearScale) zoomed() bool {
	return s.min != s.baseMin || s.max != s.baseMax
}

// niceTicks returns evenly spaced tick positions covering [min,max] using
// 1/2/5 * 10^n steps, targeting about count ticks.
func niceTicks(min, max float64, count int) []float64 {
	if count < 2 {
		count = 2
	}
	span := max - min
	if span <= 0 {
		return []float64{min}
	}

	// Round step candidates: 1, 2, 5, 10, 20, 50 ...
	raw := span / float64(count-1)
	mag := math.Pow(10, math.Floor(math.Log10(raw)))
	norm := raw / mag
	var step float64
	switch {
	case norm < 1.5:
		step = 1
	case norm < 3.5:
		step = 2
	case norm < 7.5:
		step = 5
	default:
		step = 10
	}
	step *= mag

	start := math.Ceil(min/step) * step
	var ticks []float64
	// Guard against pathological step/precision loops.
	for i := 0; i < 200; i++ {
		v := start + float64(i)*step
		if v > max+step*1e-9 {
			break
		}
		ticks = append(ticks, v)
	}
	return ticks
}

// formatTick formats a tick value compactly: integers without decimals,
// otherwise as few decimals as needed.
func formatTick(v float64) string {
	if v == math.Trunc(v) && math.Abs(v) < 1e15 {
		return fmt.Sprintf("%.0f", v)
	}
	s := fmt.Sprintf("%.4f", v)
	for len(s) > 0 && s[len(s)-1] == '0' {
		s = s[:len(s)-1]
	}
	s = strings.TrimSuffix(s, ".")
	return s
}
