package designer

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/scene"
)

type storyEdge struct {
	From  string
	To    string
	Event string
}

// storyEdges collects every navigate action in the document, including nested
// hotspots and non-tap events.
func storyEdges(doc *scene.Document) []storyEdge {
	var out []storyEdge
	if doc == nil {
		return out
	}
	doc.Walk(func(n *scene.Node) bool {
		for ev, actions := range n.On {
			for _, a := range actions {
				if a.Do == "navigate" && a.To != "" {
					from := n.ID
					if n.Name != "" {
						from = n.Name
					}
					out = append(out, storyEdge{From: from, To: a.To, Event: ev})
				}
			}
		}
		return true
	})
	return out
}

// ShowStoryMap opens a visual map of the current scene's outgoing links.
func (e *Editor) ShowStoryMap() {
	if e.opts.Window == nil {
		e.setStatus("Story map needs a host window")
		return
	}
	edges := storyEdges(e.doc)
	if len(edges) == 0 {
		dialog.ShowInformation("Story map", "This scene has no navigate actions.", e.opts.Window)
		return
	}
	targets := make(map[string]int)
	var order []string
	for _, ed := range edges {
		if _, ok := targets[ed.To]; !ok {
			targets[ed.To] = len(order)
			order = append(order, ed.To)
		}
	}
	sources := make(map[string]int)
	var sourceOrder []string
	for _, ed := range edges {
		if _, ok := sources[ed.From]; !ok {
			sources[ed.From] = len(sourceOrder)
			sourceOrder = append(sourceOrder, ed.From)
		}
	}
	const cardW, cardH, gap = 220, 44, 70
	height := float32(len(order))*gap + 160
	if h := float32(len(sourceOrder))*gap + 160; h > height {
		height = h
	}
	width := float32(cardW*2 + 360)
	board := container.NewWithoutLayout()

	sceneCard := e.card(e.doc.Title, 20, 20, cardW, cardH, true)
	board.Add(sceneCard)
	for i, from := range sourceOrder {
		y := float32(110 + i*gap)
		board.Add(e.card(from, 20, y, cardW, cardH, false))
	}
	for i, to := range order {
		y := float32(110 + i*gap)
		board.Add(e.card(to, float32(cardW+220), y, cardW, cardH, false))
	}
	for _, ed := range edges {
		sy := float32(110 + sources[ed.From]*gap + cardH/2)
		ty := float32(110 + targets[ed.To]*gap + cardH/2)
		line := canvas.NewLine(color.NRGBA{R: 59, G: 130, B: 246, A: 220})
		line.StrokeWidth = 2
		line.Position1 = fyne.NewPos(cardW+24, sy)
		line.Position2 = fyne.NewPos(cardW+220-4, ty)
		board.Add(line)
		label := canvas.NewText(ed.Event, color.NRGBA{R: 148, G: 163, B: 184, A: 255})
		label.TextSize = 11
		label.Move(fyne.NewPos(cardW+40, (sy+ty)/2-8))
		board.Add(label)
	}
	board.Resize(fyne.NewSize(width, height))
	d := dialog.NewCustom("Story map — "+e.doc.Title, "Close", container.NewScroll(container.NewGridWrap(fyne.NewSize(width, height), board)), e.opts.Window)
	d.Resize(fyne.NewSize(900, 620))
	d.Show()
}

func (e *Editor) card(text string, x, y, w, h float32, primary bool) fyne.CanvasObject {
	bg := canvas.NewRectangle(color.NRGBA{R: 30, G: 41, B: 59, A: 255})
	bg.StrokeColor = color.NRGBA{R: 59, G: 130, B: 246, A: 200}
	bg.StrokeWidth = 1
	if primary {
		bg.FillColor = color.NRGBA{R: 37, G: 99, B: 235, A: 90}
	}
	label := widget.NewLabel(text)
	label.Alignment = fyne.TextAlignCenter
	label.Wrapping = fyne.TextWrapWord
	box := container.NewStack(bg, container.NewCenter(label))
	box.Move(fyne.NewPos(x, y))
	box.Resize(fyne.NewSize(w, h))
	return box
}
