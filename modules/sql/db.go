package sql

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/microsoft/go-mssqldb"
	"github.com/xo/dburl"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

const DBConnType object.Type = "sql.conn"

type DB struct {
	conn   *sql.DB
	once   sync.Once
	closed chan bool
	stream bool // Whether to use streaming for query results
}

func (db *DB) Type() object.Type {
	return DBConnType
}

func (db *DB) Inspect() string {
	return "sql.conn"
}

func (db *DB) Interface() interface{} {
	return db.conn
}

func (db *DB) IsTruthy() bool {
	return db.conn != nil
}

func (db *DB) Cost() int {
	return 8
}

func (db *DB) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal sql.conn")
}

func (db *DB) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for sql.conn: %v", opType), nil
}

func (db *DB) Equals(other object.Object) bool {
	if other.Type() != DBConnType {
		return false
	}
	return db.conn == other.(*DB).conn
}

func (db *DB) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: sql.conn object has no attribute %q", name)
}

func (db *DB) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "query":
		return object.NewBuiltin("sql.query", db.Query), true
	case "exec":
		return object.NewBuiltin("sql.exec", db.Exec), true
	case "close":
		return object.NewBuiltin("sql.close", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, fmt.Errorf("sql.close: expected 0 arguments, got %d", len(args))
			}
			if err := db.Close(); err != nil {
				return nil, err
			}
			return object.Nil, nil
		}), true
	}
	return nil, false
}

func (db *DB) Attrs() []object.AttrSpec {
	return []object.AttrSpec{
		{Name: "query"},
		{Name: "exec"},
		{Name: "close"},
	}
}

func (db *DB) Exec(ctx context.Context, args ...object.Object) (object.Object, error) {
	numArgs := len(args)
	if numArgs < 1 {
		return nil, fmt.Errorf("sql.exec() requires at least one argument")
	}

	query, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	// Build list of query args as their Go types
	var queryArgs []interface{}
	for _, queryArg := range args[1:] {
		queryArgs = append(queryArgs, queryArg.Interface())
	}
	_, execErr := db.conn.Exec(query, queryArgs...)
	if execErr != nil {
		return nil, execErr
	}

	return object.Nil, nil
}

func (db *DB) Query(ctx context.Context, args ...object.Object) (object.Object, error) {
	numArgs := len(args)
	if numArgs < 1 {
		return nil, fmt.Errorf("sql.query() requires at least one argument")
	}

	query, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	// Build list of query args as their Go types
	var queryArgs []interface{}
	for _, queryArg := range args[1:] {
		queryArgs = append(queryArgs, queryArg.Interface())
	}

	// Start the query
	rows, queryErr := db.conn.Query(query, queryArgs...)
	if queryErr != nil || rows.Err() != nil {
		return nil, fmt.Errorf("failed to query db: %w", queryErr)
	}

	// If streaming is enabled, return a row iterator
	if db.stream {
		return NewRowIterator(ctx, rows), nil
	}

	// Otherwise, process all rows and return a list (original behavior)
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	rowList := object.NewList(make([]object.Object, 0))
	for rows.Next() {
		rowValues := make([]interface{}, len(columns))
		for i := range rowValues {
			var s interface{}
			rowValues[i] = &s
		}
		if err := rows.Scan(rowValues...); err != nil {
			return nil, err
		}

		row := object.NewMap(make(map[string]object.Object))
		for i := range rowValues {
			val := *(rowValues[i].(*interface{}))
			switch val := val.(type) {
			case []byte:
				row.Set(columns[i], object.NewString(string(val)))
			default:
				row.Set(columns[i], object.FromGoType(val))
			}
		}

		rowList.Append(row)
	}

	return rowList, nil
}

func (db *DB) Close() error {
	var err error
	db.once.Do(func() {
		err = db.conn.Close()
		close(db.closed)
	})
	return err
}

func (db *DB) waitToClose(ctx context.Context) {
	go func() {
		select {
		case <-db.closed:
		case <-ctx.Done():
			_ = db.conn.Close()
		}
	}()
}

func New(ctx context.Context, connection string, stream bool) (*DB, error) {
	db, err := dburl.Open(connection)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	obj := &DB{
		conn:   db,
		closed: make(chan bool),
		stream: stream,
	}
	obj.waitToClose(ctx)
	return obj, nil
}

func NewFromDB(ctx context.Context, db *sql.DB) *DB {
	return &DB{
		conn:   db,
		closed: make(chan bool),
		stream: true,
	}
}
