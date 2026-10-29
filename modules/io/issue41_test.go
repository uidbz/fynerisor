package io

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// This file covers issue #41: io is the single byte/text stream module.
// read/lines accept a path, bytes, or the io.stdin handle; write accepts
// a path or a writable handle; copy is the one file-copy (mode
// preserving) with cp kept as its deprecated alias.

func ioBuiltin(t *testing.T, name string) *object.Builtin {
	t.Helper()
	obj, ok := NewIO().GetAttr(name)
	if !ok {
		t.Fatalf("io.%s missing", name)
	}
	b, ok := obj.(*object.Builtin)
	if !ok {
		t.Fatalf("io.%s is %T, want *object.Builtin", name, obj)
	}
	return b
}

func TestIOReadGeneralized(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(path, []byte("from file\n"), 0644); err != nil {
		t.Fatal(err)
	}
	read := ioBuiltin(t, "read")

	res, err := read.Call(context.Background(), object.NewString(path))
	if err != nil || res.(*object.String).Value() != "from file\n" {
		t.Errorf("read(path) = %q, %v", res, err)
	}

	res, err = read.Call(context.Background(), object.NewBytes([]byte("as bytes")))
	if err != nil || res.(*object.String).Value() != "as bytes" {
		t.Errorf("read(bytes) = %q, %v", res, err)
	}

	src := &memStreams{in: "from stdin\n"}
	ctx := ContextWithStreams(context.Background(), src)
	res, err = read.Call(ctx, NewStdin())
	if err != nil || res.(*object.String).Value() != "from stdin\n" {
		t.Errorf("read(io.stdin) = %q, %v", res, err)
	}

	if _, err := read.Call(context.Background()); err == nil {
		t.Error("read() with no args should raise")
	}
}

func TestIOLinesGeneralized(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(path, []byte("a\r\nb\n"), 0644); err != nil {
		t.Fatal(err)
	}
	lines := ioBuiltin(t, "lines")

	res, err := lines.Call(context.Background(), object.NewString(path))
	if err != nil {
		t.Fatal(err)
	}
	list, ok := res.(*object.List)
	if !ok || len(list.Value()) != 2 {
		t.Fatalf("lines(path) = %s", res.Inspect())
	}
	if list.Value()[1].(*object.String).Value() != "b" {
		t.Errorf("CRLF not normalized: %s", res.Inspect())
	}

	src := &memStreams{in: "x\ny\nz"}
	ctx := ContextWithStreams(context.Background(), src)
	res, err = lines.Call(ctx, NewStdin())
	if err != nil {
		t.Fatal(err)
	}
	if len(res.(*object.List).Value()) != 3 {
		t.Errorf("lines(io.stdin) = %s", res.Inspect())
	}
}

func TestIOWriteGeneralized(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")
	write := ioBuiltin(t, "write")

	res, err := write.Call(context.Background(), object.NewString(path), object.NewString("data"))
	if err != nil {
		t.Fatal(err)
	}
	if res.Inspect() != "4" {
		t.Errorf("write(path) returned %s, want byte count 4", res.Inspect())
	}
	if data, _ := os.ReadFile(path); string(data) != "data" {
		t.Errorf("file = %q", data)
	}

	src := &memStreams{}
	ctx := ContextWithStreams(context.Background(), src)
	res, err = write.Call(ctx, NewStdout(), object.NewString("to stdout"))
	if err != nil {
		t.Fatal(err)
	}
	if res.Inspect() != "9" || src.out.String() != "to stdout" {
		t.Errorf("write(io.stdout) = %s, buffer %q", res.Inspect(), src.out.String())
	}
}

func TestIOCopyPreservesMode(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	if err := os.WriteFile(src, []byte("payload"), 0600); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"copy", "cp"} {
		target := dst + name
		if _, err := ioBuiltin(t, name).Call(context.Background(), object.NewString(src), object.NewString(target)); err != nil {
			t.Fatalf("io.%s: %v", name, err)
		}
		info, err := os.Stat(target)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0600 {
			t.Errorf("io.%s did not preserve the source mode: %v", name, info.Mode())
		}
	}
}
