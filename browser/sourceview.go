package browser

import (
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// SourceViewProvider defines the interface for providing source code to display
type SourceViewProvider interface {
	// GetCurrentSource returns the current script source code
	GetCurrentSource() string
}

// EnableSourceView adds a source view toggle feature to the browser
// sourceProvider supplies the source code to display
// onOpenInBrowser is called when user clicks "Open in Browser" button (optional)
func (b *Browser) EnableSourceView(sourceProvider SourceViewProvider, onOpenInBrowser func(url string)) {
	b.sourceProvider = sourceProvider
	b.onOpenInBrowser = onOpenInBrowser
}

// ToggleSourceView switches between normal and source view
func (b *Browser) ToggleSourceView() {
	if b.sourceProvider == nil {
		b.SetStatus("Source view not enabled")
		return
	}

	source := b.sourceProvider.GetCurrentSource()
	if source == "" {
		b.SetStatus("No source to display")
		return
	}

	b.sourceViewActive = !b.sourceViewActive
	b.renderView(source)
}

// renderView switches between normal and source view layouts
func (b *Browser) renderView(source string) {
	fyne.Do(func() {
		if b.sourceViewActive {
			// Create text grid for source code with line numbers
			grid := widget.NewTextGridFromString(source)
			grid.ShowLineNumbers = true
			grid.ShowWhitespace = false
			grid.TabWidth = 4

			// Create "Open in Browser" button for the source pane
			var openInBrowserBtn *widget.Button
			if b.onOpenInBrowser != nil {
				openInBrowserBtn = widget.NewButton("Open in Browser", func() {
					currentURL := b.GetURL()
					if currentURL == "" {
						b.SetStatus("No URL to open")
						return
					}
					// If the URL is a directory, append index.risor
					browserURL := currentURL
					if strings.HasSuffix(currentURL, "/") || !strings.Contains(filepath.Base(currentURL), ".") {
						if !strings.HasSuffix(currentURL, "/") {
							browserURL += "/"
						}
						browserURL += "index.risor"
					}
					b.onOpenInBrowser(browserURL)
				})
			}

			// Create source pane with optional button at the bottom
			var sourcePane *fyne.Container
			if openInBrowserBtn != nil {
				sourcePane = container.NewBorder(
					nil,
					openInBrowserBtn,
					nil,
					nil,
					container.NewScroll(grid),
				)
			} else {
				sourcePane = container.NewBorder(
					nil,
					nil,
					nil,
					nil,
					container.NewScroll(grid),
				)
			}

			// Create split view with the rendered app content and source code
			split := container.NewHSplit(
				b.content,
				sourcePane,
			)
			split.Offset = 0.5

			// Build source view layout
			sourceLayout := container.NewBorder(nil, nil, nil, b.sideMenu,
				container.NewBorder(b.topBar, b.statusBar, nil, nil, split))

			// Switch to source view
			b.window.FyneWindow.SetContent(sourceLayout)
			b.window.FyneWindow.Content().Refresh()
			b.SetStatus("Source view active - toggle to return to normal view")
		} else {
			// Rebuild normal view layout
			normalLayout := container.NewBorder(nil, nil, nil, b.sideMenu,
				container.NewBorder(b.topBar, b.statusBar, nil, nil, b.content))

			// Switch to normal view
			b.window.FyneWindow.SetContent(normalLayout)
			b.window.FyneWindow.Content().Refresh()
			b.SetStatus("Normal view")
		}
	})
}

// IsSourceViewActive returns whether source view is currently active
func (b *Browser) IsSourceViewActive() bool {
	return b.sourceViewActive
}
