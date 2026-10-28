// Package convert is the single JSON<->Risor conversion used by the
// modules (issue #34): the http and json modules previously carried
// copy-pasted converters that had drifted apart (missing nil handling,
// %v-stringified unknown types, and every JSON number decoded as a
// float). DecodeJSON uses json.Number so integral JSON numbers map to
// Risor ints: {"id": 3} parses to 3, not 3.0.
package convert

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// ToGo converts a Risor object into a Go value suitable for
// encoding/json. Types with no JSON representation are an error rather
// than a silent %v rendering.
func ToGo(obj object.Object) (any, error) {
	switch v := obj.(type) {
	case nil:
		return nil, nil
	case *object.NilType:
		return nil, nil
	case *object.String:
		return v.Value(), nil
	case *object.Int:
		return v.Value(), nil
	case *object.Float:
		return v.Value(), nil
	case *object.Bool:
		return v.Value(), nil
	case *object.Time:
		return v.Value(), nil
	case *object.Bytes:
		return v.Value(), nil
	case *object.List:
		items := v.Value()
		result := make([]any, len(items))
		for i, item := range items {
			val, err := ToGo(item)
			if err != nil {
				return nil, err
			}
			result[i] = val
		}
		return result, nil
	case *object.Map:
		result := make(map[string]any, len(v.Value()))
		for k, val := range v.Value() {
			converted, err := ToGo(val)
			if err != nil {
				return nil, err
			}
			result[k] = converted
		}
		return result, nil
	default:
		return nil, fmt.Errorf("json: %s has no JSON representation", obj.Type())
	}
}

// FromGo converts a JSON-decoded Go value to a Risor object. json.Number
// (from DecodeJSON's decoder) maps to Int when integral and Float
// otherwise.
func FromGo(data any) (object.Object, error) {
	switch v := data.(type) {
	case nil:
		return object.Nil, nil
	case string:
		return object.NewString(v), nil
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return object.NewInt(i), nil
		}
		f, err := v.Float64()
		if err != nil {
			return nil, fmt.Errorf("json: bad number %q", v.String())
		}
		return object.NewFloat(f), nil
	case float64:
		return object.NewFloat(v), nil
	case int64:
		return object.NewInt(v), nil
	case int:
		return object.NewInt(int64(v)), nil
	case bool:
		return object.NewBool(v), nil
	case []any:
		items := make([]object.Object, len(v))
		for i, item := range v {
			obj, err := FromGo(item)
			if err != nil {
				return nil, err
			}
			items[i] = obj
		}
		return object.NewList(items), nil
	case map[string]any:
		result := make(map[string]object.Object, len(v))
		for k, val := range v {
			obj, err := FromGo(val)
			if err != nil {
				return nil, err
			}
			result[k] = obj
		}
		return object.NewMap(result), nil
	default:
		return nil, fmt.Errorf("json: cannot convert Go type %T", data)
	}
}

// DecodeJSON parses JSON into a Risor object with integral numbers as
// ints (the UseNumber path that fixes the float64-default).
func DecodeJSON(data []byte) (object.Object, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var decoded any
	if err := dec.Decode(&decoded); err != nil {
		return nil, err
	}
	return FromGo(decoded)
}

// MarshalJSON encodes a Risor object, optionally indented.
func MarshalJSON(obj object.Object, indent string) ([]byte, error) {
	data, err := ToGo(obj)
	if err != nil {
		return nil, err
	}
	if indent != "" {
		return json.MarshalIndent(data, "", indent)
	}
	return json.Marshal(data)
}
