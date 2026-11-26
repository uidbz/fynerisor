package risorchart

import (
	"context"
	"strings"
	"testing"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// TestParseHLineOpts covers the shared AddHLine option parser directly. It
// runs on the caller's goroutine with no GUI driver involved, which is what
// guarantees validation errors are returned to the script (#65): the
// BarChart variant used to validate inside the asynchronous guithread.Do and
// lose them.
func TestParseHLineOpts(t *testing.T) {
	col, w, err := parseHLineOpts(nil)
	if err != nil || col != colorRed || w != 1.5 {
		t.Fatalf("defaults: got (%v, %v, %v), want (red, 1.5, nil)", col, w, err)
	}
	col, w, err = parseHLineOpts([]object.Object{object.NewString("green")})
	if err != nil || col == colorRed || w != 1.5 {
		t.Fatalf("color only: got (%v, %v, %v)", col, w, err)
	}
	_, w, err = parseHLineOpts([]object.Object{object.NewString("blue"), object.NewInt(3)})
	if err != nil || w != 3 {
		t.Fatalf("explicit width: got (%v, %v)", w, err)
	}
	for _, bound := range []float64{0, 20} {
		if _, w, err = parseHLineOpts([]object.Object{object.NewString("red"), object.NewFloat(bound)}); err != nil || w != float32(bound) {
			t.Errorf("width %v must be accepted, got (%v, %v)", bound, w, err)
		}
	}

	bad := []struct {
		name string
		args []object.Object
		want string
	}{
		{"bad color", []object.Object{object.NewString("bogus")}, "AddHLine: "},
		{"non-number width", []object.Object{object.NewString("red"), object.NewString("wide")}, "AddHLine: line width must be a number (got string)"},
		{"negative width", []object.Object{object.NewString("red"), object.NewFloat(-1)}, "chart error: AddHLine: line width must be in [0, 20] (got -1)"},
		{"huge width", []object.Object{object.NewString("red"), object.NewFloat(21)}, "chart error: AddHLine: line width must be in [0, 20] (got 21)"},
	}
	for _, tc := range bad {
		_, _, err := parseHLineOpts(tc.args)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: got err %v, want containing %q", tc.name, err, tc.want)
		}
	}
}

// hlineCallErr calls AddHLine and returns the surfaced error (Go error or
// Risor error value), or nil.
func hlineCallErr(t *testing.T, obj interface {
	GetAttr(string) (object.Object, bool)
}, args ...object.Object) error {
	t.Helper()
	a, ok := obj.GetAttr("AddHLine")
	if !ok {
		t.Fatal("AddHLine not bound")
	}
	res, err := a.(*object.Builtin).Call(context.Background(), args...)
	if err != nil {
		return err
	}
	if e, ok := res.(*object.Error); ok {
		return e.Value()
	}
	return nil
}

// TestAddHLineRejectsBadOptions checks both charts return every validation
// error and leave the chart untouched.
func TestAddHLineRejectsBadOptions(t *testing.T) {
	bc, err := NewBarChart("monitor", "T2", []string{"s0", "s1"}, []float64{2, 4})
	if err != nil {
		t.Fatal(err)
	}
	defer mountChart(t, bc, 460, 340)()
	lc, err := NewLineChart("monitor", "sample", "T2", []float64{1, 2}, []float64{5, 2})
	if err != nil {
		t.Fatal(err)
	}
	defer mountChart(t, lc, 460, 340)()

	cases := [][]object.Object{
		{object.NewFloat(1), object.NewString("bogus")},
		{object.NewFloat(1), object.NewString("red"), object.NewString("wide")},
		{object.NewFloat(1), object.NewString("red"), object.NewFloat(-1)},
		{object.NewFloat(1), object.NewString("red"), object.NewFloat(25)},
	}
	for i, args := range cases {
		if err := hlineCallErr(t, bc, args...); err == nil {
			t.Errorf("bar case %d: AddHLine must return the validation error", i)
		}
		if err := hlineCallErr(t, lc, args...); err == nil {
			t.Errorf("line case %d: AddHLine must return the validation error", i)
		}
	}
	if len(bc.hLines) != 0 {
		t.Errorf("rejected bar AddHLine calls must not add lines, got %d", len(bc.hLines))
	}
	if len(lc.series) != 1 {
		t.Errorf("rejected line AddHLine calls must not add series, got %d", len(lc.series))
	}

	if err := hlineCallErr(t, bc, object.NewFloat(3), object.NewString("green"), object.NewFloat(4)); err != nil {
		t.Fatalf("valid bar AddHLine: %v", err)
	}
	if len(bc.hLines) != 1 || bc.hLines[0].width != 4 || bc.hLines[0].y != 3 {
		t.Errorf("bar line not applied: %+v", bc.hLines)
	}
}
