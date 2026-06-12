// Package os provides basic OS functionality for Risor scripts.
package os

import (
	"context"
	"os"
	"os/exec"
	"os/user"
	"runtime"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

// Module returns the os module for risor
func Module() *object.Module {
	return object.NewBuiltinsModule("os", map[string]object.Object{
		"goos":         object.NewBuiltin("os.goos", goos),
		"current_user": object.NewBuiltin("os.current_user", currentUser),
		"open_browser": object.NewBuiltin("os.open_browser", openBrowser),
		"read_file":    object.NewBuiltin("os.read_file", readFile),
		"write_file":   object.NewBuiltin("os.write_file", writeFile),
		"read_dir":     object.NewBuiltin("os.read_dir", readDir),
		"mkdir_all":    object.NewBuiltin("os.mkdir_all", mkdirAll),
		"is_dir":       object.NewBuiltin("os.is_dir", isDir),
		"getwd":        object.NewBuiltin("os.getwd", getwd),
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
