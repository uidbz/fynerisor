//go:build windows

package hidewindow

import (
	"os/exec"
	"syscall"
)

// Win32 process creation flags (winbase.h).
const (
	detachedProcess  = 0x00000008 // DETACHED_PROCESS
	createNewConsole = 0x00000010 // CREATE_NEW_CONSOLE
	createNoWindow   = 0x08000000 // CREATE_NO_WINDOW
)

var procGetConsoleWindow = syscall.NewLazyDLL("kernel32.dll").NewProc("GetConsoleWindow")

// hostHasConsole reports whether this process is attached to a console
// (a console-subsystem build run from a terminal, or one that called
// AllocConsole). GUI builds (-H windowsgui) have none.
func hostHasConsole() bool {
	if procGetConsoleWindow.Find() != nil {
		return false
	}
	h, _, _ := procGetConsoleWindow.Call()
	return h != 0
}

// Apply sets CREATE_NO_WINDOW on cmd when the host has no console, so a
// console-subsystem child (a converter, a sidecar daemon, cmd.exe) runs
// without opening a visible console window. GUI-subsystem children ignore
// the flag and show their windows normally; this is why it is used instead
// of SysProcAttr.HideWindow, which forces SW_HIDE on the child's first
// ShowWindow and leaves launched GUI programs invisible. When the host has
// a console (CLI builds) children keep sharing it, so nothing changes.
//
// Apply must be called before cmd is started. It allocates cmd.SysProcAttr
// if nil and only ORs into CreationFlags, preserving every other field the
// caller set; it is idempotent. It leaves the flags alone if the caller
// asked for DETACHED_PROCESS or CREATE_NEW_CONSOLE (Windows ignores
// CREATE_NO_WINDOW with either).
func Apply(cmd *exec.Cmd) {
	apply(cmd, hostHasConsole())
}

func apply(cmd *exec.Cmd, hasConsole bool) {
	if hasConsole {
		return
	}
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	if cmd.SysProcAttr.CreationFlags&(detachedProcess|createNewConsole) != 0 {
		return
	}
	cmd.SysProcAttr.CreationFlags |= createNoWindow
}
