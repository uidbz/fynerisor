package csv

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
	if mod.Name().Value() != "csv" {
		t.Errorf("expected module name 'csv', got: %s", mod.Name().Value())
	}
	for _, fn := range []string{"parse", "format", "read", "write"} {
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

func TestParseWithHeader(t *testing.T) {
	ctx := context.Background()
	res, err := parse(ctx, object.NewString("name,age\nAda,36\nBob,40"))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	list, ok := res.(*object.List)
	if !ok {
		t.Fatalf("expected list, got %T", res)
	}
	if len(list.Value()) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(list.Value()))
	}
	row0 := list.Value()[0].(*object.Map)
	if row0.Get("name").(*object.String).Value() != "Ada" {
		t.Errorf("row0.name = %v", row0.Get("name"))
	}
	if row0.Get("age").(*object.String).Value() != "36" {
		t.Errorf("row0.age = %v", row0.Get("age"))
	}
}

func TestParseNoHeader(t *testing.T) {
	ctx := context.Background()
	opts := object.NewMap(map[string]object.Object{"header": object.NewBool(false)})
	res, err := parse(ctx, object.NewString("a,b\nc,d"), opts)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	list := res.(*object.List)
	if len(list.Value()) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(list.Value()))
	}
	row0 := list.Value()[0].(*object.List)
	if row0.Value()[0].(*object.String).Value() != "a" {
		t.Errorf("row0[0] = %v", row0.Value()[0])
	}
}

func TestFormatMapsRoundTrip(t *testing.T) {
	ctx := context.Background()
	parsed, _ := parse(ctx, object.NewString("name,age\nAda,36\nBob,40"))
	out, err := format(ctx, parsed)
	if err != nil {
		t.Fatalf("format error: %v", err)
	}
	// Columns are sorted alphabetically -> age,name
	want := "age,name\n36,Ada\n40,Bob\n"
	if got := out.(*object.String).Value(); got != want {
		t.Errorf("format output = %q, want %q", got, want)
	}
}

func TestFormatWithColumns(t *testing.T) {
	ctx := context.Background()
	parsed, _ := parse(ctx, object.NewString("name,age\nAda,36"))
	cols := object.NewList([]object.Object{object.NewString("name"), object.NewString("age")})
	opts := object.NewMap(map[string]object.Object{"columns": cols})
	out, err := format(ctx, parsed, opts)
	if err != nil {
		t.Fatalf("format error: %v", err)
	}
	want := "name,age\nAda,36\n"
	if got := out.(*object.String).Value(); got != want {
		t.Errorf("format output = %q, want %q", got, want)
	}
}

func TestFormatLists(t *testing.T) {
	ctx := context.Background()
	rows := object.NewList([]object.Object{
		object.NewList([]object.Object{object.NewString("a"), object.NewString("b")}),
		object.NewList([]object.Object{object.NewString("c"), object.NewString("d")}),
	})
	out, err := format(ctx, rows)
	if err != nil {
		t.Fatalf("format error: %v", err)
	}
	want := "a,b\nc,d\n"
	if got := out.(*object.String).Value(); got != want {
		t.Errorf("format output = %q, want %q", got, want)
	}
}
