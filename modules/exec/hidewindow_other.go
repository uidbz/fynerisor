//go:build !windows

package exec

import "os/exec"

func hideWindow(_ *exec.Cmd) {}
