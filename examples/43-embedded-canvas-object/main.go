package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/gui"
)

// This example embeds a Risor script as a regular Fyne CanvasObject inside an
// application whose window and surrounding UI are written in plain Go.
func main() {
	a := app.New()
	host := a.NewWindow("fynerisor embedded in a Fyne app")

	script := `
		require(["v0.9", "@gui"])

		let count = 0
		let label = widget.NewLabel("The script has not been clicked")
		let button = widget.NewButton("Script button", () => {
			count = count + 1
			label.SetText(sprintf("Script clicked %d times", count))
			window.SetStatus(sprintf("count=%d", count))
		})

		window.SetContent(container.NewVBox(label, button))
	`

	// WithHostWindow gives the script access to dialogs, shortcuts and menus.
	// The view never replaces host content.
	view := gui.NewCanvasObject(script, gui.WithHostWindow(host))
	view.Execute()

	header := widget.NewLabel("This header is plain Fyne; the box below is Risor.")
	status := widget.NewLabel("Go side is independent")
	goButton := widget.NewButton("Go button", func() {
		status.SetText("Go button clicked")
	})

	host.SetContent(container.NewBorder(
		container.NewVBox(header, goButton, status, widget.NewSeparator()),
		nil, nil, nil,
		view,
	))
	host.Resize(fyne.NewSize(720, 480))
	host.ShowAndRun()
}
