package io

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	stdio "io"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

// Per-runtime text streams (issue #42): io.stdin, io.stdout and
// io.stderr give scripts the universal text interface. They resolve to
// the *runtime's* streams — a StreamSource carried in the eval context —
// so fynerisor scripts compose as pipeline filters: WithStdin feeds
// data, WithStdout/WithStderr (and window.SetStdout) steer output, and
// the process's own streams are never touched by another runtime.

// maxStdinRead bounds one io.stdin read (a GUI accidentally reading an
// endless pipe should fail, not exhaust memory).
const maxStdinRead = 32 << 20 // 32 MiB

// StreamSource provides the text streams of a running script. gui.Window
// and core.Context implement it; builtins resolve the streams at call
// time, so redirection installed mid-run is visible.
type StreamSource interface {
	Stdin() stdio.Reader
	Stdout() stdio.Writer
	Stderr() stdio.Writer
}

type streamsKey struct{}

// ContextWithStreams binds a runtime's stream source to the eval
// context; io.stdin/stdout/stderr prefer it over the process defaults.
func ContextWithStreams(ctx context.Context, src StreamSource) context.Context {
	if src == nil {
		return ctx
	}
	return context.WithValue(ctx, streamsKey{}, src)
}

// StreamsFromContext returns the runtime's stream source, or a fallback
// to the process std streams when the module is used without one.
func StreamsFromContext(ctx context.Context) StreamSource {
	if ctx != nil {
		if src, ok := ctx.Value(streamsKey{}).(StreamSource); ok && src != nil {
			return src
		}
	}
	return stdioSource{}
}

type stdioSource struct{}

func (stdioSource) Stdin() stdio.Reader  { return os.Stdin }
func (stdioSource) Stdout() stdio.Writer { return os.Stdout }
func (stdioSource) Stderr() stdio.Writer { return os.Stderr }

// readStream drains the runtime's stdin up to maxStdinRead.
func readStream(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	r := bufio.NewReader(StreamsFromContext(ctx).Stdin())
	data, err := stdio.ReadAll(stdio.LimitReader(r, maxStdinRead))
	if err != nil {
		return "", fmt.Errorf("io.stdin: %w", err)
	}
	return string(data), nil
}

// --- io.stdin ----------------------------------------------------------

// StdinObject is the script-visible io.stdin handle: read() and lines().
type StdinObject struct{}

func NewStdin() *StdinObject { return &StdinObject{} }

const stdinType object.Type = "io.stdin"

func (s *StdinObject) Type() object.Type { return stdinType }
func (s *StdinObject) Inspect() string   { return "io.stdin" }
func (s *StdinObject) String() string    { return "io.stdin" }
func (s *StdinObject) Interface() any    { return s }
func (s *StdinObject) IsTruthy() bool    { return true }
func (s *StdinObject) Cost() int         { return 0 }
func (s *StdinObject) MarshalJSON() ([]byte, error) {
	return []byte(`"io.stdin"`), nil
}
func (s *StdinObject) Equals(other object.Object) bool {
	_, ok := other.(*StdinObject)
	return ok
}
func (s *StdinObject) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for io.stdin: %v", opType)
}

func (s *StdinObject) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "read":
		return object.NewBuiltin("io.stdin.read", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.NewArgsError("io.stdin.read", 0, len(args))
			}
			data, err := readStream(ctx)
			if err != nil {
				return nil, err
			}
			return object.NewString(data), nil
		}), true
	case "lines":
		return object.NewBuiltin("io.stdin.lines", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 0 {
				return nil, object.NewArgsError("io.stdin.lines", 0, len(args))
			}
			data, err := readStream(ctx)
			if err != nil {
				return nil, err
			}
			data = strings.TrimSuffix(data, "\n")
			if data == "" {
				return object.NewList([]object.Object{}), nil
			}
			lines := strings.Split(data, "\n")
			items := make([]object.Object, len(lines))
			for i, l := range lines {
				items[i] = object.NewString(strings.TrimSuffix(l, "\r"))
			}
			return object.NewList(items), nil
		}), true
	}
	return nil, false
}

func (s *StdinObject) SetAttr(name string, value object.Object) error {
	return object.Errorf("type error: io.stdin attributes are read-only")
}

func (s *StdinObject) Attrs() []object.AttrSpec {
	return []object.AttrSpec{
		{Name: "read", Doc: "Read all of stdin and return it as a string"},
		{Name: "lines", Doc: "Read all of stdin and return its lines as a list"},
	}
}

// --- io.stdout / io.stderr ---------------------------------------------

// OutputStream is the script-visible writer handle for one runtime
// stream ("stdout" or "stderr"): write(text) and println(values...).
type OutputStream struct {
	name string // "stdout" or "stderr"
}

func NewStdout() *OutputStream { return &OutputStream{name: "stdout"} }
func NewStderr() *OutputStream { return &OutputStream{name: "stderr"} }

const outType object.Type = "io.stream"

func (s *OutputStream) Type() object.Type { return outType }
func (s *OutputStream) Inspect() string   { return "io." + s.name }
func (s *OutputStream) String() string    { return "io." + s.name }
func (s *OutputStream) Interface() any    { return s }
func (s *OutputStream) IsTruthy() bool    { return true }
func (s *OutputStream) Cost() int         { return 0 }
func (s *OutputStream) MarshalJSON() ([]byte, error) {
	return []byte(`"io.` + s.name + `"`), nil
}
func (s *OutputStream) Equals(other object.Object) bool {
	o, ok := other.(*OutputStream)
	return ok && o.name == s.name
}
func (s *OutputStream) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for io.%s: %v", s.name, opType)
}

func (s *OutputStream) writer(ctx context.Context) stdio.Writer {
	src := StreamsFromContext(ctx)
	if s.name == "stderr" {
		return src.Stderr()
	}
	return src.Stdout()
}

func (s *OutputStream) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "write":
		return object.NewBuiltin("io."+s.name+".write", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return nil, object.NewArgsError("io."+s.name+".write", 1, len(args))
			}
			text, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			n, werr := s.writer(ctx).Write([]byte(text))
			if werr != nil {
				return nil, fmt.Errorf("io.%s: %w", s.name, werr)
			}
			return object.NewInt(int64(n)), nil
		}), true
	case "println":
		return object.NewBuiltin("io."+s.name+".println", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			values := make([]any, len(args))
			for i, a := range args {
				values[i] = object.PrintableValue(a)
			}
			fmt.Fprintln(s.writer(ctx), values...)
			return object.Nil, nil
		}), true
	}
	return nil, false
}

func (s *OutputStream) SetAttr(name string, value object.Object) error {
	return object.Errorf("type error: io.%s attributes are read-only", s.name)
}

func (s *OutputStream) Attrs() []object.AttrSpec {
	return []object.AttrSpec{
		{Name: "write", Doc: "Write a string to the stream; returns the byte count"},
		{Name: "println", Doc: "Write values followed by a newline"},
	}
}
