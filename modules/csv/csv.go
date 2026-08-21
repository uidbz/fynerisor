// Package csv provides CSV encoding and decoding functions for Risor scripts.
package csv

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// Module returns the csv module for risor.
func Module() *object.Module {
	return object.NewBuiltinsModule("csv", map[string]object.Object{
		"parse":  object.NewBuiltin("csv.parse", parse),
		"format": object.NewBuiltin("csv.format", format),
		"read":   object.NewBuiltin("csv.read", read),
		"write":  object.NewBuiltin("csv.write", write),
	})
}

type options struct {
	header    bool
	delimiter rune
	columns   []string
}

// parseOptions reads an optional trailing options map. When present the map may
// contain: header (bool), delimiter (1-char string), columns (list of strings).
func parseOptions(arg object.Object) (options, error) {
	opts := options{header: true, delimiter: ','}
	if arg == nil {
		return opts, nil
	}
	m, ok := arg.(*object.Map)
	if !ok {
		return opts, fmt.Errorf("expected options map, got %s", arg.Type())
	}
	if v := m.Get("header"); v != object.Nil {
		b, ok := v.(*object.Bool)
		if !ok {
			return opts, fmt.Errorf("option 'header' must be a bool")
		}
		opts.header = b.Value()
	}
	if v := m.Get("delimiter"); v != object.Nil {
		s, ok := v.(*object.String)
		if !ok {
			return opts, fmt.Errorf("option 'delimiter' must be a string")
		}
		r := []rune(s.Value())
		if len(r) != 1 {
			return opts, fmt.Errorf("option 'delimiter' must be a single character")
		}
		opts.delimiter = r[0]
	}
	if v := m.Get("columns"); v != object.Nil {
		l, ok := v.(*object.List)
		if !ok {
			return opts, fmt.Errorf("option 'columns' must be a list of strings")
		}
		for _, item := range l.Value() {
			s, err := object.AsString(item)
			if err != nil {
				return opts, err
			}
			opts.columns = append(opts.columns, s)
		}
	}
	return opts, nil
}

func optArg(args []object.Object, idx int) object.Object {
	if len(args) > idx {
		return args[idx]
	}
	return nil
}

// parseCSV parses CSV text into a Risor object using the given options.
func parseCSV(text string, opts options) (object.Object, error) {
	r := csv.NewReader(strings.NewReader(text))
	r.Comma = opts.delimiter
	r.FieldsPerRecord = -1
	records, err := r.ReadAll()
	if err != nil {
		return object.NewError(err), nil
	}

	if !opts.header {
		rows := make([]object.Object, len(records))
		for i, rec := range records {
			cells := make([]object.Object, len(rec))
			for j, c := range rec {
				cells[j] = object.NewString(c)
			}
			rows[i] = object.NewList(cells)
		}
		return object.NewList(rows), nil
	}

	if len(records) == 0 {
		return object.NewList(nil), nil
	}
	headers := records[0]
	rows := make([]object.Object, 0, len(records)-1)
	for _, rec := range records[1:] {
		m := make(map[string]object.Object, len(headers))
		for j, h := range headers {
			if j < len(rec) {
				m[h] = object.NewString(rec[j])
			} else {
				m[h] = object.NewString("")
			}
		}
		rows = append(rows, object.NewMap(m))
	}
	return object.NewList(rows), nil
}

// formatCSV encodes a Risor list of rows (each a list or a map) into CSV text.
func formatCSV(rowsObj object.Object, opts options) (object.Object, error) {
	list, err := object.AsList(rowsObj)
	if err != nil {
		return nil, err
	}
	rows := list.Value()

	var sb strings.Builder
	w := csv.NewWriter(&sb)
	w.Comma = opts.delimiter

	if len(rows) == 0 {
		w.Flush()
		if err := w.Error(); err != nil {
			return object.NewError(err), nil
		}
		return object.NewString(sb.String()), nil
	}

	switch rows[0].(type) {
	case *object.Map:
		columns := opts.columns
		if len(columns) == 0 {
			seen := make(map[string]bool)
			for _, r := range rows {
				m, ok := r.(*object.Map)
				if !ok {
					return object.Errorf("csv.format: mixed row types (expected map)"), nil
				}
				for k := range m.Value() {
					seen[k] = true
				}
			}
			for k := range seen {
				columns = append(columns, k)
			}
			sort.Strings(columns)
		}
		if opts.header {
			if err := w.Write(columns); err != nil {
				return object.NewError(err), nil
			}
		}
		for _, r := range rows {
			m, ok := r.(*object.Map)
			if !ok {
				return object.Errorf("csv.format: mixed row types (expected map)"), nil
			}
			rec := make([]string, len(columns))
			for i, col := range columns {
				v := m.Get(col)
				if v == object.Nil {
					rec[i] = ""
					continue
				}
				s, err := object.AsString(v)
				if err != nil {
					return nil, err
				}
				rec[i] = s
			}
			if err := w.Write(rec); err != nil {
				return object.NewError(err), nil
			}
		}
	default:
		for _, r := range rows {
			cells, err := object.AsList(r)
			if err != nil {
				return object.Errorf("csv.format: mixed row types (expected list)"), nil
			}
			rec := make([]string, len(cells.Value()))
			for i, c := range cells.Value() {
				s, err := object.AsString(c)
				if err != nil {
					return nil, err
				}
				rec[i] = s
			}
			if err := w.Write(rec); err != nil {
				return object.NewError(err), nil
			}
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return object.NewError(err), nil
	}
	return object.NewString(sb.String()), nil
}

func parse(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 1 || len(args) > 2 {
		return object.Errorf("csv.parse: expected 1 or 2 arguments, got %d", len(args)), nil
	}
	text, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}
	opts, err := parseOptions(optArg(args, 1))
	if err != nil {
		return object.NewError(err), nil
	}
	return parseCSV(text, opts)
}

func format(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 1 || len(args) > 2 {
		return object.Errorf("csv.format: expected 1 or 2 arguments, got %d", len(args)), nil
	}
	opts, err := parseOptions(optArg(args, 1))
	if err != nil {
		return object.NewError(err), nil
	}
	return formatCSV(args[0], opts)
}

func read(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 1 || len(args) > 2 {
		return object.Errorf("csv.read: expected 1 or 2 arguments, got %d", len(args)), nil
	}
	path, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}
	opts, err := parseOptions(optArg(args, 1))
	if err != nil {
		return object.NewError(err), nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return object.NewError(err), nil
	}
	return parseCSV(string(content), opts)
}

func write(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 2 || len(args) > 3 {
		return object.Errorf("csv.write: expected 2 or 3 arguments, got %d", len(args)), nil
	}
	path, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}
	opts, err := parseOptions(optArg(args, 2))
	if err != nil {
		return object.NewError(err), nil
	}
	result, err := formatCSV(args[1], opts)
	if err != nil {
		return nil, err
	}
	if errObj, ok := result.(*object.Error); ok {
		return errObj, nil
	}
	text, err := object.AsString(result)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, []byte(text), 0644); err != nil {
		return object.NewError(err), nil
	}
	return object.Nil, nil
}
