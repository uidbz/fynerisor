package widget

import (
	"context"
	"fmt"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/vmexec"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"

	risorcanvas "github.com/uidbz/fynerisor/gui/canvas"
	"github.com/uidbz/fynerisor/gui/widget/tablewidget"
)

const TableType object.Type = "widget.Table"

type Table struct {
	instance *tablewidget.TableWidget
	Columns  func() []string
	w        WindowInterface

	// last-known getter results, reused during transient VM/render contention
	// (vmexec.ErrBusy) so the grid never collapses to empty for a frame; the
	// table is refreshed once the VM is idle (Callback.RefreshWhenBusy).
	cacheMu          sync.Mutex // guards the last* fields
	lastColumns      []string
	lastRowCount     int
	lastData         *tablewidget.TableData
	lastHeaderLevels [][]string
}

// cached reads a last-known getter result under cacheMu.
func (obj *Table) cached(get func() any) any {
	obj.cacheMu.Lock()
	defer obj.cacheMu.Unlock()
	return get()
}

// setCache updates last-known getter results under cacheMu.
func (obj *Table) setCache(set func()) {
	obj.cacheMu.Lock()
	defer obj.cacheMu.Unlock()
	set()
}

// refresh re-renders the table with fresh getter results; scheduled by the
// getters' callbacks when they hit VM contention.
func (obj *Table) refresh() {
	guithread.Do(obj.instance.Refresh)
}

func (obj *Table) CanvasObject() fyne.CanvasObject {
	return obj.instance.Instance
}

func (obj *Table) Type() object.Type {
	return TableType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *Table) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *Table) inspectUI() string {
	return "widget.Table"
}

func (obj *Table) Interface() interface{} {
	return obj.instance.Instance
}

func (obj *Table) IsTruthy() bool {
	return true
}

func (obj *Table) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'widget.Table'")
}

func (obj *Table) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for %s: %v", TableType, opType), nil
}

func (obj *Table) Equals(other object.Object) bool {
	return obj == other
}

func (obj *Table) Attrs() []object.AttrSpec {
	return nil
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *Table) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *Table) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", TableType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *Table) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

func (obj *Table) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "Columns":
		return object.NewBuiltin("widget.Table.Columns", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=exactly 1", len(args)), nil
			}
			cb, err := vmexec.NewCallback(ctx, "widget.Table.Columns", args[0])
			if err != nil {
				return nil, err
			}
			cb.RefreshWhenBusy(obj.refresh)

			obj.Columns = func() []string {
				o, err := cb.Call()
				if err != nil {
					cb.Report(err)
					return obj.cached(func() any { return obj.lastColumns }).([]string)
				}
				s, err2 := object.AsStringSlice(o)
				if err2 != nil {
					cb.Report(fmt.Errorf("expected []string: %w", err2))
					return obj.cached(func() any { return obj.lastColumns }).([]string)
				}
				obj.setCache(func() { obj.lastColumns = s })
				return s
			}
			return object.Nil, nil
		}), true

	case "HeaderLevels":
		return object.NewBuiltin("widget.Table.HeaderLevels", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=exactly 1", len(args)), nil
			}
			cb, err := vmexec.NewCallback(ctx, "widget.Table.HeaderLevels", args[0])
			if err != nil {
				return nil, err
			}
			cb.RefreshWhenBusy(obj.refresh)

			obj.instance.HeaderLevels = func() [][]string {
				o, err := cb.Call()
				if err != nil {
					cb.Report(err)
					return obj.cached(func() any { return obj.lastHeaderLevels }).([][]string)
				}
				rows, err2 := stringRowsFromObject(o)
				if err2 != nil {
					cb.Report(fmt.Errorf("expected list of lists of strings: %w", err2))
					return obj.cached(func() any { return obj.lastHeaderLevels }).([][]string)
				}
				obj.setCache(func() { obj.lastHeaderLevels = rows })
				return rows
			}
			return object.Nil, nil
		}), true

	case "RowCount":
		return object.NewBuiltin("widget.Table.RowCount", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=exactly 1", len(args)), nil
			}
			cb, err := vmexec.NewCallback(ctx, "widget.Table.RowCount", args[0])
			if err != nil {
				return nil, err
			}
			cb.RefreshWhenBusy(obj.refresh)

			obj.instance.RowCount = func() int {
				o, err := cb.Call()
				if err != nil {
					cb.Report(err)
					return obj.cached(func() any { return obj.lastRowCount }).(int)
				}
				s, err2 := object.AsInt(o)
				if err2 != nil {
					cb.Report(fmt.Errorf("expected int: %w", err2))
					return obj.cached(func() any { return obj.lastRowCount }).(int)
				}
				obj.setCache(func() { obj.lastRowCount = int(s) })
				return obj.cached(func() any { return obj.lastRowCount }).(int)
			}
			return object.Nil, nil
		}), true

	case "Refresh":
		return object.NewBuiltin("widget.Table.Refresh", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() { obj.instance.Refresh() })
			return object.Nil, nil
		}), true

	case "SetColumnWidth":
		return object.NewBuiltin("widget.Table.SetColumWidth", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			column, err := object.AsInt(args[0])
			if err != nil {
				return nil, err
			}
			width, err := object.AsFloat(args[1])
			if err != nil {
				return nil, err
			}
			guithread.Do(func() { obj.instance.SetColumnWidth(int(column), float32(width)) })
			return object.Nil, nil
		}), true

	case "SetOnClick":
		return object.NewBuiltin("widget.Table.SetOnClick", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=exactly 1", len(args)), nil
			}
			cb, err := vmexec.NewCallback(ctx, "widget.Table.SetOnClick", args[0])
			if err != nil {
				return nil, err
			}

			// Event: queued on the script's VM goroutine.
			obj.instance.SetOnClick(func(row, col int) {
				cb.Fire(object.NewInt(int64(row)), object.NewInt(int64(col)))
			})
			return object.Nil, nil
		}), true

	case "Data":
		return object.NewBuiltin("widget.Table.Data", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=exactly 1", len(args)), nil
			}
			cb, err := vmexec.NewCallback(ctx, "widget.Table.Data", args[0])
			if err != nil {
				return nil, err
			}
			cb.RefreshWhenBusy(obj.refresh)

			obj.instance.Data = func(offset, limit int) *tablewidget.TableData {
				lastOrEmpty := func() *tablewidget.TableData {
					if last := obj.cached(func() any { return obj.lastData }).(*tablewidget.TableData); last != nil {
						return last
					}
					return tablewidget.NewTableData(obj.instance.Title)
				}
				o, err := cb.Call(object.NewInt(int64(offset)), object.NewInt(int64(limit)))
				if err != nil {
					cb.Report(err)
					return lastOrEmpty()
				}
				outer, err2 := object.AsList(o)
				if err2 != nil {
					cb.Report(fmt.Errorf("expected list: %w", err2))
					return lastOrEmpty()
				}
				td := tablewidget.NewTableData(obj.instance.Title)
				for _, inner := range outer.Value() {
					switch inner := inner.(type) {
					case *object.List:
						td.AddStringRow(obj.Columns(), stringsFromList(inner))
					default:
						cb.Report(fmt.Errorf("expected list of rows, got row of type %s", inner.Type()))
						return lastOrEmpty()
					}
				}
				obj.setCache(func() { obj.lastData = td })
				return td
			}
			return object.Nil, nil
		}), true

	case "CreateCell":
		return object.NewBuiltin("widget.Table.CreateCell", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			cb, err := vmexec.NewCallback(ctx, "widget.Table.CreateCell", args[0])
			if err != nil {
				return nil, err
			}
			cb.RefreshWhenBusy(obj.refresh)

			createFunc := func(col, row int) fyne.CanvasObject {
				// Getter: Fyne calls it synchronously on the UI thread.
				result, err := cb.Call(object.NewInt(int64(col)), object.NewInt(int64(row)))
				if err != nil {
					if vmexec.Quiet(err) {
						// VM busy or run stopped: nil tells the flex table
						// to leave the cell empty and ask again on a later
						// refresh; a placeholder here would be pooled and
						// then updated forever in place of the script's
						// widget (issue #50).
						return nil
					}
					cb.Report(err)
					return widget.NewLabel("")
				}
				if widgetObj, ok := result.(interface{ CanvasObject() fyne.CanvasObject }); ok {
					return widgetObj.CanvasObject()
				}
				if errObj, ok := result.(*object.Error); ok {
					cb.Report(errObj.Value())
				} else {
					cb.Report(fmt.Errorf("result does not implement CanvasObject() - got type: %T", result))
				}
				return widget.NewLabel("")
			}

			guithread.Do(func() {
				obj.instance.GetFlexTable().SetCreateCell(createFunc)
			})
			return object.Nil, nil
		}), true

	case "UpdateCell":
		return object.NewBuiltin("widget.Table.UpdateCell", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			cb, err := vmexec.NewCallback(ctx, "widget.Table.UpdateCell", args[0])
			if err != nil {
				return nil, err
			}
			cb.RefreshWhenBusy(obj.refresh)

			updateFunc := func(col, row int, canvasObj fyne.CanvasObject) {
				// Getter-like: Fyne calls it synchronously on the UI thread.
				_, err := cb.Call(object.NewInt(int64(col)), object.NewInt(int64(row)), wrapCanvasObjectTable(canvasObj))
				cb.Report(err)
			}

			guithread.Do(func() {
				obj.instance.GetFlexTable().SetUpdateCell(updateFunc)
			})
			return object.Nil, nil
		}), true

	case "Hide":
		return object.NewBuiltin("widget.Table.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("widget.Table.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			guithread.Do(func() {
				obj.instance.Instance.Show()
			})
			return object.Nil, nil
		}), true
	}
	return nil, false
}

// stringRowsFromObject converts a Risor list of string lists — the shape tie's
// read_table reports header_levels in — to rows of strings.
func stringRowsFromObject(o object.Object) ([][]string, error) {
	outer, err := object.AsList(o)
	if err != nil {
		return nil, err
	}
	rows := make([][]string, 0, len(outer.Value()))
	for _, inner := range outer.Value() {
		list, ok := inner.(*object.List)
		if !ok {
			return nil, fmt.Errorf("type error: expected list (got %s)", inner.Type())
		}
		rows = append(rows, stringsFromList(list))
	}
	return rows, nil
}

func stringsFromList(l *object.List) []string {
	var strs []string
	for _, item := range l.Value() {
		switch item := item.(type) {
		case fmt.Stringer:
			strs = append(strs, item.String())
		default:
			strs = append(strs, item.Inspect())
		}
	}
	return strs
}

func NewTable(title string, pageSize int, w WindowInterface) *Table {
	tw := tablewidget.NewTableWidget(title, pageSize)
	tw.SetErrorSink(w.Report) // per-runtime diagnostics (issue #31)
	return &Table{instance: tw, w: w}
}

// wrapCanvasObjectTable wraps a fyne.CanvasObject so Risor scripts can access it in table cells
func wrapCanvasObjectTable(obj fyne.CanvasObject) object.Object {
	switch v := obj.(type) {
	case *widget.Label:
		return &Label{instance: v}
	case *widget.Button:
		return &Button{instance: v}
	case *widget.Entry:
		return &Entry{instance: v}
	case *widget.Select:
		return &Select{instance: v}
	case *widget.Check:
		return &Check{instance: v}
	case *widget.Icon:
		return &Icon{instance: v}
	case *widget.RadioGroup:
		return &RadioGroup{instance: v}
	case *widget.CheckGroup:
		return &CheckGroup{instance: v}
	case *widget.Slider:
		return &Slider{instance: v}
	case *widget.ProgressBar:
		return &ProgressBar{instance: v}
	case *widget.ProgressBarInfinite:
		return &ProgressBarInfinite{instance: v}
	case *widget.Hyperlink:
		return &Hyperlink{instance: v}
	case *widget.Separator:
		return &Separator{instance: v}
	case *widget.Card:
		return &Card{instance: v}
	case *widget.Form:
		return &Form{instance: v}
	case *widget.Accordion:
		return &Accordion{instance: v}
	// Canvas objects
	case *canvas.Image:
		// Return proper Image wrapper with SetImageFromURI support
		return risorcanvas.NewImage(v)
	case *canvas.Text:
		return &GenericCanvasObjectTable{obj: obj}
	case *canvas.Rectangle:
		return &GenericCanvasObjectTable{obj: obj}
	case *canvas.Circle:
		return &GenericCanvasObjectTable{obj: obj}
	case *canvas.Line:
		return &GenericCanvasObjectTable{obj: obj}
	case *container.Scroll:
		// Handle containers - wrap as generic
		return &GenericCanvasObjectTable{obj: obj}
	case *fyne.Container:
		// Check if container holds an image (from canvas.NewImageFromURI)
		if len(v.Objects) > 0 {
			if img, ok := v.Objects[0].(*canvas.Image); ok {
				// Create wrapper that will modify the existing container/image in place
				return &ImageCellWrapper{container: v, image: img}
			}
		}
		// Generic container handling
		return &GenericCanvasObjectTable{obj: obj}
	default:
		return &GenericCanvasObjectTable{obj: obj}
	}
}

// ImageCellWrapper wraps an image in a table cell, allowing it to be updated
type ImageCellWrapper struct {
	container *fyne.Container
	image     *canvas.Image
}

func (g *ImageCellWrapper) Type() object.Type {
	return "canvas.Image"
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (g *ImageCellWrapper) Inspect() string {
	return guithread.Inspect(g.inspectUI)
}

func (g *ImageCellWrapper) inspectUI() string {
	return "canvas.Image"
}

func (g *ImageCellWrapper) Interface() interface{} {
	return g.container
}

func (g *ImageCellWrapper) CanvasObject() fyne.CanvasObject {
	return g.container
}

func (g *ImageCellWrapper) IsTruthy() bool {
	return true
}

func (g *ImageCellWrapper) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal canvas.Image")
}

func (g *ImageCellWrapper) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for canvas.Image"), nil
}

func (g *ImageCellWrapper) Equals(other object.Object) bool {
	return g == other
}

func (g *ImageCellWrapper) Attrs() []object.AttrSpec {
	return nil
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (g *ImageCellWrapper) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, g.setAttrUI)
}

func (g *ImageCellWrapper) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: canvas.Image has no settable attributes")
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (g *ImageCellWrapper) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, g.getAttrUI)
}

func (g *ImageCellWrapper) getAttrUI(name string) (object.Object, bool) {
	switch name {
	case "SetImageFromURI":
		return object.NewBuiltin("canvas.Image.SetImageFromURI", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			path, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}

			uri, err := storage.ParseURI(path)
			if err != nil {
				return object.NewError(err), nil
			}

			guithread.Do(func() {
				// Update the existing image instead of creating a new one
				newImg := canvas.NewImageFromURI(uri)
				newImg.FillMode = canvas.ImageFillOriginal
				g.image.Resource = newImg.Resource
				g.image.File = newImg.File
				g.image.FillMode = newImg.FillMode
				g.image.Refresh()
			})

			return object.Nil, nil
		}), true
	}
	return nil, false
}

// GenericCanvasObjectTable is a minimal wrapper for canvas objects in tables
type GenericCanvasObjectTable struct {
	obj fyne.CanvasObject
}

func (g *GenericCanvasObjectTable) Type() object.Type {
	return "canvas.Object"
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (g *GenericCanvasObjectTable) Inspect() string {
	return guithread.Inspect(g.inspectUI)
}

func (g *GenericCanvasObjectTable) inspectUI() string {
	return "canvas.Object"
}

func (g *GenericCanvasObjectTable) Interface() interface{} {
	return g.obj
}

func (g *GenericCanvasObjectTable) CanvasObject() fyne.CanvasObject {
	return g.obj
}

func (g *GenericCanvasObjectTable) IsTruthy() bool {
	return true
}

func (g *GenericCanvasObjectTable) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal canvas.Object")
}

func (g *GenericCanvasObjectTable) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for canvas.Object"), nil
}

func (g *GenericCanvasObjectTable) Equals(other object.Object) bool {
	return g == other
}

func (g *GenericCanvasObjectTable) Attrs() []object.AttrSpec {
	return nil
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (g *GenericCanvasObjectTable) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, g.setAttrUI)
}

func (g *GenericCanvasObjectTable) setAttrUI(name string, value object.Object) error {
	return fmt.Errorf("attribute error: canvas.Object has no settable attributes")
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (g *GenericCanvasObjectTable) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, g.getAttrUI)
}

func (g *GenericCanvasObjectTable) getAttrUI(name string) (object.Object, bool) {
	return nil, false
}
