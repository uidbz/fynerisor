package builtin

import (
	"fmt"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/gui/widget/tablewidget"
	"github.com/uidbz/fynerisor/scene/render"
)

type tableFixture struct {
	t    *testing.T
	v    *render.View
	w    fyne.Window
	tl   *tableLive
	taps []map[string]any
}

func newTableFixture(t *testing.T) *tableFixture {
	t.Helper()
	a := test.NewApp()
	t.Cleanup(a.Quit)
	v := render.NewView(dataDoc(), render.Options{Registry: NewRegistry()})
	w := test.NewWindow(v)
	t.Cleanup(w.Close)
	w.Resize(fyne.NewSize(1000, 700))
	l, _ := v.Live("tbl")
	f := &tableFixture{t: t, v: v, w: w, tl: render.UnwrapLive(l).(*tableLive)}
	v.On("tbl", "cellTapped", func(evt map[string]any) { f.taps = append(f.taps, evt) })
	return f
}

// rows returns n rows {"v<i>", "w<i>"}.
func rows(n int) [][]string {
	out := make([][]string, n)
	for i := range out {
		out[i] = []string{fmt.Sprintf("v%d", i), fmt.Sprintf("w%d", i)}
	}
	return out
}

// objects walks the table's object tree through widget renderers
// (test.LaidOutObjects does not reach widget.Table's cells).
func (f *tableFixture) objects() []fyne.CanvasObject {
	var out []fyne.CanvasObject
	var walk func(o fyne.CanvasObject)
	walk = func(o fyne.CanvasObject) {
		out = append(out, o)
		switch x := o.(type) {
		case *fyne.Container:
			for _, c := range x.Objects {
				walk(c)
			}
		case fyne.Widget:
			for _, c := range test.WidgetRenderer(x).Objects() {
				walk(c)
			}
		}
	}
	walk(f.tl.tw.Instance)
	return out
}

func (f *tableFixture) button(icon fyne.Resource) *widget.Button {
	f.t.Helper()
	for _, o := range f.objects() {
		if b, ok := o.(*widget.Button); ok && b.Icon != nil && b.Icon.Name() == icon.Name() {
			return b
		}
	}
	f.t.Fatalf("no button with icon %s", icon.Name())
	return nil
}

// tapCell taps the visible cell at displayed (row, col).
func (f *tableFixture) tapCell(row, col int) map[string]any {
	f.t.Helper()
	for _, o := range f.objects() {
		if c, ok := o.(*tablewidget.TableCell); ok && c.Visible() && c.Id.Row == row && c.Id.Col == col {
			before := len(f.taps)
			test.Tap(c)
			if len(f.taps) != before+1 {
				f.t.Fatalf("tap emitted %d events", len(f.taps)-before)
			}
			return f.taps[len(f.taps)-1]
		}
	}
	f.t.Fatalf("no visible cell (%d,%d)", row, col)
	return nil
}

// Shrinking the rows while a later page is shown must not panic (#64.1:
// slice bounds out of range [10:3]) and lands on the last valid page.
func TestTableRowsShrinkOnLaterPage(t *testing.T) {
	f := newTableFixture(t)
	f.tl.SetRows(rows(25))
	test.Tap(f.button(theme.NavigateNextIcon()))
	test.Tap(f.button(theme.NavigateNextIcon()))
	if f.tl.tw.Offset != 20 {
		t.Fatalf("offset after two next = %d, want 20", f.tl.tw.Offset)
	}
	f.tl.SetRows(rows(3)) // panicked
	if f.tl.tw.Offset != 0 {
		t.Fatalf("offset after shrink = %d, want 0", f.tl.tw.Offset)
	}
	if got := f.tapCell(2, 0); got["text"] != "v2" {
		t.Fatalf("tap after shrink = %v", got)
	}
	// An out-of-range offset is clamped by page() itself as well.
	if td := f.tl.page(10, 10); td.RowCount() != 0 {
		t.Fatalf("page(10,10) over 3 rows = %d rows", td.RowCount())
	}
}

// cellTapped reports the displayed text and the absolute dataRow on later
// pages, after filtering and after sorting (#64.2).
func TestTableCellTappedTextAndDataRow(t *testing.T) {
	f := newTableFixture(t)
	f.tl.SetRows(rows(25))

	got := f.tapCell(0, 1)
	if got["text"] != "w0" || got["dataRow"] != 0 || got["row"] != 0 || got["column"] != "B" {
		t.Fatalf("page 1 tap = %v", got)
	}

	test.Tap(f.button(theme.NavigateNextIcon()))
	got = f.tapCell(0, 0)
	if got["text"] != "v10" || got["dataRow"] != 10 {
		t.Fatalf("page 2 tap = %v, want text v10 dataRow 10", got)
	}

	// Filter page 2 down to v15 only: displayed row 0 is source row 15.
	var filter *widget.Entry
	for _, o := range f.objects() {
		if e, ok := o.(*widget.Entry); ok && e.PlaceHolder == "Filter in page" {
			filter = e
		}
	}
	if filter == nil {
		t.Fatal("no filter entry")
	}
	test.Type(filter, "v15")
	got = f.tapCell(0, 1)
	if got["text"] != "w15" || got["dataRow"] != 15 {
		t.Fatalf("filtered tap = %v, want text w15 dataRow 15", got)
	}
	filter.SetText("")

	// Sort page 2 descending by A (tap the header twice: asc, then desc):
	// displayed row 0 is v19.
	var header *tablewidget.Header
	for _, o := range f.objects() {
		if h, ok := o.(*tablewidget.Header); ok && h.Visible() {
			header = h
			break
		}
	}
	if header == nil {
		t.Fatal("no header")
	}
	test.Tap(header)
	test.Tap(header)
	got = f.tapCell(0, 0)
	want := f.tl.rows[got["dataRow"].(int)][0]
	if got["text"] != want {
		t.Fatalf("sorted tap = %v, but rows[dataRow][0] = %q", got, want)
	}
	if got["text"] != "v19" || got["dataRow"] != 19 {
		t.Fatalf("sorted tap = %v, want text v19 dataRow 19", got)
	}
	// dataRow addresses SetCell correctly.
	if !f.tl.SetCell(got["dataRow"].(int), 1, "edited") || f.tl.rows[got["dataRow"].(int)][1] != "edited" {
		t.Fatal("SetCell(dataRow) did not edit the tapped row")
	}
}

// Changing title or pageSize at runtime updates in place: the rows fed with
// SetRows survive and the new values apply (#63.2).
func TestTableTitlePageSizeKeepRows(t *testing.T) {
	f := newTableFixture(t)
	f.tl.SetRows(rows(5))
	if !f.v.RuntimeSet("tbl", "title", "New") {
		t.Fatal("RuntimeSet(title) failed")
	}
	if !f.v.RuntimeSet("tbl", "pageSize", 2.0) {
		t.Fatal("RuntimeSet(pageSize) failed")
	}
	l, _ := f.v.Live("tbl")
	tl := render.UnwrapLive(l).(*tableLive)
	if len(tl.rows) != 5 {
		t.Fatalf("rows after title/pageSize change = %d, want 5", len(tl.rows))
	}
	if tl.tw.Title != "New" {
		t.Fatalf("title = %q", tl.tw.Title)
	}
	if tl.tw.Limit != 2 {
		t.Fatalf("page size = %d, want 2", tl.tw.Limit)
	}
}

// Rows shorter than the column set keep later cells in their own row.
func TestTableShortRowsStayAligned(t *testing.T) {
	f := newTableFixture(t)
	f.tl.SetRows([][]string{{"a", "b"}, {"c"}, {"d", "e"}})
	td := f.tl.page(0, 10)
	if td.RowCount() != 3 || td.Get(1, 1) != "" || td.Get(1, 2) != "e" {
		t.Fatalf("column B = %q %q %q", td.Get(1, 0), td.Get(1, 1), td.Get(1, 2))
	}
}
