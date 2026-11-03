package browser

import (
	"context"
	"errors"
	"log"
	"net/url"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/gui"
	"github.com/uidbz/fynerisor/gui/guithread"
)

// ErrUnauthorized is the sentinel a LoadContent implementation returns
// for an HTTP 401 (Basic auth) response. Detect it with
// errors.Is(err, browser.ErrUnauthorized) — never error strings (issue #28).
var ErrUnauthorized = errors.New("Unauthorized")

// Config configures a Browser instance
type Config struct {
	// Required interfaces
	MenuProvider MenuProvider
	AuthProvider AuthProvider

	// LoadContent loads a classic Risor page. It runs off the UI thread on a
	// navigation goroutine and must honor ctx: a superseded navigation is
	// cancelled so slow fetches die instead of piling up (issue #28).
	// LoadPage supersedes it when set.
	LoadContent func(ctx context.Context, url, username, password string) (string, error)

	// LoadPage loads a page of either kind (script or scene). When set it is
	// used instead of LoadContent.
	LoadPage LoadPageFunc

	// RenderScene displays scene pages. Without it a scene page reports an
	// error instead of rendering.
	RenderScene RenderSceneFunc

	// Optional callbacks
	OnNavigate         func(url string) error
	OnNavigateComplete func(url string)
	OnNavigateError    func(url string, err error)
	OnStatusChange     func(status string)

	// Optional settings
	HomeURL string
}

// Browser is a generic browser UI for Fyne applications
type Browser struct {
	window           *gui.Window
	history          *History
	url              binding.String
	status           binding.String
	saveState        binding.String
	config           Config
	content          *fyne.Container
	topBar           fyne.CanvasObject
	statusBar        *widget.Label
	saveLabel        *widget.Label
	sideMenu         *fyne.Container
	homeURL          string
	params           url.Values
	sourceProvider   SourceViewProvider
	onOpenInBrowser  func(url string)
	sourceViewActive bool

	currentPage *Page
	editor      EditorProvider
	editing     bool
	editorObj   fyne.CanvasObject

	// Navigation state (issue #28): fetches run on goroutines tagged with
	// a generation id; ctx cancels the superseded one, and results apply
	// on the UI thread only while their generation is current.
	navMu     sync.Mutex
	navCancel context.CancelFunc
	navID     uint64
}

// New creates a new Browser instance
func New(window *gui.Window, config Config) *Browser {
	b := &Browser{
		window:    window,
		config:    config,
		url:       binding.NewString(),
		status:    binding.NewString(),
		saveState: binding.NewString(),
		homeURL:   config.HomeURL,
		params:    url.Values{},
	}

	// Create history with callbacks
	b.history = NewHistory("", b.navigate, b.SetStatus)

	// Build the UI
	b.buildUI()

	return b
}

// buildUI constructs the browser interface
func (b *Browser) buildUI() {
	// Get content container from fynerisor window
	b.content = b.window.GetContentContainer()

	// Navigation toolbar
	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.NavigateBackIcon(), b.history.Back),
		widget.NewToolbarAction(theme.NavigateNextIcon(), b.history.Forward),
		widget.NewToolbarAction(theme.ViewRefreshIcon(), func() { b.Navigate(b.history.Current()) }),
		widget.NewToolbarAction(theme.HomeIcon(), func() {
			if b.homeURL != "" {
				b.navigate(b.homeURL, true)
			}
		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.DocumentCreateIcon(), b.ToggleEdit),
	)
	b.installEditShortcut()

	// Address bar
	addressBar := widget.NewEntryWithData(b.url)
	addressBar.OnSubmitted = func(text string) {
		b.navigate(text, true)
	}

	// Build side menu
	b.buildSideMenu()

	// Burger button
	burgerButton := widget.NewButton("☰", func() {
		if b.sideMenu.Visible() {
			b.sideMenu.Hide()
		} else {
			b.sideMenu.Show()
		}
	})

	// Top bar layout
	b.topBar = container.NewBorder(nil, nil, toolbar, burgerButton, addressBar)

	// Status bar, with a persistent save-state indicator on the right.
	b.statusBar = widget.NewLabelWithData(b.status)
	b.saveLabel = widget.NewLabelWithData(b.saveState)
	b.saveLabel.Importance = widget.LowImportance
	statusRow := container.NewBorder(nil, nil, nil, b.saveLabel, b.statusBar)

	// Main layout
	mainLayout := container.NewBorder(nil, nil, nil, b.sideMenu,
		container.NewBorder(b.topBar, statusRow, nil, nil, b.content))

	b.window.FyneWindow.SetContent(mainLayout)
}

// buildSideMenu creates the side menu from MenuProvider
func (b *Browser) buildSideMenu() {
	menuItems := []fyne.CanvasObject{
		widget.NewLabel("Menu"),
	}

	if b.config.MenuProvider != nil {
		for _, item := range b.config.MenuProvider.GetMenuItems() {
			// Capture item in closure
			action := item.Action
			btn := widget.NewButton(item.Label, func() {
				action()
				b.sideMenu.Hide()
			})
			menuItems = append(menuItems, btn)
		}
	}

	b.sideMenu = container.NewStack(container.NewVBox(menuItems...))
	b.sideMenu.Hide()
}

// SetHomeURL sets the home URL for the home button and navigates to it.
func (b *Browser) SetHomeURL(url string) {
	b.homeURL = url
	b.navigate(url, true)
}

// SetHome sets the home URL for the home button without navigating to it.
// Use this when the startup page differs from the home page.
func (b *Browser) SetHome(url string) {
	b.homeURL = url
}

// Navigate navigates to the given URL and records it in history
func (b *Browser) Navigate(url string) {
	b.navigate(url, true)
}

// Back navigates to the previous history entry. While editing it asks the
// editor first: moving the history cursor before the editor agrees left the
// cursor pointing at a page that was never shown when the user cancelled.
func (b *Browser) Back() {
	if b.editing && b.editor != nil {
		b.editor.BeforeLeave(func(ok bool) {
			if !ok {
				b.SetStatus("Navigation cancelled")
				return
			}
			b.exitEditor()
			b.history.Back()
		})
		return
	}
	b.history.Back()
}

// Forward navigates to the next history entry.
func (b *Browser) Forward() {
	if b.editing && b.editor != nil {
		b.editor.BeforeLeave(func(ok bool) {
			if !ok {
				b.SetStatus("Navigation cancelled")
				return
			}
			b.exitEditor()
			b.history.Forward()
		})
		return
	}
	b.history.Forward()
}

// navigate is the internal navigation method. While an editor is open it asks
// the editor whether unsaved work may be discarded; the navigation is
// cancelled when the editor says no.
func (b *Browser) navigate(url string, recordVisit bool) {
	if b.editing && b.editor != nil {
		b.editor.BeforeLeave(func(ok bool) {
			if !ok {
				b.SetStatus("Navigation cancelled")
				return
			}
			b.exitEditor()
			b.navigateNow(url, recordVisit)
		})
		return
	}
	b.navigateNow(url, recordVisit)
}

// navigateNow is the internal navigation method. The fetch runs on its own
// goroutine so the UI thread never blocks on the network (issue #28);
// starting a navigation cancels the previous one, and results apply
// only while their generation is still current.
func (b *Browser) navigateNow(url string, recordVisit bool) {
	// Correct relative URLs (tie: pages resolve like file paths, not like
	// directory URLs, so links can point at sibling scenes).
	url = resolveRef(b.history.Current(), url)
	log.Println("browser: navigating to", url)

	// Parse query parameters from the URL and expose them to scripts
	b.params = parseParams(url)

	// Store display URL (without index.risor suffix)
	b.url.Set(displayURLFor(url))

	// Call OnNavigate hook
	if b.config.OnNavigate != nil {
		if err := b.config.OnNavigate(url); err != nil {
			b.SetStatus("Navigation cancelled: " + err.Error())
			return
		}
	}

	// Supersede any in-flight navigation (issue #28).
	b.navMu.Lock()
	if b.navCancel != nil {
		b.navCancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	b.navCancel = cancel
	b.navID++
	id := b.navID
	b.navMu.Unlock()

	go func() {
		page, err := b.loadPageContent(ctx, url, "", "")
		if errors.Is(err, ErrUnauthorized) {
			b.ifCurrent(id, func() { b.showAuthDialog(ctx, id, url) })
			return
		}
		if err != nil {
			b.ifCurrent(id, func() {
				if b.config.OnNavigateError != nil {
					b.config.OnNavigateError(url, err)
				}
				b.showErrorPage(url, err)
			})
			return
		}
		b.ifCurrent(id, func() { b.loadPageObject(url, page, recordVisit) })
	}()
}

// ifCurrent runs fn on the UI thread if (and only if) gen is still the
// newest navigation; results of cancelled navigations are dropped.
func (b *Browser) ifCurrent(gen uint64, fn func()) {
	b.navMu.Lock()
	current := gen == b.navID
	b.navMu.Unlock()
	if !current {
		return
	}
	guithread.Do(fn)
}

// loadPage loads a fetched page into the script window (the single
// place that clears, loads, executes, refreshes source view, records
// the visit and reports completion — navigate and the auth retry share
// it now, issue #28). Runs on the UI thread.
func (b *Browser) loadPageObject(url string, page *Page, recordVisit bool) {
	b.disposePage()
	b.currentPage = page
	b.window.Clear()
	if page.Kind == KindScene {
		if b.config.RenderScene == nil {
			b.showErrorPage(url, errors.New("no scene renderer configured"))
		} else if err := b.config.RenderScene(b, page); err != nil {
			if b.config.OnNavigateError != nil {
				b.config.OnNavigateError(url, err)
			}
			b.showErrorPage(url, err)
		}
	} else {
		// Record the page URL as the script base so relative imports resolve
		// against the page, same-origin only (issue #8).
		b.window.LoadScriptFrom(string(page.Body), url)
		b.window.Execute()
	}

	// If source view is active, refresh it with new content
	if b.sourceViewActive && b.sourceProvider != nil {
		newSource := b.sourceProvider.GetCurrentSource()
		if newSource != "" {
			b.renderView(newSource)
		}
	}

	// Record visit
	if recordVisit {
		b.history.Visit(displayURLFor(url))
	}

	// Call completion hook
	if b.config.OnNavigateComplete != nil {
		b.config.OnNavigateComplete(url)
	}
}

// displayURLFor strips the implicit index.risor page name from a URL
// for the address bar and history.
func displayURLFor(url string) string {
	if strings.HasSuffix(url, "/index.risor") {
		return strings.TrimSuffix(url, "/index.risor")
	}
	return url
}

// showAuthDialog asks the AuthProvider for credentials (callback-based,
// never blocking), then retries the fetch on its own goroutine with the
// same generation guard (issue #28). Runs on the UI thread.
func (b *Browser) showAuthDialog(ctx context.Context, gen uint64, url string) {
	if b.config.AuthProvider == nil {
		b.SetStatus("Authentication required but no auth provider configured")
		return
	}

	b.config.AuthProvider.ShowAuthDialog(ctx, url, b.window.FyneWindow,
		func(username, password string, ok bool) {
			if !ok {
				b.SetStatus("Authentication cancelled")
				return
			}
			go func() {
				page, err := b.loadPageContent(ctx, url, username, password)
				if err != nil {
					b.ifCurrent(gen, func() {
						b.SetStatus("Authentication failed: " + err.Error())
					})
					return
				}
				b.ifCurrent(gen, func() { b.loadPageObject(url, page, true) })
			}()
		})
}

// GetURL returns the current URL
func (b *Browser) GetURL() string {
	url, _ := b.url.Get()
	return url
}

// GetParams returns the query parameters of the current URL.
// A copy is returned so callers cannot mutate the browser's state.
func (b *Browser) GetParams() url.Values {
	out := make(url.Values, len(b.params))
	for k, v := range b.params {
		vals := make([]string, len(v))
		copy(vals, v)
		out[k] = vals
	}
	return out
}

// GetParam returns the first value for the given query parameter,
// or an empty string if it is not present.
func (b *Browser) GetParam(name string) string {
	return b.params.Get(name)
}

// SetStatus updates the status bar
func (b *Browser) SetStatus(text string) {
	guithread.Do(func() {
		b.status.Set(text)
	})
	if b.config.OnStatusChange != nil {
		b.config.OnStatusChange(text)
	}
}

// SetSaveState updates the persistent save-state indicator in the status bar
// ("" hides it). Editors use it so "Saving…" and "Save failed" do not scroll
// away with the transient status line.
func (b *Browser) SetSaveState(text string) {
	guithread.Do(func() {
		b.saveState.Set(text)
	})
}

// disposePage releases the live objects of the page being replaced (scene
// views own async loads and galleries that must stop with the page).
func (b *Browser) disposePage() {
	if b.currentPage != nil && b.currentPage.Dispose != nil {
		b.currentPage.Dispose()
		b.currentPage.Dispose = nil
	}
}

// showErrorPage replaces the content with an actionable error panel instead of
// leaving the previous page (or a blank area) on screen.
func (b *Browser) showErrorPage(url string, err error) {
	b.disposePage()
	b.window.Release()
	msg := widget.NewLabel(err.Error())
	msg.Wrapping = fyne.TextWrapWord
	retry := widget.NewButton("Retry", func() { b.navigate(url, false) })
	buttons := []fyne.CanvasObject{retry}
	if b.currentPage != nil && b.editor != nil && b.editor.CanEdit(b.currentPage) {
		buttons = append(buttons, widget.NewButton("Edit page", b.ToggleEdit))
	}
	panel := container.NewVBox(
		widget.NewLabelWithStyle("Could not load this page", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		msg,
		container.NewHBox(buttons...),
	)
	b.content.Objects = []fyne.CanvasObject{container.NewPadded(container.NewCenter(panel))}
	b.content.Refresh()
	b.SetStatus("Error: " + err.Error())
}

// ShowAndRun shows the window and starts the app
func (b *Browser) ShowAndRun() {
	b.window.ShowAndRun()
}

// GetContentContainer returns the content container for direct manipulation
func (b *Browser) GetContentContainer() *fyne.Container {
	return b.content
}

// GetWindow returns the underlying fynerisor window
func (b *Browser) GetWindow() *gui.Window {
	return b.window
}

// NavigateToURL navigates to the specified URL (for programmatic navigation from scripts)
func (b *Browser) NavigateToURL(url string) {
	b.Navigate(url)
}

// CorrectRelativeURL is a convenience export
func CorrectRelativeURL(base, target string) string {
	return correctRelativeURL(base, target)
}

// FixURL is a convenience export
func FixURL(url string) (string, error) {
	return fixURL(url)
}
