// Package json provides JSON encoding and decoding functions for Risor scripts.
package json

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// Module returns the json module for risor.
func Module() *object.Module {
	return object.NewBuiltinsModule("json", map[string]object.Object{
		"parse":          object.NewBuiltin("json.parse", parse),
		"marshal":        object.NewBuiltin("json.marshal", marshal),
		"marshal_indent": object.NewBuiltin("json.marshal_indent", marshalIndent),
		"valid":          object.NewBuiltin("json.valid", valid),
		"read":           object.NewBuiltin("json.read", read),
		"write":          object.NewBuiltin("json.write", write),
	})
}

func parse(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return object.Errorf("json.parse: expected 1 argument, got %d", len(args)), nil
	}
	text, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}
	var data any
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		return object.NewError(err), nil
	}
	return interfaceToObject(data)
}

func marshal(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return object.Errorf("json.marshal: expected 1 argument, got %d", len(args)), nil
	}
	data, err := objectToInterface(args[0])
	if err != nil {
		return nil, err
	}
	out, err := json.Marshal(data)
	if err != nil {
		return object.NewError(err), nil
	}
	return object.NewString(string(out)), nil
}

func marshalIndent(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 1 || len(args) > 2 {
		return object.Errorf("json.marshal_indent: expected 1 or 2 arguments, got %d", len(args)), nil
	}
	indent := "  "
	if len(args) == 2 {
		s, err := object.AsString(args[1])
		if err != nil {
			return nil, err
		}
		indent = s
	}
	data, err := objectToInterface(args[0])
	if err != nil {
		return nil, err
	}
	out, err := json.MarshalIndent(data, "", indent)
	if err != nil {
		return object.NewError(err), nil
	}
	return object.NewString(string(out)), nil
}

func valid(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return object.Errorf("json.valid: expected 1 argument, got %d", len(args)), nil
	}
	text, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}
	return object.NewBool(json.Valid([]byte(text))), nil
}

func read(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return object.Errorf("json.read: expected 1 argument, got %d", len(args)), nil
	}
	path, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return object.NewError(err), nil
	}
	var data any
	if err := json.Unmarshal(content, &data); err != nil {
		return object.NewError(err), nil
	}
	return interfaceToObject(data)
}

func write(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 2 || len(args) > 3 {
		return object.Errorf("json.write: expected 2 or 3 arguments, got %d", len(args)), nil
	}
	path, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}
	data, err := objectToInterface(args[1])
	if err != nil {
		return nil, err
	}
	var out []byte
	if len(args) == 3 {
		indent, err := object.AsString(args[2])
		if err != nil {
			return nil, err
		}
		out, err = json.MarshalIndent(data, "", indent)
		if err != nil {
			return object.NewError(err), nil
		}
	} else {
		out, err = json.Marshal(data)
		if err != nil {
			return object.NewError(err), nil
		}
	}
	if err := os.WriteFile(path, out, 0644); err != nil {
		return object.NewError(err), nil
	}
	return object.Nil, nil
}

// objectToInterface converts a Risor object to a Go interface for JSON marshaling.
func objectToInterface(obj object.Object) (any, error) {
	switch v := obj.(type) {
	case *object.String:
		return v.Value(), nil
	case *object.Int:
		return v.Value(), nil
	case *object.Float:
		return v.Value(), nil
	case *object.Bool:
		return v.Value(), nil
	case *object.NilType:
		return nil, nil
	case *object.List:
		result := make([]any, len(v.Value()))
		for i, item := range v.Value() {
			val, err := objectToInterface(item)
			if err != nil {
				return nil, err
			}
			result[i] = val
		}
		return result, nil
	case *object.Map:
		result := make(map[string]any)
		for k, val := range v.Value() {
			converted, err := objectToInterface(val)
			if err != nil {
				return nil, err
			}
			result[k] = converted
		}
		return result, nil
	default:
		return fmt.Sprintf("%v", obj), nil
	}
}

// interfaceToObject converts a Go interface (from JSON) to a Risor object.
func interfaceToObject(data any) (object.Object, error) {
	switch v := data.(type) {
	case string:
		return object.NewString(v), nil
	case float64:
		return object.NewFloat(v), nil
	case int64:
		return object.NewInt(v), nil
	case int:
		return object.NewInt(int64(v)), nil
	case bool:
		return object.NewBool(v), nil
	case nil:
		return object.Nil, nil
	case []any:
		items := make([]object.Object, len(v))
		for i, item := range v {
			obj, err := interfaceToObject(item)
			if err != nil {
				return nil, err
			}
			items[i] = obj
		}
		return object.NewList(items), nil
	case map[string]any:
		objMap := make(map[string]object.Object)
		for k, val := range v {
			obj, err := interfaceToObject(val)
			if err != nil {
				return nil, err
			}
			objMap[k] = obj
		}
		return object.NewMap(objMap), nil
	default:
		return object.NewString(fmt.Sprintf("%v", v)), nil
	}
}
