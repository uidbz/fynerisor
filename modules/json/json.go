// Package json provides JSON encoding and decoding functions for Risor scripts.
// Conversion is shared with the http module through internal/convert (issue #34):
// integral JSON numbers decode to ints, and types without a JSON representation
// are an error instead of a %v rendering.
package json

import (
	"context"
	"encoding/json"
	"os"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"

	"github.com/uidbz/fynerisor/internal/convert"
)

// Module returns the json module for risor. json is a pure transform
// module (issue #41): parse/format move between strings and values and
// compose with io/http/exec. read/write remain as thin deprecated sugar
// over io behaviour; prefer json.parse(io.read(p)) /
// io.write(p, json.format(v)).
func Module() *object.Module {
	return object.NewBuiltinsModule("json", map[string]object.Object{
		"parse":          object.NewBuiltin("json.parse", parse),
		"marshal":        object.NewBuiltin("json.marshal", marshal),
		"format":         object.NewBuiltin("json.format", format),
		"marshal_indent": object.NewBuiltin("json.marshal_indent", marshalIndent),
		"valid":          object.NewBuiltin("json.valid", valid),
		// Deprecated sugar (issue #41): use json.parse(io.read(p)) and
		// io.write(p, json.format(v)) so data can come from and go to
		// any stream, not only files.
		"read":  object.NewBuiltin("json.read", read),
		"write": object.NewBuiltin("json.write", write),
	})
}

func parse(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return nil, object.Errorf("json.parse: expected 1 argument, got %d", len(args))
	}
	text, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}
	result, err := convert.DecodeJSON([]byte(text))
	if err != nil {
		return nil, err
	}
	return result, nil
}

func marshal(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return nil, object.Errorf("json.marshal: expected 1 argument, got %d", len(args))
	}
	out, err := convert.MarshalJSON(args[0], "")
	if err != nil {
		return nil, err
	}
	return object.NewString(string(out)), nil
}

func marshalIndent(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, object.Errorf("json.marshal_indent: expected 1 or 2 arguments, got %d", len(args))
	}
	indent := "  "
	if len(args) == 2 {
		s, err := object.AsString(args[1])
		if err != nil {
			return nil, err
		}
		indent = s
	}
	out, err := convert.MarshalJSON(args[0], indent)
	if err != nil {
		return nil, err
	}
	return object.NewString(string(out)), nil
}

func valid(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return nil, object.Errorf("json.valid: expected 1 argument, got %d", len(args))
	}
	text, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}
	return object.NewBool(json.Valid([]byte(text))), nil
}

// format is the encode half of the transform pair (issue #41):
// format(value) encodes compactly, format(value, indent) with an
// indent string — equivalent to marshal / marshal_indent.
func format(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, object.Errorf("json.format: expected 1 or 2 arguments, got %d", len(args))
	}
	indent := ""
	if len(args) == 2 {
		s, err := object.AsString(args[1])
		if err != nil {
			return nil, err
		}
		indent = s
	}
	out, err := convert.MarshalJSON(args[0], indent)
	if err != nil {
		return nil, err
	}
	return object.NewString(string(out)), nil
}

func read(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return nil, object.Errorf("json.read: expected 1 argument, got %d", len(args))
	}
	path, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	result, err := convert.DecodeJSON(content)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func write(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, object.Errorf("json.write: expected 2 or 3 arguments, got %d", len(args))
	}
	path, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}
	indent := ""
	if len(args) == 3 {
		indent, err = object.AsString(args[2])
		if err != nil {
			return nil, err
		}
	}
	out, err := convert.MarshalJSON(args[1], indent)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, out, 0644); err != nil {
		return nil, err
	}
	return object.Nil, nil
}
