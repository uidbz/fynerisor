// Package builtin provides the Fyne implementations of the standard scene
// components. It binds factories to the pure-data specs in scene/builtin, so
// validation, the inspector and the renderer all agree on names and defaults.
package builtin

import (
	scenebuiltin "github.com/uidbz/fynerisor/scene/builtin"
	"github.com/uidbz/fynerisor/scene/render"
)

// NewRegistry returns a render registry with every standard component
// registered.
func NewRegistry() *render.Registry {
	specs := scenebuiltin.Registry()
	r := render.NewRegistry(specs)
	RegisterFactories(r)
	return r
}

// RegisterFactories binds the standard factories to an existing registry. The
// registry must already contain the matching specs.
func RegisterFactories(r *render.Registry) {
	// Layout.
	r.Factory("free", newFree)
	for _, kind := range []string{"vbox", "hbox", "grid", "gridwrap", "stack", "center", "padded"} {
		r.Factory(kind, newOrdered(kind))
	}
	r.Factory("border", newBorder)
	r.Factory("hsplit", newSplit(true))
	r.Factory("vsplit", newSplit(false))
	r.Factory("scroll", newScroll)
	r.Factory("card", newCard)
	r.Factory("tabs", newTabs)
	r.Factory("accordion", newAccordion)

	// Display.
	r.Factory("label", newLabel)
	r.Factory("text", newText)
	r.Factory("markdown", newMarkdown)
	r.Factory("image", newImage)
	r.Factory("icon", newIcon)
	r.Factory("separator", newSeparator)
	r.Factory("progress", newProgress)
	r.Factory("activity", newActivity)

	// Shapes.
	r.Factory("rect", newRect)
	r.Factory("circle", newCircle)
	r.Factory("line", newLine)

	// Input.
	r.Factory("button", newButton)
	r.Factory("hyperlink", newHyperlink)
	r.Factory("entry", newEntry)
	r.Factory("check", newCheck)
	r.Factory("select", newSelect)
	r.Factory("radio", newRadio)
	r.Factory("slider", newSlider)

	// Navigation and composition.
	r.Factory("hotspot", newHotspot)
	r.Factory("link", newLink)
	r.Factory("embed", newEmbed)
	r.Factory("slot", newSlot)
}
