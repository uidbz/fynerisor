package io

import (
	"bytes"
	"context"
	stdio "io"
	"strings"
	"testing"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// This file covers issue #42: the stream builtins must resolve to the
// StreamSource carried in the eval context — per-runtime streams, no
// process-global hijacking.

type memStreams struct {
	in   string
	out  bytes.Buffer
	errs bytes.Buffer
}

func (m *memStreams) Stdin() stdio.Reader  { return strings.NewReader(m.in) }
func (m *memStreams) Stdout() stdio.Writer { return &m.out }
func (m *memStreams) Stderr() stdio.Writer { return &m.errs }

// callAttr invokes a stream object's attribute as a builtin.
func callAttr(obj object.Object, attr string, ctx context.Context, args ...object.Object) (object.Object, error) {
	a, ok := obj.GetAttr(attr)
	if !ok {
		panic("missing attr " + attr)
	}
	b, ok := a.(*object.Builtin)
	if !ok {
		panic("attr is not builtin")
	}
	fn := b.Value()
	return fn(ctx, args...)
}

func TestStdinLinesFromContext(t *testing.T) {
	src := &memStreams{in: "a,b\r\nc\n"}
	ctx := ContextWithStreams(context.Background(), src)

	res, err := callAttr(NewStdin(), "lines", ctx)
	if err != nil {
		t.Fatalf("lines: %v", err)
	}
	want := `["a,b", "c"]`
	if got := res.Inspect(); got != want {
		t.Errorf("lines = %s, want %s", got, want)
	}

	res, err = callAttr(NewStdin(), "read", ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got, _ := object.AsString(res); got != "a,b\r\nc\n" {
		t.Errorf("read = %q, want %q", got, "a,b\r\nc\n")
	}
}

func TestOutputStreamsFromContext(t *testing.T) {
	src := &memStreams{}
	ctx := ContextWithStreams(context.Background(), src)

	res, err := callAttr(NewStdout(), "write", ctx, object.NewString("to-out"))
	if err != nil {
		t.Fatalf("stdout.write: %v", err)
	}
	if n, _ := object.AsInt(res); n != 6 {
		t.Errorf("write returned %d, want 6", n)
	}
	if _, err := callAttr(NewStderr(), "println", ctx, object.NewString("to-err")); err != nil {
		t.Fatalf("stderr.println: %v", err)
	}

	if src.out.String() != "to-out" {
		t.Errorf("stdout buffer = %q, want %q", src.out.String(), "to-out")
	}
	if src.errs.String() != "to-err\n" {
		t.Errorf("stderr buffer = %q, want %q", src.errs.String(), "to-err\n")
	}
}

// Two runtimes bound to one context chain never see each other's streams.
func TestStreamsArePerContext(t *testing.T) {
	srcA := &memStreams{in: "A"}
	srcB := &memStreams{in: "B"}
	ctxA := ContextWithStreams(context.Background(), srcA)
	ctxB := ContextWithStreams(context.Background(), srcB)

	resA, err := callAttr(NewStdin(), "read", ctxA)
	if err != nil {
		t.Fatal(err)
	}
	resB, err := callAttr(NewStdin(), "read", ctxB)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := object.AsString(resA); got != "A" {
		t.Errorf("context A stdin = %q, want A", got)
	}
	if got, _ := object.AsString(resB); got != "B" {
		t.Errorf("context B stdin = %q, want B", got)
	}
}

// Argument errors raise (issue #30 convention).
func TestStreamArgsErrorsRaise(t *testing.T) {
	ctx := ContextWithStreams(context.Background(), &memStreams{})
	if _, err := callAttr(NewStdin(), "read", ctx, object.NewInt(1)); err == nil {
		t.Error("io.stdin.read with args should raise")
	}
	if _, err := callAttr(NewStdout(), "write", ctx); err == nil {
		t.Error("io.stdout.write without args should raise")
	}
}
