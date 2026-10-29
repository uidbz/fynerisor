package io

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

const IOType object.Type = "io"

type IO struct{}

func NewIO() *IO {
	return &IO{}
}

func (i *IO) Type() object.Type {
	return IOType
}

func (i *IO) Inspect() string {
	return "io"
}

func (i *IO) String() string {
	return "io"
}

func (i *IO) Interface() interface{} {
	return i
}

func (i *IO) Equals(other object.Object) bool {
	_, ok := other.(*IO)
	return ok
}

func (i *IO) IsTruthy() bool {
	return true
}

func (i *IO) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for io: %v", opType)
}

func (i *IO) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "read":
		return object.NewBuiltin("io.read", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.NewArgsError("io.read", 1, len(args))
			}
			s, err := readSource(ctx, args[0])
			if err != nil {
				return nil, err
			}
			return object.NewString(s), nil
		}), true

	case "lines":
		return object.NewBuiltin("io.lines", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.NewArgsError("io.lines", 1, len(args))
			}
			s, err := readSource(ctx, args[0])
			if err != nil {
				return nil, err
			}
			return splitLines(s), nil
		}), true

	case "read_all":
		return object.NewBuiltin("io.read_all", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.NewArgsError("io.read_all", 1, len(args))
			}
			s, err := readSource(ctx, args[0])
			if err != nil {
				return nil, err
			}
			return object.NewString(s), nil
		}), true

	case "write":
		return object.NewBuiltin("io.write", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return nil, object.NewArgsError("io.write", 2, len(args))
			}
			data, err := object.AsString(args[1])
			if err != nil {
				return nil, err
			}
			n, err := writeSink(ctx, args[0], data)
			if err != nil {
				return nil, err
			}
			return object.NewInt(n), nil
		}), true

	case "copy":
		return object.NewBuiltin("io.copy", copyFn("io.copy")), true

	case "cp":
		// Deprecated alias for io.copy (issue #41); io.copy preserves the
		// source mode, cp never did.
		return object.NewBuiltin("io.cp", copyFn("io.cp")), true

	case "stdin":
		return NewStdin(), true
	case "stdout":
		return NewStdout(), true
	case "stderr":
		return NewStderr(), true
	}
	return nil, false
}

func (i *IO) SetAttr(name string, value object.Object) error {
	return object.Errorf("type error: io object attributes are read-only")
}

// readSource resolves the argument of io.read / io.lines: a file path, a
// bytes value, or the readable io.stdin handle (issue #41 — streams and
// paths share one interface so pipelines compose).
func readSource(ctx context.Context, arg object.Object) (string, error) {
	if _, ok := arg.(*StdinObject); ok {
		return readStream(ctx)
	}
	if b, ok := arg.(*object.Bytes); ok {
		return string(b.Value()), nil
	}
	path, err := object.AsString(arg)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// splitLines turns stream content into a list of lines (trailing newline
// dropped, CRLF normalized), shared by io.lines and io.stdin.lines.
func splitLines(data string) object.Object {
	data = strings.TrimSuffix(data, "\n")
	if data == "" {
		return object.NewList([]object.Object{})
	}
	lines := strings.Split(data, "\n")
	items := make([]object.Object, len(lines))
	for i, l := range lines {
		items[i] = object.NewString(strings.TrimSuffix(l, "\r"))
	}
	return object.NewList(items)
}

// writeSink resolves the first argument of io.write: a file path or a
// writable handle (io.stdout, io.stderr). Writing a path creates or
// truncates it; the byte count is returned either way (issue #41).
func writeSink(ctx context.Context, arg object.Object, data string) (int64, error) {
	if s, ok := arg.(*OutputStream); ok {
		n, err := s.writer(ctx).Write([]byte(data))
		if err != nil {
			return 0, fmt.Errorf("io.%s: %w", s.name, err)
		}
		return int64(n), nil
	}
	path, err := object.AsString(arg)
	if err != nil {
		return 0, err
	}
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		return 0, err
	}
	return int64(len(data)), nil
}

// copyFn implements io.copy (and the deprecated io.cp alias): byte copy
// that also preserves the source's permission bits — the behaviour
// os.copy_file had and io.cp lacked (issue #41).
func copyFn(name string) func(ctx context.Context, args ...object.Object) (object.Object, error) {
	return func(ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) != 2 {
			return nil, object.NewArgsError(name, 2, len(args))
		}
		src, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}
		dst, err := object.AsString(args[1])
		if err != nil {
			return nil, err
		}

		srcFile, err := os.Open(src)
		if err != nil {
			return nil, err
		}
		defer srcFile.Close()

		info, err := srcFile.Stat()
		if err != nil {
			return nil, err
		}

		dstFile, err := os.Create(dst)
		if err != nil {
			return nil, err
		}
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return nil, err
		}
		if err := os.Chmod(dst, info.Mode().Perm()); err != nil {
			return nil, err
		}
		return object.Nil, nil
	}
}

func (i *IO) Attrs() []object.AttrSpec {
	return []object.AttrSpec{
		{Name: "read", Doc: "Read a file path, bytes, or io.stdin; return the content as a string (issue #41)"},
		{Name: "lines", Doc: "Read a file path or io.stdin; return its lines as a list (issue #41)"},
		{Name: "read_all", Doc: "Alias of read"},
		{Name: "write", Doc: "Write text to a file path or io.stdout/io.stderr; returns the byte count (issue #41)"},
		{Name: "copy", Doc: "Copy src file to dst file, preserving the source mode (issue #41)"},
		{Name: "cp", Doc: "Deprecated alias of copy"},
		{Name: "stdin", Doc: "Standard input handle: read() and lines() (issue #42)"},
		{Name: "stdout", Doc: "Standard output handle: write(text) and println(...) (issue #42)"},
		{Name: "stderr", Doc: "Standard error handle: write(text) and println(...) (issue #42)"},
	}
}
