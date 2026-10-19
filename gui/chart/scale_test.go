package risorchart

import (
	"math"
	"testing"
)

func TestLinearScaleRoundTrip(t *testing.T) {
	s := newLinearScale(10, 50)
	s.setRange(0, 400)
	for _, v := range []float64{10, 22.5, 30, 49.9} {
		px := s.toPixel(v)
		back := s.fromPixel(px)
		if math.Abs(back-v) > 1e-9 {
			t.Errorf("round trip failed for %v: got %v", v, back)
		}
	}
	if got := s.toPixel(10); got != 0 {
		t.Errorf("domain min should map to range lo, got %v", got)
	}
	if got := s.toPixel(50); got != 400 {
		t.Errorf("domain max should map to range hi, got %v", got)
	}
}

func TestLinearScaleDegenerate(t *testing.T) {
	s := newLinearScale(5, 5)
	if s.min >= s.max {
		t.Fatalf("degenerate domain not widened: %v..%v", s.min, s.max)
	}
	z := newLinearScale(0, 0)
	if z.min != -1 || z.max != 1 {
		t.Errorf("zero domain should widen to -1..1, got %v..%v", z.min, z.max)
	}
}

func TestZoomAtKeepsAnchor(t *testing.T) {
	s := newLinearScale(0, 100)
	s.setRange(0, 400)
	anchor := 60.0
	anchorPx := s.toPixel(anchor)

	s.zoomAt(0.5, anchor) // zoom in 2x

	if w := s.max - s.min; math.Abs(w-50) > 1e-9 {
		t.Errorf("expected domain width 50, got %v", w)
	}
	if got := s.toPixel(anchor); math.Abs(float64(got-anchorPx)) > 0.5 {
		t.Errorf("anchor pixel drifted: %v -> %v", anchorPx, got)
	}
	if s.min < s.baseMin || s.max > s.baseMax {
		t.Errorf("zoomed domain %v..%v outside base %v..%v", s.min, s.max, s.baseMin, s.baseMax)
	}
}

func TestZoomLimitsAndReset(t *testing.T) {
	s := newLinearScale(0, 100)
	s.setRange(0, 400)

	for i := 0; i < 40; i++ {
		s.zoomAt(0.5, 50) // overscroll zoom-in past the 64x limit
	}
	if w := s.max - s.min; w < 100.0/64.0-1e-9 {
		t.Errorf("zoom-in limit exceeded: width %v", w)
	}
	for i := 0; i < 10; i++ {
		s.zoomAt(4, 50) // overscroll zoom-out
	}
	if s.min != 0 || s.max != 100 {
		t.Errorf("zoom-out should clamp to base domain, got %v..%v", s.min, s.max)
	}

	s.zoomAt(0.5, 50)
	if !s.zoomed() {
		t.Error("zoomed() should report true after zoom-in")
	}
	s.reset()
	if s.zoomed() {
		t.Error("reset() should clear zoomed state")
	}
}

func TestPanClampsToBase(t *testing.T) {
	s := newLinearScale(0, 100)
	s.setRange(0, 400)
	s.zoomAt(0.5, 50) // width 50

	s.pan(-1000)
	if s.min != 0 {
		t.Errorf("pan left should clamp min to baseMin, got %v", s.min)
	}
	s.pan(1000)
	if s.max != 100 {
		t.Errorf("pan right should clamp max to baseMax, got %v", s.max)
	}
}

func TestNiceTicks(t *testing.T) {
	cases := []struct {
		min, max float64
		wantStep float64
	}{
		{0, 100, 20},
		{0, 3, 0.5},
		{-7.5, 7.5, 2},
		{0.1, 0.5, 0.1},
	}
	for _, c := range cases {
		tk := niceTicks(c.min, c.max, 6)
		if len(tk) < 2 {
			t.Fatalf("niceTicks(%v,%v): got %d ticks", c.min, c.max, len(tk))
		}
		step := tk[1] - tk[0]
		if math.Abs(step-c.wantStep) > 1e-9 {
			t.Errorf("niceTicks(%v,%v): step %v, want %v (ticks %v)", c.min, c.max, step, c.wantStep, tk)
		}
		for _, v := range tk {
			if v < c.min-1e-9 || v > c.max+1e-9 {
				t.Errorf("tick %v outside domain [%v, %v]", v, c.min, c.max)
			}
		}
	}
}

func TestFormatTick(t *testing.T) {
	cases := map[float64]string{
		0:      "0",
		150:    "150",
		12.5:   "12.5",
		-3.25:  "-3.25",
		2.0000: "2",
	}
	for v, want := range cases {
		if got := formatTick(v); got != want {
			t.Errorf("formatTick(%v) = %q, want %q", v, got, want)
		}
	}
}
