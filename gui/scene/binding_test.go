package scene

import (
	"context"
	"reflect"
	"testing"

	scenedoc "github.com/uidbz/fynerisor/scene"
)

func TestParseTemplate(t *testing.T) {
	lit := func(s string) segment { return segment{text: s} }
	ex := func(s string) segment { return segment{text: s, expr: true} }
	cases := []struct {
		in   string
		want []segment
	}{
		{"plain", []segment{lit("plain")}},
		{"${a}", []segment{ex("a")}},
		{"x ${a} y ${b}", []segment{lit("x "), ex("a"), lit(" y "), ex("b")}},
		// Braces nest inside an expression (#66: the old [^{}]+ regexp
		// rejected these).
		{"${ ({n: 1}).n }", []segment{ex(" ({n: 1}).n ")}},
		{`${"}" + a}!`, []segment{ex(`"}" + a`), lit("!")}},
		{"${'{'}", []segment{ex("'{'")}},
		// Unterminated and empty templates stay literal.
		{"${a", []segment{lit("${a")}},
		{"${}", []segment{lit("${}")}},
		{"a ${} ${b}", []segment{lit("a ${} "), ex("b")}},
	}
	for _, c := range cases {
		if got := parseTemplate(c.in); !reflect.DeepEqual(got, c.want) {
			t.Errorf("parseTemplate(%q) = %#v, want %#v", c.in, got, c.want)
		}
	}
}

// Binding expressions see Risor's builtins (#66: risor.WithEnv replaced them,
// so ${len(items)} failed with undefined variable "len").
func TestBindingBuiltins(t *testing.T) {
	var c exprCache
	env := newBindEnv(map[string]any{"items": []any{"a", "b", "c"}, "name": "bob"})
	got, err := c.resolve(context.Background(), parseTemplate("${len(items)} for ${name.to_upper()}"), scenedoc.PropString, env)
	if err != nil {
		t.Fatal(err)
	}
	if got != "3 for BOB" {
		t.Fatalf("got %q", got)
	}
	// A values key shadows a builtin of the same name.
	env = newBindEnv(map[string]any{"len": 7})
	got, err = c.resolve(context.Background(), parseTemplate("${len}"), scenedoc.PropInt, env)
	if err != nil || got != int64(7) {
		t.Fatalf("shadowed len = %#v, %v", got, err)
	}
}

// Expressions are compiled once per (expression, global-name set), not on
// every refresh (#66), and new values with the same keys reuse the code.
func TestBindingCompileCache(t *testing.T) {
	var c exprCache
	ctx := context.Background()
	segs := parseTemplate("${count + 1}")
	for i := 0; i < 3; i++ {
		got, err := c.resolve(ctx, segs, scenedoc.PropInt, newBindEnv(map[string]any{"count": int64(i)}))
		if err != nil {
			t.Fatal(err)
		}
		if got != int64(i+1) {
			t.Fatalf("refresh %d: got %#v", i, got)
		}
	}
	if len(c.codes) != 1 {
		t.Fatalf("cache has %d entries after 3 refreshes, want 1", len(c.codes))
	}
	// A different key set compiles separately (bytecode binds globals by
	// index) and still evaluates correctly.
	got, err := c.resolve(ctx, segs, scenedoc.PropInt, newBindEnv(map[string]any{"count": int64(10), "other": "x"}))
	if err != nil || got != int64(11) {
		t.Fatalf("new key set: %#v, %v", got, err)
	}
	if len(c.codes) != 2 {
		t.Fatalf("cache has %d entries, want 2", len(c.codes))
	}
}
