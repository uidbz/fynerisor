package hidewindow

import (
	"os/exec"
	"runtime"
	"testing"
)

// Outside Windows Apply must not touch the command at all.
func TestApplyOtherPlatformsNoop(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("covered by hidewindow_windows_test.go")
	}
	cmd := exec.Command("true")
	Apply(cmd)
	if cmd.SysProcAttr != nil {
		t.Errorf("SysProcAttr must stay nil, got %+v", cmd.SysProcAttr)
	}
}
