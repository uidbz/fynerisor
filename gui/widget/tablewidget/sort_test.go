package tablewidget

import "testing"

func columnValues(td *TableData, name string) []string {
	cells := td.GetColumn(name)
	out := make([]string, len(cells))
	for i, c := range cells {
		out[i] = c.StringValue
	}
	return out
}

func TestSortKeepsRowsAligned(t *testing.T) {
	td := NewTableData("t")
	td.AddStringRow([]string{"Name", "Age"}, []string{"Charlie", "30"})
	td.AddStringRow([]string{"Name", "Age"}, []string{"Alice", "25"})
	td.AddStringRow([]string{"Name", "Age"}, []string{"Bob", "40"})

	td.Sort("Name", true)

	want := map[string]string{"Alice": "25", "Bob": "40", "Charlie": "30"}
	names := columnValues(td, "Name")
	ages := columnValues(td, "Age")
	for i, n := range names {
		if ages[i] != want[n] {
			t.Errorf("row %d: %s should pair with %s, got %s", i, n, want[n], ages[i])
		}
	}
}

func TestSortDescending(t *testing.T) {
	td := NewTableData("t")
	td.AddStringRow([]string{"Name"}, []string{"Alice"})
	td.AddStringRow([]string{"Name"}, []string{"Charlie"})
	td.AddStringRow([]string{"Name"}, []string{"Bob"})

	td.Sort("Name", false)
	got := columnValues(td, "Name")
	want := []string{"Charlie", "Bob", "Alice"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("pos %d: want %s got %s", i, want[i], got[i])
		}
	}
}

// Empty cells in the sort column must not scramble paired columns (Bug 1),
// and empties sort last (Bug 2).
func TestSortWithEmptyCellsKeepsAlignment(t *testing.T) {
	td := NewTableData("t")
	td.InsertStringCell("Name", 0, "Charlie")
	td.InsertStringCell("Name", 2, "Alice") // row 1 empty in Name
	td.InsertStringCell("Tag", 0, "charlie-tag")
	td.InsertStringCell("Tag", 1, "empty-tag")
	td.InsertStringCell("Tag", 2, "alice-tag")

	pair := map[string]string{}
	names := columnValues(td, "Name")
	tags := columnValues(td, "Tag")
	for i := range names {
		pair[names[i]] = tags[i]
	}

	td.Sort("Name", true)

	names = columnValues(td, "Name")
	tags = columnValues(td, "Tag")
	for i, n := range names {
		if tags[i] != pair[n] {
			t.Errorf("row scrambled: %q paired with %q, was %q", n, tags[i], pair[n])
		}
	}
	if names[len(names)-1] != "" {
		t.Errorf("empty cell should sort last, got order %v", names)
	}
}

// Comparator must be internally consistent: never Less(i,j) && Less(j,i).
func TestSortLessConsistent(t *testing.T) {
	td := NewTableData("t")
	td.InsertStringCell("Name", 0, "Alice")
	td.InsertStringCell("Name", 2, "Bob") // row 1 empty
	td.Sort("Name", true)                 // pads columns

	ts := td.tableSorter
	for i := 0; i < ts.Len(); i++ {
		for j := 0; j < ts.Len(); j++ {
			if i != j && ts.Less(i, j) && ts.Less(j, i) {
				t.Errorf("inconsistent Less at (%d,%d)", i, j)
			}
		}
	}
}

// Ragged columns must not panic and must sort all rows (Bug 3).
func TestSortRaggedColumns(t *testing.T) {
	td := NewTableData("t")
	td.InsertStringCell("Key", 0, "b")
	td.InsertStringCell("Key", 1, "a")
	td.InsertStringCell("Val", 0, "1")
	td.InsertStringCell("Val", 1, "2")
	td.InsertStringCell("Val", 2, "3") // Val longer than Key

	td.Sort("Key", true)

	if td.rowCount != 3 {
		t.Fatalf("rowCount = %d, want 3", td.rowCount)
	}
	keys := columnValues(td, "Key")
	if len(keys) != 3 {
		t.Fatalf("Key len = %d, want 3 (padded)", len(keys))
	}
	// Non-empty keys ascending, empty last.
	if keys[0] != "a" || keys[1] != "b" || keys[2] != "" {
		t.Errorf("ragged sort wrong order: %v", keys)
	}
}
