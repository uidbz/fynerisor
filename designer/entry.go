package designer

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// editorEntry is an Entry that lets the editor see shortcuts and keys first.
//
// The GLFW driver offers every shortcut to the focused widget before the
// canvas, and widget.Entry silently swallows unknown ones — so Ctrl+S did
// nothing while an inspector field had focus. editorEntry forwards the editor
// bindings (Ctrl/Cmd+S, Ctrl/Cmd+E, F4) and otherwise behaves like a normal
// entry.
//
// The Entry is constructed as a zero value and ExtendBaseWidget is called with
// the outer widget: widget.NewEntry would bind BaseWidget to the inner Entry,
// so refreshes would target a second, unpainted renderer.
type editorEntry struct {
	widget.Entry
	ed       *Editor
	onCommit func()
	onCancel func()
}

func newEditorEntry(ed *Editor) *editorEntry {
	e := &editorEntry{ed: ed}
	e.Entry.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
	e.ExtendBaseWidget(e)
	return e
}

func newEditorMultiLineEntry(ed *Editor) *editorEntry {
	e := newEditorEntry(ed)
	e.MultiLine = true
	return e
}

// TypedShortcut forwards editor shortcuts before the Entry handles its own.
func (e *editorEntry) TypedShortcut(s fyne.Shortcut) {
	if e.ed != nil && e.ed.handleShortcut(s) {
		return
	}
	e.Entry.TypedShortcut(s)
}

// TypedKey forwards editor keys (F4, Esc for inline editing) first.
func (e *editorEntry) TypedKey(ev *fyne.KeyEvent) {
	if ev != nil && ev.Name == fyne.KeyEscape && e.onCancel != nil {
		e.onCancel()
		return
	}
	if e.ed != nil && e.ed.handleKey(ev) {
		return
	}
	e.Entry.TypedKey(ev)
}

// FocusLost commits an inline edit like the old inlineEntry did.
func (e *editorEntry) FocusLost() {
	e.Entry.FocusLost()
	if e.onCommit != nil {
		e.onCommit()
	}
}

// FocusGained is the Entry's; kept explicit so the inline hooks above stay
// visible in the type's method set.
func (e *editorEntry) FocusGained() { e.Entry.FocusGained() }
