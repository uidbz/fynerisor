package json

import (
	"context"
	"testing"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

func TestModule(t *testing.T) {
	mod := Module()
	if mod == nil {
		t.Fatal("Module() returned nil")
	}
	if mod.Name().Value() != "json" {
		t.Errorf("expected module name 'json', got: %s", mod.Name().Value())
	}
	for _, fn := range []string{"parse", "marshal", "marshal_indent", "valid", "read", "write"} {
		obj, ok := mod.GetAttr(fn)
		if !ok {
			t.Errorf("expected '%s' function in module", fn)
			continue
		}
		if obj.Type() != "builtin" {
			t.Errorf("expected builtin type for %s, got: %s", fn, obj.Type())
		}
	}
}

func TestParseMarshalRoundTrip(t *testing.T) {
	ctx := context.Background()

	parsed, err := parse(ctx, object.NewString(`{"name":"Ada","age":36,"tags":["x","y"]}`))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	m, ok := parsed.(*object.Map)
	if !ok {
		t.Fatalf("expected map, got %T", parsed)
	}
	if got := m.Get("name"); got.(*object.String).Value() != "Ada" {
		t.Errorf("name = %v", got)
	}

	out, err := marshal(ctx, parsed)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	reparsed, err := parse(ctx, out)
	if err != nil {
		t.Fatalf("reparse error: %v", err)
	}
	if _, ok := reparsed.(*object.Map); !ok {
		t.Fatalf("expected map after round trip, got %T", reparsed)
	}
}

func TestValid(t *testing.T) {
	ctx := context.Background()
	res, _ := valid(ctx, object.NewString(`{"a":1}`))
	if !res.(*object.Bool).Value() {
		t.Error("expected valid JSON")
	}
	res, _ = valid(ctx, object.NewString(`{bad`))
	if res.(*object.Bool).Value() {
		t.Error("expected invalid JSON")
	}
}

func TestParseInvalidRaises(t *testing.T) {
	// Issue #30 convention: runtime failures raise (catchable with try),
	// they are not error values the script silently continues past.
	ctx := context.Background()
	_, err := parse(ctx, object.NewString(`{not json`))
	if err == nil {
		t.Fatal("expected a raised error for invalid JSON")
	}
}

func TestParseKeepsIntegralNumbersInts(t *testing.T) {
	// Issue #34: JSON numbers used to decode as float64, so {"id": 3}
	// became 3.0. Integral numbers must stay ints.
	ctx := context.Background()
	parsed, err := parse(ctx, object.NewString(`{"id": 3, "pi": 3.5}`))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	m := parsed.(*object.Map)
	if id, ok := m.Get("id").(*object.Int); !ok || id.Value() != 3 {
		t.Errorf("id = %s (%v), want int 3", m.Get("id").Type(), m.Get("id"))
	}
	if pi, ok := m.Get("pi").(*object.Float); !ok || pi.Value() != 3.5 {
		t.Errorf("pi = %v, want float 3.5", m.Get("pi"))
	}
}

func TestMarshalUnconvertibleRaises(t *testing.T) {
	// Issue #34: unknown types used to be stringified with %v.
	ctx := context.Background()
	_, err := marshal(ctx, object.NewBuiltin("x", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		return object.Nil, nil
	}))
	if err == nil {
		t.Fatal("expected a raised error marshaling a builtin")
	}
}
