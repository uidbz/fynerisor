package tablewidget

import (
	"reflect"
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// levelsData builds the two-level example table: Sample plus two replicates under
// each of two temperatures, with the merged parent cells forward-filled the way a
// converter must emit them.
func levelsData() (*TableData, [][]string) {
	cols := []string{"Sample", "Temperature (20°C)\x1fReplicate 1", "Temperature (20°C)\x1fReplicate 2",
		"Temperature (37°C)\x1fReplicate 1", "Temperature (37°C)\x1fReplicate 2"}
	td := NewTableData("levels")
	td.AddStringRow(cols, []string{"S1", "4.2", "4.4", "5.9", "6.1"})
	td.AddStringRow(cols, []string{"S2", "4.1", "4.3", "5.8", "6.0"})
	rows := [][]string{
		{"Sample", "Temperature (20°C)", "Temperature (20°C)", "Temperature (37°C)", "Temperature (37°C)"},
		{"", "Replicate 1", "Replicate 2", "Replicate 1", "Replicate 2"},
	}
	return td, rows
}

// renderedHeader forces the header's renderer to exist, which is what creates the
// sort icon. Fyne does this by refreshing the header container before it calls
// UpdateHeader; a test driving UpdateHeader directly has to do it itself.
func renderedHeader(h *Header) *Header {
	test.WidgetRenderer(h)
	return h
}

func TestSpanHeaderRuns(t *testing.T) {
	_, rows := levelsData()
	got := spanHeaderRuns(rows)
	want := [][]string{
		// The forward-filled parent is drawn once per run, at its first column.
		{"Sample", "Temperature (20°C)", "", "Temperature (37°C)", ""},
		// Replicate 1/2 repeat under both parents, but the run broke above them.
		{"", "Replicate 1", "Replicate 2", "Replicate 1", "Replicate 2"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("spanHeaderRuns:\ngot  %q\nwant %q", got, want)
	}
}

func TestSpanHeaderRunsPadsRaggedRows(t *testing.T) {
	got := spanHeaderRuns([][]string{{"A", "A", "B"}, {"x"}})
	want := [][]string{{"A", "", "B"}, {"x", "", ""}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("spanHeaderRuns:\ngot  %q\nwant %q", got, want)
	}
	if spanHeaderRuns(nil) != nil {
		t.Fatal("spanHeaderRuns(nil) should be nil, so the single-row default applies")
	}
}

// TestHeaderLabels covers what each header cell draws, including the two fallbacks:
// no levels at all, and a levels grid narrower than the data.
func TestHeaderLabels(t *testing.T) {
	test.NewApp()
	td, rows := levelsData()
	ft := NewFlexTable(td, func(c *TableCell) {})

	if got := ft.headerLabels(1); !reflect.DeepEqual(got, []string{td.Columns[1]}) {
		t.Fatalf("without levels a column should show its own name, got %q", got)
	}

	ft.SetHeaderLevels(rows)
	for col, want := range [][]string{
		{"Sample", ""},
		{"Temperature (20°C)", "Replicate 1"},
		{"", "Replicate 2"},
		{"Temperature (37°C)", "Replicate 1"},
		{"", "Replicate 2"},
	} {
		if got := ft.headerLabels(col); !reflect.DeepEqual(got, want) {
			t.Errorf("headerLabels(%d) = %q, want %q", col, got, want)
		}
	}

	// A grid out of step with the data must still name the columns it misses.
	ft.SetHeaderLevels(rows[:1])
	if got := ft.headerLabels(9); !reflect.DeepEqual(got, []string{""}) {
		t.Errorf("headerLabels past both grid and data = %q, want one empty level", got)
	}
}

// TestHeaderCellsStackLevels drives the header the way Fyne does — a template from
// CreateHeader, then UpdateHeader per column — and checks the labels that land on
// each level, that the header row grows taller, and that the column identity used
// for sorting is still the canonical column key rather than the leaf label.
func TestHeaderCellsStackLevels(t *testing.T) {
	test.NewApp()
	td, rows := levelsData()
	ft := NewFlexTable(td, func(c *TableCell) {})

	flatHeight := ft.table.CreateHeader().MinSize().Height
	ft.SetHeaderLevels(rows)
	if h := ft.table.CreateHeader().MinSize().Height; h <= flatHeight {
		t.Fatalf("two-level header template height %v should exceed the one-level %v", h, flatHeight)
	}

	header := renderedHeader(ft.table.CreateHeader().(*Header))
	ft.table.UpdateHeader(widget.TableCellID{Row: -1, Col: 2}, header)
	if len(header.levels) != 2 {
		t.Fatalf("header has %d level labels, want 2", len(header.levels))
	}
	if header.levels[0].Text != "" || header.levels[1].Text != "Replicate 2" {
		t.Errorf("header levels = %q/%q, want \"\"/\"Replicate 2\"", header.levels[0].Text, header.levels[1].Text)
	}
	if header.colId != td.Columns[2] {
		t.Errorf("header sorts on %q, want the column key %q", header.colId, td.Columns[2])
	}

	// Fyne pools header cells, so a cell built when the table was flat gets reused
	// for a hierarchical one and has to grow its stack.
	reused := renderedHeader(NewHeader("Sample", nil, ft))
	ft.table.UpdateHeader(widget.TableCellID{Row: -1, Col: 1}, reused)
	if len(reused.levels) != 2 || reused.levels[0].Text != "Temperature (20°C)" {
		t.Errorf("reused header did not grow: %d levels, top %q", len(reused.levels), reused.levels[0].Text)
	}

	// And back: dropping the levels returns the cell to a single label.
	ft.SetHeaderLevels(nil)
	ft.table.UpdateHeader(widget.TableCellID{Row: -1, Col: 1}, reused)
	if len(reused.levels) != 1 || reused.levels[0].Text != td.Columns[1] {
		t.Errorf("reused header did not shrink: %d levels, top %q", len(reused.levels), reused.levels[0].Text)
	}
}

// TestHeaderTapSortsUnderLevels guards the interaction a multi-row header could
// plausibly break: the sort key is the column, not the header text.
func TestHeaderTapSortsUnderLevels(t *testing.T) {
	test.NewApp()
	td, rows := levelsData()
	ft := NewFlexTable(td, func(c *TableCell) {})
	ft.SetHeaderLevels(rows)

	header := renderedHeader(ft.table.CreateHeader().(*Header))
	ft.table.UpdateHeader(widget.TableCellID{Row: -1, Col: 1}, header)
	header.Tapped(nil)

	if td.columnToSortBy != td.Columns[1] {
		t.Fatalf("sorted by %q, want %q", td.columnToSortBy, td.Columns[1])
	}
	if got := td.Get(1, 0); got != "4.1" {
		t.Fatalf("first row of the sorted column = %q, want the ascending minimum 4.1", got)
	}
}
