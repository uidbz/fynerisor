package tablewidget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"image/color"

	"fyne.io/fyne/v2/theme"
)

type FlexTable struct {
	widget.BaseWidget
	data             *TableData
	table            *widget.Table
	OnClick          func(cell *TableCell)
	OnClickSecondary func(cell *TableCell)
	selectedRow      int
	minWidth         float32
	scroll           *container.Scroll
	w                fyne.Window
	SelectionColor   color.Color
	CellBgColor      color.Color
	CellBgColorAlt   color.Color
	headerBgColor    color.Color
}

func NewFlexTable(data *TableData, onClick func(cell *TableCell)) *FlexTable {
	table := &FlexTable{
		OnClick:        onClick,
		selectedRow:    -1,
		minWidth:       float32(data.ColumnCount() * 200),
		SelectionColor: theme.Color(theme.ColorNameSelection),
		CellBgColor:    theme.Color(theme.ColorNameBackground),
		CellBgColorAlt: theme.Color(theme.ColorNameInputBackground),
		// headerBgColor:  color.RGBA{89, 89, 89, 255},
		// headerBgColor: color.RGBA{85, 170, 127, 255},
	}

	table.table = widget.NewTable(
		func() (int, int) {
			return table.data.RowCount(), table.data.ColumnCount()
		},
		func() fyne.CanvasObject {
			return NewCell(table)
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {
			cell := o.(*TableCell)
			txt := table.data.Get(id.Col, id.Row)
			if cell.label.Text != txt {
				cell.label.SetText(txt)
			}
			cell.Id = id
			if id.Row == table.selectedRow {
				cell.background.FillColor = table.SelectionColor
			} else {
				if id.Row%2 == 0 {
					cell.background.FillColor = table.CellBgColor
				} else {
					cell.background.FillColor = table.CellBgColorAlt
				}
			}
		})
	table.table.ShowHeaderRow = true

	table.SetData(data)

	table.ExtendBaseWidget(table)

	return table
}

func (t *FlexTable) SetData(data *TableData) {
	t.data = data
	t.table.CreateHeader = func() fyne.CanvasObject {
		return NewHeader("", t.headerBgColor, t)
	}
	t.table.UpdateHeader = func(id widget.TableCellID, o fyne.CanvasObject) {
		h := o.(*Header)
		h.SetText(data.Columns[id.Col])
		h.colId = data.Columns[id.Col]
		h.RefreshSortIcon()
	}
	// for i, _ := range data.Columns {
	// 	t.table.SetColumnWidth(i, 300)
	// }
}

func (t *FlexTable) SetColumnWidth(id int, width float32) {
	t.table.SetColumnWidth(id, width)
}

func (t *FlexTable) Refresh() {
	t.table.Refresh()
}

type tableRenderer struct {
	table   *widget.Table
	objects []fyne.CanvasObject
}

func (tr *tableRenderer) Destroy() {
}

func (tr *tableRenderer) Layout(size fyne.Size) {
	tr.table.Resize(size)
}

func (tr *tableRenderer) MinSize() fyne.Size {
	s := fyne.NewSize(200, 200)
	return s
}

func (tr *tableRenderer) Objects() []fyne.CanvasObject {
	return tr.objects
}

func (tr *tableRenderer) Refresh() {
	canvas.Refresh(tr.table)
}

func (t *FlexTable) CreateRenderer() fyne.WidgetRenderer {
	tr := &tableRenderer{
		table: t.table,
	}
	tr.objects = []fyne.CanvasObject{t.table}
	return tr
}

type TableCell struct {
	widget.BaseWidget
	background *canvas.Rectangle
	label      *widget.Label
	table      *FlexTable
	Id         widget.TableCellID
	isEmpty    bool
	rowCells   []*TableCell
}

func NewCell(table *FlexTable) *TableCell {
	cell := &TableCell{
		background: canvas.NewRectangle(theme.Color(theme.ColorNameBackground)),
		label:      widget.NewLabel(""),
		table:      table,
		isEmpty:    true,
	}
	cell.label.Truncation = fyne.TextTruncateEllipsis

	cell.ExtendBaseWidget(cell)
	return cell
}

func (c *TableCell) CreateRenderer() fyne.WidgetRenderer {
	item := container.NewStack(c.background, c.label)
	return widget.NewSimpleRenderer(item)
}

func (c *TableCell) Tapped(_ *fyne.PointEvent) {
	if c.table.selectedRow == c.Id.Row {
		c.table.selectedRow = -1
	} else {
		c.table.selectedRow = c.Id.Row
	}
	c.table.OnClick(c)
	c.table.Refresh()
}

func (c *TableCell) TappedSecondary(_ *fyne.PointEvent) {
	if c.table.selectedRow == c.Id.Row {
		c.table.selectedRow = -1
	} else {
		c.table.selectedRow = c.Id.Row
	}
	c.table.OnClickSecondary(c)
	c.table.Refresh()
}

func (c *TableCell) ColumnName() string {
	return c.table.data.Columns[c.Id.Col]
}

func (c *TableCell) Text() string {
	return c.label.Text
}

func (c *TableCell) IsEmpty() bool {
	if c.Text() == "" {
		return true
	}
	return false
}

// func (c *TableCell) SetText(text string) {
// 	c.label.Text = text
// 	c.label.SetText()
// 	c.label.Refresh()
// }

type Header struct {
	TableCell
	bgColor color.Color
	icon    *widget.Icon
	colId   string
}

func NewHeader(columnName string, bgColor color.Color, table *FlexTable) *Header {
	header := &Header{}
	header.background = canvas.NewRectangle(bgColor)
	// header.label = canvas.NewText(columnName, color.RGBA{211, 211, 233, 255})
	header.label = widget.NewLabel(columnName)
	header.label.TextStyle.Bold = true
	header.label.Importance = widget.HighImportance
	// header.label.TextSize = theme.TextSize() * 1.2
	header.table = table
	header.bgColor = theme.Color(theme.ColorNameBackground)

	header.ExtendBaseWidget(header)
	return header
}

func (h *Header) SetText(columnName string) {
	h.label.Text = columnName
	h.label.Refresh()
}

func (h *Header) CreateRenderer() fyne.WidgetRenderer {
	h.icon = widget.NewIcon(nil)
	item := container.NewStack(h.background, container.NewBorder(nil, nil, nil, h.icon, h.label))
	return widget.NewSimpleRenderer(item)
}

func (h *Header) RefreshSortIcon() {
	defer h.Refresh()

	if h.table.data.columnToSortBy != h.colId {
		h.icon.Hide()
		return
	}
	if h.table.data.sortAscending {
		h.icon.SetResource(theme.MenuDropDownIcon())
	} else {
		h.icon.SetResource(theme.MenuDropUpIcon())
	}
	h.icon.Show()
}

func (h *Header) Tapped(_ *fyne.PointEvent) {
	h.table.data.sortAscending = !h.table.data.sortAscending
	h.table.data.Sort(h.colId, h.table.data.sortAscending)
	h.table.Refresh()
}
