package browser

import (
	"context"

	"fyne.io/fyne/v2"
)

// PageKind distinguishes the two page formats a browser can display.
type PageKind int

const (
	// KindRisor is a Risor script page, the classic fynerisor browser page.
	KindRisor PageKind = iota
	// KindScene is a fynerisor scene document (JSON).
	KindScene
)

// Page is one loaded document. Version and Writable are used by editors for
// optimistic concurrency and to decide whether an edit command is offered.
type Page struct {
	URL      string // the URL the page was loaded from (not the display URL)
	Kind     PageKind
	Body     []byte
	Version  string
	Writable bool
	// Script is an optional sidecar script, already fetched off the UI thread.
	// ScriptErr reports why it is missing; the page still renders data-only.
	Script    []byte
	ScriptErr error
	// Dispose, when set by the renderer, releases the page's live objects
	// (scene views own async loads that must stop when the page is replaced).
	Dispose func()
}

// EditorProvider turns a page into an editor UI. The browser stops the page's
// script before calling Open and reloads the page after the editor closes.
//
// BeforeLeave is called before a navigation would replace an open editor. It
// must call done exactly once; done(false) cancels the navigation.
type EditorProvider interface {
	CanEdit(p *Page) bool
	Open(p *Page, b *Browser) fyne.CanvasObject
	BeforeLeave(done func(ok bool))
}

// LoadPageFunc loads one document. It runs off the UI thread and must honour
// ctx. It replaces the simple LoadContent callback when a host supports scene
// pages or wants version/metadata information.
type LoadPageFunc func(ctx context.Context, url, username, password string) (*Page, error)

// RenderSceneFunc displays a scene page. It is called on the UI thread with the
// page's bytes and must arrange for the scene to appear in the browser window
// (usually by building a render.View and executing a page script). Returning an
// error shows the navigation error.
type RenderSceneFunc func(b *Browser, p *Page) error
