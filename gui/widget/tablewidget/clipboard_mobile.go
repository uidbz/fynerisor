//go:build android || ios

package tablewidget

// writeClipboard is a no-op on mobile platforms.
//
// The desktop implementation uses golang.design/x/clipboard, which on mobile
// pulls in the standalone golang.org/x/mobile/app package. That package defines
// the same C entry points (ANativeActivity_onCreate, JNI_OnLoad, ...) as Fyne's
// own mobile driver, causing duplicate-symbol link errors. Excluding it here
// keeps mobile builds linkable; table clipboard copy is simply unavailable.
func writeClipboard(text string) {}
