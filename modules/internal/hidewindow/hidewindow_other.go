//go:build !windows

package hidewindow

import "os/exec"

// Apply is a no-op outside Windows.
func Apply(_ *exec.Cmd) {}
