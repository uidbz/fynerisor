package browser

import (
	"context"
	"errors"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"

	"github.com/uidbz/fynerisor/gui"
)

// EnableEditor installs the page editor provider used by the toolbar pencil and
// F4. Call it after New (the toolbar action always exists and reports when no
// provider is configured).
func (b *Browser) EnableEditor(provider EditorProvider) { b.editor = provider }

// IsEditing reports whether the editor is currently replacing the page view.
func (b *Browser) IsEditing() bool { return b.editing }

// CurrentPage returns the page currently displayed, or nil.
func (b *Browser) CurrentPage() *Page { return b.currentPage }

// Window returns the fynerisor window (editor providers and scene renderers
// need it for scripts and dialogs).
func (b *Browser) Window() *gui.Window { return b.window }

// ToggleEdit enters or leaves edit mode. A page is only editable when the
// provider accepts it. Leaving goes through the editor's own close path when it
// offers one, so unsaved changes are not discarded silently.
func (b *Browser) ToggleEdit() {
	if b.editing {
		if closer, ok := b.editor.(interface{ RequestClose() }); ok {
			closer.RequestClose()
			return
		}
		b.CloseEditor(true)
		return
	}
	if b.editor == nil {
		b.SetStatus("No editor configured")
		return
	}
	p := b.currentPage
	if p == nil {
		b.SetStatus("No page loaded")
		return
	}
	if !b.editor.CanEdit(p) {
		b.SetStatus("This page is not editable")
		return
	}
	// Stop the page script first: the designer must be the only owner of the
	// document while it is open, and no widget callbacks may race the editor.
	// The rendered view is replaced by the editor; release its async loads.
	b.window.Release()
	b.disposePage()
	obj := b.editor.Open(p, b)
	if obj == nil {
		b.SetStatus("Editor failed to open")
		b.navigateNow(p.URL, false)
		return
	}
	b.editing = true
	b.editorObj = obj
	content := b.window.GetContentContainer()
	content.Objects = []fyne.CanvasObject{obj}
	content.Refresh()
	b.SetStatus("Editing — Done saves and reloads the page")
}

// CloseEditor leaves edit mode and reloads the current page from its store.
// It is called by the editor (Done) and by ToggleEdit.
func (b *Browser) CloseEditor(reload bool) {
	if !b.editing {
		return
	}
	p := b.currentPage
	b.editing = false
	b.editorObj = nil
	// The provider may not know the editor is gone (a programmatic close, the
	// window closing). Tell it, so its autosave goroutine stops instead of
	// writing drafts for a dead editor.
	if c, ok := b.editor.(interface{ EditorClosed() }); ok {
		c.EditorClosed()
	}
	b.SetSaveState("")
	if p == nil {
		return
	}
	b.SetStatus("Reloading page…")
	b.navigateNow(p.URL, false)
}

func (b *Browser) exitEditor() {
	b.editing = false
	b.editorObj = nil
	if c, ok := b.editor.(interface{ EditorClosed() }); ok {
		c.EditorClosed()
	}
}

func (b *Browser) installEditShortcut() {
	if b.window == nil || b.window.FyneWindow == nil {
		return
	}
	c := b.window.FyneWindow.Canvas()
	toggle := func(fyne.Shortcut) { b.ToggleEdit() }
	// The GLFW driver only builds custom shortcuts for keys with a non-Shift
	// modifier, so F4 could never be a canvas shortcut. Ctrl/Cmd+E is the
	// reliable binding; F4 is handled as a plain key when nothing is focused
	// (the designer overlay and its fields handle it themselves while editing).
	c.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyE, Modifier: fyne.KeyModifierControl}, toggle)
	c.AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyE, Modifier: fyne.KeyModifierSuper}, toggle)
	c.SetOnTypedKey(func(ev *fyne.KeyEvent) {
		if ev != nil && ev.Name == fyne.KeyF4 {
			b.ToggleEdit()
		}
	})
}

// loadPageContent fetches a page of either kind. It runs off the UI thread.
func (b *Browser) loadPageContent(ctx context.Context, url, username, password string) (*Page, error) {
	if b.config.LoadPage != nil {
		p, err := b.config.LoadPage(ctx, url, username, password)
		if err != nil {
			return nil, err
		}
		if p == nil {
			return nil, errors.New("browser: LoadPage returned no page")
		}
		if p.URL == "" {
			p.URL = url
		}
		if p.Kind == KindRisor && (isSceneURL(p.URL) || looksLikeScene(p.Body)) {
			p.Kind = KindScene
		}
		return p, nil
	}
	if b.config.LoadContent == nil {
		return nil, errors.New("browser: no LoadContent or LoadPage configured")
	}
	content, err := b.config.LoadContent(ctx, url, username, password)
	if err != nil {
		return nil, err
	}
	p := &Page{URL: url, Kind: KindRisor, Body: []byte(content)}
	if isSceneURL(url) || looksLikeScene(p.Body) {
		p.Kind = KindScene
	}
	return p, nil
}

func isSceneURL(url string) bool {
	u := url
	if i := strings.IndexAny(u, "?#"); i >= 0 {
		u = u[:i]
	}
	return strings.HasSuffix(u, ".scene.json")
}

func looksLikeScene(body []byte) bool {
	s := strings.TrimSpace(string(body))
	return strings.HasPrefix(s, "{") && strings.Contains(s[:minInt(len(s), 160)], `"fynerisor.scene/`)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
