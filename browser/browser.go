package browser

import (
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/gui"
)

// Config configures a Browser instance
type Config struct {
	// Required interfaces
	MenuProvider MenuProvider
	AuthProvider AuthProvider

	// Required callback
	LoadContent func(url, username, password string) (string, error)

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
	config           Config
	content          *fyne.Container
	topBar           fyne.CanvasObject
	statusBar        *widget.Label
	sideMenu         *fyne.Container
	homeURL          string
	sourceProvider   SourceViewProvider
	onOpenInBrowser  func(url string)
	sourceViewActive bool
}

// New creates a new Browser instance
func New(window *gui.Window, config Config) *Browser {
	b := &Browser{
		window:  window,
		config:  config,
		url:     binding.NewString(),
		status:  binding.NewString(),
		homeURL: config.HomeURL,
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
	)

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

	// Status bar
	b.statusBar = widget.NewLabelWithData(b.status)

	// Main layout
	mainLayout := container.NewBorder(nil, nil, nil, b.sideMenu,
		container.NewBorder(b.topBar, b.statusBar, nil, nil, b.content))

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

// navigate is the internal navigation method
func (b *Browser) navigate(url string, recordVisit bool) {
	// Correct relative URLs
	url = CorrectRelativeURL(b.history.Current(), url)
	log.Println("browser: navigating to", url)

	// Store display URL (without index.risor suffix)
	displayURL := url
	if strings.HasSuffix(url, "/index.risor") {
		displayURL = strings.TrimSuffix(url, "/index.risor")
	}
	b.url.Set(displayURL)

	// Call OnNavigate hook
	if b.config.OnNavigate != nil {
		if err := b.config.OnNavigate(url); err != nil {
			b.SetStatus("Navigation cancelled: " + err.Error())
			return
		}
	}

	// Load content
	content, err := b.config.LoadContent(url, "", "")
	if err != nil {
		// Check if it's an auth error
		if err.Error() == "Unauthorized" {
			b.showAuthDialog(url)
			return
		}

		// Call error hook
		if b.config.OnNavigateError != nil {
			b.config.OnNavigateError(url, err)
		}
		b.SetStatus("Error: " + err.Error())
		return
	}

	// Clear and load the script
	b.window.Clear()
	b.window.LoadScript(content)
	b.window.Execute()

	// If source view is active, refresh it with new content
	if b.sourceViewActive && b.sourceProvider != nil {
		newSource := b.sourceProvider.GetCurrentSource()
		if newSource != "" {
			b.renderView(newSource)
		}
	}

	// Record visit
	if recordVisit {
		b.history.Visit(displayURL)
	}

	// Call completion hook
	if b.config.OnNavigateComplete != nil {
		b.config.OnNavigateComplete(url)
	}
}

// showAuthDialog displays the authentication dialog
func (b *Browser) showAuthDialog(url string) {
	if b.config.AuthProvider == nil {
		b.SetStatus("Authentication required but no auth provider configured")
		return
	}

	username, password, submitted := b.config.AuthProvider.ShowAuthDialog(url, b.content)
	if !submitted {
		b.SetStatus("Authentication cancelled")
		return
	}

	// Retry with credentials
	content, err := b.config.LoadContent(url, username, password)
	if err != nil {
		b.SetStatus("Authentication failed: " + err.Error())
		return
	}

	// Load the authenticated content
	b.window.Clear()
	b.window.LoadScript(content)
	b.window.Execute()

	// If source view is active, refresh it with new content
	if b.sourceViewActive && b.sourceProvider != nil {
		newSource := b.sourceProvider.GetCurrentSource()
		if newSource != "" {
			b.renderView(newSource)
		}
	}

	// Record visit
	displayURL := url
	if strings.HasSuffix(url, "/index.risor") {
		displayURL = strings.TrimSuffix(url, "/index.risor")
	}
	b.history.Visit(displayURL)

	if b.config.OnNavigateComplete != nil {
		b.config.OnNavigateComplete(url)
	}
}

// GetURL returns the current URL
func (b *Browser) GetURL() string {
	url, _ := b.url.Get()
	return url
}

// SetStatus updates the status bar
func (b *Browser) SetStatus(text string) {
	b.status.Set(text)
	if b.config.OnStatusChange != nil {
		b.config.OnStatusChange(text)
	}
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
