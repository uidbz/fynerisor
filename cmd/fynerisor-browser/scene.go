package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"fyne.io/fyne/v2"

	"github.com/uidbz/fynerisor/browser"
	"github.com/uidbz/fynerisor/designer"
	guiscene "github.com/uidbz/fynerisor/gui/scene"
	"github.com/uidbz/fynerisor/scene"
	renderbuiltin "github.com/uidbz/fynerisor/scene/render/builtin"
)

// loadPage implements browser.Config.LoadPage. Scene pages and script pages
// share the loader; the browser detects scene documents by extension or by the
// format marker in the body.
func (b *BrowserApp) loadPage(ctx context.Context, rawURL, username, password string) (*browser.Page, error) {
	body, err := b.loader.LoadScript(ctx, rawURL, username, password)
	if err != nil {
		return nil, err
	}
	data := []byte(body)
	p := &browser.Page{
		URL:      rawURL,
		Body:     data,
		Version:  scene.VersionOf(data),
		Writable: strings.HasPrefix(rawURL, "file://"),
	}
	if strings.HasSuffix(strings.SplitN(rawURL, "?", 2)[0], ".scene.json") {
		p.Kind = browser.KindScene
	}
	// Source view shows the page's own source in both cases.
	b.source = body
	return p, nil
}

// renderScene implements browser.Config.RenderScene.
func (b *BrowserApp) renderScene(br *browser.Browser, p *browser.Page) error {
	reg := renderbuiltin.NewRegistry()
	doc, issues, err := scene.DecodeValidated(p.Body, reg.Registry)
	if err != nil {
		return err
	}
	if scene.Errors(issues) > 0 {
		return fmt.Errorf("scene validation: %s", firstSceneError(issues))
	}

	cfg := guiscene.Config{
		Registry:  reg,
		Window:    br.Window(),
		Navigator: browserNav{br},
		Opener:    browserOpener{},
		Status:    br.SetStatus,
	}
	page, err := guiscene.BuildPage(context.Background(), doc, cfg, p.URL, p.Version)
	if err != nil {
		return err
	}

	script := "window.SetContent(page)\n"
	if doc.Script != "" {
		ref, err := scene.ResolveRef(p.URL, doc.Script)
		if err != nil {
			return fmt.Errorf("scene script: %w", err)
		}
		src, err := b.loader.LoadScript(context.Background(), ref, "", "")
		if err != nil {
			return fmt.Errorf("scene script: %w", err)
		}
		script = src + "\nwindow.SetContent(page)\n"
	}
	br.Window().LoadScript(script)
	br.Window().ExecuteWith(map[string]any{
		"page":  page,
		"scene": guiscene.NewScene(cfg, p.URL),
	})
	return nil
}

func firstSceneError(issues []scene.Issue) string {
	for _, i := range issues {
		if i.Severity == scene.Error {
			return i.String()
		}
	}
	if len(issues) > 0 {
		return issues[0].String()
	}
	return "unknown error"
}

// browserNav connects declarative navigate/back actions to the browser.
type browserNav struct{ b *browser.Browser }

func (n browserNav) Navigate(ref string) { n.b.Navigate(ref) }
func (n browserNav) Back()               { n.b.Back() }

// browserOpener opens external URLs in the system browser. Media playback is a
// tie-browser concern and is registered there through its own Opener.
type browserOpener struct{}

func (browserOpener) Open(ref string) {
	if u, err := url.Parse(ref); err == nil {
		_ = fyne.CurrentApp().OpenURL(u)
	}
}
func (browserOpener) Play(ref string) { browserOpener{}.Open(ref) }
func (browserOpener) Pause()          {}

// BrowserEditorProvider implements browser.EditorProvider for scene documents.
type BrowserEditorProvider struct{ app *BrowserApp }

func (p *BrowserEditorProvider) CanEdit(page *browser.Page) bool {
	return page != nil && page.Kind == browser.KindScene && page.Writable
}

func (p *BrowserEditorProvider) Open(page *browser.Page, br *browser.Browser) fyne.CanvasObject {
	reg := renderbuiltin.NewRegistry()
	doc, err := scene.Decode(page.Body)
	if err != nil {
		br.SetStatus("Cannot edit: " + err.Error())
		return nil
	}
	var store scene.Store = scene.FileStore{}
	if strings.HasPrefix(page.URL, "http://") || strings.HasPrefix(page.URL, "https://") {
		store = scene.NewHTTPStore(true)
	}
	ed := designer.New(doc, designer.Options{
		Registry:  reg,
		Store:     store,
		Ref:       page.URL,
		Version:   page.Version,
		Window:    br.Window().FyneWindow,
		Navigator: browserNav{br},
		Opener:    browserOpener{},
		Status:    br.SetStatus,
		OnClose: func(reload bool) {
			p.app.editor = nil
			br.CloseEditor(reload)
		},
		OnDirty: func(dirty bool) {
			if dirty {
				br.SetStatus("Unsaved changes — press Done to save")
			}
		},
	})
	p.app.editor = ed
	return ed.Object()
}

func (p *BrowserEditorProvider) BeforeLeave(done func(ok bool)) {
	if p.app.editor == nil || p.app.editor.Closed() {
		p.app.editor = nil
		done(true)
		return
	}
	p.app.editor.BeforeLeave(done)
}

// EditorClosed is called by the browser when it abandons an editor because the
// user chose to navigate away.
func (p *BrowserEditorProvider) EditorClosed() { p.app.editor = nil }
