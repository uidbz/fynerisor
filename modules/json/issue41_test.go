package json

import (
	"context"
	"strings"
	"testing"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// This file covers issue #41: json is a pure transform module. format is
// the encode half of parse/format so pipelines compose as
// json.parse(io.read(p)) / io.write(p, json.format(v)).

func TestJSONFormat(t *testing.T) {
	rows := object.NewMap(map[string]object.Object{
		"name": object.NewString("Ada"),
		"age":  object.NewInt(36),
	})

	res, err := format(context.Background(), rows)
	if err != nil {
		t.Fatal(err)
	}
	if res.(*object.String).Value() != `{"age":36,"name":"Ada"}` {
		t.Errorf("format = %q", res.(*object.String).Value())
	}

	res, err = format(context.Background(), rows, object.NewString("  "))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.(*object.String).Value(), "\n  ") {
		t.Errorf("format(v, indent) not indented: %q", res.(*object.String).Value())
	}

	if _, err := format(context.Background()); err == nil {
		t.Error("format() with no args should raise")
	}
}
