package tablewidget

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// makeData builds a small two-column TableData with n rows.
func makeData(n int) *TableData {
	td := NewTableData("test")
	cols := []string{"A", "B"}
	for i := 0; i < n; i++ {
		td.AddStringRow(cols, []string{"a", "b"})
	}
	return td
}

// TestResizeSkipsUpdateCell verifies the core performance fix: a column-resize
// (SetColumnWidth) re-renders visible cells for geometry but must NOT re-run the
// script UpdateCell callback, whereas an explicit Refresh (data change) must.
func TestResizeSkipsUpdateCell(t *testing.T) {
	test.NewApp()

	ft := NewFlexTable(makeData(3), func(c *TableCell) {})

	updateCalls := 0
	ft.SetCreateCell(func(col, row int) fyne.CanvasObject {
		return widget.NewLabel("")
	})
	ft.SetUpdateCell(func(col, row int, obj fyne.CanvasObject) {
		updateCalls++
	})

	// Render the table through a real window so the inner widget.Table drives
	// CreateCell/UpdateCell for visible cells.
	w := test.NewWindow(ft)
	defer w.Close()
	w.Resize(fyne.NewSize(400, 300))

	if updateCalls == 0 {
		t.Fatal("expected UpdateCell to run during initial render, got 0")
	}

	// A column-resize must not trigger any further UpdateCell calls.
	before := updateCalls
	ft.SetColumnWidth(0, 250)
	ft.SetColumnWidth(0, 120)
	ft.SetColumnWidth(1, 200)
	if updateCalls != before {
		t.Fatalf("column resize re-ran UpdateCell %d time(s); expected 0 (geometry-only)", updateCalls-before)
	}

	// An explicit Refresh (e.g. data change / sort) must re-run UpdateCell.
	before = updateCalls
	ft.Refresh()
	if updateCalls <= before {
		t.Fatalf("Refresh did not re-run UpdateCell (before=%d after=%d)", before, updateCalls)
	}
}
