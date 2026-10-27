package tablewidget

import (
	"sync/atomic"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

// TestBusyCreateCellIsRetried covers issue #50 at the table level: when
// the script's CreateCell cannot answer (its VM was busy), it returns nil
// instead of a fake. The cell must stay empty — the script must never be
// handed a placeholder to update forever — and the real widget must be
// created on a later pass once the callback can answer.
func TestBusyCreateCellIsRetried(t *testing.T) {
	test.NewApp()

	ft := NewFlexTable(makeData(2), func(c *TableCell) {})

	var busy atomic.Bool
	busy.Store(true) // stand-in for a parked VM: CreateCell cannot run
	creates := 0
	seenReal := 0
	seenFake := 0
	ft.SetCreateCell(func(col, row int) fyne.CanvasObject {
		creates++
		if busy.Load() {
			return nil // what widget.Table.CreateCell returns on ErrBusy/ErrStopped
		}
		return widget.NewLabel("real")
	})
	ft.SetUpdateCell(func(col, row int, obj fyne.CanvasObject) {
		if lbl, ok := obj.(*widget.Label); ok && lbl.Text == "real" {
			seenReal++
		} else {
			seenFake++
		}
	})

	w := test.NewWindow(ft)
	defer w.Close()
	w.Resize(fyne.NewSize(400, 300))

	// While CreateCell is silenced, UpdateCell must not run against any
	// stand-in content.
	if seenReal != 0 || seenFake != 0 {
		t.Fatalf("UpdateCell ran while CreateCell was silent: real=%d fake=%d", seenReal, seenFake)
	}
	firstRound := creates

	// The VM frees up; the next refresh must build and update real cells.
	busy.Store(false)
	ft.Refresh()

	if creates <= firstRound {
		t.Fatalf("cells were not recreated after the VM freed up (creates=%d)", creates)
	}
	if seenFake != 0 {
		t.Fatalf("%d UpdateCell calls saw placeholder content instead of the script's widget", seenFake)
	}
	if want := 4; seenReal != want {
		t.Fatalf("UpdateCell saw the real widget in %d cells, want %d", seenReal, want)
	}
}
