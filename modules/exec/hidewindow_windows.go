//go:build windows

package exec

import (
	"os/exec"
	"syscall"
)

// hideWindow stops a spawned child (console-subsystem binary such as a
// sidecar daemon or a converter) from flashing a black console window when
// the host app is a GUI build. Stdio pipes are unaffected.
func hideWindow(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
}
