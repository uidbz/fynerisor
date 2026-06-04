package exec

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

type Result struct {
	cmd *exec.Cmd
}

func (r *Result) Type() object.Type {
	return "exec.result"
}

func (r *Result) Inspect() string {
	if r.cmd.Process != nil {
		return fmt.Sprintf("exec.result(pid: %d)", r.cmd.Process.Pid)
	}
	return "exec.result(no process)"
}

func (r *Result) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "pid":
		if r.cmd.Process != nil {
			return object.NewInt(int64(r.cmd.Process.Pid)), true
		}
		return object.Nil, true
	case "stdout":
		return r.Stdout(), true
	case "stderr":
		return r.Stderr(), true
	case "json":
		return object.NewBuiltin("json",
			func(ctx context.Context, args ...object.Object) (object.Object, error) {
				if len(args) != 0 {
					return nil, fmt.Errorf("json: expected 0 arguments, got %d", len(args))
				}
				return r.JSON()
			},
		), true
	}
	return nil, false
}

func (r *Result) Stdout() object.Object {
	value := r.cmd.Stdout
	switch value := value.(type) {
	case *bytes.Buffer:
		return object.NewString(value.String())
	default:
		return object.Nil
	}
}

func (r *Result) Stderr() object.Object {
	value := r.cmd.Stderr
	switch value := value.(type) {
	case *bytes.Buffer:
		return object.NewString(value.String())
	default:
		return object.Nil
	}
}

func (r *Result) JSON() (object.Object, error) {
	var data []byte
	switch stdout := r.cmd.Stdout.(type) {
	case *bytes.Buffer:
		data = stdout.Bytes()
	default:
		return nil, fmt.Errorf("exec.result.json: does not support stdout type %T", stdout)
	}

	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, fmt.Errorf("json.unmarshal failed: %s", err.Error())
	}

	scriptObj := object.FromGoType(obj)
	if scriptObj == nil {
		return nil, fmt.Errorf("json.unmarshal: failed to convert type")
	}
	return scriptObj, nil
}

func (r *Result) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("exec.result does not support attribute assignment")
}

func (r *Result) Interface() interface{} {
	return r.cmd
}

func (r *Result) Equals(other object.Object) bool {
	return r == other
}

func (r *Result) IsTruthy() bool {
	return true
}

func (r *Result) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, fmt.Errorf("type error: unsupported operation for exec.result: %v", opType)
}

func (r *Result) Cost() int {
	return 0
}

func (r *Result) MarshalJSON() ([]byte, error) {
	pid := 0
	if r.cmd.Process != nil {
		pid = r.cmd.Process.Pid
	}
	return json.Marshal(struct {
		Stdout string `json:"stdout"`
		Stderr string `json:"stderr"`
		Pid    int    `json:"pid"`
	}{
		Stdout: r.Stdout().Inspect(),
		Stderr: r.Stderr().Inspect(),
		Pid:    pid,
	})
}

func (r *Result) String() string {
	return r.Inspect()
}

func (r *Result) Attrs() []object.AttrSpec {
	return []object.AttrSpec{
		{Name: "pid", Doc: "Process ID of the command"},
		{Name: "stdout", Doc: "Standard output produced by the command"},
		{Name: "stderr", Doc: "Standard error produced by the command"},
		{Name: "json", Doc: "Parse stdout as JSON"},
	}
}

func NewResult(cmd *exec.Cmd) *Result {
	return &Result{cmd: cmd}
}
