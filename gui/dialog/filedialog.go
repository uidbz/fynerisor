package dialog

import (
	"context"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"github.com/uidbz/fyne-addons/kdefilepicker"
	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/gui/objattrs"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &FileDialog{}

const FileDialogType object.Type = "dialog.FileDialog"

// FileDialog wraps the kdefilepicker KDE-style file dialog for Risor
// scripts. The picker is presented as an overlay on the parent window with
// the same lifecycle the stock Fyne file dialog had: Show/Hide may be used
// repeatedly, and accepting or cancelling dismisses the overlay and fires
// the script callback with the chosen path.
type FileDialog struct {
	instance *kdefilepicker.Dialog
	parent   fyne.Window
}

func (obj *FileDialog) Type() object.Type {
	return FileDialogType
}

// Inspect runs on Fyne's UI thread; see guithread.Inspect (issue #49).
func (obj *FileDialog) Inspect() string {
	return guithread.Inspect(obj.inspectUI)
}

func (obj *FileDialog) inspectUI() string {
	return "dialog.FileDialog"
}

func (obj *FileDialog) Interface() interface{} {
	return obj.instance
}

func (obj *FileDialog) IsTruthy() bool {
	return true
}

func (obj *FileDialog) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for %s: %v", FileDialogType, opType)
}

func (obj *FileDialog) Equals(other object.Object) bool {
	return obj == other
}

func (obj *FileDialog) Attrs() []object.AttrSpec {
	return fileDialogAttrs.Specs()
}

// SetAttr runs on Fyne's UI thread; see guithread.SetAttr (issue #49).
func (obj *FileDialog) SetAttr(name string, value object.Object) error {
	return guithread.SetAttr(name, value, obj.setAttrUI)
}

func (obj *FileDialog) setAttrUI(name string, value object.Object) error {
	switch name {
	case "FileName":
		filename, err := object.AsString(value)
		if err != nil {
			return fmt.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
			obj.instance.SetFileName(filename)
		})
		return nil
	}
	return fmt.Errorf("attribute error: %s object has no attribute %q", FileDialogType, name)
}

// GetAttr runs on Fyne's UI thread; see guithread.GetAttr (issue #49).
func (obj *FileDialog) GetAttr(name string) (object.Object, bool) {
	return guithread.GetAttr(name, obj.getAttrUI)
}

var fileDialogAttrs = objattrs.Register(string(FileDialogType), object.NewAttrRegistry[*FileDialog](string(FileDialogType)))

func (obj *FileDialog) getAttrUI(name string) (object.Object, bool) {
	return fileDialogAttrs.GetAttr(obj, name)
}

// NewFileOpen creates an open-file picker. The callback receives the
// selected path, or an empty path with a nil error when the dialog was
// cancelled.
func NewFileOpen(callback func(path string, err error), parent fyne.Window) *FileDialog {
	return newFileDialog(kdefilepicker.ModeOpen, callback, parent)
}

// NewFileSave creates a save-file picker.
func NewFileSave(callback func(path string, err error), parent fyne.Window) *FileDialog {
	return newFileDialog(kdefilepicker.ModeSave, callback, parent)
}

// NewFolderOpen creates a folder picker.
func NewFolderOpen(callback func(path string, err error), parent fyne.Window) *FileDialog {
	return newFileDialog(kdefilepicker.ModeDirectory, callback, parent)
}

// newFileDialog builds the picker up front so Interface() always returns a
// live canvas object. Construction reads the filesystem (initial listing,
// user places) but never touches the canvas, so it is safe to call from a
// script goroutine; presentation happens on the GUI thread via ShowUI.
func newFileDialog(mode kdefilepicker.Mode, callback func(path string, err error), parent fyne.Window) *FileDialog {
	var instance *kdefilepicker.Dialog
	switch mode {
	case kdefilepicker.ModeSave:
		instance = kdefilepicker.NewSaveDialog("", "")
	case kdefilepicker.ModeDirectory:
		instance = kdefilepicker.NewDirectoryDialog("")
	default:
		instance = kdefilepicker.NewOpenDialog("")
	}
	instance.SetOnAccept(func(path string) {
		if callback != nil {
			callback(path, nil)
		}
	})
	instance.SetOnCancel(func() {
		if callback != nil {
			callback("", nil)
		}
	})
	return &FileDialog{instance: instance, parent: parent}
}

// ShowUI presents the dialog as an overlay on the parent window. Safe to
// call from any goroutine; it is the Go-side entry point used by the
// one-shot ShowFileOpen/ShowFileSave/ShowFolderOpen builtins.
func (obj *FileDialog) ShowUI() {
	guithread.Do(func() { obj.instance.ShowOverlay(obj.parent) })
}

// extensionPatterns converts a stock-style extension filter such as
// ".txt" (or "txt", "*.txt", space/semicolon/comma separated lists) into
// kdefilepicker glob patterns.
func extensionPatterns(filter string) []string {
	fields := strings.FieldsFunc(filter, func(r rune) bool {
		return r == ' ' || r == ';' || r == ','
	})
	patterns := make([]string, 0, len(fields))
	for _, f := range fields {
		if !strings.HasPrefix(f, "*") {
			if !strings.HasPrefix(f, ".") {
				f = "." + f
			}
			f = "*" + f
		}
		patterns = append(patterns, f)
	}
	return patterns
}

func init() {
	fileDialogAttrs.
		Define("Show").Args().Returns("nil").Impl(func(obj *FileDialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		obj.ShowUI()
		return object.Nil, nil
	})
	fileDialogAttrs.
		Define("Hide").Args().Returns("nil").Impl(func(obj *FileDialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		guithread.Do(func() { obj.instance.Dismiss() })
		return object.Nil, nil
	})
	fileDialogAttrs.
		Define("Refresh").Args().Returns("nil").Impl(func(obj *FileDialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		guithread.Do(func() { obj.instance.RefreshDir() })
		return object.Nil, nil
	})
	fileDialogAttrs.
		Define("SetFileName").Args("arg1").Returns("nil").Impl(func(obj *FileDialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		filename, err := object.AsString(args[0])
		if err != nil {
			return nil, object.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
			obj.instance.SetFileName(filename)
		})
		return object.Nil, nil
	})
	fileDialogAttrs.
		Define("SetFilter").Args("arg1").Returns("nil").Impl(func(obj *FileDialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		filter, err := object.AsString(args[0])
		if err != nil {
			return nil, object.Errorf("type error: %v", err)
		}
		patterns := extensionPatterns(filter)
		guithread.Do(func() {
			if len(patterns) == 0 {
				obj.instance.SetFilters([]kdefilepicker.Filter{kdefilepicker.AllFiles})
				return
			}
			// Keep "All Files" selectable in the combo, like the stock dialog
			// kept its tree/list toggle, and select the custom filter.
			obj.instance.SetFilters([]kdefilepicker.Filter{
				kdefilepicker.AllFiles,
				{Name: "Filtered", Patterns: patterns},
			})
			obj.instance.SetSelectedFilter(1)
		})
		return object.Nil, nil
	})
	fileDialogAttrs.
		Define("SetLocation").Args("arg1").Returns("nil").Impl(func(obj *FileDialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		path, err := object.AsString(args[0])
		if err != nil {
			return nil, object.Errorf("type error: %v", err)
		}
		guithread.Do(func() {
			obj.instance.SetDirectory(path)
		})
		return object.Nil, nil
	})
	fileDialogAttrs.
		Define("SetConfirmText").Args("arg1").Returns("nil").Impl(func(obj *FileDialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		text, err := object.AsString(args[0])
		if err != nil {
			return nil, object.Errorf("type error: %v", err)
		}
		guithread.Do(func() { obj.instance.SetOpenLabel(text) })
		return object.Nil, nil
	})
	fileDialogAttrs.
		Define("SetDismissText").Args("arg1").Returns("nil").Impl(func(obj *FileDialog, ctx context.Context, args ...object.Object) (object.Object, error) {
		text, err := object.AsString(args[0])
		if err != nil {
			return nil, object.Errorf("type error: %v", err)
		}
		guithread.Do(func() { obj.instance.SetCancelLabel(text) })
		return object.Nil, nil
	})
}
