package designer

import (
	"context"
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/scene"
	"github.com/uidbz/fynerisor/scene/render"
)

type storyEdge struct {
	From  string
	To    string
	Event string
}

type storyScene struct {
	Ref   string
	Title string
	Depth int
	Edges []storyEdge
}

// storyEdges collects every navigate action in a document.
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

// storyGraph loads the current scene and up to two levels of linked scenes.
func (e *Editor) storyGraph(ctx context.Context) ([]storyScene, error) {
	resolver := e.opts.Embed
	if resolver == nil {
		resolver = &render.StoreResolver{Store: e.opts.Store}
	}
	type queued struct {
		ref, title string
		depth      int
		doc        *scene.Document
	}
	seen := map[string]bool{e.opts.Ref: true}
	queue := []queued{{ref: e.opts.Ref, title: e.doc.Title, depth: 0, doc: e.doc}}
	var out []storyScene
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		s := storyScene{Ref: cur.ref, Title: cur.title, Depth: cur.depth}
		if s.Title == "" {
			s.Title = cur.ref
		}
		for _, ed := range storyEdges(cur.doc) {
			s.Edges = append(s.Edges, ed)
			target, err := scene.ResolveRef(cur.ref, ed.To)
			if err != nil || seen[target] || cur.depth >= 2 {
				continue
			}
			seen[target] = true
			doc, base, err := resolver.LoadEmbed(ctx, cur.ref, ed.To)
			if err != nil {
				queue = append(queue, queued{ref: target, title: target, depth: cur.depth + 1})
				continue
			}
			title := ""
			if doc != nil {
				title = doc.Title
			}
			queue = append(queue, queued{ref: base, title: title, depth: cur.depth + 1, doc: doc})
		}
		out = append(out, s)
	}
	return out, nil
}

// ShowStoryMap opens a scene-level map of the document and its linked scenes.
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
	e.setStatus("Loading story map…")
	go func() {
		graph, err := e.storyGraph(context.Background())
		guithread.Do(func() {
			if err != nil {
				e.setStatus("Story map: " + err.Error())
				return
			}
			e.showGraph(graph)
		})
	}()
}

func (e *Editor) showGraph(graph []storyScene) {
	const cardW, cardH, colGap, rowGap = 240, 70, 150, 90
	byDepth := make(map[int][]storyScene)
	maxDepth := 0
	for _, s := range graph {
		byDepth[s.Depth] = append(byDepth[s.Depth], s)
		if s.Depth > maxDepth {
			maxDepth = s.Depth
		}
	}
	cardPos := make(map[string]fyne.Position)
	width := float32(40 + (maxDepth+1)*cardW + maxDepth*colGap)
	height := float32(80)
	for d := 0; d <= maxDepth; d++ {
		for i, s := range byDepth[d] {
			x := float32(40 + d*(cardW+colGap))
			y := float32(40 + i*(cardH+rowGap))
			cardPos[s.Ref] = fyne.NewPos(x, y)
			if y+cardH+40 > height {
				height = y + cardH + 40
			}
		}
	}
	board := container.NewWithoutLayout()
	for _, s := range graph {
		pos := cardPos[s.Ref]
		board.Add(e.card(s.Title, pos.X, pos.Y, cardW, cardH, s.Depth == 0))
	}
	blue := color.NRGBA{R: 59, G: 130, B: 246, A: 220}
	for _, s := range graph {
		from := cardPos[s.Ref]
		for _, ed := range s.Edges {
			target, err := scene.ResolveRef(s.Ref, ed.To)
			if err != nil {
				continue
			}
			to, ok := cardPos[target]
			if !ok {
				continue
			}
			line := canvas.NewLine(blue)
			line.StrokeWidth = 2
			line.Position1 = fyne.NewPos(from.X+cardW, from.Y+cardH/2)
			line.Position2 = fyne.NewPos(to.X, to.Y+cardH/2)
			board.Add(line)
			label := canvas.NewText(ed.From+" · "+ed.Event, color.NRGBA{R: 148, G: 163, B: 184, A: 255})
			label.TextSize = 11
			label.Move(fyne.NewPos(from.X+cardW+8, (from.Y+cardH/2+to.Y+cardH/2)/2-8))
			board.Add(label)
		}
	}
	d := dialog.NewCustom("Story map — "+e.doc.Title, "Close",
		container.NewScroll(container.NewGridWrap(fyne.NewSize(width, height), board)), e.opts.Window)
	d.Resize(fyne.NewSize(960, 640))
	d.Show()
	e.setStatus(fmt.Sprintf("Story map: %d scene(s)", len(graph)))
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
