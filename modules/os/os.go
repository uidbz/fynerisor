// Package os provides basic OS functionality for Risor scripts.
package os

import (
	"context"
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
