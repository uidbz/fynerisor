package widget

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"

	"github.com/uidbz/fynerisor/gui/widget/tablewidget"
)

const TableType object.Type = "widget.Table"

type Table struct {
	instance *tablewidget.TableWidget
	Columns  func() []string
	w WindowInterface
}

func (obj *Table) CanvasObject() fyne.CanvasObject {
	return obj.instance.Instance
}

func (obj *Table) Type() object.Type {
	return TableType
}

func (obj *Table) Inspect() string {
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

func (obj *Table) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", TableType, name)
}

func (obj *Table) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "Columns":
		return object.NewBuiltin("widget.Table.Columns", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=exactly 1", len(args)), nil
			}
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("table (%s): unable to get call function", name), nil
			}

			obj.Columns = func() []string {
				o, err := callFunc(ctx, fn, []object.Object{})
				if err != nil {
					fmt.Println("ERROR:", err)
				}
				s, err2 := object.AsStringSlice(o)
				if err2 != nil {
					fmt.Println("ERROR:", err)
				}
				return s
			}
			return object.Nil, nil
		}), true

	case "RowCount":
		return object.NewBuiltin("widget.Table.RowCount", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=exactly 1", len(args)), nil
			}
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("table (%s): unable to get call function", name), nil
			}

			obj.instance.RowCount = func() int {
				o, err := callFunc(ctx, fn, []object.Object{})
				if err != nil {
					fmt.Println("RowCount: ERROR:", err)
					return 0
				}
				s, err2 := object.AsInt(o)
				if err2 != nil {
					fmt.Println("RowCount: ERROR:", err)
					return 0
				}
				return int(s)
			}
			return object.Nil, nil
		}), true

	case "Refresh":
		return object.NewBuiltin("widget.Table.Refresh", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() { obj.instance.Refresh() })
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
			fyne.Do(func() { obj.instance.SetColumnWidth(int(column), float32(width)) })
			return object.Nil, nil
		}), true

	case "SetOnClick":
		return object.NewBuiltin("widget.Table.SetOnClick", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=exactly 1", len(args)), nil
			}
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("table (%s): unable to get call function", name), nil
			}

			obj.instance.SetOnClick(func(row, col int) {
				_, err := callFunc(ctx, fn, []object.Object{object.NewInt(int64(row)), object.NewInt(int64(col))})
				if err != nil {
					fmt.Println("widget.Table.SetOnClick: ERROR:", err)
					return
				}
			})
			return object.Nil, nil
		}), true

	case "Data":
		return object.NewBuiltin("widget.Table.Data", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=exactly 1", len(args)), nil
			}
			fn, ok := args[0].(*object.Closure)
			if !ok {
				return object.Errorf("argument error: expected function, got %s", args[0].Type()), nil
			}
			callFunc, ok := object.GetCallFunc(ctx)
			if !ok {
				return object.Errorf("table (%s): unable to get call function", name), nil
			}

			obj.instance.Data = func(offset, limit int) *tablewidget.TableData {
				o, err := callFunc(ctx, fn, []object.Object{object.NewInt(int64(offset)), object.NewInt(int64(limit))})
				if err != nil {
					fmt.Println("widget.Table.Data: ERROR:", err)
					return tablewidget.NewTableData(obj.instance.Title)
				}
				outer, err2 := object.AsList(o)
				if err2 != nil {
					fmt.Println("ERROR:", err2)
				}
				td := tablewidget.NewTableData(obj.instance.Title)
				for _, inner := range outer.Value() {
					switch inner := inner.(type) {
					case *object.List:
						td.AddStringRow(obj.Columns(), stringsFromList(inner))
					default:
						fmt.Printf("type error: expected list (got %s)\n", inner.Type())
						return tablewidget.NewTableData(obj.instance.Title)
					}
				}
				return td
			}
			return object.Nil, nil
		}), true
	case "Hide":
		return object.NewBuiltin("widget.Table.Hide", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Instance.Hide()
			})
			return object.Nil, nil
		}), true
	case "Show":
		return object.NewBuiltin("widget.Table.Show", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return object.Errorf("wrong number of arguments. got=%d, want=0", len(args)), nil
			}
			fyne.Do(func() {
				obj.instance.Instance.Show()
			})
			return object.Nil, nil
		}), true
	}
	return nil, false
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
	return &Table{instance: tablewidget.NewTableWidget(title, pageSize), w: w}
}
