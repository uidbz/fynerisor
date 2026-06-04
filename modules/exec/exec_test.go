package exec

import (
	"bytes"
	"context"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

func TestExec(t *testing.T) {
	ctx := context.Background()

	t.Run("simple echo command", func(t *testing.T) {
		args := []object.Object{
			object.NewList([]object.Object{
				object.NewString("echo"),
				object.NewString("hello"),
			}),
		}

		result, err := Exec(ctx, args...)
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		r, ok := result.(*Result)
		if !ok {
			t.Fatalf("expected *Result, got %T", result)
		}

		stdout := r.Stdout().(*object.String).Value()
		if !strings.Contains(stdout, "hello") {
			t.Errorf("expected stdout to contain 'hello', got: %s", stdout)
		}
	})

	t.Run("command with working directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		args := []object.Object{
			object.NewList([]object.Object{
				object.NewString("pwd"),
			}),
			object.NewMap(map[string]object.Object{
				"dir": object.NewString(tmpDir),
			}),
		}

		result, err := Exec(ctx, args...)
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		r := result.(*Result)
		stdout := strings.TrimSpace(r.Stdout().(*object.String).Value())
		if !strings.Contains(stdout, tmpDir) {
			t.Errorf("expected stdout to contain %s, got: %s", tmpDir, stdout)
		}
	})

	t.Run("command with stdin", func(t *testing.T) {
		var cmd string
		if runtime.GOOS == "windows" {
			t.Skip("skipping on windows")
		}
		cmd = "cat"

		args := []object.Object{
			object.NewList([]object.Object{
				object.NewString(cmd),
			}),
			object.NewMap(map[string]object.Object{
				"stdin": object.NewString("test input"),
			}),
		}

		result, err := Exec(ctx, args...)
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		r := result.(*Result)
		stdout := r.Stdout().(*object.String).Value()
		if stdout != "test input" {
			t.Errorf("expected 'test input', got: %s", stdout)
		}
	})

	t.Run("command with environment", func(t *testing.T) {
		var cmd, arg string
		if runtime.GOOS == "windows" {
			cmd = "cmd"
			arg = "/c"
		} else {
			cmd = "sh"
			arg = "-c"
		}

		args := []object.Object{
			object.NewList([]object.Object{
				object.NewString(cmd),
				object.NewString(arg),
				object.NewString("echo $TEST_VAR"),
			}),
			object.NewMap(map[string]object.Object{
				"env": object.NewMap(map[string]object.Object{
					"TEST_VAR": object.NewString("test_value"),
				}),
			}),
		}

		result, err := Exec(ctx, args...)
		if err != nil {
			t.Fatalf("Exec failed: %v", err)
		}

		r := result.(*Result)
		stdout := strings.TrimSpace(r.Stdout().(*object.String).Value())
		if !strings.Contains(stdout, "test_value") {
			t.Errorf("expected stdout to contain 'test_value', got: %s", stdout)
		}
	})

	t.Run("invalid arguments", func(t *testing.T) {
		_, err := Exec(ctx)
		if err == nil {
			t.Error("expected error with no arguments")
		}

		_, err = Exec(ctx, object.NewInt(42))
		if err == nil {
			t.Error("expected error with invalid argument type")
		}
	})
}

func TestCommandFunc(t *testing.T) {
	ctx := context.Background()

	t.Run("list form", func(t *testing.T) {
		args := []object.Object{
			object.NewList([]object.Object{
				object.NewString("echo"),
				object.NewString("test"),
			}),
		}

		result, err := CommandFunc(ctx, args...)
		if err != nil {
			t.Fatalf("CommandFunc failed: %v", err)
		}

		cmd, ok := result.(*Command)
		if !ok {
			t.Fatalf("expected *Command, got %T", result)
		}

		if !strings.Contains(cmd.value.Path, "echo") {
			t.Errorf("expected path to contain 'echo', got: %s", cmd.value.Path)
		}
	})

	t.Run("variadic form", func(t *testing.T) {
		args := []object.Object{
			object.NewString("ls"),
			object.NewString("-l"),
		}

		result, err := CommandFunc(ctx, args...)
		if err != nil {
			t.Fatalf("CommandFunc failed: %v", err)
		}

		cmd := result.(*Command)
		if !strings.Contains(cmd.value.Path, "ls") {
			t.Errorf("expected path to contain 'ls', got: %s", cmd.value.Path)
		}
		if len(cmd.value.Args) != 2 {
			t.Errorf("expected 2 args, got: %d", len(cmd.value.Args))
		}
	})
}

func TestLookPath(t *testing.T) {
	ctx := context.Background()

	t.Run("find echo", func(t *testing.T) {
		args := []object.Object{
			object.NewString("echo"),
		}

		result, err := LookPath(ctx, args...)
		if err != nil {
			t.Fatalf("LookPath failed: %v", err)
		}

		path := result.(*object.String).Value()
		if path == "" {
			t.Error("expected non-empty path")
		}
	})

	t.Run("nonexistent command", func(t *testing.T) {
		args := []object.Object{
			object.NewString("this_command_does_not_exist_12345"),
		}

		_, err := LookPath(ctx, args...)
		if err == nil {
			t.Error("expected error for nonexistent command")
		}
	})
}

func TestCommand(t *testing.T) {
	ctx := context.Background()

	t.Run("run method", func(t *testing.T) {
		args := []object.Object{
			object.NewList([]object.Object{
				object.NewString("echo"),
				object.NewString("test"),
			}),
		}

		result, _ := CommandFunc(ctx, args...)
		cmd := result.(*Command)

		err := cmd.Run(ctx)
		if err != nil {
			t.Fatalf("Run failed: %v", err)
		}

		stdout := cmd.Stdout().(*object.String).Value()
		if !strings.Contains(stdout, "test") {
			t.Errorf("expected stdout to contain 'test', got: %s", stdout)
		}
	})

	t.Run("output method", func(t *testing.T) {
		args := []object.Object{
			object.NewList([]object.Object{
				object.NewString("echo"),
				object.NewString("output_test"),
			}),
		}

		result, _ := CommandFunc(ctx, args...)
		cmd := result.(*Command)

		output, err := cmd.value.Output()
		if err != nil {
			t.Fatalf("Output failed: %v", err)
		}

		if !strings.Contains(string(output), "output_test") {
			t.Errorf("expected output to contain 'output_test', got: %s", output)
		}
	})

	t.Run("set attributes", func(t *testing.T) {
		args := []object.Object{
			object.NewList([]object.Object{
				object.NewString("echo"),
			}),
		}

		result, _ := CommandFunc(ctx, args...)
		cmd := result.(*Command)

		tmpDir := t.TempDir()
		err := cmd.SetAttr("dir", object.NewString(tmpDir))
		if err != nil {
			t.Fatalf("SetAttr failed: %v", err)
		}

		if cmd.value.Dir != tmpDir {
			t.Errorf("expected dir %s, got: %s", tmpDir, cmd.value.Dir)
		}
	})

	t.Run("get attributes", func(t *testing.T) {
		args := []object.Object{
			object.NewList([]object.Object{
				object.NewString("echo"),
			}),
		}

		result, _ := CommandFunc(ctx, args...)
		cmd := result.(*Command)

		pathObj, ok := cmd.GetAttr("path")
		if !ok {
			t.Fatal("expected path attribute")
		}

		path := pathObj.(*object.String).Value()
		if !strings.Contains(path, "echo") {
			t.Errorf("expected path to contain 'echo', got: %s", path)
		}
	})
}

func TestResult(t *testing.T) {
	ctx := context.Background()

	t.Run("stdout and stderr", func(t *testing.T) {
		args := []object.Object{
			object.NewList([]object.Object{
				object.NewString("echo"),
				object.NewString("hello"),
			}),
		}

		execResult, _ := Exec(ctx, args...)
		result := execResult.(*Result)

		stdout := result.Stdout()
		if stdout == object.Nil {
			t.Error("expected non-nil stdout")
		}

		stderr := result.Stderr()
		if stderr == object.Nil {
			t.Error("expected non-nil stderr")
		}
	})

	t.Run("json parsing", func(t *testing.T) {
		jsonStr := `{"key":"value","number":42}`

		args := []object.Object{
			object.NewList([]object.Object{
				object.NewString("echo"),
				object.NewString(jsonStr),
			}),
		}

		execResult, _ := Exec(ctx, args...)
		result := execResult.(*Result)

		jsonObj, err := result.JSON()
		if err != nil {
			t.Fatalf("JSON parsing failed: %v", err)
		}

		m, ok := jsonObj.(*object.Map)
		if !ok {
			t.Fatalf("expected map, got %T", jsonObj)
		}

		keyVal, exists := m.Value()["key"]
		if !exists {
			t.Error("expected 'key' in parsed JSON")
		}

		keyStr := keyVal.(*object.String).Value()
		if keyStr != "value" {
			t.Errorf("expected 'value', got: %s", keyStr)
		}
	})

	t.Run("attributes", func(t *testing.T) {
		args := []object.Object{
			object.NewList([]object.Object{
				object.NewString("echo"),
				object.NewString("test"),
			}),
		}

		execResult, _ := Exec(ctx, args...)
		result := execResult.(*Result)

		pid, ok := result.GetAttr("pid")
		if !ok {
			t.Error("expected pid attribute")
		}

		if pid == object.Nil {
			t.Error("expected non-nil pid")
		}
	})
}

func TestConfigureCommand(t *testing.T) {
	ctx := context.Background()
	cmd := newTestCmd(ctx, "echo", "test")

	t.Run("set directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		params := object.NewMap(map[string]object.Object{
			"dir": object.NewString(tmpDir),
		})

		err := configureCommand(cmd, params)
		if err != nil {
			t.Fatalf("configureCommand failed: %v", err)
		}

		if cmd.Dir != tmpDir {
			t.Errorf("expected dir %s, got: %s", tmpDir, cmd.Dir)
		}
	})

	t.Run("set stdin", func(t *testing.T) {
		cmd := newTestCmd(ctx, "cat")
		params := object.NewMap(map[string]object.Object{
			"stdin": object.NewString("test input"),
		})

		err := configureCommand(cmd, params)
		if err != nil {
			t.Fatalf("configureCommand failed: %v", err)
		}

		if cmd.Stdin == nil {
			t.Error("expected stdin to be set")
		}

		buf := cmd.Stdin.(*bytes.Buffer)
		if buf.String() != "test input" {
			t.Errorf("expected 'test input', got: %s", buf.String())
		}
	})

	t.Run("set environment", func(t *testing.T) {
		cmd := newTestCmd(ctx, "env")
		params := object.NewMap(map[string]object.Object{
			"env": object.NewMap(map[string]object.Object{
				"VAR1": object.NewString("value1"),
				"VAR2": object.NewString("value2"),
			}),
		})

		err := configureCommand(cmd, params)
		if err != nil {
			t.Fatalf("configureCommand failed: %v", err)
		}

		if len(cmd.Env) != 2 {
			t.Errorf("expected 2 env vars, got: %d", len(cmd.Env))
		}
	})

	t.Run("invalid key", func(t *testing.T) {
		cmd := newTestCmd(ctx, "echo")
		params := object.NewMap(map[string]object.Object{
			"invalid_key": object.NewString("value"),
		})

		err := configureCommand(cmd, params)
		if err == nil {
			t.Error("expected error for invalid key")
		}
	})
}

func newTestCmd(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

func TestModule(t *testing.T) {
	mod := Module()

	if mod == nil {
		t.Fatal("Module() returned nil")
	}

	if mod.Name().Value() != "exec" {
		t.Errorf("expected module name 'exec', got: %s", mod.Name().Value())
	}

	// Check that module has expected functions
	cmd, ok := mod.GetAttr("command")
	if !ok {
		t.Error("expected 'command' function in module")
	}
	if cmd.Type() != "builtin" {
		t.Errorf("expected builtin type, got: %s", cmd.Type())
	}

	lookPath, ok := mod.GetAttr("look_path")
	if !ok {
		t.Error("expected 'look_path' function in module")
	}
	if lookPath.Type() != "builtin" {
		t.Errorf("expected builtin type, got: %s", lookPath.Type())
	}
}
