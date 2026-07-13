package tablewidget

import (
	"errors"
	"sort"
	"strconv"
)

type CellType int

const (
	cellIsEmpty = iota
	cellIsString
	cellIsInt
	cellIsFloat
)

type TableData struct {
	TableName      string
	TableId        string
	data           map[string][]DataCell // map key is column, slice is rows
	Columns        []string              // Used to save the order of columns
	rowCount       int
	RowIds         []string
	RowCategory    string
	columnToSortBy string
	sortAscending  bool
	tableSorter    *tableSorter
	RowMapping     []int // Maps filtered row index to original row index (for filtering)
}

// First ID in list must be the tableId, following 1 for each row
func (td *TableData) SetIds(ids []string) error {
	if len(ids) != td.rowCount+1 {
		return errors.New("SetIDs needs exactly:" +
			strconv.Itoa(td.rowCount+1) + " IDs for this table (1 for tableId and 1 for each row). " +
			"Got " + strconv.Itoa(len(ids)) + " IDs")
	}

	td.TableId = ids[0]
	td.RowIds = ids[1:]

	return nil
}

type tableSorter struct {
	tableData *TableData
}

type DataCell struct {
	cellType    CellType
	StringValue string
	Row         int
}

func NewTableData(tableName string) *TableData {
	table := &TableData{
		TableName: tableName,
		data:      make(map[string][]DataCell, 0),
		rowCount:  0,
	}
	table.tableSorter = &tableSorter{table}

	return table
}

func (td *TableData) AddColumnFromTable(newColumnName, oldColumnName string, otherTable *TableData) error {
	td.Columns = append(td.Columns, newColumnName)
	if col, ok := otherTable.data[oldColumnName]; !ok {
		return errors.New("Old column not found")
	} else {
		td.data[newColumnName] = col
		if td.rowCount < len(col) {
			td.rowCount = len(col)
		}
		td.RowIds = otherTable.RowIds
		td.RowCategory = otherTable.RowCategory
	}
	return nil
}

func (td *TableData) RowCount() int {
	return td.rowCount
}

func (td *TableData) ColumnCount() int {
	return len(td.Columns)
}

// Remove rows and columns
func (td *TableData) Clear() {
	td.data = make(map[string][]DataCell, 0)
	td.RowIds = make([]string, 0)
	td.Columns = make([]string, 0)
	td.rowCount = 0
}

func (td *TableData) AddStringRow(columns []string, row []string) {
	for i, c := range columns {
		td.AddStringCell(c, row[i])
	}
}

func (td *TableData) AddStringCell(column string, value string) {
	if _, ok := td.data[column]; !ok {
		td.Columns = append(td.Columns, column)
	}
	newCell := DataCell{
		cellType:    cellIsString,
		StringValue: value,
		Row:         len(td.data[column]),
	}
	td.data[column] = append(td.data[column], newCell)
	if len(td.data[column]) > td.rowCount {
		td.rowCount = len(td.data[column])
	}
}

func (td *TableData) InsertStringCell(column string, row int, value string) {
	if _, ok := td.data[column]; !ok {
		td.Columns = append(td.Columns, column)
	}
	for len(td.data[column]) <= row {
		td.data[column] = append(td.data[column], DataCell{cellType: cellIsEmpty})
	}
	td.data[column][row] = DataCell{
		cellType:    cellIsString,
		StringValue: value,
		Row:         row,
	}
	if len(td.data[column]) > td.rowCount {
		td.rowCount = len(td.data[column])
	}
}

func (td *TableData) GetRows(column string) []DataCell {
	if rows, ok := td.data[column]; ok {
		return rows
	} else {
		return make([]DataCell, 0)
	}
}

func (td *TableData) Get(col, row int) string {
	if col < len(td.Columns) {
		return td.GetFromColumn(td.Columns[col], row)
	}
	return ""
}

func (td *TableData) GetColumn(columnName string) []DataCell {
	if _, ok := td.data[columnName]; ok {
		return td.data[columnName]
	}
	return []DataCell{}
}

func (td *TableData) GetFromColumn(column string, row int) string {
	if _, ok := td.data[column]; ok {
		if row < len(td.data[column]) {
			return td.data[column][row].StringValue
		}
	}
	return ""
}

func (td *TableData) RenameColumn(oldName, newName string) {
	found := false
	for i, x := range td.Columns {
		if x == oldName {
			td.Columns[i] = newName
			found = true
			break
		}
	}
	if !found {
		return
	}

	td.data[newName] = td.data[oldName]
	delete(td.data, oldName)
}

func (td *TableData) Sort(column string, ascending bool) {
	td.columnToSortBy = column
	td.sortAscending = ascending
	// Pad every column to rowCount so row i is a complete, aligned tuple across
	// all columns. Without this, a shorter column would make the sorter index
	// out of bounds or sort only a prefix of the rows.
	for _, x := range td.Columns {
		for len(td.data[x]) < td.rowCount {
			td.data[x] = append(td.data[x], DataCell{cellType: cellIsEmpty})
		}
	}
	sort.Sort(td.tableSorter)
}

func (ts *tableSorter) Len() int {
	return ts.tableData.rowCount
}

func (ts *tableSorter) Swap(i, j int) {
	// Swap every column at once so each row stays an aligned tuple. Swapping the
	// value and type together across ALL columns (including the sort column) is
	// what keeps paired columns in step; the positional Row field stays put.
	for _, x := range ts.tableData.Columns {
		col := ts.tableData.data[x]
		col[i].StringValue, col[j].StringValue = col[j].StringValue, col[i].StringValue
		col[i].cellType, col[j].cellType = col[j].cellType, col[i].cellType
	}
}

func (ts *tableSorter) Less(i, j int) bool {
	td := ts.tableData
	a := td.data[td.columnToSortBy][i]
	b := td.data[td.columnToSortBy][j]

	// Empty cells always sort last, independent of direction. The comparator
	// must be consistent: sort.Sort corrupts data if Less(i,j) and Less(j,i)
	// can both be true.
	if a.cellType == cellIsEmpty || b.cellType == cellIsEmpty {
		return b.cellType == cellIsEmpty && a.cellType != cellIsEmpty
	}

	if td.sortAscending {
		return a.StringValue < b.StringValue
	}
	return a.StringValue > b.StringValue
}
