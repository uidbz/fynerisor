package sql

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

const SQLRowIteratorType = object.Type("sql.row_iterator")

type RowIterator struct {
	ctx      context.Context
	rows     *sql.Rows
	once     sync.Once
	closed   chan bool
	isClosed bool
	columns  []string
	current  object.Object
	index    int64
}

func (ri *RowIterator) Type() object.Type {
	return SQLRowIteratorType
}

func (ri *RowIterator) Inspect() string {
	return "sql.row_iterator()"
}

func (ri *RowIterator) Interface() interface{} {
	return ri.rows
}

func (ri *RowIterator) Equals(other object.Object) bool {
	return ri == other
}

func (ri *RowIterator) IsTruthy() bool {
	return !ri.isClosed
}

func (ri *RowIterator) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "next":
		return object.NewBuiltin("sql.row_iterator.next", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, fmt.Errorf("sql.row_iterator.next: expected 0 arguments, got %d", len(args))
			}
			obj, ok := ri.Next(ctx)
			if !ok {
				return object.Nil, nil
			}
			return obj, nil
		}), true
	case "close":
		return object.NewBuiltin("sql.row_iterator.close", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, fmt.Errorf("sql.row_iterator.close: expected 0 arguments, got %d", len(args))
			}
			ri.Close()
			return object.Nil, nil
		}), true
	case "entry":
		return object.NewBuiltin("sql.row_iterator.entry", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, fmt.Errorf("sql.row_iterator.entry: expected 0 arguments, got %d", len(args))
			}
			entry, ok := ri.Entry()
			if !ok {
				return object.Nil, nil
			}
			return entry, nil
		}), true
	case "collect":
		return object.NewBuiltin("sql.row_iterator.collect", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, fmt.Errorf("sql.row_iterator.collect: expected 0 arguments, got %d", len(args))
			}
			var result []object.Object
			for {
				obj, ok := ri.Next(ctx)
				if !ok {
					break
				}
				if err, isErr := obj.(*object.Error); isErr {
					return err, nil
				}
				result = append(result, obj)
			}
			return object.NewList(result), nil
		}), true
	case "map":
		return object.NewBuiltin("sql.row_iterator.map", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("sql.row_iterator.map: expected 1 argument, got %d", len(args))
			}
			fn, ok := args[0].(object.Callable)
			if !ok {
				return nil, fmt.Errorf("sql.row_iterator.map: argument must be callable")
			}
			var result []object.Object
			for {
				obj, ok := ri.Next(ctx)
				if !ok {
					break
				}
				if err, isErr := obj.(*object.Error); isErr {
					return err, nil
				}
				mapped, err := fn.Call(ctx, obj)
				if err != nil {
					return nil, err
				}
				result = append(result, mapped)
			}
			return object.NewList(result), nil
		}), true
	case "each":
		return object.NewBuiltin("sql.row_iterator.each", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("sql.row_iterator.each: expected 1 argument, got %d", len(args))
			}
			fn, ok := args[0].(object.Callable)
			if !ok {
				return nil, fmt.Errorf("sql.row_iterator.each: argument must be callable")
			}
			for {
				obj, ok := ri.Next(ctx)
				if !ok {
					break
				}
				if err, isErr := obj.(*object.Error); isErr {
					return err, nil
				}
				_, err := fn.Call(ctx, obj)
				if err != nil {
					return nil, err
				}
			}
			return object.Nil, nil
		}), true
	}
	return nil, false
}

func (ri *RowIterator) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: sql.row_iterator object has no attribute %q", name)
}

func (ri *RowIterator) Attrs() []object.AttrSpec {
	return []object.AttrSpec{
		{Name: "next"},
		{Name: "close"},
		{Name: "entry"},
		{Name: "collect"},
		{Name: "map"},
		{Name: "each"},
	}
}

func (ri *RowIterator) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for sql.row_iterator: %v", opType), nil
}

func (ri *RowIterator) Close() {
	ri.once.Do(func() {
		ri.isClosed = true
		ri.rows.Close()
		close(ri.closed)
	})
}

func (ri *RowIterator) Cost() int {
	return 8
}

func (ri *RowIterator) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal sql.row_iterator")
}

// Next implements the iterator interface.
// Advances the iterator and returns the current row and a bool indicating success.
func (ri *RowIterator) Next(ctx context.Context) (object.Object, bool) {
	// Check if there are more rows
	if !ri.rows.Next() {
		ri.current = nil
		if ri.rows.Err() != nil {
			return object.NewError(ri.rows.Err()), true
		}
		return nil, false
	}

	// Get the values for the current row
	rowValues := make([]interface{}, len(ri.columns))
	for i := range rowValues {
		var s interface{}
		rowValues[i] = &s
	}

	if err := ri.rows.Scan(rowValues...); err != nil {
		ri.current = object.NewError(err)
		return ri.current, true
	}

	// Transform the row into a Risor map object
	row := object.NewMap(make(map[string]object.Object))
	for i := range rowValues {
		val := *(rowValues[i].(*interface{}))
		switch val := val.(type) {
		case []byte:
			row.Set(ri.columns[i], object.NewString(string(val)))
		default:
			row.Set(ri.columns[i], object.FromGoType(val))
		}
	}

	ri.index++
	ri.current = row
	return ri.current, true
}

// Entry returns the current row entry
func (ri *RowIterator) Entry() (object.Object, bool) {
	if ri.current == nil {
		return nil, false
	}
	// Return the current row
	return ri.current, true
}

func NewRowIterator(ctx context.Context, rows *sql.Rows) *RowIterator {
	columns, _ := rows.Columns()
	return &RowIterator{
		ctx:     ctx,
		rows:    rows,
		closed:  make(chan bool),
		columns: columns,
		index:   -1,
	}
}
