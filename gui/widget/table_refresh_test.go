package widget

import "testing"

// issue #58: widget.Table must keep its own Refresh (which re-runs the data
// getters), not the generic canvas verb (which only redraws). A dynamic table
// whose Refresh attribute resolved to the generic verb can never show new
// rows — this is what broke goto's lars-server tables.
func TestTableRefreshIsBespoke(t *testing.T) {
	count := 0
	for _, spec := range tableAttrs.Specs() {
		if spec.Name == "Refresh" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("widget.Table has %d Refresh registrations, want exactly 1", count)
	}
	for _, spec := range tableAttrs.Specs() {
		if spec.Name == "Refresh" && spec.Doc != "Reload the table: re-runs Columns, RowCount, Data and the cell callbacks, then redraws." {
			t.Fatalf("widget.Table.Refresh is the generic verb, not the bespoke one: doc=%q", spec.Doc)
		}
	}
}
