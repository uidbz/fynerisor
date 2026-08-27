package tablewidget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"image/color"

	"fyne.io/fyne/v2/theme"
)

// stripeAltColor returns the alternate-row background for the current theme.
// It blends the theme background toward the foreground by a small factor so the
// stripe is always visible in both light and dark variants (unlike a fixed theme
// color name such as OverlayBackground, which equals Background in light mode).
func stripeAltColor() color.Color {
	bg := theme.Color(theme.ColorNameBackground)
	fg := theme.Color(theme.ColorNameForeground)
	return blendColor(bg, fg, 0.07)
}

func blendColor(base, over color.Color, t float64) color.NRGBA {
	br, bg, bb, ba := base.RGBA()
	or, og, ob, _ := over.RGBA()
	lerp := func(a, b uint32) uint8 {
		return uint8(float64(a>>8)*(1-t) + float64(b>>8)*t)
	}
	return color.NRGBA{R: lerp(br, or), G: lerp(bg, og), B: lerp(bb, ob), A: uint8(ba >> 8)}
}

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

	// Widget mode support
	widgetMode     bool
	createCellFunc func(col, row int) fyne.CanvasObject
	updateCellFunc func(col, row int, obj fyne.CanvasObject)

	// dataGen bumps on every SetData (a real data change). A WidgetCell records
	// the generation it last ran UpdateCell at; a geometry-only refresh (e.g. a
	// column-resize drag, which forces a full re-render of every visible cell)
	// then skips the per-cell UpdateCell VM round-trip because the content is
	// unchanged. This is the difference between fluid and choppy column resizing.
	dataGen int
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
			if table.widgetMode && table.createCellFunc != nil {
				// Widget mode: create a wrapper that can be replaced
				return NewWidgetCell()
			}
			// String mode: use Label cell (current behavior)
			return NewCell(table)
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {
			if table.widgetMode && table.updateCellFunc != nil {
				// Widget mode: Replace the widget in the cell
				if widgetCell, ok := o.(*WidgetCell); ok {
					// Check if we need to recreate the widget
					needsRecreation := widgetCell.currentID.Col != id.Col ||
						widgetCell.currentID.Row != id.Row

					if needsRecreation {
						cellWidget := table.createCellFunc(id.Col, id.Row)
						widgetCell.SetWidget(cellWidget, id)
					}
					// Skip the UpdateCell VM round-trip when this cell already
					// reflects the current data at its current position. A
					// geometry-only refresh (column-resize drag) re-renders every
					// visible cell but changes no content, so re-running the
					// script callback per cell per drag-pixel is pure waste and is
					// what makes resizing choppy. Recreation or a new data
					// generation forces a real update.
					if widgetCell.content != nil && (needsRecreation || widgetCell.lastGen != table.dataGen) {
						// Map filtered row to original row if filtering is active
						originalRow := id.Row
						if table.data.RowMapping != nil && id.Row < len(table.data.RowMapping) {
							originalRow = table.data.RowMapping[id.Row]
						}
						table.updateCellFunc(id.Col, originalRow, widgetCell.content)
						widgetCell.lastGen = table.dataGen
					}
					// Apply row striping to widget-mode cells
					if id.Row == table.selectedRow {
						widgetCell.background.FillColor = table.SelectionColor
					} else {
						if id.Row%2 == 0 {
							widgetCell.background.FillColor = theme.Color(theme.ColorNameBackground)
						} else {
							widgetCell.background.FillColor = stripeAltColor()
						}
					}
					widgetCell.background.Refresh()
				}
			} else{
				// String mode: update Label (current behavior)
				// Handle both TableCell and WidgetCell (for when filtering in widget mode)
				if widgetCell, ok := o.(*WidgetCell); ok {
					// WidgetCell in string mode - replace content with a label showing string data
					txt := table.data.Get(id.Col, id.Row)
					if widgetCell.content == nil || widgetCell.currentID.Col != id.Col || widgetCell.currentID.Row != id.Row {
						label := widget.NewLabel(txt)
						widgetCell.SetWidget(label, id)
					} else if lbl, ok := widgetCell.content.(*widget.Label); ok {
						lbl.SetText(txt)
					}
				} else if cell, ok := o.(*TableCell); ok {
					txt := table.data.Get(id.Col, id.Row)
					if cell.label.Text != txt {
						cell.label.SetText(txt)
					}
					cell.Id = id
					if id.Row == table.selectedRow {
						cell.background.FillColor = table.SelectionColor
					} else {
						if id.Row%2 == 0 {
							cell.background.FillColor = theme.Color(theme.ColorNameBackground)
						} else {
							cell.background.FillColor = stripeAltColor()
						}
					}
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
	t.dataGen++
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
	t.table.Refresh()
	// Replacing the data invalidates the old scroll position. widget.Table keeps
	// a cached offset (t.offset.Y) that Refresh neither resets nor re-clamps, so
	// filtering a scrolled list to a shorter result leaves the viewport stranded
	// past the new content — the grid renders blank until a manual scroll/resize.
	// ScrollToTop zeroes both content.Offset.Y and the cached offset.
	t.table.ScrollToTop()
}

func (t *FlexTable) SetColumnWidth(id int, width float32) {
	t.table.SetColumnWidth(id, width)
}

func (t *FlexTable) Refresh() {
	// Treat an explicit Refresh as a possible content change (e.g. header-click
	// sort mutates data in place and calls this without SetData). Bumping dataGen
	// forces visible cells to re-run UpdateCell. The interactive column-resize
	// drag does NOT come through here — it drives the inner widget.Table directly
	// — so resize still skips the per-cell VM round-trip.
	t.dataGen++
	t.table.Refresh()
}


func (t *FlexTable) SetCreateCell(fn func(col, row int) fyne.CanvasObject) {
	t.createCellFunc = fn
	t.widgetMode = true
}

func (t *FlexTable) SetUpdateCell(fn func(col, row int, obj fyne.CanvasObject)) {
	t.updateCellFunc = fn
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
	// Use container.Stack for layering and container.Clip to prevent overflow
	content := container.NewStack(c.background, c.label)
	clipped := container.NewClip(content)
	return widget.NewSimpleRenderer(clipped)
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

// WidgetCell is a cell that can hold any widget (for widget mode)
type WidgetCell struct {
	widget.BaseWidget
	background *canvas.Rectangle
	content    fyne.CanvasObject
	currentID  widget.TableCellID
	lastGen    int // FlexTable.dataGen this cell last ran UpdateCell at (0 = never)
}

func NewWidgetCell() *WidgetCell {
	cell := &WidgetCell{
		background: canvas.NewRectangle(theme.Color(theme.ColorNameBackground)),
		currentID:  widget.TableCellID{Col: -1, Row: -1},
	}
	cell.ExtendBaseWidget(cell)
	return cell
}

func (w *WidgetCell) SetWidget(content fyne.CanvasObject, id widget.TableCellID) {
	// Only replace widget if the cell position changed
	if w.currentID.Col != id.Col || w.currentID.Row != id.Row {
		w.content = content
		w.currentID = id
	}
	w.Refresh()
}

func (w *WidgetCell) CreateRenderer() fyne.WidgetRenderer {
	clip := container.NewClip(w.content)
	stack := container.NewStack(w.background, clip)
	return &widgetCellRenderer{
		cell:  w,
		clip:  clip,
		stack: stack,
	}
}

type widgetCellRenderer struct {
	cell  *WidgetCell
	clip  *container.Clip
	stack *fyne.Container
}

func (r *widgetCellRenderer) Destroy() {}

func (r *widgetCellRenderer) Layout(size fyne.Size) {
	r.stack.Resize(size)
}

func (r *widgetCellRenderer) MinSize() fyne.Size {
	if r.cell.content != nil {
		return r.cell.content.MinSize()
	}
	return fyne.NewSize(0, 0)
}

func (r *widgetCellRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.stack}
}

func (r *widgetCellRenderer) Refresh() {
	// Update the clip container's content when cell content changes
	if r.clip.Content != r.cell.content {
		r.clip.Content = r.cell.content
	}
	r.cell.background.Refresh()
	r.stack.Refresh()
}
