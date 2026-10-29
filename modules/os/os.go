// Package os provides basic OS functionality for Risor scripts.
package os

import (
	"context"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// scriptArgs holds the command-line arguments passed to the running script.
// Deprecated: process-global; two runtimes in one process overwrite each
// other. Hosts should use core.WithScriptArgs / gui.WithScriptArgs, which
// bind args to the individual runtime via the eval context; this remains
// only as the fallback default for hosts that still call it (issue #44).
var scriptArgs []string

// SetScriptArgs stores the script's command-line arguments so scripts can
// access them through os.args(). It mirrors risor v1's SetScriptArgs.
//
// Deprecated: use the per-runtime WithScriptArgs options (core or gui);
// the global is consulted only when a runtime carries no args of its own.
func SetScriptArgs(args []string) {
	scriptArgs = args
}

// argsKey tags the per-runtime script arguments in the eval context
// (issue #44). Builtins receive that context, so two runtimes in one
// process answer os.args() from their own configuration.
type argsKey struct{}

// ContextWithArgs returns ctx carrying the runtime's command-line
// arguments; os.args() prefers them over the deprecated global.
func ContextWithArgs(ctx context.Context, args []string) context.Context {
	return context.WithValue(ctx, argsKey{}, args)
}

// ArgsFromContext returns the runtime arguments carried by ctx, and
// whether the runtime carried its own (a nil-but-set list means "empty",
// distinct from "unset").
func ArgsFromContext(ctx context.Context) ([]string, bool) {
	if ctx == nil {
		return nil, false
	}
	args, ok := ctx.Value(argsKey{}).([]string)
	return args, ok
}

// Module returns the os module for risor
func Module() *object.Module {
	return object.NewBuiltinsModule("os", map[string]object.Object{
		"goos":         object.NewBuiltin("os.goos", goos),
		"current_user": object.NewBuiltin("os.current_user", currentUser),
		"home_dir":     object.NewBuiltin("os.home_dir", homeDir),
		"open_browser": object.NewBuiltin("os.open_browser", openBrowser),
		"read_file":    object.NewBuiltin("os.read_file", readFile),
		"write_file":   object.NewBuiltin("os.write_file", writeFile),
		"create":       object.NewBuiltin("os.create", create),
		"copy_file":    object.NewBuiltin("os.copy_file", copyFile),
		"read_dir":     object.NewBuiltin("os.read_dir", readDir),
		"mkdir":        object.NewBuiltin("os.mkdir", mkdir),
		"mkdir_all":    object.NewBuiltin("os.mkdir_all", mkdirAll),
		"remove_all":   object.NewBuiltin("os.remove_all", removeAll),
		"is_dir":       object.NewBuiltin("os.is_dir", isDir),
		"getwd":        object.NewBuiltin("os.getwd", getwd),
		"args":         object.NewBuiltin("os.args", scriptArgsFunc),
		"exec":         object.NewBuiltin("os.exec", execCommand),
	})
}

// goos returns the operating system name
func goos(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 0 {
		return object.Errorf("os.goos: expected 0 arguments, got %d", len(args)), nil
	}
	return object.NewString(runtime.GOOS), nil
}

// currentUser returns information about the current user
func currentUser(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 0 {
		return object.Errorf("os.current_user: expected 0 arguments, got %d", len(args)), nil
	}

	u, err := user.Current()
	if err != nil {
		return object.NewError(err), nil
	}

	userMap := map[string]object.Object{
		"username": object.NewString(u.Username),
		"uid":      object.NewString(u.Uid),
		"gid":      object.NewString(u.Gid),
		"name":     object.NewString(u.Name),
		"home_dir": object.NewString(u.HomeDir),
	}

	return object.NewMap(userMap), nil
}

// homeDir returns the current user's home directory, a shortcut for
// current_user().home_dir
func homeDir(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 0 {
		return object.Errorf("os.home_dir: expected 0 arguments, got %d", len(args)), nil
	}

	u, err := user.Current()
	if err != nil {
		return object.NewError(err), nil
	}

	return object.NewString(u.HomeDir), nil
}

// openBrowser opens a URL in the default browser
func openBrowser(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return object.Errorf("os.open_browser: expected 1 argument, got %d", len(args)), nil
	}

	url, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		return object.Errorf("os.open_browser: unsupported platform: %s", runtime.GOOS), nil
	}

	if err := cmd.Start(); err != nil {
		return object.NewError(err), nil
	}

	return object.Nil, nil
}

// readFile reads the named file and returns bytes
func readFile(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return object.Errorf("os.read_file: expected 1 argument, got %d", len(args)), nil
	}

	path, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return object.NewError(err), nil
	}

	return object.NewBytes(data), nil
}

// writeFile writes data (bytes or string) to the named file
func writeFile(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 2 {
		return object.Errorf("os.write_file: expected 2 arguments, got %d", len(args)), nil
	}

	path, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	var data []byte
	switch v := args[1].(type) {
	case *object.Bytes:
		data = v.Value()
	case *object.String:
		data = []byte(v.Value())
	default:
		return object.Errorf("os.write_file: expected bytes or string as second argument, got %s", args[1].Type()), nil
	}

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return object.NewError(err), nil
	}

	return object.Nil, nil
}

// create creates an empty file (truncating it if it already exists)
func create(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return object.Errorf("os.create: expected 1 argument, got %d", len(args)), nil
	}

	path, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	f, err := os.Create(path)
	if err != nil {
		return object.NewError(err), nil
	}

	if err := f.Close(); err != nil {
		return object.NewError(err), nil
	}

	return object.Nil, nil
}

// copyFile copies a file from src to dst, preserving the source file mode
func copyFile(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 2 {
		return object.Errorf("os.copy_file: expected 2 arguments, got %d", len(args)), nil
	}

	src, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	dst, err := object.AsString(args[1])
	if err != nil {
		return nil, err
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return object.NewError(err), nil
	}

	in, err := os.Open(src)
	if err != nil {
		return object.NewError(err), nil
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, srcInfo.Mode())
	if err != nil {
		return object.NewError(err), nil
	}

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return object.NewError(err), nil
	}

	if err := out.Close(); err != nil {
		return object.NewError(err), nil
	}

	return object.Nil, nil
}

// readDir reads the directory and returns an iterator of directory entries
func readDir(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return object.Errorf("os.read_dir: expected 1 argument, got %d", len(args)), nil
	}

	path, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return object.NewError(err), nil
	}

	// Convert entries to a list of maps
	items := make([]object.Object, len(entries))
	for i, entry := range entries {
		info, _ := entry.Info()
		entryMap := map[string]object.Object{
			"name":   object.NewString(entry.Name()),
			"path":   object.NewString(filepath.Join(path, entry.Name())),
			"is_dir": object.NewBool(entry.IsDir()),
		}
		if info != nil {
			entryMap["size"] = object.NewInt(info.Size())
			entryMap["mode"] = object.NewInt(int64(info.Mode()))
		}
		items[i] = object.NewMap(entryMap)
	}

	return object.NewList(items), nil
}

// mkdirAll creates a directory along with any necessary parents
func mkdirAll(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 1 || len(args) > 2 {
		return object.Errorf("os.mkdir_all: expected 1 or 2 arguments, got %d", len(args)), nil
	}

	path, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	// Default permission is 0755
	perm := os.FileMode(0755)
	if len(args) == 2 {
		permInt, err := object.AsInt(args[1])
		if err != nil {
			return nil, err
		}
		perm = os.FileMode(permInt)
	}

	err = os.MkdirAll(path, perm)
	if err != nil {
		return object.NewError(err), nil
	}

	return object.Nil, nil
}

// mkdir creates a single directory (parents must already exist)
func mkdir(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 1 || len(args) > 2 {
		return object.Errorf("os.mkdir: expected 1 or 2 arguments, got %d", len(args)), nil
	}

	path, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	// Default permission is 0755
	perm := os.FileMode(0755)
	if len(args) == 2 {
		permInt, err := object.AsInt(args[1])
		if err != nil {
			return nil, err
		}
		perm = os.FileMode(permInt)
	}

	err = os.Mkdir(path, perm)
	if err != nil {
		return object.NewError(err), nil
	}

	return object.Nil, nil
}

// removeAll removes a directory or file along with any children it contains
func removeAll(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return object.Errorf("os.remove_all: expected 1 argument, got %d", len(args)), nil
	}

	path, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	err = os.RemoveAll(path)
	if err != nil {
		return object.NewError(err), nil
	}

	return object.Nil, nil
}

// isDir checks if a path is a directory
func isDir(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return object.Errorf("os.is_dir: expected 1 argument, got %d", len(args)), nil
	}

	path, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(path)
	if err != nil {
		// Path doesn't exist or error accessing it
		return object.NewBool(false), nil
	}

	return object.NewBool(info.IsDir()), nil
}

// getwd returns the current working directory
func getwd(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 0 {
		return object.Errorf("os.getwd: expected 0 arguments, got %d", len(args)), nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return object.NewError(err), nil
	}

	return object.NewString(wd), nil
}

// scriptArgsFunc returns the command-line arguments passed to the script,
// as populated by SetScriptArgs
func scriptArgsFunc(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 0 {
		return nil, object.Errorf("os.args: expected 0 arguments, got %d", len(args))
	}

	if runtimeArgs, ok := ArgsFromContext(ctx); ok {
		return object.NewStringList(runtimeArgs), nil
	}
	return object.NewStringList(scriptArgs), nil
}

// execCommand executes a command with arguments
func execCommand(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 1 {
		return object.Errorf("os.exec: expected at least 1 argument, got %d", len(args)), nil
	}

	command, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	var cmdArgs []string
	if len(args) > 1 {
		for _, arg := range args[1:] {
			str, err := object.AsString(arg)
			if err != nil {
				return nil, err
			}
			cmdArgs = append(cmdArgs, str)
		}
	}

	cmd := exec.Command(command, cmdArgs...)
	if err := cmd.Start(); err != nil {
		return object.NewError(err), nil
	}

	return object.Nil, nil
}
