package builtin

import (
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/gui/widget/tablewidget"
	"github.com/uidbz/fynerisor/scene"
	"github.com/uidbz/fynerisor/scene/render"
)

// ------------------------------------------------------------------- form

// formLive renders its children as one Fyne form so field labels align
// across rows, matching the look of every script-built Form. A child that
// does not provide a formItem (a bare input dropped in directly) becomes an
// unlabeled row, so the page still renders.
type formLive struct {
	form     *widget.Form
	bc       *render.BuildContext
	children []render.Live
}

// formChild is the internal contract between form and field: field hands
// over its assembled FormItem. It is looked up through render.UnwrapLive
// because children arrive wrapped in style/hint wrappers.
type formChild interface {
	formItem() (widget.FormItem, bool)
}

func newForm(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
	l := &formLive{form: widget.NewForm(), bc: bc}
	// Props must be applied at build too: the renderer only calls Update on
	// changes, so without this submit-text/cancel-text never take effect and
	// Fyne's default Submit/Cancel row leaks through on every scene form.
	l.applyButtons(n)
	return l, nil
}

func (l *formLive) Object() fyne.CanvasObject { return l.form }

func (l *formLive) Update(n *scene.Node, _ scene.Fields) error {
	l.applyButtons(n)
	l.sync()
	return nil
}

// applyButtons exposes Fyne Form's native submit/cancel row: setting a text
// installs the callback that emits the event, clearing it removes the button.
func (l *formLive) applyButtons(n *scene.Node) {
	st := l.bc.Str(n, "submit-text", "")
	ct := l.bc.Str(n, "cancel-text", "")
	id := n.ID
	l.form.SubmitText = st
	l.form.CancelText = ct
	if st != "" {
		l.form.OnSubmit = func() { l.bc.Emit(id, "submitted", nil) }
	} else {
		l.form.OnSubmit = nil
	}
	if ct != "" {
		l.form.OnCancel = func() { l.bc.Emit(id, "cancelled", nil) }
	} else {
		l.form.OnCancel = nil
	}
}

func (l *formLive) Destroy() {}

func (l *formLive) items() []*widget.FormItem {
	items := make([]*widget.FormItem, 0, len(l.children))
	for _, c := range l.children {
		if fc, ok := render.UnwrapLive(c).(formChild); ok {
			if it, ok2 := fc.formItem(); ok2 {
				items = append(items, &it)
				continue
			}
		}
		items = append(items, &widget.FormItem{Widget: c.Object()})
	}
	return items
}

func (l *formLive) sync() {
	l.form.Items = l.items()
	l.form.Refresh()
}

func (l *formLive) Insert(index int, child render.Live, p scene.Place) error {
	index = clamp(index, 0, len(l.children))
	l.children = append(l.children, nil)
	copy(l.children[index+1:], l.children[index:])
	l.children[index] = child
	render.SetSizeHint(child.Object(), p.MinW, p.MinH)
	l.sync()
	return nil
}

func (l *formLive) Remove(child render.Live) error {
	for i, c := range l.children {
		if c == child {
			l.children = append(l.children[:i], l.children[i+1:]...)
			l.sync()
			return nil
		}
	}
	return nil
}

func (l *formLive) Reorder(child render.Live, index int) error {
	from := -1
	for i, c := range l.children {
		if c == child {
			from = i
			break
		}
	}
	if from < 0 {
		return nil
	}
	index = clamp(index, 0, len(l.children)-1)
	if from == index {
		return nil
	}
	c := l.children[from]
	l.children = append(l.children[:from], l.children[from+1:]...)
	l.children = append(l.children, nil)
	copy(l.children[index+1:], l.children[index:])
	l.children[index] = c
	l.sync()
	return nil
}

func (l *formLive) SetPlace(child render.Live, p scene.Place) error {
	render.SetSizeHint(child.Object(), p.MinW, p.MinH)
	l.form.Refresh()
	return nil
}

func (l *formLive) Preview(render.Live, scene.Place) {}

func (l *formLive) Transform(fyne.Size) render.Transform { return render.Transform{SX: 1, SY: 1} }

func (l *formLive) DropSpot(local fyne.Position) (render.DropSpot, bool) {
	index := len(l.children)
	for i, c := range l.children {
		o := c.Object()
		if local.Y < o.Position().Y+o.Size().Height/2 {
			index = i
			break
		}
	}
	y := float32(0)
	if index > 0 && index <= len(l.children) {
		o := l.children[index-1].Object()
		y = o.Position().Y + o.Size().Height
	}
	return render.DropSpot{Index: index, Rect: render.NewRect(0, y, l.form.Size().Width, 2)}, true
}

// ------------------------------------------------------------------ field

// fieldLive is a labeled row. Inside a form it contributes its FormItem to
// the parent; standalone it renders as its own one-row form, so a stray
// field degrades to the same look instead of an unlabeled widget.
type fieldLive struct {
	bc          *render.BuildContext
	form        *widget.Form
	label, help string
	child       render.Live
}

func newField(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
	return &fieldLive{
		bc:    bc,
		form:  widget.NewForm(),
		label: bc.Str(n, "label", ""),
		help:  bc.Str(n, "help", ""),
	}, nil
}

func (l *fieldLive) Object() fyne.CanvasObject { return l.form }

func (l *fieldLive) Update(n *scene.Node, _ scene.Fields) error {
	l.label = l.bc.Str(n, "label", "")
	l.help = l.bc.Str(n, "help", "")
	l.sync()
	return nil
}

func (l *fieldLive) Destroy() {}

func (l *fieldLive) formItem() (widget.FormItem, bool) {
	if l.child == nil {
		return widget.FormItem{}, false
	}
	return widget.FormItem{Text: l.label, HintText: l.help, Widget: l.child.Object()}, true
}

func (l *fieldLive) sync() {
	items := []*widget.FormItem{}
	if it, ok := l.formItem(); ok {
		items = append(items, &it)
	}
	l.form.Items = items
	l.form.Refresh()
}

func (l *fieldLive) Insert(_ int, child render.Live, p scene.Place) error {
	if l.child != nil && l.child != child {
		l.child.Destroy()
	}
	l.child = child
	render.SetSizeHint(child.Object(), p.MinW, p.MinH)
	l.sync()
	return nil
}

func (l *fieldLive) Remove(child render.Live) error {
	if l.child == child {
		l.child = nil
		l.sync()
	}
	return nil
}

func (l *fieldLive) Reorder(render.Live, int) error { return nil }

func (l *fieldLive) SetPlace(child render.Live, p scene.Place) error {
	if child == l.child {
		render.SetSizeHint(child.Object(), p.MinW, p.MinH)
	}
	l.form.Refresh()
	return nil
}

func (l *fieldLive) Preview(render.Live, scene.Place) {}

func (l *fieldLive) Transform(fyne.Size) render.Transform { return render.Transform{SX: 1, SY: 1} }

func (l *fieldLive) DropSpot(local fyne.Position) (render.DropSpot, bool) {
	if l.child != nil {
		return render.DropSpot{}, false
	}
	return render.DropSpot{Index: 0, Rect: render.NewRect(0, 0, l.form.Size().Width, l.form.Size().Height)}, true
}

// ------------------------------------------------------------------ table

// tableLive wraps the same TableWidget the classic widget API uses, so a
// scene table gets the identical title/filter/sort/pager chrome and script
// apps keep one look. Rows live in the component and are fed through the
// render.RowModel capability; the TableWidget getters page over them.
type tableLive struct {
	bc   *render.BuildContext
	tw   *tablewidget.TableWidget
	cols []string
	rows [][]string
	id   string
}

func newTable(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
	pageSize := int(bc.Num(n, "pageSize", 10))
	if pageSize < 1 {
		pageSize = 10
	}
	tw := tablewidget.NewTableWidget(bc.Str(n, "title", ""), pageSize)
	tw.SetErrorSink(func(what string, err error) {
		bc.Statusf("table %s: %v", n.ID, err)
	})
	l := &tableLive{bc: bc, tw: tw, cols: bc.List(n, "columns"), id: n.ID}
	tw.RowCount = func() int { return len(l.rows) }
	tw.Data = l.page
	tw.SetOnClick(func(row, col int) {
		text := ""
		if row >= 0 && row < len(l.rows) && col >= 0 && col < len(l.rows[row]) {
			text = l.rows[row][col]
		}
		bc.Emit(l.id, "cellTapped", map[string]any{
			"row": row, "col": col, "column": l.colName(col), "text": text,
		})
	})
	return l, nil
}

func (l *tableLive) Object() fyne.CanvasObject { return l.tw.Instance }

func (l *tableLive) Update(n *scene.Node, changed scene.Fields) error {
	// Title and page size build the widget chrome; changing them rebuilds.
	if changed["title"] || changed["pageSize"] {
		return render.ErrRebuild
	}
	l.cols = l.bc.List(n, "columns")
	l.tw.Refresh()
	return nil
}

func (l *tableLive) Destroy() {}

func (l *tableLive) colName(col int) string {
	if col >= 0 && col < len(l.cols) {
		return l.cols[col]
	}
	return ""
}

// columnNames resolves the column set for one page of TableData. Without
// declared columns the data's own width gets numbered names, which keeps
// AddStringRow's column alignment well-defined.
func (l *tableLive) columnNames(row []string) []string {
	out := make([]string, len(row))
	for i := range row {
		if i < len(l.cols) {
			out[i] = l.cols[i]
		} else {
			out[i] = strconv.Itoa(i + 1)
		}
	}
	return out
}

// page builds one page of TableData (the TableWidget pulls it on Refresh).
// Sorting operates per page snapshot, matching the classic script tables.
func (l *tableLive) page(offset, limit int) *tablewidget.TableData {
	td := tablewidget.NewTableData(l.id)
	if offset < 0 {
		offset = 0
	}
	end := offset + limit
	if limit < 1 || end > len(l.rows) {
		end = len(l.rows)
	}
	for _, r := range l.rows[offset:end] {
		td.AddStringRow(l.columnNames(r), r)
	}
	return td
}

// SetRows replaces the whole row set (render.RowModel).
func (l *tableLive) SetRows(rows [][]string) {
	l.rows = make([][]string, len(rows))
	for i, r := range rows {
		l.rows[i] = append([]string(nil), r...)
	}
	l.tw.Refresh()
}

// SetCell updates one in-range cell (render.RowModel).
func (l *tableLive) SetCell(row, col int, value string) bool {
	if row < 0 || row >= len(l.rows) || col < 0 {
		return false
	}
	for col >= len(l.rows[row]) {
		l.rows[row] = append(l.rows[row], "")
	}
	l.rows[row][col] = value
	l.tw.Refresh()
	return true
}

// -------------------------------------------------------------------- log

// logLive is the scene twin of the classic widget.Log (a virtual list with a
// bounded line buffer), so a scene log matches script-built logs.
type logLive struct {
	bc    *render.BuildContext
	list  *widget.List
	lines []string
	max   int
}

func newLog(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
	l := &logLive{bc: bc, max: int(bc.Num(n, "maxLines", 1000))}
	if l.max < 1 {
		l.max = 1000
	}
	l.list = widget.NewList(
		func() int { return len(l.lines) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(l.lines) {
				obj.(*widget.Label).SetText(l.lines[id])
			}
		},
	)
	l.setLines(bc.Str(n, "text", ""))
	return l, nil
}

func (l *logLive) Object() fyne.CanvasObject { return l.list }

func (l *logLive) Update(n *scene.Node, changed scene.Fields) error {
	if !changed["text"] {
		return nil
	}
	l.setLines(l.bc.Str(n, "text", ""))
	l.list.Refresh()
	return nil
}

func (l *logLive) Destroy() {}

func (l *logLive) setLines(text string) {
	if text == "" {
		l.lines = nil
		return
	}
	l.lines = strings.Split(text, "\n")
	if over := len(l.lines) - l.max; over > 0 {
		l.lines = l.lines[over:]
	}
}

// AppendText appends one line (render.Appender).
func (l *logLive) AppendText(text string) {
	for _, line := range strings.Split(text, "\n") {
		if len(l.lines) >= l.max {
			l.lines = l.lines[1:]
		}
		l.lines = append(l.lines, line)
	}
	l.list.Refresh()
	l.list.ScrollToBottom()
}
