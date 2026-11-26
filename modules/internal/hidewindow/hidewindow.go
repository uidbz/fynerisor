// Package hidewindow keeps console-subsystem children spawned by the
// fynerisor modules (exec, os.exec, os.open_browser) from flashing a console
// window when the host is a Windows GUI build. It is a no-op elsewhere.
package hidewindow
