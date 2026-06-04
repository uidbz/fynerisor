package exec

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

func CommandFunc(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 1 || len(args) > 1000 {
		return nil, fmt.Errorf("command: expected 1 to 1000 arguments, got %d", len(args))
	}

	var strArgs []string

	// Two forms of arguments are supported:
	// 1. command(["ls", "-l"]) - preferred form
	// 2. command("ls", "-l") - original form
	if len(args) == 1 {
		if list, ok := args[0].(*object.List); ok {
			// This is form 1
			for _, arg := range list.Value() {
				argStr, err := object.AsString(arg)
				if err != nil {
					return nil, err
				}
				strArgs = append(strArgs, argStr)
			}
			if len(strArgs) == 0 {
				return nil, fmt.Errorf("command: expected at least one argument in list")
			}
			return NewCommand(exec.CommandContext(ctx, strArgs[0], strArgs[1:]...)), nil
		}
	}

	// This is form 2
	name, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}
	for _, arg := range args[1:] {
		argStr, err := object.AsString(arg)
		if err != nil {
			return nil, err
		}
		strArgs = append(strArgs, argStr)
	}
	return NewCommand(exec.CommandContext(ctx, name, strArgs...)), nil
}

func LookPath(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("look_path: expected 1 argument, got %d", len(args))
	}
	path, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}
	result, execErr := exec.LookPath(path)
	if execErr != nil {
		return nil, execErr
	}
	return object.NewString(result), nil
}

func Exec(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 1 || len(args) > 3 {
		return nil, fmt.Errorf("exec: expected 1 to 3 arguments, got %d", len(args))
	}

	var wasList bool
	var program string
	var optArgs []string

	if list, ok := args[0].(*object.List); ok {
		wasList = true
		var cmdArgs []string
		for _, arg := range list.Value() {
			argStr, err := object.AsString(arg)
			if err != nil {
				return nil, err
			}
			cmdArgs = append(cmdArgs, argStr)
		}
		if len(cmdArgs) == 0 {
			return nil, fmt.Errorf("exec: expected at least one argument in list")
		}
		program = cmdArgs[0]
		optArgs = cmdArgs[1:]
	} else {
		var err error
		program, err = object.AsString(args[0])
		if err != nil {
			return nil, err
		}
		if len(args) > 1 {
			if list, ok := args[1].(*object.List); ok {
				for _, arg := range list.Value() {
					argStr, err := object.AsString(arg)
					if err != nil {
						return nil, err
					}
					optArgs = append(optArgs, argStr)
				}
			} else {
				return nil, fmt.Errorf("exec: expected list for second argument, got %s", args[1].Type())
			}
		}
	}

	cmd := exec.CommandContext(ctx, program, optArgs...)

	mapOffset := 2
	if wasList {
		mapOffset = 1
	}

	if len(args) > mapOffset {
		params, ok := args[mapOffset].(*object.Map)
		if !ok {
			return nil, fmt.Errorf("exec: expected map for options, got %s", args[mapOffset].Type())
		}
		if err := configureCommand(cmd, params); err != nil {
			return nil, err
		}
	}

	if cmd.Stdout == nil {
		cmd.Stdout = &bytes.Buffer{}
	}
	if cmd.Stderr == nil {
		cmd.Stderr = &bytes.Buffer{}
	}

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return NewResult(cmd), nil
}

var allowedKeys = map[string]bool{
	"dir":    true,
	"stdin":  true,
	"stdout": true,
	"stderr": true,
	"env":    true,
}

func configureCommand(cmd *exec.Cmd, params *object.Map) error {
	for key := range params.Value() {
		if !allowedKeys[key] {
			return fmt.Errorf("exec: unexpected key %q", key)
		}
	}

	if stdoutObj, ok := params.Value()["stdout"]; ok {
		switch v := stdoutObj.(type) {
		case io.Writer:
			cmd.Stdout = v
		default:
			return fmt.Errorf("exec: expected io.Writer for stdout, got %s", stdoutObj.Type())
		}
	}

	if stderrObj, ok := params.Value()["stderr"]; ok {
		switch v := stderrObj.(type) {
		case io.Writer:
			cmd.Stderr = v
		default:
			return fmt.Errorf("exec: expected io.Writer for stderr, got %s", stderrObj.Type())
		}
	}

	if stdinObj, ok := params.Value()["stdin"]; ok {
		switch v := stdinObj.(type) {
		case *object.Bytes:
			cmd.Stdin = bytes.NewBuffer(v.Value())
		case *object.String:
			cmd.Stdin = bytes.NewBufferString(v.Value())
		case io.Reader:
			cmd.Stdin = v
		default:
			return fmt.Errorf("exec: expected io.Reader for stdin, got %s", stdinObj.Type())
		}
	}

	if dirObj, ok := params.Value()["dir"]; ok {
		dirStr, err := object.AsString(dirObj)
		if err != nil {
			return fmt.Errorf("exec: expected string for dir, got %s", dirObj.Type())
		}
		cmd.Dir = dirStr
	}

	if envObj, ok := params.Value()["env"]; ok {
		envMap, ok := envObj.(*object.Map)
		if !ok {
			return fmt.Errorf("exec: expected map for env, got %s", envObj.Type())
		}
		var env []string
		for key, value := range envMap.Value() {
			valueStr, err := object.AsString(value)
			if err != nil {
				return fmt.Errorf("exec: expected string for env value, got %s", value.Type())
			}
			env = append(env, fmt.Sprintf("%s=%s", key, valueStr))
		}
		cmd.Env = env
	}

	return nil
}

func Module() *object.Module {
	return object.NewBuiltinsModule("exec", map[string]object.Object{
		"command":   object.NewBuiltin("command", CommandFunc),
		"look_path": object.NewBuiltin("look_path", LookPath),
	}, Exec)
}
