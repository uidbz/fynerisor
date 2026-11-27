//go:build windows

package hidewindow

import (
	"os/exec"
	"syscall"
	"testing"
)

func TestApplyNoConsoleSetsCreateNoWindow(t *testing.T) {
	cmd := exec.Command("cmd", "/c", "exit")
	apply(cmd, false)
	if cmd.SysProcAttr == nil || cmd.SysProcAttr.CreationFlags&createNoWindow == 0 {
		t.Fatalf("CREATE_NO_WINDOW not set: %+v", cmd.SysProcAttr)
	}
	if cmd.SysProcAttr.HideWindow {
		t.Error("HideWindow must not be set: it hides GUI children too")
	}
	apply(cmd, false) // idempotent
	if cmd.SysProcAttr.CreationFlags != createNoWindow {
		t.Errorf("flags = %#x, want %#x", cmd.SysProcAttr.CreationFlags, createNoWindow)
	}
}

func TestApplyPreservesCallerAttrs(t *testing.T) {
	cmd := exec.Command("cmd", "/c", "exit")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
		CmdLine:       "cmd /c exit",
	}
	apply(cmd, false)
	want := uint32(syscall.CREATE_NEW_PROCESS_GROUP | createNoWindow)
	if cmd.SysProcAttr.CreationFlags != want || cmd.SysProcAttr.CmdLine != "cmd /c exit" {
		t.Errorf("caller attrs clobbered: %+v", cmd.SysProcAttr)
	}
}

func TestApplyRespectsConsoleChoices(t *testing.T) {
	for _, f := range []uint32{detachedProcess, createNewConsole} {
		cmd := exec.Command("cmd")
		cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: f}
		apply(cmd, false)
		if cmd.SysProcAttr.CreationFlags != f {
			t.Errorf("flags %#x changed to %#x", f, cmd.SysProcAttr.CreationFlags)
		}
	}
}

func TestApplyWithHostConsoleIsNoop(t *testing.T) {
	cmd := exec.Command("cmd")
	apply(cmd, true)
	if cmd.SysProcAttr != nil {
		t.Errorf("console host: SysProcAttr must stay nil, got %+v", cmd.SysProcAttr)
	}
}

func TestApplyRunsChild(t *testing.T) {
	cmd := exec.Command("cmd", "/c", "echo ok")
	Apply(cmd)
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	if len(out) == 0 {
		t.Error("child stdout must still be captured")
	}
}
