package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/browser"
	"github.com/uidbz/fynerisor/core"
	"github.com/uidbz/fynerisor/gui"
)

const version = "0.1.0"

func main() {
	// Parse command-line flags
	homeURL := flag.String("home", "https://example.com", "Home URL to load on startup")
	flag.Parse()

	// If URL provided as argument, use it instead
	if flag.NArg() > 0 {
		*homeURL = flag.Arg(0)
	}

	// Create the browser application
	app := NewBrowserApp("fynerisor-browser")
	app.SetHomeURL(*homeURL)
	app.ShowAndRun()
}

// BrowserApp is a minimal browser application demonstrating the browser package
type BrowserApp struct {
	app     fyne.App
	window  *gui.Window
	browser *browser.Browser
	loader  *ScriptLoader
	source  string // Current script source for source view
}

// NewBrowserApp creates a new browser application
func NewBrowserApp(title string) *BrowserApp {
	b := &BrowserApp{
		app:    app.New(),
		loader: NewScriptLoader(),
	}

	// Set version for require() checks
	core.SetAppVersion(version)

	// Create fynerisor window
	fyneWindow := b.app.NewWindow(title)
	b.window = gui.NewWindow(fyneWindow,
		gui.WithAppName("fynerisor-browser"),
		gui.WithStatusCallback(func(status string) {
			// Status updates handled by browser
		}),
		gui.WithResultCallback(func(result string) {
			// Print results to console
			if result != "" {
				fmt.Println(result)
			}
		}),
		// Add modules your scripts need
		gui.WithHTTP(),
		gui.WithOS(),
		gui.WithStrings(),
		gui.WithFilepath(),
		gui.WithTime(),
		gui.WithIO(),
	)

	// Create browser with plugins
	b.browser = browser.New(b.window, browser.Config{
		MenuProvider: &BrowserMenuProvider{app: b},
		AuthProvider: &BrowserAuthProvider{app: b},
		LoadContent:  b.loadScript,
		OnNavigate: func(url string) error {
			log.Printf("Navigating to: %s", url)
			return nil
		},
		OnNavigateError: func(url string, err error) {
			b.showError("Navigation Error", err.Error())
		},
	})

	// Register browser global for programmatic navigation from scripts
	// This allows scripts to call browser.Open(url), browser.GetURL(), browser.SetStatus()
	browserObj := browser.NewRisorBrowser(b.browser)
	b.window.RegisterGlobal("browser", browserObj)

	// Enable source view
	b.browser.EnableSourceView(b, nil)

	return b
}

// SetHomeURL sets the home URL and navigates to it
func (b *BrowserApp) SetHomeURL(url string) {
	b.browser.SetHomeURL(url)
}

// ShowAndRun displays the window and starts the app
func (b *BrowserApp) ShowAndRun() {
	b.browser.ShowAndRun()
}

// GetCurrentSource implements browser.SourceViewProvider
func (b *BrowserApp) GetCurrentSource() string {
	return b.source
}

// loadScript loads and returns script content from URL
func (b *BrowserApp) loadScript(url, username, password string) (string, error) {
	script, err := b.loader.LoadScript(url, username, password)
	if err != nil {
		return "", err
	}

	// Store for source view
	b.source = script

	return script, nil
}

// showError displays an error page
func (b *BrowserApp) showError(title, message string) {
	errorScript := fmt.Sprintf(`
let title = widget.NewLabel(%q)
title.TextStyle.Bold = true

let msg = widget.NewLabel(%q)

let content = container.NewVBox(
    title,
    widget.NewSeparator(),
    msg
)

window.SetContent(content)
`, title, message)

	b.window.Clear()
	b.source = errorScript
	b.window.LoadScript(errorScript)
	b.window.Execute()
}

// showAbout displays the about page
func (b *BrowserApp) showAbout() {
	aboutScript := fmt.Sprintf(`
let title = widget.NewLabel("fynerisor-browser")
title.TextStyle.Bold = true
title.Alignment = 1

let version_label = widget.NewLabel("Version %s")
version_label.Alignment = 1

let description = widget.NewLabel("Reference browser implementation")
description.Alignment = 1

let separator = widget.NewSeparator()

let info = widget.NewLabel("Built with:")
let fynerisor_label = widget.NewLabel("• fynerisor - Risor bindings for Fyne UI")
let risor_label = widget.NewLabel("• Risor v2 - Fast and flexible scripting")
let fyne_label = widget.NewLabel("• Fyne - Cross-platform GUI toolkit")

let separator2 = widget.NewSeparator()

let docs_link = widget.NewHyperlink("fynerisor Documentation", "https://github.com/uidbz/fynerisor")

let content = container.NewVBox(
    title,
    version_label,
    description,
    separator,
    info,
    fynerisor_label,
    risor_label,
    fyne_label,
    separator2,
    docs_link
)

window.SetContent(content)
`, version)

	b.window.Clear()
	b.source = aboutScript
	b.window.LoadScript(aboutScript)
	b.window.Execute()
}

// BrowserMenuProvider implements browser.MenuProvider
type BrowserMenuProvider struct {
	app *BrowserApp
}

func (p *BrowserMenuProvider) GetMenuItems() []browser.MenuItem {
	return []browser.MenuItem{
		{
			Label: "View Source",
			Action: func() {
				p.app.browser.ToggleSourceView()
			},
		},
		{
			Label: "About",
			Action: func() {
				p.app.showAbout()
			},
		},
		{
			Label: "Quit",
			Action: func() {
				p.app.app.Quit()
			},
		},
	}
}

// BrowserAuthProvider implements browser.AuthProvider
type BrowserAuthProvider struct {
	app *BrowserApp
}

func (p *BrowserAuthProvider) ShowAuthDialog(url string, contentContainer *fyne.Container) (string, string, bool) {
	userEntry := widget.NewEntry()
	userEntry.PlaceHolder = "Username"

	passEntry := widget.NewEntry()
	passEntry.Password = true
	passEntry.PlaceHolder = "Password"

	label := widget.NewLabel(fmt.Sprintf("Authentication required for:\n%s", url))
	label.Wrapping = fyne.TextWrapWord

	form := widget.NewForm(
		widget.NewFormItem("Username:", userEntry),
		widget.NewFormItem("Password:", passEntry),
	)
	form.SubmitText = "Login"

	submitted := false
	form.OnSubmit = func() {
		submitted = true
	}

	form.OnCancel = func() {
		submitted = false
	}

	// Display in container
	fyne.Do(func() {
		contentContainer.Objects = []fyne.CanvasObject{
			container.NewVBox(label, form),
		}
		contentContainer.Refresh()
	})

	// Note: This is a simplified implementation. In a real app, you'd need
	// proper event handling to wait for form submission before returning.
	return userEntry.Text, passEntry.Text, submitted
}

// ScriptLoader handles fetching scripts from URLs or filesystem
type ScriptLoader struct {
	httpClient *http.Client
}

// NewScriptLoader creates a new script loader
func NewScriptLoader() *ScriptLoader {
	return &ScriptLoader{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// LoadScript fetches a script from the given URL with optional authentication
func (sl *ScriptLoader) LoadScript(url, username, password string) (string, error) {
	// Normalize the URL - convert file paths to file:// URLs
	normalizedURL := normalizeURL(url)

	fixedURL, err := browser.FixURL(normalizedURL)
	if err != nil {
		return "", err
	}

	// Handle file:// URLs
	if strings.HasPrefix(fixedURL, "file://") {
		return sl.loadFromFile(fixedURL)
	}

	// Handle HTTP(S) URLs
	return sl.loadFromHTTP(fixedURL, username, password)
}

// normalizeURL converts file paths to file:// URLs
func normalizeURL(url string) string {
	// Already a URL scheme
	if strings.HasPrefix(url, "http://") ||
	   strings.HasPrefix(url, "https://") ||
	   strings.HasPrefix(url, "file://") {
		return url
	}

	// Check if it looks like a file path
	// Unix absolute path: starts with /
	// Windows absolute path: starts with X:\ or X:/ (drive letter)
	// Relative path: ./ or ../ or just a filename
	isUnixPath := strings.HasPrefix(url, "/")
	isWindowsPath := len(url) >= 3 && url[1] == ':' && (url[2] == '\\' || url[2] == '/')
	isRelativePath := strings.HasPrefix(url, "./") || strings.HasPrefix(url, "../") ||
	                  !strings.Contains(url, "://")

	if isUnixPath || isWindowsPath || isRelativePath {
		// Convert to absolute path
		absPath, err := filepath.Abs(url)
		if err != nil {
			// If we can't get absolute path, just prepend file://
			// and let the error handling deal with it later
			absPath = url
		}

		// Convert to file:// URL
		// filepath.ToSlash converts Windows backslashes to forward slashes
		absPath = filepath.ToSlash(absPath)

		// Ensure it starts with /
		if !strings.HasPrefix(absPath, "/") {
			absPath = "/" + absPath
		}

		return "file://" + absPath
	}

	// Assume it's a domain/hostname, return as-is for browser.FixURL to handle
	return url
}

func (sl *ScriptLoader) loadFromFile(fileURL string) (string, error) {
	filePath := strings.TrimPrefix(fileURL, "file://")
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return string(content), nil
}

func (sl *ScriptLoader) loadFromHTTP(url, username, password string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	if username != "" {
		req.SetBasicAuth(username, password)
	}

	resp, err := sl.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check for auth required
	if resp.StatusCode == http.StatusUnauthorized {
		authHeader := resp.Header.Get("WWW-Authenticate")
		if strings.HasPrefix(authHeader, "Basic") {
			return "", errors.New("Unauthorized")
		}
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %s", resp.Status)
	}

	// Read the body
	limitReader := io.LimitReader(resp.Body, 10*1024*1024) // 10MB limit
	bodyBytes, err := io.ReadAll(limitReader)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}
