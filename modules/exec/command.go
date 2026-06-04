package exec

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

type Command struct {
	value *exec.Cmd
}

func (c *Command) Inspect() string {
	var args []string
	for _, arg := range c.value.Args {
		args = append(args, fmt.Sprintf("%q", arg))
	}
	return fmt.Sprintf("exec.command(%s)", strings.Join(args, ", "))
}

func (c *Command) Type() object.Type {
	return "exec.command"
}

func (c *Command) Value() *exec.Cmd {
	return c.value
}

func (c *Command) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "path":
		return object.NewString(c.value.Path), true
	case "dir":
		return object.NewString(c.value.Dir), true
	case "env":
		var env []object.Object
		for _, e := range c.value.Env {
			env = append(env, object.NewString(e))
		}
		return object.NewList(env), true
	case "stdout":
		return c.Stdout(), true
	case "stderr":
		return c.Stderr(), true
	case "run":
		return object.NewBuiltin("run", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if err := c.Run(ctx); err != nil {
				return nil, err
			}
			return object.Nil, nil
		}), true
	case "combined_output":
		return object.NewBuiltin("combined_output", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			output, err := c.value.CombinedOutput()
			if err != nil {
				return nil, err
			}
			return object.NewBytes(output), nil
		}), true
	case "environ":
		return object.NewBuiltin("environ", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			env := c.value.Environ()
			var envStr []object.Object
			for _, e := range env {
				envStr = append(envStr, object.NewString(e))
			}
			return object.NewList(envStr), nil
		}), true
	case "output":
		return object.NewBuiltin("output", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			output, err := c.value.Output()
			if err != nil {
				return nil, err
			}
			return object.NewBytes(output), nil
		}), true
	case "start":
		return object.NewBuiltin("start", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if err := c.value.Start(); err != nil {
				return nil, err
			}
			return object.Nil, nil
		}), true
	case "wait":
		return object.NewBuiltin("wait", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if err := c.value.Wait(); err != nil {
				return nil, err
			}
			return object.Nil, nil
		}), true
	}
	return nil, false
}

func (c *Command) SetAttr(name string, value object.Object) error {
	switch name {
	case "path":
		path, err := object.AsString(value)
		if err != nil {
			return err
		}
		c.value.Path = path
	case "dir":
		dir, err := object.AsString(value)
		if err != nil {
			return err
		}
		c.value.Dir = dir
	case "env":
		list, ok := value.(*object.List)
		if !ok {
			return fmt.Errorf("type error: expected list for env, got %s", value.Type())
		}
		var envStr []string
		for _, e := range list.Value() {
			item, err := object.AsString(e)
			if err != nil {
				return err
			}
			envStr = append(envStr, item)
		}
		c.value.Env = envStr
	case "stdin":
		// Try to convert to reader
		switch v := value.(type) {
		case *object.String:
			c.value.Stdin = bytes.NewBufferString(v.Value())
		case *object.Bytes:
			c.value.Stdin = bytes.NewBuffer(v.Value())
		default:
			return fmt.Errorf("type error: expected string or byte_slice for stdin, got %s", value.Type())
		}
	case "stdout":
		// Stdout needs to be set to a buffer if we want to capture it
		return fmt.Errorf("type error: stdout cannot be set directly")
	case "stderr":
		// Stderr needs to be set to a buffer if we want to capture it
		return fmt.Errorf("type error: stderr cannot be set directly")
	default:
		return fmt.Errorf("type error: exec.command has no attribute %q", name)
	}
	return nil
}

func (c *Command) Stdout() object.Object {
	if c.value.Stdout == nil {
		return object.Nil
	}
	switch value := c.value.Stdout.(type) {
	case *bytes.Buffer:
		return object.NewString(value.String())
	default:
		return object.Nil
	}
}

func (c *Command) Stderr() object.Object {
	if c.value.Stderr == nil {
		return object.Nil
	}
	switch value := c.value.Stderr.(type) {
	case *bytes.Buffer:
		return object.NewString(value.String())
	default:
		return object.Nil
	}
}

func (c *Command) Run(ctx context.Context) error {
	if c.value.Stdout == nil {
		c.value.Stdout = &bytes.Buffer{}
	}
	if c.value.Stderr == nil {
		c.value.Stderr = &bytes.Buffer{}
	}
	return c.value.Run()
}

func (c *Command) Interface() interface{} {
	return c.value
}

func (c *Command) String() string {
	return c.Inspect()
}

func (c *Command) Equals(other object.Object) bool {
	return c == other
}

func (c *Command) IsTruthy() bool {
	return true
}

func (c *Command) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, fmt.Errorf("type error: unsupported operation for exec.command: %v", opType)
}

func (c *Command) Cost() int {
	return 8
}

func (c *Command) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal exec.command")
}

func (c *Command) Attrs() []object.AttrSpec {
	return []object.AttrSpec{
		{Name: "path", Doc: "Path to the executable"},
		{Name: "dir", Doc: "Working directory of the command"},
		{Name: "env", Doc: "Environment variables given to the command"},
		{Name: "stdout", Doc: "Standard output produced by the command"},
		{Name: "stderr", Doc: "Standard error produced by the command"},
		{Name: "run", Doc: "Run the command and wait for it to complete"},
		{Name: "output", Doc: "Run the command and return its standard output"},
		{Name: "combined_output", Doc: "Run the command and return combined stdout and stderr"},
		{Name: "start", Doc: "Start the command without waiting"},
		{Name: "wait", Doc: "Wait for the command to exit"},
		{Name: "environ", Doc: "Get the environment of the command"},
	}
}

func NewCommand(cmd *exec.Cmd) *Command {
	return &Command{value: cmd}
}
