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

	// headerDisplay is the multi-row header to draw, row-major and rectangular:
	// headerDisplay[level][col]. It lives on the table rather than on TableData
	// because paging and filtering rebuild TableData from scratch, while the header
	// belongs to the view. Empty means the single-row default (the column name).
	headerDisplay [][]string

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
	t.installHeader()
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

// SetHeaderLevels sets a multi-row (hierarchical) header. rows is row-major —
// rows[i][j] is level i of column j — the shape tie's read_table reports and a
// spreadsheet reads. Rows need not be rectangular; short ones are padded. Passing
// nil restores the single-row default, where each column shows its own name.
//
// Levels are display-only: the bottom level is not the column identity, so sorting,
// cell lookup and export keep using TableData.Columns.
func (t *FlexTable) SetHeaderLevels(rows [][]string) {
	display := spanHeaderRuns(rows)
	if headerGridsEqual(display, t.headerDisplay) {
		return
	}
	t.headerDisplay = display
	t.installHeader()
	t.table.Refresh()
}

func (t *FlexTable) installHeader() {
	depth := len(t.headerDisplay)
	if depth < 1 {
		depth = 1
	}
	t.table.CreateHeader = func() fyne.CanvasObject {
		return newHeaderWithDepth(depth, t.headerBgColor, t)
	}
	t.table.UpdateHeader = func(id widget.TableCellID, o fyne.CanvasObject) {
		h := o.(*Header)
		h.SetTexts(t.headerLabels(id.Col))
		h.colId = t.columnName(id.Col)
		h.RefreshSortIcon()
	}
}

// headerLabels returns what to draw on each level of one column's header. A column
// the levels grid does not reach falls back to its name on the bottom level, so a
// grid out of step with the data still names every column.
func (t *FlexTable) headerLabels(col int) []string {
	if len(t.headerDisplay) == 0 {
		return []string{t.columnName(col)}
	}
	labels := make([]string, len(t.headerDisplay))
	if col >= len(t.headerDisplay[0]) {
		labels[len(labels)-1] = t.columnName(col)
		return labels
	}
	for i, row := range t.headerDisplay {
		labels[i] = row[col]
	}
	return labels
}

func (t *FlexTable) columnName(col int) string {
	if col >= 0 && col < len(t.data.Columns) {
		return t.data.Columns[col]
	}
	return ""
}

// spanHeaderRuns pads a row-major header grid to a rectangle and blanks the
// continuation columns of every run of equal adjacent labels. That is how a merged
// parent cell reads: the converter forward-fills it across the columns it covers,
// so drawing it once at the start of the run restores the merge. Fyne cannot span
// header cells, but the label is not clipped to its column, so a run's first cell
// carries the text across the ones after it.
//
// A run breaks as soon as any level above it changes, which keeps two identically
// named groups under different parents apart.
func spanHeaderRuns(rows [][]string) [][]string {
	width := 0
	for _, row := range rows {
		if len(row) > width {
			width = len(row)
		}
	}
	if width == 0 {
		return nil
	}
	grid := make([][]string, len(rows))
	for i, row := range rows {
		grid[i] = make([]string, width)
		copy(grid[i], row)
	}
	out := make([][]string, len(grid))
	for i, row := range grid {
		out[i] = make([]string, width)
		copy(out[i], row)
		for col := 1; col < width; col++ {
			same := true
			for level := 0; level <= i; level++ {
				if grid[level][col] != grid[level][col-1] {
					same = false
					break
				}
			}
			if same {
				out[i][col] = ""
			}
		}
	}
	return out
}

func headerGridsEqual(a, b [][]string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}
		for j := range a[i] {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}
	return true
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

	// levels holds one label per header level, top to bottom. The last is
	// TableCell.label — the column's own level, where the sort icon sits.
	levels []*widget.Label
	stack  *fyne.Container
}

func NewHeader(columnName string, bgColor color.Color, table *FlexTable) *Header {
	header := newHeaderWithDepth(1, bgColor, table)
	header.SetText(columnName)
	return header
}

// newHeaderWithDepth builds a header cell that stacks depth level labels, for a
// multi-row (hierarchical) header. Fyne pools and reuses header cells across
// refreshes, so a cell also adapts its depth in SetTexts; depth here only has to
// make the template tall enough, since Fyne derives the header row height from
// CreateHeader().MinSize() on every refresh.
func newHeaderWithDepth(depth int, bgColor color.Color, table *FlexTable) *Header {
	header := &Header{}
	header.background = canvas.NewRectangle(bgColor)
	header.table = table
	header.bgColor = theme.Color(theme.ColorNameBackground)
	header.setDepth(depth)

	header.ExtendBaseWidget(header)
	return header
}

func newHeaderLabel() *widget.Label {
	label := widget.NewLabel("")
	label.TextStyle.Bold = true
	label.Importance = widget.HighImportance
	// label.TextSize = theme.TextSize() * 1.2
	return label
}

func (h *Header) setDepth(depth int) {
	if depth < 1 {
		depth = 1
	}
	if depth == len(h.levels) {
		return
	}
	for len(h.levels) < depth {
		h.levels = append(h.levels, newHeaderLabel())
	}
	h.levels = h.levels[:depth]
	h.label = h.levels[depth-1]
	h.layoutLevels()
}

func (h *Header) SetText(columnName string) {
	h.label.Text = columnName
	h.label.Refresh()
}

// SetTexts sets one label per header level, top to bottom, growing or shrinking the
// stack to match.
func (h *Header) SetTexts(labels []string) {
	h.setDepth(len(labels))
	for i, label := range h.levels {
		text := ""
		if i < len(labels) {
			text = labels[i]
		}
		if label.Text != text {
			label.Text = text
			label.Refresh()
		}
	}
}

func (h *Header) CreateRenderer() fyne.WidgetRenderer {
	h.icon = widget.NewIcon(nil)
	h.stack = container.NewVBox()
	h.layoutLevels()
	item := container.NewStack(h.background, h.stack)
	return widget.NewSimpleRenderer(item)
}

// layoutLevels stacks the level labels, the sort icon beside the bottom one. It is
// a no-op until the renderer exists, which is where the container comes from.
func (h *Header) layoutLevels() {
	if h.stack == nil {
		return
	}
	objects := make([]fyne.CanvasObject, 0, len(h.levels))
	for i, label := range h.levels {
		if i == len(h.levels)-1 {
			objects = append(objects, container.NewBorder(nil, nil, nil, h.icon, label))
			continue
		}
		objects = append(objects, label)
	}
	h.stack.Objects = objects
	h.stack.Refresh()
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
